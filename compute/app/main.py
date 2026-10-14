"""数据处理层的 FastAPI 入口。"""

from fastapi import FastAPI

SERVICE_NAME = "ycsp-数据处理层"
SERVICE_VERSION = "0.1.0"
SERVICE_HOST = "127.0.0.1"
SERVICE_PORT = 8001  # 与数据管理层 8000、前端 5173 错开

app = FastAPI(title=SERVICE_NAME, version=SERVICE_VERSION)


@app.get("/health")
def read_health() -> dict[str, str]:
    """健康检查。返回的 message 会被数据管理层转达给前端，用于验证三层连通。"""
    return {
        "service": SERVICE_NAME,
        "version": SERVICE_VERSION,
        "status": "ok",
        "message": "你好，我是数据处理层，已经准备好接收计算任务。",
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("app.main:app", host=SERVICE_HOST, port=SERVICE_PORT, reload=True)
