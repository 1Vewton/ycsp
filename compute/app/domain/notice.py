"""语言无关的说明与错误。

面向学生的文案由前端按当前语言翻译（见根 README「多语言」），
数据处理层只负责说清“是哪一条说明”以及“要填进句子的数值”。
"""

from dataclasses import dataclass, field


@dataclass(frozen=True)
class Notice:
    """一条可翻译的说明：code 决定说哪句话，params 是填进句子的数值或标识。"""

    code: str
    params: dict[str, object] = field(default_factory=dict)


class DomainError(Exception):
    """可预期的输入或数据问题。

    用异常携带 Notice，让接口层能返回结构化错误，而不是把技术细节丢给学生。
    """

    def __init__(self, notice: Notice) -> None:
        super().__init__(f"{notice.code} {notice.params}")
        self.notice = notice
