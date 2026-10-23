"""接口层应答一致性：健康检查、不支持的物质、未预期异常。

框架层的错误也必须走统一应答，否则界面上会直接出现英文堆栈。
"""

from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.api.errors import register_error_handlers
from app.constants.codes import (
    CODE_COMPUTE_HEALTHY,
    CODE_INTERNAL_ERROR,
    ERROR_COMPOSITION_NOT_SUPPORTED,
)
from app.main import app

client = TestClient(app)


def test_health_returns_language_neutral_code():
    response = client.get("/health")

    assert response.status_code == 200
    payload = response.json()
    assert payload["status"] == "ok"
    assert payload["code"] == CODE_COMPUTE_HEALTHY
    assert payload["service"]
    assert payload["version"]


def test_unsupported_composition_returns_400_with_code():
    response = client.post("/predict", json={"formula": "CaCO3"})

    assert response.status_code == 400
    payload = response.json()
    assert payload["status"] == "error"
    assert payload["code"] == ERROR_COMPOSITION_NOT_SUPPORTED
    assert payload["params"]["formula"] == "CaCO3"


def test_unexpected_error_returns_internal_error_code():
    failing_app = FastAPI()
    register_error_handlers(failing_app)

    @failing_app.get("/boom")
    def raise_unexpected_error() -> None:
        raise RuntimeError("假装数据库断了")

    failing_client = TestClient(failing_app, raise_server_exceptions=False)
    response = failing_client.get("/boom")

    assert response.status_code == 500
    payload = response.json()
    assert payload["status"] == "error"
    assert payload["code"] == CODE_INTERNAL_ERROR
