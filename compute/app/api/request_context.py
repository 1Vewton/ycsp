"""请求编号：一次预测在前端、数据管理层、数据处理层用同一个编号。

数据管理层会通过 X-Request-ID 传进来；没有或格式不合法时就生成一个。
请求结束时把编号写回响应头，并记录一行带编号的访问日志。
"""

import logging
import time
import uuid
from contextvars import ContextVar

from starlette.middleware.base import BaseHTTPMiddleware, RequestResponseEndpoint
from starlette.requests import Request
from starlette.responses import Response

REQUEST_ID_HEADER = "X-Request-ID"
REQUEST_ID_MAX_LENGTH = 64
# 只允许字母、数字与 -_.:；换行等字符会伪造日志行（日志注入）。
REQUEST_ID_SAFE_CHARACTERS = frozenset("-_.:")

request_id_context: ContextVar[str] = ContextVar("request_id", default="-")

logger = logging.getLogger(__name__)


def new_request_id() -> str:
    """生成计算层自己的请求编号。"""
    return f"compute-{uuid.uuid4().hex}"


def normalize_request_id(raw: str | None) -> str:
    """校验外部传入的编号；不合法时返回空字符串，由调用方重新生成。"""
    if raw is None:
        return ""
    value = raw.strip()
    if not value or len(value) > REQUEST_ID_MAX_LENGTH:
        return ""
    if not all(
        character.isascii() and (character.isalnum() or character in REQUEST_ID_SAFE_CHARACTERS)
        for character in value
    ):
        return ""
    return value


def current_request_id() -> str:
    """当前请求的编号；不在请求里时返回 "-"。"""
    return request_id_context.get()


class RequestIdFilter(logging.Filter):
    """给每条日志加上当前请求编号，日志格式里用 %(request_id)s 取。"""

    def filter(self, record: logging.LogRecord) -> bool:
        record.request_id = current_request_id()
        return True


class RequestContextMiddleware(BaseHTTPMiddleware):
    """确定请求编号、写回响应头，并记录请求结果与耗时。"""

    async def dispatch(
        self, request: Request, call_next: RequestResponseEndpoint
    ) -> Response:
        request_id = (
            normalize_request_id(request.headers.get(REQUEST_ID_HEADER))
            or new_request_id()
        )
        token = request_id_context.set(request_id)
        started_at = time.perf_counter()

        try:
            response = await call_next(request)
        except Exception:
            logger.exception(
                "request failed method=%s path=%s duration_ms=%.1f",
                request.method,
                request.url.path,
                (time.perf_counter() - started_at) * 1000,
            )
            raise
        else:
            response.headers[REQUEST_ID_HEADER] = request_id
            logger.info(
                "request method=%s path=%s status=%d duration_ms=%.1f",
                request.method,
                request.url.path,
                response.status_code,
                (time.perf_counter() - started_at) * 1000,
            )
            return response
        finally:
            request_id_context.reset(token)
