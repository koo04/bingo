// Utilities
import { defineStore } from 'pinia'
import axios from 'axios'
import websocketService from '@/services/websocket'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// Global error handler for authentication errors
let storeInstance = null

// Axios response interceptor for global error handling
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    // Check if this is an authentication-related error
    const isAuthError = (
      error.response?.status === 401 || 
      error.response?.status === 403 ||
      error.response?.data?.error?.toLowerCase().includes('bad token') ||
      error.response?.data?.error?.toLowerCase().includes('invalid token') ||
      error.response?.data?.error?.toLowerCase().includes('token expired') ||
      error.response?.data?.error?.toLowerCase().includes('unauthorized') ||
      error.response?.data?.message?.toLowerCase().includes('bad token') ||
      error.response?.data?.message?.toLowerCase().includes('invalid token') ||
      error.response?.data?.message?.toLowerCase().includes('token expired') ||
      error.response?.data?.message?.toLowerCase().includes('unauthorized')
    )

    if (isAuthError && storeInstance) {
      console.log('Authentication error detected, logging out user:', error.response?.data)
      storeInstance.handleAuthError(error)
    }

    return Promise.reject(error)
  }
)

export const useAppStore = defineStore('app', {
  state: () => ({
    user: null,
    token: null,
    isAdminUser: false,
    bingoCards: [],
    currentCard: null,
    loading: false,
    error: null,
    snackbar: {
      show: false,
      message: '',
      color: 'info'
    },
    globalMarkedItems: [],
    authInitialized: false,
  }),

  getters: {
    isAuthenticated: (state) => !!state.token && !!state.user,
    hasCurrentCard: (state) => !!state.currentCard,
    isItemGloballyMarked: (state) => (item) => state.globalMarkedItems.includes(item),
    isAdmin: (state) => state.isAdminUser
  },

  actions: {
    // Initialize store instance for global error handling
    init() {
      storeInstance = this
    },

    // Handle authentication errors globally
    handleAuthError(error) {
      const errorMessage = error.response?.data?.error || error.response?.data?.message || 'Authentication failed'
      console.log('Handling auth error:', errorMessage)
      
      // Show error message to user
      this.showSnackbar(`Session expired: ${errorMessage}`, 'error')
      
      // Log out the user
      this.logout()
    },

    // Helper method to check if error is authentication-related
    isAuthError(error) {
      return (
        error.response?.status === 401 || 
        error.response?.status === 403 ||
        error.response?.data?.error?.toLowerCase().includes('bad token') ||
        error.response?.data?.error?.toLowerCase().includes('invalid token') ||
        error.response?.data?.error?.toLowerCase().includes('token expired') ||
        error.response?.data?.error?.toLowerCase().includes('unauthorized') ||
        error.response?.data?.message?.toLowerCase().includes('bad token') ||
        error.response?.data?.message?.toLowerCase().includes('invalid token') ||
        error.response?.data?.message?.toLowerCase().includes('token expired') ||
        error.response?.data?.message?.toLowerCase().includes('unauthorized')
      )
    },

    // General API call method
    async apiCall(endpoint, method = 'GET', data = null) {
      try {
        const config = {
          method,
          url: `${API_BASE_URL}${endpoint}`,
        }
        
        if (data) {
          config.data = data
        }
        
        const response = await axios(config)
        return response.data
      } catch (error) {
        console.error('API call failed:', error)
        throw error
      }
    },

    showSnackbar(message, color = 'info') {
      this.snackbar = {
        show: true,
        message,
        color
      }
    },

    hideSnackbar() {
      this.snackbar.show = false
    },

    initializeWebSocket() {
      websocketService.on('initial_state', (data) => {
        this.globalMarkedItems = data.marked_items
      })
      
      websocketService.on('item_marked', (data) => {
        if (!this.globalMarkedItems.includes(data.item)) {
          this.globalMarkedItems.push(data.item)
        }
        // Update bingo cards if provided
        if (data.cards) {
          this.updateBingoCards(data.cards)
        }
      })
      
      websocketService.on('item_unmarked', (data) => {
        this.globalMarkedItems = this.globalMarkedItems.filter(item => item !== data.item)
        // Update bingo cards if provided
        if (data.cards) {
          this.updateBingoCards(data.cards)
        }
      })
      
      websocketService.connect()
    },

    updateBingoCards(newCards) {
      // Update the cards array with new data
      this.bingoCards = newCards
      
      // Update current card if it exists
      if (this.currentCard) {
        const updatedCurrentCard = newCards.find(card => card.id === this.currentCard.id)
        if (updatedCurrentCard) {
          this.currentCard = updatedCurrentCard
        }
      }
    },

    setToken(token) {
      console.log('Setting token:', !!token)
      this.token = token
      if (token) {
        console.log("setting token in localStorage and axios headers")
        localStorage.setItem('bingo_token', token)
        axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
      } else {
        console.log("removing token from localStorage and axios headers")
        localStorage.removeItem('bingo_token')
        localStorage.removeItem('bingo_user')
        delete axios.defaults.headers.common['Authorization']
      }
    },

    setUser(user) {
      console.log('Setting user:', user?.discord_id || 'null')
      this.user = user
      if (user) {
        localStorage.setItem('bingo_user', JSON.stringify(user))
      } else {
        localStorage.removeItem('bingo_user')
      }
    },

    async initializeAuth() {
      if (this.authInitialized) {
        console.log('Auth already initialized, skipping')
        return
      }
      
      console.log('Initializing authentication...')
      let token = localStorage.getItem('bingo_token')
      let tokenFromUrl = false
      
      if (!token) {
          const urlParams = new URLSearchParams(window.location.search)
          token = urlParams.get('token')
          if (token) {
            tokenFromUrl = true
            console.log('Token found in URL, will clean up after processing')
          }
      }

      const savedUser = localStorage.getItem('bingo_user')
      
      console.log(`Found token: ${!!token}, Found saved user: ${!!savedUser}`)
      
      if (token) {
        this.setToken(token)
        
        // Load saved user data first
        if (savedUser) {
          try {
            this.user = JSON.parse(savedUser)
            console.log('Loaded saved user data:', this.user?.discord_id)
          } catch (error) {
            console.error('Failed to parse saved user data:', error)
            localStorage.removeItem('bingo_user')
          }
        }
        
        // Try to fetch fresh user data
        try {
          console.log('Fetching fresh user data...')
          await this.fetchUser()
          console.log('Fresh user data fetched successfully')
          
          // URL cleanup is handled by the login page, not here
          // if (tokenFromUrl && this.isAuthenticated && typeof window !== 'undefined') {
          //   console.log('🏪 STORE: Cleaning up token from URL after successful authentication')
          //   this.cleanupTokenFromUrl()
          // }
          console.log('🏪 STORE: Skipping URL cleanup - letting login page handle it')
        } catch (error) {
          console.error('Failed to fetch fresh user data:', error)
          // If we have saved user data, continue with that
          if (!this.user) {
            console.log('No saved user data available, logging out')
            this.logout()
          } else {
            console.log('Continuing with saved user data')
            // URL cleanup is handled by the login page, not here
            // if (tokenFromUrl && typeof window !== 'undefined') {
            //   console.log('🏪 STORE: Cleaning up token from URL with saved user data')
            //   this.cleanupTokenFromUrl()
            // }
            console.log('🏪 STORE: Skipping URL cleanup - letting login page handle it')
          }
        }
      } else {
        console.log('No token found in localStorage')
      }
      
      this.authInitialized = true
      console.log(`Auth initialization complete. Authenticated: ${this.isAuthenticated}`)
    },

    async fetchUser() {
      try {
        this.loading = true
        const response = await axios.get(`${API_BASE_URL}/api/user`)
        this.setUser(response.data)
        // Check admin status after fetching user
        await this.checkAdminStatus()
      } catch (error) {
        console.error('Failed to fetch user:', error)
        // Auth errors are handled by the interceptor, so we don't need to handle them here
        throw error
      } finally {
        this.loading = false
      }
    },

    async checkAdminStatus() {
      try {
        const response = await this.apiCall('/api/admin/check')
        this.isAdminUser = response.is_admin
      } catch (error) {
        console.error('Error checking admin status:', error)
        // Only set false for non-auth errors (auth errors are handled by interceptor)
        if (!this.isAuthError(error)) {
          this.isAdminUser = false
        }
      }
    },

    async loginWithDiscord() {
      window.location.href = `${API_BASE_URL}/auth/discord`
    },

    logout() {
      console.log('Logging out user')
      this.isAdminUser = false
      this.bingoCards = []
      this.currentCard = null
      this.authInitialized = false
      this.error = null
      this.globalMarkedItems = []
      this.setUser(null)
      this.setToken(null)
      
      // Disconnect websocket if connected
      if (websocketService) {
        websocketService.disconnect()
      }
      
      // Redirect to login page
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
    },

    async generateNewBingoCard() {
      try {
        this.loading = true
        this.error = null
        const response = await axios.get(`${API_BASE_URL}/api/bingo/new`)
        this.currentCard = response.data
        await this.fetchBingoCards()
        return response.data
      } catch (error) {
        // Only handle non-auth errors (auth errors are handled by interceptor)
        if (!this.isAuthError(error)) {
          this.error = error.response?.data?.error || 'Failed to generate bingo card'
          console.error('Failed to generate bingo card:', error)
        }
        throw error
      } finally {
        this.loading = false
      }
    },

    async fetchBingoCards() {
      try {
        const response = await axios.get(`${API_BASE_URL}/api/bingo/cards`)
        this.bingoCards = response.data || []
        if (this.bingoCards.length > 0 && !this.currentCard) {
          this.currentCard = this.bingoCards[this.bingoCards.length - 1]
        }
      } catch (error) {
        console.error('Failed to fetch bingo cards:', error)
        // Only set empty array for non-auth errors (auth errors are handled by interceptor)
        if (!this.isAuthError(error)) {
          this.bingoCards = []
        }
      }
    },

    async markBingoItem(row, col) {
      if (!this.currentCard) return

      try {
        const response = await axios.post(`${API_BASE_URL}/api/bingo/mark`, {
          card_id: this.currentCard.id,
          row,
          col
        })
        
        this.currentCard = response.data
        
        // Update the card in the cards list
        const cardIndex = this.bingoCards.findIndex(card => card.id === this.currentCard.id)
        if (cardIndex !== -1) {
          this.bingoCards[cardIndex] = response.data
        }
        
        return response.data
      } catch (error) {
        // Only handle non-auth errors (auth errors are handled by interceptor)
        if (!this.isAuthError(error)) {
          this.error = error.response?.data?.error || 'Failed to mark bingo item'
          console.error('Failed to mark bingo item:', error)
        }
        throw error
      }
    },

    setCurrentCard(card) {
      this.currentCard = card
    },

    clearError() {
      this.error = null
    },

    // Clean up token from URL after successful authentication
    cleanupTokenFromUrl() {
      if (typeof window !== 'undefined') {
        try {
          console.log('🧹 CLEANUP: Current URL before cleanup:', window.location.href)
          const url = new URL(window.location)
          if (url.searchParams.has('token')) {
            console.log('🧹 CLEANUP: Found token in URL, removing...')
            url.searchParams.delete('token')
            // Construct the new URL without the token parameter
            let newUrl = url.pathname
            if (url.search && url.searchParams.toString()) {
              newUrl += '?' + url.searchParams.toString()
            }
            if (url.hash) {
              newUrl += url.hash
            }
            window.history.replaceState({}, document.title, newUrl)
            console.log('🧹 CLEANUP: Token parameter removed. New URL:', newUrl)
            console.log('🧹 CLEANUP: Current URL after cleanup:', window.location.href)
            
            // Add a small delay and check again to see if it gets restored
            setTimeout(() => {
              console.log('🧹 CLEANUP: URL check after 100ms:', window.location.href)
            }, 100)
          } else {
            console.log('🧹 CLEANUP: No token parameter found in URL')
          }
        } catch (error) {
          console.error('🧹 CLEANUP: Error cleaning up token from URL:', error)
        }
      }
    },

    // Test method to simulate auth errors (for testing purposes)
    async testAuthError() {
      try {
        // This will trigger a 401 response and should automatically log out the user
        await axios.get(`${API_BASE_URL}/api/test-invalid-token`)
      } catch (error) {
        console.log('Test auth error triggered')
      }
    }
  }
})
