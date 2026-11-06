"""请求编号：沿用客户端编号、缺失或非法时生成，并写回响应头。"""

import logging

from fastapi.testclient import TestClient

from app.api.request_context import (
    REQUEST_ID_HEADER,
    RequestIdFilter,
    normalize_request_id,
)
from app.main import app

client = TestClient(app)


def test_request_id_is_echoed():
    response = client.get("/health", headers={REQUEST_ID_HEADER: "web-test-123"})

    assert response.status_code == 200
    assert response.headers[REQUEST_ID_HEADER] == "web-test-123"


def test_request_id_is_generated_when_missing():
    response = client.get("/health")

    assert response.headers[REQUEST_ID_HEADER].startswith("compute-")


def test_invalid_request_id_is_replaced():
    response = client.get("/health", headers={REQUEST_ID_HEADER: "web 123"})

    generated = response.headers[REQUEST_ID_HEADER]
    assert generated.startswith("compute-")
    assert generated != "web 123"


def test_normalize_rejects_log_injection_and_overlong_values():
    assert normalize_request_id("web-123\nforged") == ""
    assert normalize_request_id("a" * 65) == ""
    assert normalize_request_id(None) == ""
    assert normalize_request_id("  ") == ""
    assert normalize_request_id("web-123") == "web-123"


def test_prediction_log_carries_the_request_id():
    records: list[logging.LogRecord] = []

    class CaptureHandler(logging.Handler):
        def emit(self, record: logging.LogRecord) -> None:
            records.append(record)

    handler = CaptureHandler()
    handler.addFilter(RequestIdFilter())
    logger = logging.getLogger("app.api.predict")
    logger.addHandler(handler)

    try:
        response = client.post(
            "/predict",
            json={"formula": "NaCl"},
            headers={REQUEST_ID_HEADER: "web-log-123"},
        )
    finally:
        logger.removeHandler(handler)

    assert response.status_code == 200
    prediction_records = [
        record for record in records if record.getMessage().startswith("prediction ")
    ]
    assert prediction_records, "the prediction route logged nothing"
    assert prediction_records[0].request_id == "web-log-123"
