"""数值展示约定：输出给界面的计算值统一保留固定小数位。"""

# 界面展示用的小数位。教学场景 3 位（如 5.662 Å、0.564）足够，
# 也避免把 0.5635359116… 这类浮点噪声直接显示给学生。
DISPLAY_DECIMALS = 3


def round_for_display(value: float) -> float:
    """把计算值舍入到界面展示用的位数。"""
    return round(value, DISPLAY_DECIMALS)
