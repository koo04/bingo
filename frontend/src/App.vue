<template>
  <v-app>
    <!-- Connection Status Overlay -->
    <v-overlay
      v-model="showConnectionOverlay"
      class="align-center justify-center"
      persistent
    >
      <v-card class="pa-8 text-center" min-width="300">
        <v-progress-circular
          v-if="!appStore.isServerConnected"
          indeterminate
          size="64"
          color="primary"
          class="mb-4"
        ></v-progress-circular>
        
        <div class="text-h6 mb-2">
          Connecting to Server...
        </div>
      </v-card>
    </v-overlay>

    <router-view v-if="appStore.isServerConnected" />
    
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
import { onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const showConnectionOverlay = computed(() => {
  return !appStore.isServerConnected
})

onMounted(async () => {
  console.log('App.vue mounted')

  // Start monitoring server connection immediately
  appStore.startConnectionMonitoring()
  
  // Load token from localStorage if it exists
  await appStore.loadTokenFromStorage()
})
</script>
