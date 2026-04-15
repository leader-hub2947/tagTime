import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/notes'
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { transition: 'fade' }
    },
    {
      path: '/notes',
      name: 'Notes',
      component: () => import('../views/Notes.vue'),
      meta: { requiresAuth: true, transition: 'fade', index: 1 }
    },
    {
      path: '/tasks',
      name: 'Tasks',
      component: () => import('../views/Tasks.vue'),
      meta: { requiresAuth: true, transition: 'fade', index: 2 }
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { requiresAuth: true, transition: 'fade', index: 3 }
    },
    {
      path: '/ai-crush',
      name: 'AICrush',
      component: () => import('../views/AICrush.vue'),
      meta: { requiresAuth: true, transition: 'fade', index: 4 }
    },
    {
      path: '/timer',
      name: 'Timer',
      component: () => import('../views/Timer.vue'),
      meta: { requiresAuth: true, transition: 'scale', index: 5 }
    }
  ]
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/notes')
  } else {
    next()
  }
})

export default router
