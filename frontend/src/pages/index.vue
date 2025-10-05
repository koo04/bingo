<template>
  <v-container fluid class="fill-height d-flex align-center justify-center">
    <div v-if="currentCard" class="d-flex flex-column align-center" style="max-width: 600px; width: 100%;">
      <!-- Theme Info -->
      <ThemeInfo
        :active-theme="store.themes.find(t => t.id === store.activeThemeId) || { name: 'Unknown Theme', description: '' }"
        class="mb-4 w-100" />

      <!-- Bingo Grid -->
      
      <BingoGrid
        :current-card="currentCard" :items="items"/>
      
      <v-row class="w-100" justify="center">
        <v-col cols="2"
            v-for="card in otherUserCards"
            :key="card.id">
          <BingoGridMini :card="card" :user="getUser(card.user_id)" :items="items"/>
        </v-col>
      </v-row>
      <!-- Card Info -->
      <!-- <BingoCardInfo
        :current-card="currentCard"
        class="w-100" /> -->
    </div>

    <!-- No card message -->
    <v-card v-else class="text-center pa-8" style="max-width: 600px; width: 100%;">
      <v-card-text>
        Generating card...
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useAppStore } from '@/stores/app'
import BingoGrid from '@/components/BingoGrid.vue'
import BingoCardInfo from '@/components/BingoCardInfo.vue'
import ThemeInfo from '@/components/ThemeInfo.vue'
import websocketService from '@/services/websocket'
import BingoGridMini from '@/components/BingoGridMini.vue'

const store = useAppStore()
const { currentCard, activeTheme, getUser } = storeToRefs(store)
const isInitialized = ref(false)

// Filter out current user's card from the list
const otherUserCards = computed(() => {
  if (!activeTheme.value?.cards || !store.user?.id) {
    return []
  }
  
  // Convert cards object to array and filter out current user's card
  return Object.entries(activeTheme.value.cards)
    .filter(([userId, card]) => userId !== store.user.id.toString())
    .map(([userId, card]) => card)
})

const items = computed(() => {
  return activeTheme.value?.items || []
})

const handleWebSocketMessage = (data) => {
  switch (data.type) {
    case 'item_updated':
      if (data.data) {
        store.updateItem(data.data)
      }
      break
    case 'winners':
      if (data.data) {
        const cards = data.data.cards
        for (const card of cards) {
          store.updateCard(card)
        }
      }
      break
    default:
      console.warn('Unknown WebSocket message type:', data.type)
      break
  }
}

onMounted(async () => {
  if (isInitialized.value) {
    console.log('Index page already initialized, skipping')
    return
  }

  console.log('Index page onMounted - Waiting for server connection...')

  // Wait for server connection before initializing
  const waitForConnection = () => {
    return new Promise((resolve) => {
      const checkConnection = () => {
        if (store.isServerConnected) {
          resolve()
        } else {
          setTimeout(checkConnection, 100) // Check every 100ms
        }
      }
      checkConnection()
    })
  }

  await waitForConnection()
  console.log('Server connected, starting initialization')
  isInitialized.value = true

  try {
    // Fetch themes first to know what's active
    console.log('Fetching themes...')
    try {
      await store.fetchThemes()
      console.log('Active theme ID:', store.activeThemeId)
    } catch (error) {
      console.error('Failed to fetch themes:', error)
      // Continue anyway, we might have cached data
    }

    try {
      await store.fetchUsers()
      console.log('Fetched users:', store.users)
    } catch (error) {
      console.error('Failed to fetch users:', error)
      // Continue anyway, we might have cached data
    }

    // Clear any stale errors if we have no themes or no active theme
    if (!store.themes || store.themes.length === 0 || !store.activeThemeId) {
      store.error = null
    }

    console.log('Index page initialization complete')
  } catch (error) {
    console.error('Error during index page initialization:', error)
    // Only reset initialization flag for network errors, not theme configuration errors
    const errorMessage = error.response?.data?.error || ''
    const isThemeConfigError = errorMessage.includes('No themes available') ||
      errorMessage.includes('No active theme selected') ||
      (errorMessage.includes('only has') && errorMessage.includes('items (need at least 25)'))

    if (!isThemeConfigError) {
      // Reset initialization flag so it can be retried for network errors
      isInitialized.value = false
    }
  }

  // Setup WebSocket
  websocketService.connect()
  websocketService.on('item_updated', handleWebSocketMessage)
  websocketService.on('winners', handleWebSocketMessage)
})

onUnmounted(() => {
  console.log('Index page unmounted, disconnecting WebSocket')
  websocketService.disconnect()
  websocketService.off('item_updated', handleWebSocketMessage)
  websocketService.off('winners', handleWebSocketMessage)
})
</script>

<style scoped>
/* Page-specific styles can be added here if needed */
</style>
