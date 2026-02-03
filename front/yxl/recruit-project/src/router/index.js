import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/layout/home.vue'
import UpdateResume from '@/views/layout/updateResume.vue'
import Login from '@/views/login/login.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/home',
      name: 'HomePage',
      component: Home
    },
    {
      path: '/updateResume',
      name: 'updateResume',
      component: UpdateResume
    },
    {
      path: '/login',
      name: 'login',
      component: Login
    }
  ]
})

export default router
