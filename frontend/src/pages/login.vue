<template>
    <!-- Login State -->
    <v-row justify="center" align="center" class="fill-height">
      <v-col cols="12" sm="8" md="6" lg="4">
        <v-card class="pa-6 text-center">
          <v-card-title class="text-h4 mb-4">
            🎯 Bingo Game
          </v-card-title>
          <v-card-text>
            <p class="text-body-1 mb-6">
              Welcome to the Bingo Game! Sign in with Discord to create and play bingo cards.
            </p>
            <v-btn
              color="primary"
              size="large"
              block
              @click="store.loginWithDiscord"
              prepend-icon="mdi-discord"
              :loading="store.loading"
            >
              Sign in with Discord
            </v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
</template>
<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const router = useRouter()

onMounted(async () => {
  // Check if there's a token in the URL (OAuth callback)
  const urlParams = new URLSearchParams(window.location.search)
  const urlToken = urlParams.get('token')
  
  if (urlToken) {
    console.log('🔐 LOGIN: OAuth callback detected on login page')
    console.log('🔐 LOGIN: Current URL:', window.location.href)
    
    // Wait for auth initialization to complete
    let retries = 0
    while (retries < 50 && !store.authInitialized) {
      await new Promise(resolve => setTimeout(resolve, 100))
      retries++
    }
    
    console.log('🔐 LOGIN: Auth initialization complete, authenticated:', store.isAuthenticated)
    
    // If authentication was successful, redirect to home immediately
    if (store.isAuthenticated) {
      console.log('🔐 LOGIN: Authentication successful, redirecting to home and cleaning URL')
      // Navigate directly to home, which will clean up the current URL
      router.replace('/')
    } else {
      console.log('🔐 LOGIN: Authentication failed, cleaning up URL')
      // Only clean up URL if staying on login page
      const url = new URL(window.location)
      url.searchParams.delete('token')
      const cleanUrl = url.pathname + url.search + url.hash
      window.history.replaceState({}, document.title, cleanUrl)
      console.log('🔐 LOGIN: URL cleaned up to:', cleanUrl)
    }
  }
})
</script>

<route lang="yaml">
meta:
  layout: login
</route>
