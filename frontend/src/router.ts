// 页面路由：首页、预测页、结果页，以及登录、找回密码、账号设置、预测记录。
//
// 用 hash 模式（地址形如 `/#/predict`）：静态托管时直接打开或刷新任何页面都不会 404，
// 部署方式还没定（TASKS.md Q5），这样就不用为服务器补“回退到 index.html”的规则。

import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHashHistory } from 'vue-router'

import HomeView from './views/HomeView.vue'
import AccountView from './views/AccountView.vue'
import HistoryView from './views/HistoryView.vue'
import LoginView from './views/LoginView.vue'
import PredictView from './views/PredictView.vue'
import ResetPasswordView from './views/ResetPasswordView.vue'
import ResultView from './views/ResultView.vue'

const routes: RouteRecordRaw[] = [
  { path: '/', name: 'home', component: HomeView },
  { path: '/predict', name: 'predict', component: PredictView },
  // 登录与注册共用一页：学生不用在两张表单之间找入口。
  { path: '/login', name: 'login', component: LoginView },
  // 忘记密码：先按昵称取安全问题，再用答案重置。
  { path: '/reset-password', name: 'reset-password', component: ResetPasswordView },
  // 登录后的账号设置：修改密码与安全问题。
  { path: '/account', name: 'account', component: AccountView },
  // 个人预测记录；未登录时页面会提示去登录。
  { path: '/history', name: 'history', component: HistoryView },
  // 任务编号写在地址里：刷新结果页还能回到同一个任务。
  { path: '/result/:taskId', name: 'result', component: ResultView },
  // 地址输错时不要停在空白页，直接回首页。
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})
