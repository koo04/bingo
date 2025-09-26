<template>
  <v-row justify="center">
    <v-col cols="12" lg="12" xl="6">
      <v-card>
        <v-card-text class="pa-2">
          <div class="bingo-grid">
            <div
              v-for="(row, rowIndex) in currentCard.items"
              :key="rowIndex"
              class="bingo-row"
            >
              <div
                v-for="(item, colIndex) in row"
                :key="colIndex"
                class="bingo-cell"
                :class="{
                  'marked': currentCard.marked_items[rowIndex][colIndex],
                  'global-marked': isItemGloballyMarked(item),
                  'free-space': item === 'FREE SPACE',
                  'interactive': interactive && item !== 'FREE SPACE'
                }"
                @click="handleCellClick(rowIndex, colIndex, item)"
              >
                <div class="bingo-cell-content">
                  {{ item }}
                  <v-icon
                    v-if="currentCard.marked_items[rowIndex][colIndex]"
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
</template>

<script setup>
const props = defineProps({
  currentCard: {
    type: Object,
    required: true
  },
  isItemGloballyMarked: {
    type: Function,
    required: true
  },
  interactive: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['cell-click'])

function handleCellClick(row, col, item) {
  if (!props.interactive || item === 'FREE SPACE') return
  emit('cell-click', { row, col })
}
</script>

<style scoped>
.bingo-grid {
  display: grid;
  grid-template-rows: repeat(5, 1fr);
  gap: 2px;
  background: #919191;
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
  background: rgb(88, 88, 88);
  transition: all 0.2s ease;
  min-height: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bingo-cell.interactive {
  cursor: pointer;
}

.bingo-cell.interactive:hover {
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

.bingo-cell.free-space.interactive:hover {
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
