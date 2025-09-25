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
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    hasCurrentCard: (state) => !!state.currentCard,
    isItemGloballyMarked: (state) => (item) => state.globalMarkedItems.includes(item),
    isAdmin: (state) => state.user?.is_admin || false,
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
        this.setToken(token)
        // Try to fetch user data
        try {
          await this.fetchUser()
          console.log('User data loaded from existing token')
        } catch (error) {
          console.warn('Failed to fetch user with stored token, token may be expired')
          // fetchUser already handles logout on error
        }
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
      this.token = null
      this.user = null
      this.setToken(null)
      // Redirect to login
      if (typeof window !== 'undefined') {
        window.location.href = '/login'
      }
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
        const response = await axios.get(`${API_BASE_URL}/api/bingo/cards`)
        this.bingoCards = response.data || []
        if (this.bingoCards.length > 0 && !this.currentCard) {
          this.currentCard = this.bingoCards[this.bingoCards.length - 1]
        }
      } catch (error) {
        console.error('Failed to fetch bingo cards:', error)
        this.bingoCards = []
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
    }
  }
})
