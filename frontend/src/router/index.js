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
  
  // Special handling for login page with token
  if (to.path === '/login') {
    // Check if there's a token in the URL (OAuth callback)
    const urlParams = new URLSearchParams(window.location.search)
    const token = urlParams.get('token')
    
    if (token) {
      console.log('🚦 ROUTER: Token found in URL, processing immediately')
      console.log(`🚦 ROUTER: Token (first 20 chars): ${token.substring(0, 20)}...`)
      
      // Store the token immediately
      store.setToken(token)
      
      // Clean up URL
      const cleanUrl = window.location.pathname
      window.history.replaceState({}, document.title, cleanUrl)
      console.log('🚦 ROUTER: Token captured and URL cleaned')
      
      // Fetch user data
      try {
        console.log('🚦 ROUTER: Fetching user data...')
        await store.fetchUser()
        console.log('🚦 ROUTER: User data fetched successfully')
      } catch (error) {
        console.error('🚦 ROUTER: Failed to fetch user data:', error)
        // fetchUser already handles logout on error, so just return
        return
      }
      
      // Redirect to home
      console.log('🚦 ROUTER: Redirecting to home page')
      next('/')
      return
    }
    
    // If already authenticated and no token in URL, redirect to home
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
