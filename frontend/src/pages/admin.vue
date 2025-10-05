<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4">Admin Panel</h1>

        <!-- Theme Management -->
        <ThemeManager :themes="themes" />

        <!-- Bingo Items Control -->
        <v-card class="mb-6">
          <v-card-text>
            <v-row>
              <v-col cols="2" v-for="item in items" :key="item.id" >
                <v-checkbox :label="item.name" v-model="item.marked"
                  @change="toggleItem(item.id)" :disabled="loadingItems.has(item.id)"></v-checkbox>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>
      </v-col>

      <!-- Player Cards Overview -->
      <v-col cols="12" xl="1" lg="1" md="2" v-for="card in playerCards" :key="card.id">
        <BingoGridMini :card="card" :user="getUser(card.user_id)" :items="items" />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import ThemeManager from '@/components/ThemeManager.vue'

const store = useAppStore()
const { activeTheme, getUser, themes } = storeToRefs(store)

const isAdmin = ref(false)
const items = computed(() => {
  return activeTheme.value?.items || []
})

const playerCards = computed(() => {
  return activeTheme.value ? Object.values(activeTheme.value.cards) : []
})
const loadingItems = ref(new Set())

const checkAdminAccess = async () => {
  try {
    const response = await store.apiCall('/api/admin/check')
    isAdmin.value = response.is_admin
  } catch (error) {
    console.error('Error checking admin access:', error)
    isAdmin.value = false
  }
}

const loadAdminData = async () => {
  if (!isAdmin.value) return

  try {
    const themesResponse = await store.fetchThemes()
    themes.value = themesResponse
  } catch (error) {
    console.error('Error loading admin data:', error)
  }

  try {
    await store.fetchUsers()
  } catch (error) {
    console.error('Failed to fetch users:', error)
    // Continue anyway, we might have cached data
  }
}

const toggleItem = async (itemId) => {
  loadingItems.value.add(itemId)

  try {
    const endpoint = '/api/admin/themes/' + store.activeThemeId + '/items/' + itemId + '/toggle'
    await store.apiCall(endpoint, 'POST')
  } catch (error) {
    console.error('Error toggling item:', error)
    store.showSnackbar('Error updating item', 'error')
  } finally {
    loadingItems.value.delete(itemId)
  }
}

onMounted(async () => {
  await checkAdminAccess()
  if (isAdmin.value) {
    await loadAdminData()
  } else {
    console.warn('User does not have admin access')
    store.showSnackbar('You do not have admin access', 'error')
  }
})

</script>

<style scoped>
</style>
