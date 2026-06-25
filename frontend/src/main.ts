import { createApp } from 'vue'
import App from './App.vue'

// 等待 Wails 运行时准备就绪
const startApp = () => {
  createApp(App).mount('#app')
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', startApp)
} else {
  startApp()
}
