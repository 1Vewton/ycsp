// 英文文案：同时是“有哪些文案 tag”的基准——zh.ts 按这里的键补齐，缺键或写错键在类型检查时就报错。
// 任何语言缺一条文案时，界面回退到这份英文（见 index.ts）。

export const enMessages = {
  // Backend status codes
  backend_healthy: 'The data layer is running normally.',
  compute_healthy: 'The background services are running, so you can start predicting.',
  compute_unreachable:
    'The background computing service is not responding ({url}). Please tell your teacher.',
  compute_bad_status:
    'The background computing service returned an unexpected status ({status_code}). Please tell your teacher.',
  compute_read_failed:
    'Failed to read the response from the background computing service. Please try again or tell your teacher.',
  compute_invalid_json:
    'The response from the background computing service could not be read. Please tell your teacher.',
  unknown_code: 'Unexpected response from the backend (code {code}).',
  request_failed:
    'Cannot reach the backend. Please make sure the data layer and the compute layer are running.',
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
  rule_metal_experimental: 'The structure type of a metal comes from experiment, not from a formula.',
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
  element_not_supported: 'The element {symbol} is not available yet. Try a common element instead.',
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
  compute_no_result:
    'The compute layer returned no structure. Please try again later or tell your teacher.',
  // Task statuses (shown in results)
  task_status_pending: 'Queued',
  task_status_running: 'Computing',
  task_status_succeeded: 'Done',
  task_status_failed: 'Failed',
  task_status_cancelled: 'Cancelled',
  // Accounts and sessions
  user_registered: 'Account created and signed in.',
  user_ok: 'Current account loaded.',
  signed_in: 'Signed in.',
  signed_out: 'Signed out.',
  user_name_taken: 'That nickname is already taken. Please pick another one.',
  user_name_invalid: 'The nickname must be {min_length} to {max_length} characters long.',
  password_too_weak: 'Your password does not meet the requirements yet. Please check the hints below.',
  invalid_credentials: 'Wrong nickname or password. Please try again.',
  not_signed_in: 'You are not signed in yet.',
  session_unavailable:
    'The sign-in service is unavailable right now. Please try again later or tell your teacher.',
  too_many_attempts: 'Too many attempts. Please try again in about {retry_after_minutes} minute(s).',
  // Password rules (translate each entry of params.rules)
  password_rule_too_short: 'Use at least {min_length} characters.',
  password_rule_too_long: 'The password is too long (at most {max_bytes} bytes).',
  password_rule_missing_letter: 'Include at least one English letter.',
  password_rule_missing_digit: 'Include at least one digit.',
  password_rule_missing_symbol: 'Include at least one special character, such as ! @ # ?.',
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
  language_switcher_label: 'Interface language',
  main_navigation_label: 'Main navigation',
  nav_home: 'Home',
  nav_predict: 'Start predicting',
  ping_button_idle: 'Check background services',
  ping_button_busy: 'Checking...',
  ping_waiting: 'Waiting for the backend, please hold on...',
  // Home page
  home_heading: 'See what a crystal looks like from its formula',
  home_intro:
    'Type a chemical formula and the platform predicts the crystal structure it is most likely to form, and explains why. You do not need to write code or do any long calculations.',
  home_steps_title: 'Three steps to a result',
  home_step_input_title: 'Enter a chemical formula',
  home_step_input_description:
    'Write it the way you do in class, for example NaCl, CaF2, C or Cu.',
  home_step_predict_title: 'The platform works it out',
  home_step_predict_description:
    'It first decides which kind of crystal it is (ionic, covalent or metallic), then places the atoms using the rules from your textbook.',
  home_step_read_title: 'Understand the result',
  home_step_read_description:
    'You will see how big the unit cell (the smallest repeating unit of a crystal) is, where the atoms sit, and how reliable the answer is.',
  home_supported_title: 'What it can predict today',
  home_supported_ionic:
    'Ionic crystals: positively and negatively charged ions held together by electric attraction, for example NaCl, CsCl, CaF2.',
  home_supported_covalent:
    'Covalent crystals: atoms share electron pairs, for example C (diamond), Si, SiC.',
  home_supported_metallic:
    'Metallic crystals: a single metal element, for example Cu, Fe, Mg.',
  home_start_button: 'Start predicting',
  home_check_title: 'Check the background services',
  home_check_description:
    'If nothing happens after you press a button, check here whether the background services are running, then tell your teacher.',
  // Prediction page
  predict_heading: 'Predict a crystal structure',
  predict_intro:
    'Enter a chemical formula to see which crystal structure it is most likely to form.',
  formula_input_label: 'Chemical formula',
  formula_input_placeholder: 'For example NaCl',
  formula_input_hint:
    'Use element symbols and subscript numbers, for example CaF2. You can also use brackets, for example Ca(OH)2.',
  formula_syntax_ok: 'The formula looks well formed.',
  predict_start_button: 'Start predicting',
  predict_starting: 'Creating the task...',
  predict_examples_title: 'How to write a formula',
  predict_examples_hint: 'Or tap an example to fill it in.',
  example_group_ionic: 'Ionic crystals',
  example_group_covalent: 'Covalent crystals',
  example_group_metallic: 'Metallic crystals',
  // Result page
  result_heading: 'Prediction result',
  result_intro: 'This page shows the progress and the result of your prediction.',
  result_task_label: 'Task {task_id}',
  result_loading: 'Loading the task status...',
  result_polling_hint: 'The progress updates automatically; you do not need to refresh the page.',
  result_cancel_button: 'Cancel task',
  result_cancelling: 'Cancelling...',
  result_cancelled_hint: 'The task was cancelled, so there will be no result.',
  result_failed_hint: 'This run did not finish; the reason is shown above.',
  result_duration_seconds: 'This run took {seconds} seconds.',
  result_retry_button: 'Try again',
  // Structure viewer
  structure_title: 'Crystal structure',
  structure_loading: 'Loading the structure...',
  structure_preparing: 'Preparing the 3D model...',
  structure_drag_hint: 'Drag to rotate, use the scroll wheel to zoom.',
  structure_reset_view: 'Reset view',
  structure_canvas_label: '3D view of the crystal structure',
  structure_webgl_unavailable:
    'This device cannot show the 3D model because the browser does not support WebGL. Please try another device or tell your teacher.',
  structure_render_failed:
    'The 3D model could not be loaded. Please refresh the page; if it keeps failing, tell your teacher.',
  // Shared navigation
  back_home: 'Back to home',
  predict_another: 'Predict another',
} satisfies Record<string, string>

/** 全部文案 tag；写界面文字时用它约束键名，拼错会在类型检查时报错。 */
export type MessageKey = keyof typeof enMessages
