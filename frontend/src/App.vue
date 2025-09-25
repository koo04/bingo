<template>
  <v-app>
    <router-view />
    
    <!-- Global Snackbar for error messages -->
    <v-snackbar
      v-model="appStore.snackbar.show"
      :color="appStore.snackbar.color"
      :timeout="6000"
      location="top right"
    >
      {{ appStore.snackbar.message }}
      <template v-slot:actions>
        <v-btn
          variant="text"
          @click="appStore.hideSnackbar"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </v-app>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

onMounted(async () => {
  console.log('App.vue mounted')
  // Load token from localStorage if it exists
  await appStore.loadTokenFromStorage()
  
  // Initialize WebSocket if authenticated
  if (appStore.isAuthenticated) {
    appStore.initializeWebSocket()
  }
})
</script>
