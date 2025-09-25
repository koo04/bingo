<template>
  <v-row justify="center" align="center" class="fill-height">
    <v-col cols="12" sm="8" md="6" lg="4">
      <v-card class="pa-6 text-center">
        <v-card-title class="text-h4 mb-4">
          🎯 Signing you in...
        </v-card-title>
        <v-card-text>
          <div v-if="loading">
            <v-progress-circular
              indeterminate
              color="primary"
              size="64"
              class="mb-4"
            ></v-progress-circular>
            <p class="text-body-1">{{ statusMessage }}</p>
          </div>
          
          <div v-if="error" class="text-center">
            <v-icon color="error" size="64" class="mb-4">mdi-alert-circle</v-icon>
            <p class="text-body-1 text-error mb-4">{{ error }}</p>
            <v-btn color="primary" @click="goToLogin">
              Try Again
            </v-btn>
          </div>
        </v-card-text>
      </v-card>
    </v-col>
  </v-row>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const store = useAppStore()

const loading = ref(true)
const error = ref('')
const statusMessage = ref('Connecting to Discord...')

onMounted(async () => {
  console.log('🔐 AUTH CALLBACK: Page mounted')
  console.log('🔐 AUTH CALLBACK: Current URL:', window.location.href)
  
  try {
    // Get the authorization code from URL parameters
    const urlParams = new URLSearchParams(window.location.search)
    const code = urlParams.get('code')
    const errorParam = urlParams.get('error')
    
    if (errorParam) {
      throw new Error(`Discord OAuth error: ${errorParam}`)
    }
    
    if (!code) {
      throw new Error('No authorization code received from Discord')
    }
    
    console.log('🔐 AUTH CALLBACK: Authorization code received')
    statusMessage.value = 'Processing authorization...'
    
    // Send the code to our backend
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/auth/discord/exchange`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ code }),
    })
    
    if (!response.ok) {
      const errorData = await response.json()
      throw new Error(errorData.error || 'Failed to exchange authorization code')
    }
    
    const data = await response.json()
    console.log('🔐 AUTH CALLBACK: Token received successfully')
    
    statusMessage.value = 'Completing sign in...'
    
    // Store the token and user data
    store.setToken(data.token)
    store.user = data.user
    
    console.log('🔐 AUTH CALLBACK: Authentication completed, redirecting to home')
    
    // Clean up URL and redirect to home
    window.history.replaceState({}, document.title, window.location.pathname)
    router.push('/')
    
  } catch (err) {
    console.error('🔐 AUTH CALLBACK: Authentication failed:', err)
    loading.value = false
    error.value = err.message || 'Authentication failed. Please try again.'
  }
})

function goToLogin() {
  router.push('/login')
}
</script>

<route lang="yaml">
meta:
  layout: login
</route>