"""数据处理层的 FastAPI 入口。"""

from fastapi import FastAPI

from app.env import SERVICE_HOST, SERVICE_PORT

SERVICE_NAME = "ycsp-compute"
SERVICE_VERSION = "0.1.0"
# 状态码与数据管理层保持一致；具体文案由前端按当前语言翻译（见根 README「多语言」）。
CODE_COMPUTE_HEALTHY = "compute_healthy"

app = FastAPI(title=SERVICE_NAME, version=SERVICE_VERSION)


@app.get("/health")
def read_health() -> dict[str, str]:
    """健康检查：返回语言无关的状态码，文案由前端按当前语言翻译。"""
    return {
        "service": SERVICE_NAME,
        "version": SERVICE_VERSION,
        "status": "ok",
        "code": CODE_COMPUTE_HEALTHY,
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("app.main:app", host=SERVICE_HOST, port=SERVICE_PORT, reload=True)
