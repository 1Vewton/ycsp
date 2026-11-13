// 英文文案：同时是“有哪些文案 tag”的基准——zh.ts 按这里的键补齐，缺键或写错键在类型检查时就报错。
// 任何语言缺一条文案时，界面回退到这份英文（见 index.ts）。

export const enMessages = {
  // Backend status codes
  backend_healthy: 'The data layer is running normally.',
  compute_healthy: 'The background services are running, so you can start predicting.',
  compute_unreachable:
    'The background computing service is not responding ({url}). Please tell your teacher.',
  compute_timeout:
    'This calculation is taking too long, probably because many people are using the platform. Please try again later; if it keeps happening, tell your teacher.',
  compute_bad_status:
    'The background computing service returned an unexpected status ({status_code}). Please tell your teacher.',
  compute_read_failed:
    'Failed to read the response from the background computing service. Please try again or tell your teacher.',
  compute_invalid_json:
    'The response from the background computing service could not be read. Please tell your teacher.',
  unknown_code:
    'The backend returned a message we cannot explain yet (code {code}). Please refresh the page; if it keeps happening, tell your teacher.',
  request_failed:
    'Cannot reach the background services. Please refresh the page; if it keeps failing, tell your teacher.',
  // Prediction result: reasoning (rule)
  prediction_ready: 'The predicted structure is ready.',
  rule_radius_ratio:
    'The cation radius is {ratio} times the anion radius. The smaller the ratio, the fewer opposite ions fit around it; by this ratio each cation touches {coordination_number} anions.',
  rule_experimental_override:
    'The radius-ratio rule would predict {rule_coordination_number}-fold coordination, but the measured structure is different, so the experimental data is used here (radius ratio {ratio}).',
  rule_wurtzite_experimental:
    'The measured structure is the wurtzite type (hexagonal ZnS): each ion is surrounded by {coordination_number} opposite ions in a tetrahedral arrangement. The radius ratio {ratio} is shown only for reference.',
  rule_rutile_experimental:
    'With a 1:2 ion ratio the usual prediction is the fluorite type ({rule_coordination_number}-coordination). At radius ratio {ratio} the measured structure is the rutile type, where each cation touches {coordination_number} anions, so the experimental structure is used here.',
  rule_fluorite_stoichiometry:
    'The cation to anion ratio is {cation_count}:{anion_count} and the radius ratio {ratio} falls in the 8-coordination range: the large ions form a face-centered cubic array (one at every corner and at the centre of every face), and the small ions fill the tetrahedral holes between four large ions. This is the fluorite type.',
  rule_antifluorite_stoichiometry:
    'The anion to cation ratio is {anion_count}:{cation_count}; the structure is the same as fluorite with the roles of the two ions swapped (antifluorite type).',
  rule_diamond_covalent:
    'This is a covalent crystal of one element: every atom shares electron pairs with 4 neighbours, forming the diamond structure.',
  rule_zinc_blende_covalent:
    'The two elements differ by only {electronegativity_difference} in electronegativity (how strongly an atom pulls electrons), so they bond covalently and alternate in the zinc blende (cubic ZnS) arrangement.',
  rule_metal_experimental: 'The structure type of a metal comes from experiment, not from a formula.',
  // Prediction result: caveats
  caveat_radius_ratio_empirical:
    'The radius-ratio rule is only a rule of thumb: it gives the most likely coordination number, while the real structure is decided by experiment.',
  caveat_near_boundary:
    'The radius ratio {ratio} is very close to the boundary {boundary}, so either coordination number is possible.',
  caveat_confirmed_by_experiment:
    'The measured structure agrees with the reasoning above, so it can be trusted.',
  caveat_polymorph:
    'A substance can have more than one crystal structure (polymorphism); this result uses the experimentally measured form. ZnS, for example, also has a cubic zinc blende form.',
  caveat_low_ionic_character:
    'The electronegativity difference (how differently the two atoms pull electrons) is only {difference} (below {threshold}), so the bonding is noticeably covalent; such crystals sit between ionic and covalent crystals.',
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
    'The background service hit an unexpected error. Please try again later; if it keeps failing, tell your teacher.',
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
    'The background computing service returned no structure. Please try again later or tell your teacher.',
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
  password_changed: 'Your password has been updated. Use the new password next time you sign in.',
  password_reset: 'Your password has been reset and you are signed in.',
  password_reset_question: 'Please answer your security question.',
  security_question_updated: 'Your security question has been saved.',
  current_password_incorrect: 'The current password is not correct. Please try again.',
  password_unchanged: 'The new password is the same as the old one. Please choose another.',
  security_question_invalid:
    'The security question must be {min_length} to {max_length} characters long.',
  security_answer_invalid:
    'The answer must be at least {min_length} characters and at most {max_bytes} bytes.',
  security_setup_incomplete: 'Fill in both the security question and the answer, or leave both empty.',
  password_reset_unavailable:
    'This nickname cannot reset the password by itself (it may not have a security question). Check the nickname or ask your teacher for help.',
  password_reset_failed:
    'The nickname or the security answer is not correct. Please check and try again; if you forgot it, ask your teacher for help.',
  task_history: 'Prediction history loaded.',
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
  prototype_wurtzite: 'Wurtzite (hexagonal ZnS) type',
  prototype_fluorite: 'Fluorite type',
  prototype_antifluorite: 'Antifluorite type',
  prototype_rutile: 'Rutile type',
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
    'Ionic crystals: positively and negatively charged ions held together by electric attraction, for example NaCl, CsCl, CaF2, ZnO.',
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
    'This device cannot show the 3D model because the browser is too old for WebGL 2. Please try another device or tell your teacher.',
  structure_render_failed:
    'The 3D model could not be loaded. Please refresh the page; if it keeps failing, tell your teacher.',
  // Explanation
  explain_title: 'What the result means',
  explain_family_label: 'Crystal type',
  explain_prototype_label: 'Structure type',
  explain_coordination_label: 'Coordination number',
  explain_coordination_note:
    'Coordination number: how many nearest neighbours an atom or ion has.',
  explain_cell_title: 'How big is the unit cell',
  explain_edge_title: 'Edge lengths',
  explain_angle_title: 'Angles between the edges',
  explain_unit_note:
    'Edge lengths are in Å (angstrom, 1 Å = 0.1 nm); angles are in degrees (°).',
  explain_rule_title: 'Why this structure',
  explain_caveats_title: 'Things to keep in mind',
  family_ionic:
    'Ionic crystal: positively charged cations and negatively charged anions held together by electric attraction.',
  family_covalent: 'Covalent crystal: atoms are held together by shared electron pairs.',
  family_metallic: 'Metallic crystal: metal atoms packed together by metallic bonding.',
  cell_shape_cubic:
    'This is a cubic unit cell: all three edges are equal and all angles are 90°.',
  cell_shape_tetragonal:
    'This is a tetragonal unit cell: the two base edges are equal, the third edge is different, and all angles are 90°.',
  cell_shape_hexagonal:
    'This is a hexagonal unit cell: two edges of the base are equal with a 120° angle, and the third edge is perpendicular to the base.',
  cell_shape_general:
    'The shape of the unit cell is set by the three edges and three angles below.',
  // Sign-in page and account menu
  nav_login: 'Sign in',
  nav_logout: 'Sign out',
  nav_signing_out: 'Signing out…',
  nav_account_label: 'Account: {name}',
  auth_checking_session: 'Checking whether you are signed in…',
  auth_signed_in_title: 'You are signed in',
  auth_signed_in_as: 'Signed in as {name}.',
  auth_continue_button: 'Start predicting',
  auth_sign_out_button: 'Sign out',
  auth_sign_out_busy: 'Signing out…',
  auth_fields_required: 'Please fill in the nickname and password.',
  auth_nickname_required: 'Please enter your nickname first.',
  auth_reset_fields_required: 'Please fill in the answer and the new password.',
  auth_password_mismatch: 'The two passwords do not match. Please type them again.',
  login_heading: 'Sign in or create an account',
  login_intro:
    'Signing in keeps your predictions with your account. You can also predict without signing in.',
  login_mode_label: 'Sign in or create an account',
  login_mode_sign_in: 'Sign in',
  login_mode_register: 'Create an account',
  login_display_name_label: 'Nickname',
  login_display_name_placeholder: 'For example: xiaoming',
  login_display_name_hint: '2 to 32 characters. You will also use it to sign in.',
  login_password_label: 'Password',
  login_password_placeholder: 'Enter your password',
  login_password_confirm_label: 'Password again',
  login_password_confirm_placeholder: 'Enter the same password again',
  login_password_policy_hint:
    'At least 8 characters, with an English letter, a digit and a special character.',
  login_submit_sign_in: 'Sign in',
  login_submit_register: 'Create account and sign in',
  login_submitting: 'Please wait…',
  login_switch_to_register: 'No account yet? Create one',
  login_switch_to_sign_in: 'Already have an account? Sign in',
  login_anonymous_hint: 'No account needed to try a prediction.',
  login_anonymous_link: 'Predict without signing in',
  nav_history: 'My history',
  login_forgot_password_link: 'Forgot your password?',
  login_security_question_label: 'Security question (optional)',
  login_security_question_placeholder: 'For example: What is my favourite element?',
  login_security_answer_label: 'Security answer (optional)',
  login_security_answer_placeholder: 'For example: sodium',
  login_security_hint:
    'Set both to reset the password yourself later. The answer ignores case and extra spaces.',
  auth_security_incomplete: 'Fill in both the security question and the answer, or leave both empty.',
  account_heading: 'Account settings',
  account_intro: 'Change your password or set the security question used to reset it.',
  account_signed_out_hint: 'Sign in to change your account settings.',
  account_sign_in_link: 'Sign in',
  account_password_title: 'Change password',
  account_password_intro:
    'Enter your current password first. You stay signed in on this device after the change.',
  account_current_password_label: 'Current password',
  account_new_password_label: 'New password',
  account_new_password_confirm_label: 'New password again',
  account_change_password_button: 'Save new password',
  account_password_submitting: 'Saving…',
  account_password_changed: 'Password updated. Use the new password next time you sign in.',
  account_security_title: 'Security question',
  account_security_intro:
    'With a security question you can reset the password yourself. Changing it needs your current password.',
  account_security_current_question: 'Current question: {question}',
  account_security_not_set: 'No security question has been set yet.',
  account_security_question_label: 'Security question',
  account_security_question_placeholder: 'For example: What is my favourite element?',
  account_security_answer_label: 'Answer',
  account_security_answer_placeholder: 'For example: sodium',
  account_security_setup_button: 'Save security question',
  account_security_submitting: 'Saving…',
  account_security_updated: 'Security question saved.',
  reset_heading: 'Reset your password',
  reset_intro:
    'Enter your nickname, answer the security question you set, then choose a new password.',
  reset_display_name_label: 'Nickname',
  reset_lookup_button: 'Next',
  reset_lookuping: 'Looking up the security question…',
  reset_question_intro: 'Your security question is:',
  reset_answer_label: 'Answer',
  reset_answer_placeholder: 'Enter your answer',
  reset_new_password_label: 'New password',
  reset_new_password_confirm_label: 'New password again',
  reset_submit_button: 'Reset password and sign in',
  reset_submitting: 'Resetting…',
  reset_back_to_login: 'Remembered it? Go to sign in',
  history_heading: 'My prediction history',
  history_intro:
    'Only predictions created while signed in are listed; anonymous predictions are not saved to an account.',
  history_sign_in_hint: 'Sign in to see your own prediction history.',
  history_sign_in_link: 'Sign in',
  history_empty: 'No predictions yet.',
  history_empty_link: 'Make your first prediction',
  history_load_more: 'Load more',
  history_loading_more: 'Loading…',
  history_showing: 'Showing {shown} of {total}',
  history_open_button: 'View result',
  // Shared navigation
  back_home: 'Back to home',
  predict_another: 'Predict another',
} satisfies Record<string, string>

/** 全部文案 tag；写界面文字时用它约束键名，拼错会在类型检查时报错。 */
export type MessageKey = keyof typeof enMessages
