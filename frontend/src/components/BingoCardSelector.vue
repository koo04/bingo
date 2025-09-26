<template>
  <v-row>
    <v-col cols="12" md="6" lg="4">
      <v-card>
        <v-card-title class="d-flex align-center">
          <v-icon class="mr-2">mdi-cards</v-icon>
          Your Bingo Cards
        </v-card-title>
        <v-card-text>
          <v-select
            :model-value="selectedCardId"
            :items="cardSelectItems"
            item-title="text"
            item-value="value"
            @update:model-value="handleCardChange"
            variant="outlined"
            density="compact"
          >
          </v-select>
        </v-card-text>
      </v-card>
    </v-col>
  </v-row>
</template>

<script setup>
const props = defineProps({
  selectedCardId: {
    type: String,
    required: true
  },
  cardSelectItems: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['card-change'])

function handleCardChange(cardId) {
  emit('card-change', cardId)
}
</script>

<style scoped>
/* Component-specific styles can be added here if needed */
</style>
