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
          v-if="appStore.connectionStatus === 'checking'"
          indeterminate
          size="64"
          color="primary"
          class="mb-4"
        ></v-progress-circular>
        
        <v-icon
          v-else
          size="64"
          color="warning"
          class="mb-4"
        >
          mdi-server-network-off
        </v-icon>
        
        <div class="text-h6 mb-2">
          {{ connectionMessage }}
        </div>
        
        <div class="text-body-2 text-medium-emphasis">
          {{ connectionSubMessage }}
        </div>
        
        <v-btn
          v-if="appStore.connectionStatus === 'disconnected'"
          color="primary"
          class="mt-4"
          @click="appStore.checkServerConnection"
          :loading="appStore.isCheckingConnection"
        >
          Retry Connection
        </v-btn>
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
import { onMounted, computed, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

const showConnectionOverlay = computed(() => {
  return !appStore.isServerConnected || appStore.isCheckingConnection
})

const connectionMessage = computed(() => {
  switch (appStore.connectionStatus) {
    case 'checking':
      return 'Connecting to Server...'
    case 'disconnected':
      return 'Server Not Available'
    default:
      return 'Connected'
  }
})

const connectionSubMessage = computed(() => {
  switch (appStore.connectionStatus) {
    case 'checking':
      return 'Please wait while we establish connection'
    case 'disconnected':
      return 'Make sure the backend server is running and try again'
    default:
      return ''
  }
})

onMounted(async () => {
  console.log('App.vue mounted')
  
  // Start monitoring server connection immediately
  appStore.startConnectionMonitoring()
  
  // Load token from localStorage if it exists
  await appStore.loadTokenFromStorage()
  
  // Initialize WebSocket if authenticated
  if (appStore.isAuthenticated) {
    appStore.initializeWebSocket()
  }
})

onUnmounted(() => {
  // Clean up connection monitoring when app unmounts
  appStore.stopConnectionMonitoring()
})
</script>
