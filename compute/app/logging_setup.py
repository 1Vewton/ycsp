"""日志配置：每条应用日志都带请求编号，便于按编号串起三层。"""

import logging

from app.api.request_context import RequestIdFilter

LOG_FORMAT = "%(asctime)s %(levelname)s %(name)s request_id=%(request_id)s %(message)s"


def configure_logging() -> None:
    """配置根日志与 request_id 过滤器。

    uvicorn 自己的访问日志保持原样；应用日志（app.*）会经过这里，
    因此异常与预测结果都能带上同一个 request_id。
    """
    root_logger = logging.getLogger()
    if not root_logger.handlers:
        logging.basicConfig(level=logging.INFO, format=LOG_FORMAT)

    for handler in root_logger.handlers:
        handler.setFormatter(logging.Formatter(LOG_FORMAT))
        if not any(isinstance(existing, RequestIdFilter) for existing in handler.filters):
            handler.addFilter(RequestIdFilter())
    root_logger.setLevel(logging.INFO)
