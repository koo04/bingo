<template>
  <v-app id="inspire">
    <v-app-bar flat>
      <v-container class="mx-auto d-flex align-center justify-center">
        <v-avatar
          class="me-4"
          color="grey-darken-1"
          size="32"
        >
          <v-img
            v-if="store.user?.avatar"
            :src="`https://cdn.discordapp.com/avatars/${store.user.discord_id}/${store.user.avatar}.png`"
          ></v-img>
        </v-avatar>

        <v-btn
          v-for="link in visibleLinks"
          :key="link.text"
          :text="link.text"
          :to="link.href"
          variant="text"
        ></v-btn>

        <v-spacer></v-spacer>

        <v-btn
          color="error"
          variant="outlined"
          @click="store.logout"
          append-icon="mdi-logout"
          class="me-4"
          size="small"
        >
          Logout
        </v-btn>
      </v-container>
    </v-app-bar>

    <v-main>
      <!-- Loading State -->
      <v-container v-if="store.loading" class="fill-height">
        <v-row justify="center" align="center" class="fill-height">
          <v-col cols="auto">
            <v-progress-circular indeterminate size="64" color="primary"></v-progress-circular>
            <div class="text-center mt-4">Loading...</div>
          </v-col>
        </v-row>
      </v-container>

      <!-- Main Content Layout -->
      <v-container v-else>
        <v-row>
          <v-col>
            <v-sheet>
              <router-view />
            </v-sheet>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script setup>
import { onMounted, computed, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const links = [
  { text: 'Home', href: '/' },
  { text: 'Admin', href: '/admin', adminOnly: true }
]

const visibleLinks = computed(() => {
  return links.filter(link => !link.adminOnly || (link.adminOnly && store.isAuthenticated))
})

onMounted(async () => {
  // Initialize WebSocket if authenticated
  if (store.isAuthenticated) {
    store.initializeWebSocket()
  }
})

onUnmounted(() => {
  // Clean up connection monitoring when app unmounts
  store.stopConnectionMonitoring()
})
</script>
