<template>
  <v-app>
    <router-view />
  </v-app>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

onMounted(async () => {
  // Initialize authentication
  await appStore.initializeAuth()
  
  // Initialize WebSocket if authenticated
  if (appStore.isAuthenticated) {
    appStore.initializeWebSocket()
  }
})
</script>
