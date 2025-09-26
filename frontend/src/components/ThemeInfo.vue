<template>
    <v-row>
        <v-col cols="12">
            <v-card class="mb-4 text-center" variant="outlined">
                <v-card-text class="d-flex justify-space-between align-center">
                    <div>
                        <div class="text-h6">{{ activeTheme.name }}</div>
                        <div class="text-body-2 text-medium-emphasis">{{ activeTheme.description }}</div>
                    </div>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>
</template>

<script setup>
const props = defineProps({
    activeTheme: {
        type: Object,
        required: true
    },
}) 
</script>
