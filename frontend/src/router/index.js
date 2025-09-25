/**
 * router/index.ts
 *
 * Automatic routes for `./src/pages/*.vue`
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router/auto'
import { setupLayouts } from 'virtual:generated-layouts'
import { routes } from 'vue-router/auto-routes'
import { useAppStore } from '@/stores/app'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: setupLayouts(routes),
})

// Authentication guard
router.beforeEach(async (to, from, next) => {
  const store = useAppStore()
  
  console.log(`🚦 ROUTER: Navigating to: ${to.path}, Authenticated: ${store.isAuthenticated}`)
  console.log(`🚦 ROUTER: Current URL: ${window.location.href}`)
  
  // Special handling for auth callback page
  if (to.path === '/auth/callback') {
    console.log('🚦 ROUTER: Auth callback page, allowing access')
    next()
    return
  }
  
  // Special handling for login page
  if (to.path === '/login') {
    // If already authenticated, redirect to home
    if (store.isAuthenticated) {
      console.log('🚦 ROUTER: Already authenticated, redirecting to home')
      next('/')
    } else {
      console.log('🚦 ROUTER: Not authenticated, showing login page')
      next()
    }
    return
  }
  
  // For all other pages, require authentication
  if (!store.isAuthenticated) {
    console.log('🚦 ROUTER: Not authenticated, redirecting to login')
    next('/login')
  } else {
    console.log('🚦 ROUTER: Authenticated, allowing access')
    next()
  }
})

// Workaround for https://github.com/vitejs/vite/issues/11804
router.onError((err, to) => {
  if (err?.message?.includes?.('Failed to fetch dynamically imported module')) {
    if (localStorage.getItem('vuetify:dynamic-reload')) {
      console.error('Dynamic import error, reloading page did not fix it', err)
    } else {
      console.log('Reloading page to fix dynamic import error')
      localStorage.setItem('vuetify:dynamic-reload', 'true')
      location.assign(to.fullPath)
    }
  } else {
    console.error(err)
  }
})

router.isReady().then(() => {
  localStorage.removeItem('vuetify:dynamic-reload')
})

export default router
