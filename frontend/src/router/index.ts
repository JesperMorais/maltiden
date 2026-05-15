import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

// Lazy load all views for code splitting
const LandingView = () => import('@/views/LandingView.vue')
const OnboardingView = () => import('@/views/OnboardingView.vue')
const LoginView = () => import('@/views/LoginView.vue')
const DashboardView = () => import('@/views/DashboardView.vue')
const GenerateMenuView = () => import('@/views/GenerateMenuView.vue')
const AboutView = () => import('@/views/AboutView.vue')
const OffersView = () => import('@/views/OffersView.vue')
const RecipesView = () => import('@/views/RecipesView.vue')
const ShoppingListView = () => import('@/views/ShoppingListView.vue')
const AccountView = () => import('@/views/AccountView.vue')
const ForgotPasswordView = () => import('@/views/ForgotPasswordView.vue')
const ResetPasswordView = () => import('@/views/ResetPasswordView.vue')
const MenuHistoryView = () => import('@/views/MenuHistoryView.vue')

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
      path: '/forgot-password',
      name: 'forgot-password',
      component: ForgotPasswordView,
      meta: { transition: 'page-slide' }
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: ResetPasswordView,
      meta: { transition: 'page-slide' }
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: DashboardView,
      meta: { transition: 'page-fade', requiresAuth: true }
    },
    {
      path: '/menu/generate',
      name: 'generate-menu',
      component: GenerateMenuView,
      meta: { transition: 'page-slide', requiresAuth: true, requiresMember: true }
    },
    {
      path: '/about',
      name: 'about',
      component: AboutView,
      meta: { transition: 'page-fade' }
    },
    {
      path: '/recipes',
      name: 'recipes',
      component: RecipesView,
      meta: { transition: 'page-slide', requiresAuth: true, requiresMember: true }
    },
    {
      path: '/recipes/parse',
      redirect: '/recipes'
    },
    {
      path: '/shopping-list',
      name: 'shopping-list',
      component: ShoppingListView,
      meta: { transition: 'page-slide', requiresAuth: true, requiresMember: true }
    },
    {
      path: '/konto',
      name: 'account',
      component: AccountView,
      meta: { transition: 'page-slide', requiresAuth: true }
    },
    {
      path: '/offers-poc',
      name: 'offers-poc',
      component: OffersView,
      meta: { transition: 'page-fade' }
    },
    {
      path: '/menus',
      name: 'menu-history',
      component: MenuHistoryView,
      meta: { transition: 'page-slide', requiresAuth: true, requiresMember: true }
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/ProfileView.vue'),
      meta: { transition: 'page-slide', requiresAuth: true }
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('@/views/SettingsView.vue'),
      meta: { transition: 'page-slide', requiresAuth: true }
    },
  ],
})

// Authentication guard
router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth && !userStore.isAuthenticated) {
    // Redirect to login if not authenticated
    next({ name: 'login', query: { redirect: to.fullPath } })
  } else if (to.meta.requiresMember && !userStore.isMember) {
    // Redirect to dashboard if member access required but user is guest
    next({ name: 'dashboard' })
  } else if (to.name === 'login' && userStore.isAuthenticated) {
    // Redirect to dashboard if already authenticated
    next({ name: 'dashboard' })
  } else {
    next()
  }
})

export default router
