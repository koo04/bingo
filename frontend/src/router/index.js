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

// Navigation guard to redirect unauthenticated users to login
router.beforeEach(async (to, from, next) => {
  const store = useAppStore()
  
  console.log(`🚦 ROUTER: Navigating from ${from.path} to: ${to.path}`)
  console.log(`🚦 ROUTER: Current URL: ${window.location.href}`)
  console.log(`🚦 ROUTER: Auth initialized: ${store.authInitialized}, Token: ${!!store.token}, User: ${!!store.user}, Authenticated: ${store.isAuthenticated}`)
  
  // Special handling for login page
  if (to.path === '/login') {
    // Check for OAuth callback token in URL
    const urlToken = to.query.token || new URLSearchParams(window.location.search).get('token')
    
    if (urlToken) {
      console.log('OAuth callback detected, allowing access to login page')
      // Allow login page to handle the OAuth callback
      next()
      return
    }
    
    // Wait for auth initialization if not done yet
    if (!store.authInitialized) {
      console.log('Waiting for auth initialization on login page...')
      let retries = 0
      while (retries < 50 && !store.authInitialized) {
        await new Promise(resolve => setTimeout(resolve, 100))
        retries++
      }
    }
    
    // If already authenticated and no token in URL, redirect to home
    if (store.isAuthenticated) {
      console.log('Already authenticated, redirecting to home from login page')
      next('/')
    } else {
      console.log('Not authenticated, showing login page')
      next()
    }
    return
  }
  
  // For all other pages, wait for auth initialization
  if (!store.authInitialized) {
    console.log('Waiting for auth initialization...')
    let retries = 0
    while (retries < 50 && !store.authInitialized) {
      await new Promise(resolve => setTimeout(resolve, 100))
      retries++
    }
  }
  
  console.log(`After waiting - Authenticated: ${store.isAuthenticated}`)
  
  // For admin page, check if user is authenticated and has admin privileges
  if (to.path === '/admin') {
    if (!store.isAuthenticated) {
      console.log('Not authenticated for admin page, redirecting to login')
      next('/login')
      return
    }
    // Check admin status if not already checked
    if (store.user && !store.hasOwnProperty('isAdminUser')) {
      await store.checkAdminStatus()
    }
    // Allow access regardless of admin status - the page itself will show access denied
    next()
    return
  }
  
  // For all other pages, check if user is authenticated
  if (!store.isAuthenticated) {
    console.log('Not authenticated, redirecting to login.')
    next('/login')
  } else {
    console.log('Authenticated, allowing access')
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
