import { createApp } from 'vue'
import { createPinia } from 'pinia'
import persist from 'pinia-plugin-persistedstate'
import App from './App.vue'
import router from './router'
import './style.css'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
const app = createApp(App)
const initDarkMode = () => {
  const savedMode = localStorage.getItem('element-plus-dark-mode')
  let isDark = false
  
  if (savedMode !== null) {
    isDark = savedMode === 'true'
  } 
  if (isDark) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  
  return isDark
}
initDarkMode()

window.toggleDarkMode = () => {
  const isDark = document.documentElement.classList.contains('dark')
  const newMode = !isDark
  if (newMode) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  localStorage.setItem('element-plus-dark-mode', newMode.toString())
  return newMode
}
window.getDarkMode = () => {
  return document.documentElement.classList.contains('dark')
}

app.use(ElementPlus)
app.use(createPinia().use(persist))
app.use(router)
app.mount('#app')
