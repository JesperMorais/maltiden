import { createRouter, createWebHistory } from 'vue-router'
import LandingView from '@/views/LandingView.vue'
import OnboardingView from '@/views/OnboardingView.vue'
import LoginView from '@/views/LoginView.vue'
import DashboardView from '@/views/DashboardView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'landing',
      component: LandingView,
      meta: { transition: 'page-fade' }
    },
    {
      path: '/register',
      name: 'register',
      component: OnboardingView,
      meta: { transition: 'page-slide' }
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { transition: 'page-slide' }
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
      meta: { transition: 'page-fade' }
    },
  ],
})

export default router
