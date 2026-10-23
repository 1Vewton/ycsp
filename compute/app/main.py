"""数据处理层的服务入口：组装 FastAPI 应用并启动 uvicorn。"""

import uvicorn
from fastapi import FastAPI

from app import SERVICE_NAME, SERVICE_VERSION
from app.api.errors import register_error_handlers
from app.api.health import router as health_router
from app.api.predict import router as predict_router
from app.env import SERVICE_HOST, SERVICE_PORT

app = FastAPI(title=SERVICE_NAME, version=SERVICE_VERSION)
app.include_router(health_router)
app.include_router(predict_router)
register_error_handlers(app)


def main() -> None:
    """启动开发服务器（改代码自动重载）。"""
    uvicorn.run("app.main:app", host=SERVICE_HOST, port=SERVICE_PORT, reload=True)


if __name__ == "__main__":
    main()
