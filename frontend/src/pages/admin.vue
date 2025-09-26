<template>
  <v-container>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-4">Admin Panel</h1>
        
        <v-alert v-if="!isAdmin" type="error" class="mb-4">
          Access denied. You need admin privileges to view this page.
        </v-alert>
        
        <div v-if="isAdmin">
          <!-- Theme Management -->
          <ThemeManager />

          <!-- Bingo Items Control -->
          <v-card class="mb-6">
            <v-card-title>
              <v-icon class="mr-2">mdi-checkbox-marked</v-icon>
              Bingo Items Control
            </v-card-title>
            <v-card-text>
              <v-row>
                <v-col 
                  v-for="item in allItems" 
                  :key="item" 
                  cols="12" 
                  md="6" 
                  lg="4"
                >
                  <v-card 
                    :class="{ 'bg-success': isItemMarked(item) }"
                    variant="outlined"
                  >
                    <v-card-text class="d-flex justify-space-between align-center">
                      <span :class="{ 'text-decoration-line-through': isItemMarked(item) }">
                        {{ item }}
                      </span>
                      <v-btn
                        :color="isItemMarked(item) ? 'error' : 'success'"
                        :icon="isItemMarked(item) ? 'mdi-close' : 'mdi-check'"
                        size="small"
                        @click="toggleItem(item)"
                        :loading="loadingItems.has(item)"
                      />
                    </v-card-text>
                  </v-card>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>

          <!-- Player Cards Overview -->
          <v-card>
            <v-card-title>
              <v-icon class="mr-2">mdi-account-group</v-icon>
              Player Cards Overview
            </v-card-title>
            <v-card-text>
              <v-row>
                <v-col 
                  v-for="card in playerCards" 
                  :key="card.id" 
                  cols="12" 
                  md="6" 
                  lg="4"
                >
                  <v-card :class="{ 'border-success': card.is_winner }" variant="outlined">
                    <v-card-title class="d-flex justify-space-between">
                      <span>{{ getUserName(card.user_id) }}</span>
                      <v-chip
                        v-if="card.is_winner"
                        color="success"
                        size="small"
                      >
                        WINNER
                      </v-chip>
                    </v-card-title>
                    <v-card-text>
                      <div class="bingo-grid">
                        <div 
                          v-for="(row, i) in card.items" 
                          :key="i"
                          class="bingo-row"
                        >
                          <div 
                            v-for="(item, j) in row" 
                            :key="j"
                            class="bingo-cell"
                            :class="{ 
                              'marked': card.marked_items[i][j],
                              'global-marked': isItemMarked(item)
                            }"
                          >
                            {{ item }}
                          </div>
                        </div>
                      </div>
                    </v-card-text>
                  </v-card>
                </v-col>
              </v-row>
            </v-card-text>
          </v-card>
        </div>
      </v-col>
    </v-row>


  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import websocketService from '@/services/websocket'
import ThemeManager from '@/components/ThemeManager.vue'

const store = useAppStore()
const isAdmin = ref(false)
const allItems = ref([])

const themes = ref([])
const markedItems = ref([])
const playerCards = ref([])
const users = ref([])
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
    const themes = await store.fetchThemes()
    themes.value = themes

    const itemsResponse = await store.apiCall('/api/admin/items')
    markedItems.value = itemsResponse.marked_items
    allItems.value = itemsResponse.all_items

    const cardsResponse = await store.apiCall('/api/admin/cards')
    playerCards.value = cardsResponse.cards
    users.value = cardsResponse.users
  } catch (error) {
    console.error('Error loading admin data:', error)
  }
}

const isItemMarked = (item) => {
  return markedItems.value.includes(item)
}

const getUserName = (userId) => {
  const user = users.value.find(u => u.id === userId)
  return user ? user.username : 'Unknown User'
}

const toggleItem = async (item) => {
  loadingItems.value.add(item)
  
  try {
    const endpoint = isItemMarked(item) ? '/api/admin/items/unmark' : '/api/admin/items/mark'
    await store.apiCall(endpoint, 'POST', { item })
    
    // Update local state
    if (isItemMarked(item)) {
      markedItems.value = markedItems.value.filter(i => i !== item)
    } else {
      markedItems.value.push(item)
    }
  } catch (error) {
    console.error('Error toggling item:', error)
    store.showSnackbar('Error updating item', 'error')
  } finally {
    loadingItems.value.delete(item)
  }
}

const handleWebSocketMessage = (data) => {
  switch (data.type) {
    case 'initial_state':
      markedItems.value = data.marked_items
      break
    case 'item_marked':
      if (!markedItems.value.includes(data.item)) {
        markedItems.value.push(data.item)
      }
      break
    case 'item_unmarked':
      markedItems.value = markedItems.value.filter(item => item !== data.item)
      break
  }
}

onMounted(async () => {
  await checkAdminAccess()
  if (isAdmin.value) {
    await loadAdminData()

    // Setup WebSocket
    websocketService.connect()
    websocketService.on('initial_state', handleWebSocketMessage)
    websocketService.on('item_marked', handleWebSocketMessage)
    websocketService.on('item_unmarked', handleWebSocketMessage)
  }
})

onUnmounted(() => {
  websocketService.off('initial_state', handleWebSocketMessage)
  websocketService.off('item_marked', handleWebSocketMessage)
  websocketService.off('item_unmarked', handleWebSocketMessage)
})
</script>

<style scoped>
.bingo-grid {
  display: grid;
  gap: 2px;
  max-width: 300px;
}

.bingo-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 2px;
}

.bingo-cell {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #f5f5f5;
  border-radius: 4px;
  font-size: 0.7rem;
  text-align: center;
  padding: 2px;
  overflow: hidden;
}

.bingo-cell.marked {
  background-color: #4caf50;
  color: white;
}

.bingo-cell.global-marked {
  background-color: #2196f3;
  color: white;
  font-weight: bold;
}

.border-success {
  border: 2px solid #4caf50 !important;
}
</style>
