<template>
  <v-container fluid class="fill-height d-flex align-center justify-center">
    <div v-if="store.currentCard" class="d-flex flex-column align-center" style="max-width: 600px; width: 100%;">
      <!-- Theme Info -->
      <ThemeInfo
        :active-theme="store.themes.find(t => t.id === store.activeThemeId) || { name: 'Unknown Theme', description: '' }"
        class="mb-4 w-100" />

      <!-- Bingo Grid -->
      <BingoGrid 
        :current-card="store.currentCard" 
        :is-item-globally-marked="store.isItemGloballyMarked"
        class="mb-4" />

      <!-- Card Info -->
      <BingoCardInfo 
        :current-card="store.currentCard" 
        :total-cards="store.bingoCards.length" 
        :win-count="winCount"
        class="w-100" />
    </div>

    <!-- No card message -->
    <v-card v-else class="text-center pa-8" style="max-width: 600px; width: 100%;">
      <v-card-text>
        <v-icon size="64" :color="store.error ? 'error' : 'grey'" class="mb-4">
          {{ store.error ? 'mdi-alert-circle' : 'mdi-cards-outline' }}
        </v-icon>

        <div class="text-h6 mb-2">
          {{ getCardStatusTitle() }}
        </div>

        <div class="text-body-1 mb-4">
          {{ getCardStatusMessage() }}
        </div>

        <v-btn v-if="!isThemeConfigError" color="primary" size="large" @click="generateNewCard"
          :loading="store.loading">
          {{ store.error ? 'Retry' : 'Generate Bingo Card' }}
        </v-btn>

        <div v-if="isThemeConfigError" class="text-caption text-medium-emphasis mt-2">
          Please contact an administrator to resolve this issue
        </div>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useAppStore } from '@/stores/app'
import BingoGrid from '@/components/BingoGrid.vue'
import BingoCardInfo from '@/components/BingoCardInfo.vue'
import ThemeInfo from '@/components/ThemeInfo.vue'

const store = useAppStore()
const isInitialized = ref(false)

const winCount = computed(() => {
  return store.bingoCards.filter(card => card.is_winner).length
})

const isThemeConfigError = computed(() => {
  // Check if it's a theme configuration error OR if there are simply no themes available
  const hasNoThemes = !store.themes || store.themes.length === 0
  const hasNoActiveTheme = !store.activeThemeId

  return hasNoThemes || hasNoActiveTheme || (store.error && (
    store.error.includes('No themes available') ||
    store.error.includes('No active theme selected') ||
    store.error.includes('only has') && store.error.includes('items (need at least 25)')
  ))
})

function getCardStatusTitle() {
  if (!store.themes || store.themes.length === 0) {
    return 'No Themes Available'
  }
  if (!store.activeThemeId) {
    return 'No Active Theme'
  }
  if (store.error) {
    return 'Unable to Load Bingo Card'
  }
  return 'No Bingo Card Yet'
}

function getCardStatusMessage() {
  if (!store.themes || store.themes.length === 0) {
    return 'An administrator needs to create themes before you can play bingo.'
  }
  if (!store.activeThemeId) {
    return 'An administrator needs to select an active theme before you can generate a bingo card.'
  }
  if (store.error) {
    return store.error
  }
  return 'Generate your first bingo card to start playing!'
}

async function generateNewCard() {
  try {
    await store.generateNewCard()
  } catch (error) {
    console.error('Failed to generate new card:', error)
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

    // Fetch bingo cards (user is already authenticated via router guard)
    console.log('Fetching bingo cards...')
    try {
      await store.fetchBingoCards()
      console.log('Current card:', store.currentCard?.id, 'Theme ID:', store.currentCard?.theme_id)
    } catch (error) {
      console.error('Failed to fetch bingo cards, will try to generate new one')
      // If fetching fails, we'll try to generate a new card below
    }

    // Auto-generate a card if user doesn't have one for the current theme
    // Note: Cards created before theme support won't have theme_id, so we should be lenient
    const hasThemes = store.themes && store.themes.length > 0
    const hasActiveTheme = !!store.activeThemeId
    const cardHasThemeId = !!store.currentCard?.theme_id
    const themesMismatch = hasActiveTheme && cardHasThemeId && (store.currentCard.theme_id !== store.activeThemeId)

    // Only try to generate a card if we have themes available
    const needsNewCard = hasThemes && hasActiveTheme && (!store.currentCard || themesMismatch)

    console.log('Checking if new card needed:')
    console.log('- Has themes:', hasThemes)
    console.log('- Has current card:', !!store.currentCard)
    console.log('- Active theme ID:', store.activeThemeId)
    console.log('- Current card theme ID:', store.currentCard?.theme_id)
    console.log('- Has active theme:', hasActiveTheme)
    console.log('- Card has theme ID:', cardHasThemeId)
    console.log('- Themes mismatch:', themesMismatch)
    console.log('- Needs new card:', needsNewCard)

    if (needsNewCard) {
      console.log('Generating new card...')
      try {
        await generateNewCard()
      } catch (error) {
        console.error('Failed to generate new card:', error)
        // Don't retry automatically for theme configuration errors
        const errorMessage = error.response?.data?.error || ''
        if (errorMessage.includes('No themes available') ||
          errorMessage.includes('No active theme selected') ||
          errorMessage.includes('only has') && errorMessage.includes('items (need at least 25)')) {
          console.error('Theme configuration issue - stopping auto-retry')
          // Don't reset initialization flag - we're done trying
          return
        }
      }
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
})
</script>

<style scoped>
/* Page-specific styles can be added here if needed */
</style>
