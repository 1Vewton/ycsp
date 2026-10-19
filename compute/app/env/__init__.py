"""环境配置：统一读取环境变量，未设置时使用默认值。

.env 位于 compute/ 根目录；文件不存在时不影响运行，全部使用下面的默认值。
"""

import os
from pathlib import Path

from dotenv import load_dotenv

# 本文件在 compute/app/env/__init__.py，向上三层就是 compute/ 根目录。
COMPUTE_ROOT = Path(__file__).resolve().parents[2]
ENV_FILE = COMPUTE_ROOT / ".env"
load_dotenv(ENV_FILE)

# 默认监听地址与端口；集中在此，方便一眼看出兜底值。
DEFAULT_SERVICE_HOST = "127.0.0.1"
DEFAULT_SERVICE_PORT = 8001


def get_env(key: str, fallback: str) -> str:
    """读取环境变量；未设置或为空时返回 fallback。"""
    value = os.getenv(key)
    return value if value else fallback


def get_env_int(key: str, fallback: int) -> int:
    """读取整数环境变量；未设置或为空时返回 fallback。"""
    raw = get_env(key, str(fallback))
    try:
        return int(raw)
    except ValueError as error:
        raise ValueError(f"环境变量 {key} 必须是整数，收到 {raw!r}。") from error


SERVICE_HOST = get_env("YCSP_COMPUTE_HOST", DEFAULT_SERVICE_HOST)
SERVICE_PORT = get_env_int("YCSP_COMPUTE_PORT", DEFAULT_SERVICE_PORT)
