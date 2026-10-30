// 中文文案（默认语言）。键必须与 en.ts 完全一致：少写、写错或多写都会让 `npm run type-check` 失败。

import type { MessageKey } from './en'

export const zhMessages = {
  // 后端状态码
  backend_healthy: '数据管理层运行正常。',
  compute_healthy: '后台服务正常，可以开始预测了。',
  compute_unreachable: '后台计算服务没有响应（{url}），请告诉老师。',
  compute_bad_status: '后台计算服务返回了异常状态 {status_code}，请告诉老师。',
  compute_read_failed: '读取后台计算服务的应答时出错，请稍后重试或告诉老师。',
  compute_invalid_json: '后台计算服务的应答看不懂，请告诉老师。',
  unknown_code: '出现了未预期的错误（代码 {code}）。',
  request_failed: '暂时联系不上后端，请确认数据管理层与数据处理层都已启动。',
  // 预测结果：结构依据（rule）
  prediction_ready: '已生成预测结果。',
  rule_radius_ratio:
    '阳离子半径是阴离子半径的 {ratio} 倍。半径比越小，周围能挤下的相反电荷离子越少；按这个比值，每个阳离子周围有 {coordination_number} 个阴离子。',
  rule_experimental_override:
    '半径比规则会推算成 {rule_coordination_number} 配位，但实验测定的真实结构不是这样，这里采用实验数据（半径比 {ratio}）。',
  rule_fluorite_stoichiometry:
    '阳离子与阴离子的个数比是 {cation_count}:{anion_count}，半径比 {ratio} 落在 8 配位区间：大离子按面心立方排列（立方体的顶点和每个面的中心各放一个），小离子填进它们围出的四面体空隙（四个大离子之间的空位），这就是萤石型。',
  rule_antifluorite_stoichiometry:
    '阴离子与阳离子的个数比是 {anion_count}:{cation_count}，结构与萤石型相同，只是阴阳离子位置对调，称反萤石型。',
  rule_diamond_covalent: '这是单质共价晶体：每个原子与 4 个相邻原子共用电子对，形成金刚石型结构。',
  rule_zinc_blende_covalent:
    '两种元素的电负性（原子拉电子的能力）只差 {electronegativity_difference}，以共价键结合，按闪锌矿（立方 ZnS）型交替排列。',
  rule_metal_experimental: '金属的结构型来自实验测定，不是用公式推导出来的。',
  // 预测结果：结果附注（caveat）
  caveat_radius_ratio_empirical: '半径比规则是经验规律，只说明“最可能”的配位数，实际结构以实验测定为准。',
  caveat_near_boundary: '半径比 {ratio} 离分界值 {boundary} 很近，两边的配位数都有可能。',
  caveat_confirmed_by_experiment: '实验测定的结构与上面的推算一致，可以放心采用。',
  caveat_low_ionic_character:
    '两种元素的电负性（原子拉电子的能力）只差 {difference}（小于 {threshold}），键的共价性比较明显，这类晶体介于离子晶体和共价晶体之间。',
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
  // 账号与会话
  user_registered: '注册成功，已经帮你登录了。',
  user_ok: '已获取当前账号。',
  signed_in: '登录成功。',
  signed_out: '已退出登录。',
  user_name_taken: '这个昵称已经有人用了，换一个试试。',
  user_name_invalid: '昵称要 {min_length} 到 {max_length} 个字符。',
  password_too_weak: '密码还不符合要求，请按下面的提示修改。',
  invalid_credentials: '昵称或密码不对，请重新输入。',
  not_signed_in: '还没有登录，请先登录。',
  session_unavailable: '登录服务暂时不可用，请稍后再试或告诉老师。',
  too_many_attempts: '尝试次数太多了，请约 {retry_after_minutes} 分钟后再试。',
  // 密码规则（逐条翻译 password_too_weak 的 params.rules）
  password_rule_too_short: '密码至少 {min_length} 位。',
  password_rule_too_long: '密码太长了（最多 {max_bytes} 字节）。',
  password_rule_missing_letter: '密码里要有一个英文字母。',
  password_rule_missing_digit: '密码里要有一个数字。',
  password_rule_missing_symbol: '密码里要有一个特殊符号，例如 ! @ # ?。',
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
  language_switcher_label: '界面语言',
  main_navigation_label: '主导航',
  nav_home: '首页',
  nav_predict: '开始预测',
  ping_button_idle: '检查后台服务',
  ping_button_busy: '正在检查……',
  ping_waiting: '正在等待后台应答，请稍候……',
  // 首页
  home_heading: '用化学式，看见晶体的样子',
  home_intro:
    '输入一个化学式，平台会推算出它最可能形成的晶体结构，并告诉你为什么是这样。整个过程不需要写代码，也不需要做复杂的计算。',
  home_steps_title: '三步就能看到结果',
  home_step_input_title: '输入化学式',
  home_step_input_description: '像课本上那样写，例如 NaCl、CaF2、C 或 Cu。',
  home_step_predict_title: '平台推算结构',
  home_step_predict_description:
    '平台会先判断它属于哪一类晶体（离子晶体、共价晶体还是金属晶体），再按课本里的规律排出原子的位置。',
  home_step_read_title: '看懂结果',
  home_step_read_description:
    '你会看到晶胞（构成晶体的最小重复单元）有多大、原子在哪里，以及这个结论有多可靠。',
  home_supported_title: '目前能预测这些',
  home_supported_ionic:
    '离子晶体：带正电和带负电的离子靠静电吸引结合，例如 NaCl、CsCl、CaF2。',
  home_supported_covalent: '共价晶体：原子之间共用电子对，例如 C（金刚石）、Si、SiC。',
  home_supported_metallic: '金属晶体：只含一种金属元素，例如 Cu、Fe、Mg。',
  home_start_button: '开始预测',
  home_check_title: '检查后台服务',
  home_check_description:
    '如果点了按钮一直没有反应，可以先点这里检查后台服务是否正常，再告诉老师。',
  // 预测页
  predict_heading: '预测晶体结构',
  predict_intro: '输入一个化学式，看看它最可能形成哪种晶体结构。',
  formula_input_label: '化学式',
  formula_input_placeholder: '例如 NaCl',
  formula_input_hint: '只写元素符号和下标，例如 CaF2；需要时也可以用括号，例如 Ca(OH)2。',
  formula_syntax_ok: '写法没问题，可以继续。',
  predict_start_button: '开始预测',
  predict_starting: '正在创建任务……',
  predict_examples_title: '化学式可以这样写',
  predict_examples_hint: '也可以直接点一个例子，马上填进输入框。',
  example_group_ionic: '离子晶体',
  example_group_covalent: '共价晶体',
  example_group_metallic: '金属晶体',
  // 结果页
  result_heading: '预测结果',
  result_intro: '这里显示这次预测的进度和结果。',
  result_task_label: '任务编号：{task_id}',
  result_loading: '正在读取任务状态……',
  result_polling_hint: '进度会自动刷新，不用手动刷新页面。',
  result_cancel_button: '取消任务',
  result_cancelling: '正在取消……',
  result_cancelled_hint: '任务已取消，不会再有结果。',
  result_failed_hint: '这次计算没有完成，原因见上面的提示。',
  result_duration_seconds: '这次计算用了 {seconds} 秒。',
  result_retry_button: '再试一次',
  // 三维结构
  structure_title: '晶体结构',
  structure_loading: '正在读取结构……',
  structure_preparing: '正在准备三维模型……',
  structure_drag_hint: '按住拖动可以旋转，滚轮可以缩放。',
  structure_reset_view: '重置视角',
  structure_canvas_label: '晶体结构三维示意图',
  structure_webgl_unavailable:
    '这台设备暂时显示不了三维模型（浏览器不支持 WebGL），请换一台设备或告诉老师。',
  structure_render_failed: '三维模型加载失败，请刷新页面再试；如果一直失败，请告诉老师。',
  // 结果解释
  explain_title: '结果解释',
  explain_family_label: '晶体类型',
  explain_prototype_label: '结构类型',
  explain_coordination_label: '配位数',
  explain_coordination_note: '配位数：一个原子（或离子）周围最近邻的原子（或离子）个数。',
  explain_cell_title: '晶胞有多大',
  explain_edge_title: '三条棱的长度',
  explain_angle_title: '三个夹角',
  explain_unit_note: '棱长单位是 Å（埃，1 Å = 0.1 纳米），夹角单位是度（°）。',
  explain_rule_title: '为什么是这个结构',
  explain_caveats_title: '要注意的地方',
  family_ionic: '离子晶体：带正电的阳离子和带负电的阴离子靠静电吸引结合。',
  family_covalent: '共价晶体：原子之间共用电子对结合。',
  family_metallic: '金属晶体：金属原子靠金属键堆积在一起。',
  cell_shape_cubic: '这是一个立方晶胞：三条棱一样长，三个夹角都是 90°。',
  cell_shape_hexagonal:
    '这是一个六方晶胞：底面两条棱一样长、夹角 120°，第三条棱与底面垂直。',
  cell_shape_general: '晶胞的形状由下面三条棱和三个夹角决定。',
  // 页面导航
  back_home: '回到首页',
  predict_another: '再预测一个',
} satisfies Record<MessageKey, string>
