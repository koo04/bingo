<template>
  <v-main>
    <router-view />
  </v-main>

  <AppFooter />
  
  <v-snackbar
    v-model="appStore.snackbar.show"
    :color="appStore.snackbar.color"
    timeout="4000"
    @click:outside="appStore.hideSnackbar"
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
