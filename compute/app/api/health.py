"""健康检查接口。"""

from fastapi import APIRouter

from app import SERVICE_NAME, SERVICE_VERSION
from app.constants.codes import CODE_COMPUTE_HEALTHY

router = APIRouter()


@router.get("/health")
def read_health() -> dict[str, str]:
    """健康检查：返回语言无关的状态码。"""
    return {
        "service": SERVICE_NAME,
        "version": SERVICE_VERSION,
        "status": "ok",
        "code": CODE_COMPUTE_HEALTHY,
    }
