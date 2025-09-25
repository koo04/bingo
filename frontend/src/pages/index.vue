<template>
  <v-container fluid class="fill-height">
    <div v-if="store.currentCard">
      <!-- Card Controls -->
      <BingoCardSelector
        :selected-card-id="selectedCardId"
        :card-select-items="cardSelectItems"
        @card-change="switchCard"
      />

      <!-- Bingo Grid -->
      <BingoGrid
        :current-card="store.currentCard"
        :is-item-globally-marked="store.isItemGloballyMarked"
      />

      <!-- Card Info -->
      <BingoCardInfo
        :current-card="store.currentCard"
        :total-cards="store.bingoCards.length"
        :win-count="winCount"
      />
    </div>
  </v-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import BingoGrid from '@/components/BingoGrid.vue'
import BingoCardSelector from '@/components/BingoCardSelector.vue'
import BingoCardInfo from '@/components/BingoCardInfo.vue'

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
/* Page-specific styles can be added here if needed */
</style>
