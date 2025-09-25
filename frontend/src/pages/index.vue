<template>
  <v-container fluid class="fill-height">
    <div>
      <!-- Error Display -->
      <v-row v-if="store.error">
        <v-col cols="12">
          <v-alert
            type="error"
            closable
            @click:close="store.clearError"
          >
            {{ store.error }}
          </v-alert>
        </v-col>
      </v-row>

      <!-- No Current Card State -->
      <v-row v-if="!store.hasCurrentCard" justify="center">
      </v-row>

      <!-- Bingo Card Display -->
      <div v-else>
        <!-- Card Controls -->
        <v-row class="mb-4">
          <v-col cols="12" md="6">
            <v-card>
              <v-card-title class="d-flex align-center">
                <v-icon class="mr-2">mdi-cards</v-icon>
                Your Bingo Cards
              </v-card-title>
              <v-card-text>
                <v-select
                  v-model="selectedCardId"
                  :items="cardSelectItems"
                  item-title="text"
                  item-value="value"
                  @update:model-value="switchCard"
                  variant="outlined"
                  density="compact"
                >
                </v-select>
              </v-card-text>
            </v-card>
          </v-col>
          <v-col cols="12" md="6">
            <v-card>
              <v-card-title>Actions</v-card-title>
              <v-card-text>
                <div class="d-flex flex-wrap gap-2">
                  <v-btn
                    color="primary"
                    @click="generateCard"
                    :loading="store.loading"
                    prepend-icon="mdi-plus"
                  >
                    New Card
                  </v-btn>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <!-- Bingo Grid -->
        <v-row justify="center">
          <v-col cols="12" lg="8" xl="6">
            <v-card>
              <v-card-title v-if="store.currentCard?.is_winner" class="text-center text-h4 pa-6">
                🎉 BINGO 🎉
              </v-card-title>
              <v-card-title v-else class="text-center text-h4 pa-6">
                BINGO
              </v-card-title>
              <v-card-text class="pa-2">
                <div class="bingo-grid">
                  <div
                    v-for="(row, rowIndex) in store.currentCard.items"
                    :key="rowIndex"
                    class="bingo-row"
                  >
                    <div
                      v-for="(item, colIndex) in row"
                      :key="colIndex"
                      class="bingo-cell"
                      :class="{
                        'marked': store.currentCard.marked_items[rowIndex][colIndex],
                        'global-marked': store.isItemGloballyMarked(item),
                        'free-space': item === 'FREE SPACE'
                      }"
                      @click="markItem(rowIndex, colIndex)"
                    >
                      <div class="bingo-cell-content">
                        {{ item }}
                        <v-icon
                          v-if="store.currentCard.marked_items[rowIndex][colIndex]"
                          class="check-icon"
                          size="large"
                        >
                          mdi-check-circle
                        </v-icon>
                      </div>
                    </div>
                  </div>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <!-- Card Info -->
        <v-row class="mt-4">
          <v-col cols="12">
            <v-card>
              <v-card-title>Card Information</v-card-title>
              <v-card-text>
                <v-row>
                  <v-col cols="6" sm="3">
                    <div class="text-caption">Created</div>
                    <div>{{ formatDate(store.currentCard.created_at) }}</div>
                  </v-col>
                  <v-col cols="6" sm="3">
                    <div class="text-caption">Status</div>
                    <v-chip
                      :color="store.currentCard.is_winner ? 'success' : 'default'"
                      size="small"
                    >
                      {{ store.currentCard.is_winner ? 'Winner!' : 'In Progress' }}
                    </v-chip>
                  </v-col>
                  <v-col cols="6" sm="3">
                    <div class="text-caption">Total Cards</div>
                    <div>{{ store.bingoCards.length }}</div>
                  </v-col>
                  <v-col cols="6" sm="3">
                    <div class="text-caption">Wins</div>
                    <div>{{ winCount }}</div>
                  </v-col>
                </v-row>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </div>
    </div>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const selectedCardId = ref(null)

const cardSelectItems = computed(() => {
  return store.bingoCards.map((card, index) => ({
    text: `Card ${index + 1} - ${formatDate(card.created_at)} ${card.is_winner ? '🏆' : ''}`,
    value: card.id
  }))
})

const winCount = computed(() => {
  return store.bingoCards.filter(card => card.is_winner).length
})

onMounted(async () => {
  // Fetch bingo cards (user is already authenticated via router guard)
  await store.fetchBingoCards()
  if (store.currentCard) {
    selectedCardId.value = store.currentCard.id
  }
})

watch(() => store.currentCard, (newCard) => {
  if (newCard) {
    selectedCardId.value = newCard.id
  }
}, { immediate: true })

async function generateCard() {
  try {
    await store.generateNewBingoCard()
    selectedCardId.value = store.currentCard.id
  } catch (error) {
    // Error is handled in store
  }
}

async function markItem(row, col) {
  if (store.currentCard.items[row][col] === 'FREE SPACE') return
  
  try {
    await store.markBingoItem(row, col)
  } catch (error) {
    // Error is handled in store
  }
}

function switchCard(cardId) {
  const card = store.bingoCards.find(c => c.id === cardId)
  if (card) {
    store.setCurrentCard(card)
  }
}

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.bingo-grid {
  display: grid;
  grid-template-rows: repeat(5, 1fr);
  gap: 2px;
  background: #ccc;
  border-radius: 8px;
  overflow: hidden;
}

.bingo-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 2px;
}

.bingo-cell {
  position: relative;
  aspect-ratio: 1;
  background: white;
  cursor: pointer;
  transition: all 0.2s ease;
  min-height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bingo-cell:hover {
  background: #f5f5f5;
  transform: scale(1.02);
}

.bingo-cell.marked {
  background: #4caf50;
  color: white;
}

.bingo-cell.global-marked {
  background: #2196f3;
  color: white;
  border: 2px solid #1976d2;
}

.bingo-cell.marked.global-marked {
  background: #4caf50;
  border: 2px solid #1976d2;
}

.bingo-cell.free-space {
  background: #2196f3;
  color: white;
  cursor: default;
}

.bingo-cell.free-space:hover {
  transform: none;
}

.bingo-cell-content {
  position: relative;
  text-align: center;
  padding: 8px;
  font-size: 0.85rem;
  line-height: 1.2;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  word-wrap: break-word;
  hyphens: auto;
}

.check-icon {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  opacity: 0.9;
}

@media (max-width: 768px) {
  .bingo-cell {
    min-height: 60px;
  }
  
  .bingo-cell-content {
    font-size: 0.75rem;
    padding: 4px;
  }
}

@media (max-width: 480px) {
  .bingo-cell {
    min-height: 50px;
  }
  
  .bingo-cell-content {
    font-size: 0.7rem;
    padding: 2px;
  }
}
</style>
