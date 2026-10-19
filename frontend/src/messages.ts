// 界面文案，以及“后端状态码 → 文案”的映射。
//
// 文案集中在语言包里、不写死在组件中（见根 README「多语言」）。
// 语言切换在阶段 3 接入，这里先把中/英两种语言的文案都备好。

export type Locale = 'zh' | 'en'

// 当前语言；阶段 3 会改成可由用户切换的状态。
export const CurrentLocale: Locale = 'zh'

type MessageTable = Record<string, string>

// 匹配 "{name}" 形式的占位符，用于把参数填进文案。
const FIELD_PATTERN = /\{(\w+)\}/g

const ZH_MESSAGES: MessageTable = {
  // 后端状态码
  backend_healthy: '数据管理层运行正常。',
  compute_healthy: '已连通数据处理层。',
  compute_unreachable: '联系不上数据处理层（{url}），请确认它已经启动。',
  compute_bad_status: '数据处理层返回了异常状态码 {status_code}，请查看它的运行日志。',
  compute_read_failed: '读取数据处理层的应答时出错，请稍后重试。',
  compute_invalid_json: '数据处理层的应答不是合法 JSON，请检查它的服务日志。',
  unknown_code: '出现了未预期的错误（代码 {code}）。',
  request_failed: '暂时联系不上后端，请确认数据管理层与数据处理层都已启动。',
  // 界面文案
  page_title: 'ycsp 晶体结构预测教学平台',
  page_intro: '这是阶段 0 的骨架页面，用来检查「前端 → 数据管理 → 数据处理」三层是否连通。',
  ping_button_idle: '和数据处理层打个招呼',
  ping_button_busy: '正在联系……',
  ping_waiting: '正在等待后端应答，请稍候……',
}

const EN_MESSAGES: MessageTable = {
  // Backend status codes
  backend_healthy: 'The data layer is running normally.',
  compute_healthy: 'Connected to the compute layer.',
  compute_unreachable: 'Cannot reach the compute layer ({url}). Please make sure it is running.',
  compute_bad_status: 'The compute layer returned status code {status_code}. Please check its logs.',
  compute_read_failed: 'Failed to read the response from the compute layer. Please try again later.',
  compute_invalid_json: 'The compute layer response is not valid JSON. Please check its service logs.',
  unknown_code: 'Unexpected response from the backend (code {code}).',
  request_failed: 'Cannot reach the backend. Please make sure the data layer and the compute layer are running.',
  // Interface
  page_title: 'ycsp - Crystal Structure Prediction for Students',
  page_intro: 'This is the stage-0 skeleton page, used to check whether the three layers are connected.',
  ping_button_idle: 'Say hello to the compute layer',
  ping_button_busy: 'Contacting...',
  ping_waiting: 'Waiting for the backend, please hold on...',
}

const MESSAGE_TABLES: Record<Locale, MessageTable> = {
  zh: ZH_MESSAGES,
  en: EN_MESSAGES,
}

/**
 * 取出当前语言下 key 对应的文案，并把 {name} 占位符替换成参数值。
 *
 * 找不到 key 时退回 unknown_code，并把 key 当作 {code} 填进去。
 */
export function translate(key: string, params: Record<string, unknown> = {}): string {
  const table = MESSAGE_TABLES[CurrentLocale]
  const template = table[key]
  const values: Record<string, unknown> = template === undefined ? { code: key, ...params } : params

  return (template ?? table.unknown_code).replace(
    FIELD_PATTERN,
    (placeholder: string, fieldName: string): string => {
      return fieldName in values ? String(values[fieldName]) : placeholder
    },
  )
}
