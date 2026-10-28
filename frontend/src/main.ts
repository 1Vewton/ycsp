import { createApp } from 'vue'

import { config } from '@fortawesome/fontawesome-svg-core'
import '@fortawesome/fontawesome-svg-core/styles.css'

import App from './App.vue'
import { router } from './router'
import './styles.css'

// 图标样式已经静态引入，关掉运行时自动注入，避免同一份 CSS 出现两次。
config.autoAddCss = false

createApp(App).use(router).mount('#app')
