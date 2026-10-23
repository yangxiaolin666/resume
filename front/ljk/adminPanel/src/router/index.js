import { createRouter, createWebHistory } from 'vue-router'

import Home from '../views/Home.vue'

const router = createRouter({
  
  history: createWebHistory('/admin/'),
  routes: [
    {
      path: '/',
      name: 'adminHome',
      component: Home,
      meta: { 
        requiresAuth: true 
      }
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ],
})

//  线上入口可被 VITE_LOGIN_URL 覆盖，本地启动时指向 5173 主站
const loginUrl = import.meta.env.VITE_LOGIN_URL || 'http://47.93.133.36'

router.beforeEach((to, from, next) => {
  const urlParams = new URLSearchParams(window.location.search);
  const encodedToken = urlParams.get('token');
  
  if (encodedToken) {
    try {
      const token = decodeURIComponent(encodedToken);
      // console.log('token:', token);
      localStorage.setItem('token', token);
      window.history.replaceState({}, document.title, '/admin/');
      window.location.replace('/admin/');
      return;
    } catch (e) {
      // console.error('Token解析失败', e);
      window.location.replace(loginUrl);
      return;
    }
  } else {
    const gotToken = localStorage.getItem('token');
    // console.log('本地存储的Token:', gotToken);
    if (to.meta.requiresAuth) {
      gotToken ? next() : window.location.replace(loginUrl);
    } else {
      next();
    }
  }
})



export default router
