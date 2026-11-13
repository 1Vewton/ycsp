> English | [简体中文](./README.zh-CN.md)

# Frontend · Pages (frontend/src/views)

Each file is a page component; routes are registered in [`../router.ts`](../router.ts).

| File | Route | What it does now | What is still missing |
| --- | --- | --- | --- |
| `HomeView.vue` | `/#/` | Explains what the platform does, three steps to start, supported range; background service connectivity check | — |
| `PredictView.vue` | `/#/predict` | Formula input (instant syntax validation) + example shortcuts + start prediction (creates a task and navigates to the result page) | — |
| `ResultView.vue` | `/#/result/:taskId` | Polls task progress (queued → running → finished), cancels queued tasks, shows failure reason and duration, retries after network loss; on success shows the 3D structure and result explanation | — |
| `LoginView.vue` | `/#/login` | Login and registration share one form (mode switch, repeat password, optional security question), password rules shown one by one, forgot-password link, current account and sign-out when signed in | — |
| `ResetPasswordView.vue` | `/#/reset-password` | Looks up the security question by nickname, collects the answer and a new password; the backend signs the student in on success | — |
| `AccountView.vue` | `/#/account` | Signed-in account settings: change password and set / update the security question | — |
| `HistoryView.vue` | `/#/history` | Lists the signed-in account's predictions newest first, links to each result and loads more pages | — |

## Conventions

- Pages fetch text only through `t()` from `../locales`; never hard-code text in any language.
- Sign-in state comes only from `signedInUser` in [`../session.ts`](../session.ts); pages do not parse cookies or store accounts themselves.
- After sign-in, return to the in-site address from before the login page (`?redirect=`); external addresses are always ignored.
- Shared styles such as colors, buttons, and page containers live in `../styles.css`; styles specific to one page belong in the component's `<style scoped>`.
- The result page address contains the task ID (`/#/result/<id>`): the ID comes from the route parameter, so refreshing returns to the same task.
- History only contains tasks created while signed in; anonymous tasks have no owner and are not shown.
- Password change and security-question setup both require the current password; the forms stay on the page and show translatable errors.
- Unfinished areas of the page skeleton use a `placeholder-panel` to explain "what will be here later"; no blank space, and no pretending it already works.
