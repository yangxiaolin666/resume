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

//  47.93.133.36
router.beforeEach((to, from, next) => {
  const urlParams = new URLSearchParams(window.location.search);
  const encodedToken = urlParams.get('token');
  
  if (encodedToken) {
    try {
      const token = decodeURIComponent(encodedToken);
      // console.log('token:', token);
      localStorage.setItem('token', token);
      window.history.replaceState({}, document.title, '/admin/');
      window.location.replace('http://47.93.133.36/admin/');
      return;
    } catch (e) {
      // console.error('Token解析失败', e);
      window.location.replace("http://47.93.133.36/");
      return;
    }
  } else {
    const gotToken = localStorage.getItem('token');
    // console.log('本地存储的Token:', gotToken);
    if (to.meta.requiresAuth) {
      gotToken ? next() : window.location.replace("http://47.93.133.36/");
    } else {
      next();
    }
  }
})



export default router
