import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/layout/home.vue'
import UpdateResume from '@/views/layout/updateResume.vue'
import Login from '@/views/login/index.vue'
import { useUserStore } from '@/stores/user'
import Admin from '@/views/user/admin.vue'
import Student from '@/views/user/student.vue'
import Company from '@/views/user/company/index.vue'
import CompanyChange from '@/views/user/company/change.vue'
import CompanyLike from '@/views/user/company/like.vue'
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'Main',
      component: Login
    },
    {
      path: '/home',
      name: 'Home',
      component: Home
    },
    {
      path: '/updateResume',
      name: 'updateResume',
      component: UpdateResume
    },
    {
      path: '/login',
      name: 'Login',
      component: Login
    },
    {
      path: '/company',
      name: 'Company',
      component: Company,
      redirect: '/like',
      children: [
        {
          path: '/change',
          name: 'CompanyChange',
          component: CompanyChange
        },
        {
          path: '/like',
          name: 'CompanyLike',
          component: CompanyLike
        }
      ]
    },
    {
      path: '/admin',
      name: 'Admin',
      component: Admin
    },
    {
      path: '/student',
      name: 'Student',
      component: Student
    }
  ]
})
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const whiteList = ['/login']
  if (whiteList.includes(to.path)) {
    next()
    return
  }
  if (!userStore.token) {
    next('/login')
    return
  }
  next()
})

export default router
