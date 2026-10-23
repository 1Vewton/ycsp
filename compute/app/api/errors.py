"""把框架层的错误也转成统一应答（code + params），避免界面上出现英文报错。"""

import logging
from http import HTTPStatus

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from app import SERVICE_NAME
from app.constants.codes import CODE_INTERNAL_ERROR, CODE_REQUEST_INVALID_BODY

logger = logging.getLogger(__name__)


def register_error_handlers(application: FastAPI) -> None:
    """统一错误应答：只返回语言无关的 code，文案由前端按当前语言翻译。"""
    application.add_exception_handler(RequestValidationError, _handle_invalid_request_body)
    application.add_exception_handler(Exception, _handle_unexpected_error)


async def _handle_invalid_request_body(
    _request: Request, _error: RequestValidationError
) -> JSONResponse:
    """请求体不符合约定（如缺 formula 字段）时，用统一的 code 应答。"""
    return _error_response(HTTPStatus.BAD_REQUEST, CODE_REQUEST_INVALID_BODY)


async def _handle_unexpected_error(_request: Request, error: Exception) -> JSONResponse:
    """未预期的异常：完整堆栈只写进服务日志，学生看到一句能翻译的提示。"""
    logger.exception("未预期的错误：%s", error)
    return _error_response(HTTPStatus.INTERNAL_SERVER_ERROR, CODE_INTERNAL_ERROR)


def _error_response(status_code: HTTPStatus, code: str) -> JSONResponse:
    return JSONResponse(
        status_code=status_code,
        content={"service": SERVICE_NAME, "status": "error", "code": code},
    )
