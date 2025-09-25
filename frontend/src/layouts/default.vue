<template>
  <v-main>
    <v-app-bar flat>
      <v-container class="mx-auto d-flex align-center justify-center">
        <v-avatar
          class="me-4 "
          color="grey-darken-1"
          size="42"
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
        >
          Logout
        </v-btn>
      </v-container>
    </v-app-bar>

    <!-- Loading State -->
    <v-row v-if="store.loading" justify="center" align="center" class="fill-height">
      <v-col cols="auto">
        <v-progress-circular indeterminate size="64" color="primary"></v-progress-circular>
        <div class="text-center mt-4">Loading...</div>
      </v-col>
    </v-row>

    <router-view />
  </v-main>
</template>

<script setup>
import { onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const links = [
  { text: 'Home', href: '/' },
  { text: 'Admin', href: '/admin', adminOnly: true }
]

const visibleLinks = computed(() => {
  return links.filter(link => !link.adminOnly || (link.adminOnly && store.isAuthenticated))
})

onMounted(() => {
  // Initialize WebSocket when the app loads
  if (store.isAuthenticated) {
    store.initializeWebSocket()
  }
})
</script>
