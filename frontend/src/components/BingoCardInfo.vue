<template>
  <v-row class="mt-4">
    <v-col cols="12">
      <v-card>
        <v-card-title>Card Information</v-card-title>
        <v-card-text>
          <v-row>
            <v-col cols="6" sm="3">
              <div class="text-caption">Created</div>
              <div>{{ formatDate(currentCard.created_at) }}</div>
            </v-col>
            <v-col cols="6" sm="3">
              <div class="text-caption">Status</div>
              <v-chip
                :color="currentCard.is_winner ? 'success' : 'default'"
                size="small"
              >
                {{ currentCard.is_winner ? 'Winner!' : 'In Progress' }}
              </v-chip>
            </v-col>
            <v-col cols="6" sm="3">
              <div class="text-caption">Total Cards</div>
              <div>{{ totalCards }}</div>
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
</template>

<script setup>
const props = defineProps({
  currentCard: {
    type: Object,
    required: true
  },
  totalCards: {
    type: Number,
    required: true
  },
  winCount: {
    type: Number,
    required: true
  }
})

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
/* Component-specific styles can be added here if needed */
</style>