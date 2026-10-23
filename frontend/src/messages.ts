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
  // 预测结果：结构依据（rule）
  prediction_ready: '已生成预测结果。',
  rule_radius_ratio:
    '阳离子半径是阴离子半径的 {ratio} 倍。半径比越小，周围能挤下的反向离子越少；按这个比值，每个阳离子周围有 {coordination_number} 个阴离子。',
  rule_experimental_override:
    '半径比规则会推算成 {rule_coordination_number} 配位，但实验测定的真实结构不是这样，这里采用实验数据（半径比 {ratio}）。',
  rule_fluorite_stoichiometry:
    '阳离子与阴离子的个数比是 {cation_count}:{anion_count}，半径比 {ratio} 落在 8 配位区间：大离子做面心立方堆积，小离子填满四面体空隙，这就是萤石型。',
  rule_antifluorite_stoichiometry:
    '阴离子与阳离子的个数比是 {anion_count}:{cation_count}，结构与萤石型相同，只是阴阳离子位置对调，称反萤石型。',
  rule_diamond_covalent: '这是单质共价晶体：每个原子与 4 个相邻原子共用电子对，形成金刚石型结构。',
  rule_zinc_blende_covalent:
    '两种元素的电负性只差 {electronegativity_difference}，以共价键结合，按闪锌矿（立方 ZnS）型交替排列。',
  rule_metal_experimental: '金属的结构型来自实验测定，不是用公式推导出来的。',
  // 预测结果：结果附注（caveat）
  caveat_radius_ratio_empirical: '半径比规则是经验规律，只说明“最可能”的配位数，实际结构以实验测定为准。',
  caveat_near_boundary: '半径比 {ratio} 离分界值 {boundary} 很近，两边的配位数都有可能。',
  caveat_confirmed_by_experiment: '实验测定的结构与上面的推算一致，可以放心采用。',
  caveat_low_ionic_character:
    '两种元素的电负性只差 {difference}（小于 {threshold}），键的共价性比较明显，这类晶体介于离子晶体和共价晶体之间。',
  caveat_metal_structure_experimental: '金属的堆积方式不能只由成分推出，这里的结构型来自实验数据。',
  caveat_stoichiometry_outside_band:
    '按离子个数比，这个式子应形成萤石型；但半径比 {ratio} 不在理想的 8 配位区间（0.732~1），实际结构可能不同，请以实验测定为准。',
  // 预测结果：错误提示
  formula_empty: '请先输入化学式，例如 NaCl、CaF2 或 C。',
  formula_invalid_character:
    '化学式第 {position} 个字符“{character}”看不懂。只支持元素符号和下标数字，例如 CaF2。',
  formula_unbalanced_parentheses: '化学式 {formula} 的括号没有配对，请检查左右括号。',
  formula_zero_count: '化学式 {formula} 里有下标 0，元素个数至少是 1。',
  formula_charge_unbalanced: '化学式 {formula} 的正负电荷加起来不是 0，请检查下标。',
  element_not_supported: '还没有收录元素 {symbol}，可以换一种常见元素试试。',
  element_data_missing: '元素 {symbol} 的数据还不全，暂时算不出它的结构。',
  composition_not_supported:
    '暂时还不能处理 {formula} 这类物质。目前支持：离子晶体（如 NaCl、CsCl、CaF2）、共价晶体（C、Si、Ge、SiC）、金属（如 Cu、Fe、Mg）。',
  radius_ratio_too_small: '半径比 {ratio} 太小，需要 3 配位的结构，本平台暂不覆盖。',
  request_invalid_body: '请求格式不正确，请刷新页面后重试。',
  internal_error: '数据处理层出现了意外错误，请稍后再试；如果一直失败，请告诉老师。',
  // 任务接口
  task_created: '任务已创建，正在排队计算。',
  task_ok: '已获取任务状态。',
  task_result: '已获取计算结果。',
  task_cancelled: '任务已取消。',
  task_not_found: '找不到这个任务（编号 {task_id}），请刷新页面重试。',
  task_not_ready: '结果还没准备好，请稍等片刻再试。',
  task_cancel_conflict: '任务已经开始或结束，不能再取消。',
  task_formula_empty: '请先输入化学式，例如 NaCl、CaF2 或 C。',
  task_formula_too_long: '化学式太长了（最多 {max_length} 个字符），请检查一下。',
  compute_no_result: '数据处理层没有返回结构，请稍后重试或告诉老师。',
  // 任务状态（结果显示用）
  task_status_pending: '排队中',
  task_status_running: '计算中',
  task_status_succeeded: '已完成',
  task_status_failed: '计算失败',
  task_status_cancelled: '已取消',
  // 结构型名称（结果显示用）
  prototype_rock_salt: '氯化钠型',
  prototype_cesium_chloride: '氯化铯型',
  prototype_zinc_blende: '闪锌矿型（立方 ZnS）',
  prototype_fluorite: '萤石型',
  prototype_antifluorite: '反萤石型',
  prototype_diamond: '金刚石型',
  prototype_face_centered_cubic: '面心立方最密堆积',
  prototype_body_centered_cubic: '体心立方堆积',
  prototype_hexagonal_close_packed: '六方最密堆积',
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
  // Prediction result: reasoning (rule)
  prediction_ready: 'The predicted structure is ready.',
  rule_radius_ratio:
    'The cation radius is {ratio} times the anion radius. The smaller the ratio, the fewer opposite ions fit around it; by this ratio each cation touches {coordination_number} anions.',
  rule_experimental_override:
    'The radius-ratio rule would predict {rule_coordination_number}-fold coordination, but the measured structure is different, so the experimental data is used here (radius ratio {ratio}).',
  rule_fluorite_stoichiometry:
    'The cation to anion ratio is {cation_count}:{anion_count} and the radius ratio {ratio} falls in the 8-coordination range: large ions form a face-centered cubic array and small ions fill the tetrahedral holes (fluorite type).',
  rule_antifluorite_stoichiometry:
    'The anion to cation ratio is {anion_count}:{cation_count}; the structure is the same as fluorite with the roles of the two ions swapped (antifluorite type).',
  rule_diamond_covalent:
    'This is a covalent crystal of one element: every atom shares electron pairs with 4 neighbours, forming the diamond structure.',
  rule_zinc_blende_covalent:
    'The two elements differ by only {electronegativity_difference} in electronegativity, so they bond covalently and alternate in the zinc blende (cubic ZnS) arrangement.',
  rule_metal_experimental:
    'The structure type of a metal comes from experiment, not from a formula.',
  // Prediction result: caveats
  caveat_radius_ratio_empirical:
    'The radius-ratio rule is only a rule of thumb: it gives the most likely coordination number, while the real structure is decided by experiment.',
  caveat_near_boundary:
    'The radius ratio {ratio} is very close to the boundary {boundary}, so either coordination number is possible.',
  caveat_confirmed_by_experiment:
    'The measured structure agrees with the reasoning above, so it can be trusted.',
  caveat_low_ionic_character:
    'The electronegativity difference is only {difference} (below {threshold}), so the bonding is noticeably covalent; such crystals sit between ionic and covalent crystals.',
  caveat_metal_structure_experimental:
    'The packing of a metal cannot be derived from its composition alone; this structure type comes from experimental data.',
  caveat_stoichiometry_outside_band:
    'By ion count this formula should form the fluorite type, but the radius ratio {ratio} lies outside the ideal 8-coordination range (0.732 to 1); the real structure may differ, so trust the measured structure.',
  // Prediction result: error messages
  formula_empty: 'Please enter a chemical formula first, for example NaCl, CaF2 or C.',
  formula_invalid_character:
    'Character "{character}" at position {position} is not understood. Only element symbols and subscript numbers are supported, for example CaF2.',
  formula_unbalanced_parentheses: 'The brackets in {formula} are not balanced. Please check them.',
  formula_zero_count: 'The formula {formula} contains a 0 subscript; every element count is at least 1.',
  formula_charge_unbalanced:
    'The positive and negative charges in {formula} do not cancel out. Please check the subscripts.',
  element_not_supported:
    'The element {symbol} is not available yet. Try a common element instead.',
  element_data_missing:
    'There is not enough data for {symbol} yet, so its structure cannot be predicted.',
  composition_not_supported:
    'Substances like {formula} are not supported yet. Currently supported: ionic crystals (e.g. NaCl, CsCl, CaF2), covalent crystals (C, Si, Ge, SiC) and metals (e.g. Cu, Fe, Mg).',
  radius_ratio_too_small:
    'The radius ratio {ratio} is too small and would need 3-fold coordination, which is not covered yet.',
  request_invalid_body: 'The request was not well formed. Please refresh the page and try again.',
  internal_error:
    'The compute layer hit an unexpected error. Please try again later; if it keeps failing, tell your teacher.',
  // Task endpoints
  task_created: 'Task created and queued for computing.',
  task_ok: 'Task status loaded.',
  task_result: 'Result loaded.',
  task_cancelled: 'Task cancelled.',
  task_not_found: 'Task {task_id} was not found. Please refresh the page and try again.',
  task_not_ready: 'The result is not ready yet. Please wait a moment and try again.',
  task_cancel_conflict: 'The task has already started or finished, so it cannot be cancelled.',
  task_formula_empty: 'Please enter a chemical formula first, for example NaCl, CaF2 or C.',
  task_formula_too_long: 'The formula is too long (at most {max_length} characters).',
  compute_no_result: 'The compute layer returned no structure. Please try again later or tell your teacher.',
  // Task statuses (shown in results)
  task_status_pending: 'Queued',
  task_status_running: 'Computing',
  task_status_succeeded: 'Done',
  task_status_failed: 'Failed',
  task_status_cancelled: 'Cancelled',
  // Structure type names (shown in results)
  prototype_rock_salt: 'Rock salt type',
  prototype_cesium_chloride: 'Cesium chloride type',
  prototype_zinc_blende: 'Zinc blende (cubic ZnS) type',
  prototype_fluorite: 'Fluorite type',
  prototype_antifluorite: 'Antifluorite type',
  prototype_diamond: 'Diamond type',
  prototype_face_centered_cubic: 'Face-centered cubic close packing',
  prototype_body_centered_cubic: 'Body-centered cubic packing',
  prototype_hexagonal_close_packed: 'Hexagonal close packing',
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
