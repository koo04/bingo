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
              @click="loginWithDiscord"
              prepend-icon="mdi-discord"
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

onMounted(() => {
  console.log('🔐 LOGIN PAGE: Login page mounted')
  console.log('🔐 LOGIN PAGE: Current URL:', window.location.href)
  console.log('🔐 LOGIN PAGE: Authenticated:', store.isAuthenticated)
  
  // Token processing is now handled by the router guard
  // This page should only show if user is not authenticated and there's no token
})

function loginWithDiscord() {
  window.location.href = `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/auth/discord`
}
</script>

<route lang="yaml">
meta:
  layout: login
</route>
