import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
// 懒加载组件
const Home = () => import('../views/layout/home.vue')
const MakeResume=() =>import('@/views/user/student/makeResume.vue')
const LayOut = () => import('@/views/layout/index.vue')
const UpdateResume = () => import('@/views/user/student/updateResume.vue')
const Login = () => import('@/views/login/index.vue')
const Admin = () => import('@/views/user/admin.vue')
const Student = () => import('@/views/user/student/index.vue')
const Company = () => import('@/views/user/company/index.vue')
const CompanyChange = () => import('@/views/user/company/change.vue')
const CompanyLike = () => import('@/views/user/company/like.vue')
const studentDetail = () => import('@/views/studentDetail/index.vue')
const SelfMessage = () => import('@/views/user/student/selfMessage.vue')
const Test = () => import('@/views/test/test.vue')
const ShowData = () => import('@/views/layout/showData.vue')
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'Main',
      component: Login
    },
    {
      path: '/layout',
      name: 'Layout',
      component: LayOut,
      redirect:'/showData',
      children:[
        {
          path: '/home',
          name: 'Home',
          component: Home
        },
        {
          path: '/showData',
          name: 'showData',
          component: ShowData
        },
      ]
    },
    {
      path: '/login',
      name: 'Login',
      component: Login
    },
    {
      path: '/test',
      name: 'Test',
      component: Test
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
      component: Student,
      redirect:'/selfMessage',
      children:[
        {
          path:'/selfMessage',
          name:'SelfMessage',
          component:SelfMessage
        },
        {
          path: '/updateResume',
          name: 'updateResume',
          component: UpdateResume
        },
        {
          path: '/makeResume',
          name: 'makeResume',
          component: MakeResume
        },
      ]
    },
    {
      path: '/studentDetail',
      name: 'StudentDetail',
      component: studentDetail,
      props:true
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
