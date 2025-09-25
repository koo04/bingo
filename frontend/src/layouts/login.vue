<template>
  <v-main>
    <router-view />
  </v-main>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

onMounted(() => {
  // Initialize WebSocket when the app loads
  if (appStore.isAuthenticated) {
    appStore.initializeWebSocket()
  }
})
</script>
