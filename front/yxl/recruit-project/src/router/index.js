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
const CompanyMyJobs = () => import('@/views/user/company/myJobs.vue')
const studentDetail = () => import('@/views/studentDetail/index.vue')
const SelfMessage = () => import('@/views/user/student/selfMessage.vue')
const Test = () => import('@/views/test/test.vue')
const ShowData = () => import('@/views/layout/showData.vue')
const Job = () => import('@/views/layout/Job.vue')
const Chat = () => import('@/views/chat/index.vue')

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
      redirect:'/chat',
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
        {
          path:'/job',
          name:'job',
          component:Job 
        },
        {
          path: '/chat',
          name: 'Chat',
          component: Chat
        }
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
        },
        {
          path: '/myJobs',
          name: 'CompanyMyJobs',
          component: CompanyMyJobs
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
          path: '/studentSetting',
          name: 'StudentSetting',
          component: CompanyChange
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
// 判断本地保存的 JWT 是否已过期
const isTokenExpired = (token) => {
  if (!token) return true
  try {
    const part = token.split('.')[1] || ''
    const base64 = part.replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4)
    const payload = JSON.parse(atob(padded))
    return !payload.exp || payload.exp * 1000 <= Date.now()
  } catch (e) {
    return true
  }
}

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  const whiteList = ['/login']
  if (whiteList.includes(to.path)) {
    next()
    return
  }
  if (!userStore.token || !userStore.userType || isTokenExpired(userStore.token)) {
    userStore.clear()
    next('/login')
    return
  }
  next()
})

export default router
