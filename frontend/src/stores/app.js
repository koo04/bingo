// Utilities
import { defineStore } from 'pinia'
import axios from 'axios'
import websocketService from '@/services/websocket'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const useAppStore = defineStore('app', {
  state: () => ({
    // Authentication
    token: null,
    user: null,

    // Bingo game
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

    // Themes
    themes: [],
    activeThemeId: null,

    // Server connectivity
    isServerConnected: false,
    isCheckingConnection: false,
    connectionCheckInterval: null,

    // Internal state
    wsInitialized: false,
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    hasCurrentCard: (state) => !!state.currentCard,
    isItemGloballyMarked: (state) => (item) => state.globalMarkedItems.includes(item),
    isAdmin: (state) => state.user?.is_admin || false,
    activeTheme: (state) => state.themes.find(theme => theme.id === state.activeThemeId),
    hasActiveTheme: (state) => !!state.activeThemeId,
    isAppReady: (state) => state.isServerConnected && !!state.token,
    connectionStatus: (state) => {
      if (state.isCheckingConnection) return 'checking'
      if (state.isServerConnected) return 'connected'
      return 'disconnected'
    },
  },

  actions: {
    // Authentication methods
    setToken(token) {
      this.token = token
      if (token) {
        localStorage.setItem('bingo_token', token)
        axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
      } else {
        localStorage.removeItem('bingo_token')
        delete axios.defaults.headers.common['Authorization']
      }
    },

    async loadTokenFromStorage() {
      const token = localStorage.getItem('bingo_token')
      if (token) {
        console.log('Found token in localStorage, setting it')
        this.setToken(token)
        // Try to fetch user data
        try {
          await this.fetchUser()
          console.log('User data loaded from existing token')
        } catch (error) {
          console.warn('Failed to fetch user with stored token, token may be expired or invalid')
          // fetchUser already handles logout on error
        }
      } else {
        console.log('No token found in localStorage')
      }
    },

    async fetchUser() {
      if (!this.token) {
        throw new Error('No token available')
      }

      try {
        const response = await this.apiCall('/api/user')
        this.user = response
        console.log('User fetched successfully:', this.user?.username)
        return response
      } catch (error) {
        console.error('Failed to fetch user:', error)
        // If user fetch fails, clear token (might be invalid)
        this.logout()
        throw error
      }
    },

    logout() {
      console.log('Logging out user')
      this.token = null
      this.user = null
      this.bingoCards = []
      this.currentCard = null
      this.themes = []
      this.activeThemeId = null
      this.globalMarkedItems = []
      this.error = null

      this.setToken(null)

      // Stop WebSocket and connection monitoring
      this.stopConnectionMonitoring()
      if (this.wsInitialized) {
        websocketService.disconnect()
        this.wsInitialized = false
      }

      // Redirect to login
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
    },

    // Debug method to force clear invalid tokens
    clearStorageAndReload() {
      console.log('Clearing all storage and reloading')
      localStorage.clear()
      sessionStorage.clear()
      if (typeof window !== 'undefined') {
        window.location.reload()
      }
    },    // General API call method
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

        // Handle 401 Unauthorized errors
        if (error.response?.status === 401) {
          console.log('Received 401 Unauthorized - token may be invalid, logging out')
          this.logout()
          return // Don't re-throw for 401, just logout
        }

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
      if (this.wsInitialized) {
        console.log('WebSocket already initialized, skipping')
        return
      }

      console.log('Initializing WebSocket...')
      this.wsInitialized = true

      websocketService.on('initial_state', (data) => {
        console.log('WebSocket initial_state received')
        this.globalMarkedItems = data.marked_items
      })

      websocketService.on('item_marked', (data) => {
        console.log('WebSocket item_marked received:', data.item)
        if (!this.globalMarkedItems.includes(data.item)) {
          this.globalMarkedItems.push(data.item)
        }
        // Update bingo cards if provided
        if (data.cards) {
          this.updateBingoCards(data.cards)
        }
      })

      websocketService.on('item_unmarked', (data) => {
        console.log('WebSocket item_unmarked received:', data.item)
        this.globalMarkedItems = this.globalMarkedItems.filter(item => item !== data.item)
        // Update bingo cards if provided
        if (data.cards) {
          this.updateBingoCards(data.cards)
        }
      })

      websocketService.on('theme_changed', (data) => {
        console.log('Theme changed via WebSocket:', data)
        this.activeThemeId = data.item // The item field contains the new theme ID
        this.showSnackbar('Admin changed the active theme. You may need to generate a new card.', 'info')
        // Don't automatically refresh themes as it might cause unnecessary rerenders
        // this.fetchThemes()
      })

      websocketService.on('theme_updated', (data) => {
        console.log('Theme updated via WebSocket:', data)
        // Update the theme in the local themes array
        if (data.item && data.item.id) {
          const themeIndex = this.themes.findIndex(theme => theme.id === data.item.id)
          if (themeIndex !== -1) {
            this.themes[themeIndex] = data.item
            this.showSnackbar('Theme updated successfully', 'success')
          }
        }
      })

      websocketService.on('theme_created', (data) => {
        console.log('Theme created via WebSocket:', data)
        // Add the new theme to the local themes array
        if (data.item && data.item.id) {
          this.themes.push(data.item)
          this.showSnackbar(`New theme created: ${data.item.name}`, 'success')
        }
      })

      websocketService.on('theme_deleted', (data) => {
        console.log('Theme deleted via WebSocket:', data)
        // Remove the theme from the local themes array
        if (data.item && data.item.id) {
          this.themes = this.themes.filter(theme => theme.id !== data.item.id)
          this.showSnackbar(`Theme deleted: ${data.item.name}`, 'info')

          // Clear active theme if it was deleted
          if (this.activeThemeId === data.item.id) {
            this.activeThemeId = null
          }
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

    async generateNewBingoCard() {
      try {
        this.loading = true
        this.error = null
        const response = await axios.get(`${API_BASE_URL}/api/bingo/new`)
        this.currentCard = response.data
        await this.fetchBingoCards()
        return response.data
      } catch (error) {
        this.error = error.response?.data?.error || 'Failed to generate bingo card'
        console.error('Failed to generate bingo card:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    async fetchBingoCards() {
      try {
        console.log('Fetching bingo cards from API...')
        const response = await axios.get(`${API_BASE_URL}/api/bingo/cards`)
        this.bingoCards = response.data || []
        console.log('Fetched', this.bingoCards.length, 'bingo cards')

        // Only set current card if we don't have one
        if (this.bingoCards.length > 0 && !this.currentCard) {
          const latestCard = this.bingoCards[this.bingoCards.length - 1]
          this.currentCard = latestCard
          console.log('Set current card to:', latestCard.id)
        }
      } catch (error) {
        console.error('Failed to fetch bingo cards:', error)
        console.error('Error details:', {
          status: error.response?.status,
          message: error.response?.data?.error || error.message,
          url: error.config?.url
        })

        this.bingoCards = []

        // Show user-friendly error message
        if (error.response?.status === 401) {
          this.showSnackbar('Authentication failed. Please log in again.', 'error')
        } else if (error.response?.status >= 500) {
          this.showSnackbar('Server error. Please try again later.', 'error')
        } else {
          this.showSnackbar('Failed to load bingo cards. Check server connection.', 'error')
        }

        throw error // Re-throw so the caller knows it failed
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
        this.error = error.response?.data?.error || 'Failed to mark bingo item'
        console.error('Failed to mark bingo item:', error)
        throw error
      }
    },

    setCurrentCard(card) {
      this.currentCard = card
    },

    clearError() {
      this.error = null
    },

    // Theme methods
    async fetchThemes() {
      try {
        const response = await axios.get(`${API_BASE_URL}/api/themes`)
        this.themes = response.data.themes || []
        this.activeThemeId = response.data.active_theme_id
      } catch (error) {
        console.error('Failed to fetch themes:', error)
        this.themes = []
        this.activeThemeId = null
      }
    },

    async setActiveTheme(themeId) {
      try {
        const response = await axios.post(`${API_BASE_URL}/api/admin/themes/active`, {
          theme_id: themeId
        })
        this.activeThemeId = themeId
        this.showSnackbar('Theme changed successfully', 'success')

        // Refresh bingo cards since they might need to be regenerated for the new theme
        await this.fetchBingoCards()

        return response.data
      } catch (error) {
        this.error = error.response?.data?.error || 'Failed to set active theme'
        this.showSnackbar(this.error, 'error')
        console.error('Failed to set active theme:', error)
        throw error
      }
    },

    async createTheme(themeData) {
      try {
        const response = await axios.post(`${API_BASE_URL}/api/admin/themes`, themeData)
        this.themes.push(response.data)
        this.showSnackbar('Theme created successfully', 'success')
        return response.data
      } catch (error) {
        this.error = error.response?.data?.error || 'Failed to create theme'
        this.showSnackbar(this.error, 'error')
        console.error('Failed to create theme:', error)
        throw error
      }
    },

    async generateNewCard() {
      try {
        this.loading = true
        const response = await axios.get(`${API_BASE_URL}/api/bingo/new`)
        const newCard = response.data

        // Add the new card to the beginning of the array
        this.bingoCards.unshift(newCard)
        this.currentCard = newCard

        this.showSnackbar('New bingo card generated!', 'success')
        return newCard
      } catch (error) {
        const errorMsg = error.response?.data?.error || 'Failed to generate new card'
        this.error = errorMsg

        console.error('Failed to generate new card:', error)
        console.error('Error details:', {
          status: error.response?.status,
          message: errorMsg,
          url: error.config?.url
        })

        // Handle specific errors
        if (errorMsg.includes('Not enough bingo items') || errorMsg.includes('Server configuration error')) {
          this.showSnackbar('Server configuration error: Not enough bingo items available. Please contact administrator.', 'error')
        } else if (error.response?.status === 401) {
          this.showSnackbar('Authentication failed. Please log in again.', 'error')
        } else {
          this.showSnackbar(errorMsg, 'error')
        }

        throw error
      } finally {
        this.loading = false
      }
    },

    // Server connectivity methods
    async checkServerConnection() {
      if (this.isCheckingConnection) return

      this.isCheckingConnection = true
      try {
        console.log('Checking server connection...')
        const response = await axios.get(`${API_BASE_URL}/api/health`, {
          timeout: 3000 // 3 second timeout
        })

        const wasDisconnected = !this.isServerConnected
        this.isServerConnected = true

        if (wasDisconnected) {
          console.log('Server connection restored')
          this.showSnackbar('Connected to server', 'success')

          // If we have a token but no user data, try to fetch it
          if (this.token && !this.user) {
            try {
              await this.fetchUser()
            } catch (error) {
              console.warn('Failed to fetch user after reconnection:', error)
            }
          }
        }

        return true
      } catch (error) {
        const wasConnected = this.isServerConnected
        this.isServerConnected = false

        if (wasConnected) {
          console.log('Server connection lost')
          this.showSnackbar('Lost connection to server. Retrying...', 'warning')
        }

        return false
      } finally {
        this.isCheckingConnection = false
      }
    },

    startConnectionMonitoring() {
      if (this.connectionCheckInterval) {
        clearInterval(this.connectionCheckInterval)
      }

      console.log('Starting connection monitoring')

      // Check immediately
      this.checkServerConnection()

      // Then check every 5 seconds
      this.connectionCheckInterval = setInterval(() => {
        this.checkServerConnection()
      }, 5000)
    },

    stopConnectionMonitoring() {
      if (this.connectionCheckInterval) {
        console.log('Stopping connection monitoring')
        clearInterval(this.connectionCheckInterval)
        this.connectionCheckInterval = null
      }
    },

    // Theme management methods
    async createTheme(themeData) {
      try {
        const response = await this.apiCall('/api/admin/themes', 'POST', themeData)
        this.showSnackbar('Theme created successfully!', 'success')
        return response
      } catch (error) {
        this.showSnackbar('Failed to create theme', 'error')
        throw error
      }
    },

    async updateTheme(themeId, themeData) {
      try {
        const response = await this.apiCall(`/api/admin/themes/${themeId}`, 'PUT', themeData)
        this.showSnackbar('Theme updated successfully!', 'success')
        return response
      } catch (error) {
        this.showSnackbar('Failed to update theme', 'error')
        throw error
      }
    },

    async deleteTheme(themeId) {
      try {
        await this.apiCall(`/api/admin/themes/${themeId}`, 'DELETE')
        this.showSnackbar('Theme deleted successfully!', 'success')

        // Remove from local state
        this.themes = this.themes.filter(theme => theme.id !== themeId)

        // Clear active theme if it was deleted
        if (this.activeThemeId === themeId) {
          this.activeThemeId = ''
        }
      } catch (error) {
        // Extract the error message from the response
        const errorMessage = error.response?.data?.error || 'Failed to delete theme'
        this.showSnackbar(errorMessage, 'error')
        throw error
      }
    },

    async setActiveTheme(themeId) {
      try {
        await this.apiCall('/api/admin/themes/active', 'POST', { theme_id: themeId })
        this.activeThemeId = themeId

        // Find the theme name for the message
        const theme = this.themes.find(t => t.id === themeId)
        this.showSnackbar(`Active theme set to: ${theme?.name || 'Unknown'}`, 'success')
      } catch (error) {
        this.showSnackbar('Failed to set active theme', 'error')
        throw error
      }
    },

    async markThemeComplete(themeId) {
      console.log(`Marking theme ${themeId} as complete`)
      try {
        await this.apiCall(`/api/admin/themes/${themeId}/complete`, 'POST', { is_complete: true })
        this.showSnackbar(`Theme marked as complete!`, 'success')
      } catch (error) {
        this.showSnackbar('Failed to update theme status', 'error')
        throw error
      }
    }
  }
})
