import { createRouter, createWebHistory } from 'vue-router'

import Home from '../views/Home.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home,
      meta: { 
        requiresAuth: false
      }
    }
  ],
})

/*

router.beforeEach(async (to, from, next) => {
  // 1. 获取本地 Token（可能是登录页跳转过来时存储的）
  const token = getToken()

  // 2. 如果访问需要登录的页面（你的 Home 页）
  if (to.meta.requiresAuth) {
    if (token) {
      // 3. 有 Token → 校验 Token 是否有效
      const isTokenValid = await validateToken(token)
      if (isTokenValid) {
        // ✅ Token 有效 → 放行到 Home 页
        next()
      } else {
        // ❌ Token 无效 → 删除本地 Token + 跳回登录页
        removeToken()
        next({ path: '/login', query: { reason: 'token_invalid' } })
      }
    } else {
      // ❌ 无 Token → 直接跳登录页
      next('/login')
    }
  } else {
    // 访问登录页 → 额外校验：如果有 Token 且有效，自动跳 Home 页（优化体验）
    if (token) {
      const isTokenValid = await validateToken(token)
      if (isTokenValid) {
        next('/') // 已登录且 Token 有效，直接跳 Home
      } else {
        removeToken() // Token 无效，清空后留在登录页
        next()
      }
    } else {
      next() // 无 Token，放行到登录页
    }
  }
})

*/

export default router
