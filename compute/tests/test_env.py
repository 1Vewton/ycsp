"""环境变量读取：缺省值、空值都回落到默认值，非法整数要立刻报错。"""

import pytest

from app import env


def test_missing_variable_uses_fallback(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.delenv("YCSP_TEST_VALUE", raising=False)

    assert env.get_env("YCSP_TEST_VALUE", "default") == "default"


def test_empty_variable_uses_fallback(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setenv("YCSP_TEST_VALUE", "")

    assert env.get_env("YCSP_TEST_VALUE", "default") == "default"


def test_existing_variable_wins(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setenv("YCSP_TEST_VALUE", "custom")

    assert env.get_env("YCSP_TEST_VALUE", "default") == "custom"


def test_integer_variable_is_parsed(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setenv("YCSP_TEST_PORT", "9000")

    assert env.get_env_int("YCSP_TEST_PORT", 8001) == 9000


def test_missing_integer_variable_uses_fallback(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.delenv("YCSP_TEST_PORT", raising=False)

    assert env.get_env_int("YCSP_TEST_PORT", 8001) == 8001


def test_invalid_integer_reports_the_variable_name(monkeypatch: pytest.MonkeyPatch):
    monkeypatch.setenv("YCSP_TEST_PORT", "not-a-number")

    with pytest.raises(ValueError) as error_info:
        env.get_env_int("YCSP_TEST_PORT", 8001)

    assert "YCSP_TEST_PORT" in str(error_info.value)
