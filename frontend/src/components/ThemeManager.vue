<template>
  <div>
    <v-card class="mb-6">
      <v-card-title class="d-flex justify-space-between align-center">
        <span>Theme Management</span>
        <v-btn color="primary" @click="showCreateDialog = true">
          <v-icon left>mdi-plus</v-icon>
          Create Theme
        </v-btn>
      </v-card-title>
      
      <v-card-text>
        <div v-if="store.themes.length === 0" class="text-center py-8">
          <v-icon size="64" color="grey" class="mb-4">mdi-palette-outline</v-icon>
          <div class="text-h6 mb-2">No Themes Available</div>
          <div class="text-body-2 text-medium-emphasis mb-4">
            Create your first theme to get started with bingo games.
          </div>
          <v-btn color="primary" @click="showCreateDialog = true">
            Create First Theme
          </v-btn>
        </div>
        
        <div v-else>
          <v-row>
            <v-col
              v-for="theme in store.themes"
              :key="theme.id"
              cols="12"
              md="6"
              lg="4"
            >
              <v-card
                :color="theme.id === store.activeThemeId ? 'primary' : ''"
                :variant="theme.id === store.activeThemeId ? 'flat' : 'outlined'"
                class="theme-card"
              >
                <v-card-text>
                  <div class="d-flex justify-space-between align-start mb-2">
                    <div>
                      <div class="text-h6">{{ theme.name }}</div>
                      <div class="text-body-2 text-medium-emphasis">
                        {{ theme.description || 'No description' }}
                      </div>
                    </div>
                    <div class="d-flex gap-2">
                      <v-chip
                        v-if="theme.id === store.activeThemeId"
                        color="white"
                        size="small"
                        variant="flat"
                      >
                        Active
                      </v-chip>
                      <v-chip
                        v-if="theme.is_complete"
                        color="grey"
                        size="small"
                        variant="flat"
                      >
                        Complete
                      </v-chip>
                    </div>
                  </div>
                  
                  <div class="text-caption mb-3">
                    {{ theme.items.length }} items
                    <v-chip
                      :color="theme.items.length >= 25 ? 'success' : 'warning'"
                      size="x-small"
                      class="ml-2"
                    >
                      {{ theme.items.length >= 25 ? 'Ready' : 'Need ' + (25 - theme.items.length) + ' more' }}
                    </v-chip>
                  </div>
                  
                  <div class="d-flex gap-2 flex-wrap">
                    <v-btn
                      v-if="theme.id !== store.activeThemeId && theme.items.length >= 25 && !theme.is_complete"
                      size="small"
                      color="primary"
                      @click="setActiveTheme(theme.id)"
                    >
                      Make Active
                    </v-btn>
                    
                    <v-btn
                      v-if="!theme.is_complete"
                      size="small"
                      color="info"
                      @click="editTheme(theme)"
                    >
                      <v-icon left size="small">mdi-pencil</v-icon>
                      Edit
                    </v-btn>
                    
                    <v-btn
                      size="small"
                      :color="theme.is_complete ? 'warning' : 'success'"
                      @click="toggleThemeComplete(theme)"
                    >
                      <v-icon left size="small">
                        {{ theme.is_complete ? 'mdi-undo' : 'mdi-check' }}
                      </v-icon>
                      {{ theme.is_complete ? 'Reopen' : 'Complete' }}
                    </v-btn>
                    
                    <v-btn
                      v-if="!theme.is_complete && theme.id !== store.activeThemeId"
                      size="small"
                      color="error"
                      @click="confirmDeleteTheme(theme)"
                    >
                      <v-icon left size="small">mdi-delete</v-icon>
                      Delete
                    </v-btn>
                    
                    <v-tooltip v-if="!theme.is_complete && theme.id === store.activeThemeId" bottom>
                      <template v-slot:activator="{ props }">
                        <v-btn
                          v-bind="props"
                          size="small"
                          color="error"
                          disabled
                          variant="outlined"
                        >
                          <v-icon left size="small">mdi-delete</v-icon>
                          Delete
                        </v-btn>
                      </template>
                      <span>Cannot delete the active theme. Set a different theme as active first.</span>
                    </v-tooltip>
                  </div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>
        </div>
      </v-card-text>
    </v-card>

    <!-- Create/Edit Theme Dialog -->
    <v-dialog v-model="showCreateDialog" max-width="800px" scrollable>
      <v-card>
        <v-card-title>
          {{ editingTheme ? 'Edit Theme' : 'Create New Theme' }}
        </v-card-title>
        
        <v-card-text>
          <v-form v-model="formValid" @submit.prevent="saveTheme">
            <v-text-field
              v-model="themeForm.name"
              label="Theme Name"
              :rules="[v => !!v || 'Name is required']"
              required
            />
            
            <v-textarea
              v-model="themeForm.description"
              label="Description"
              rows="2"
              counter="200"
              :rules="[v => !v || v.length <= 200 || 'Description must be less than 200 characters']"
            />
            
            <div class="text-h6 mb-3">
              Bingo Items
              <v-chip
                :color="itemsFromText.length >= 25 ? 'success' : 'warning'"
                size="small"
                class="ml-2"
              >
                {{ itemsFromText.length }}/25+ items
              </v-chip>
            </div>
            
            <div class="mb-4">
              <v-textarea
                v-model="itemsText"
                label="Items (one per line)"
                rows="10"
                placeholder="Enter each bingo item on a new line..."
                :rules="[validateItems]"
                hint="Enter at least 25 items, one per line"
                persistent-hint
                auto-grow
                no-resize
                spellcheck="false"
              />
            </div>
            
            <v-alert
              v-if="itemsFromText.length < 25"
              type="warning"
              variant="tonal"
              class="mb-4"
            >
              You need at least 25 items to create a playable bingo theme.
              Currently have {{ itemsFromText.length }} items.
            </v-alert>
          </v-form>
        </v-card-text>
        
        <v-card-actions>
          <v-spacer />
          <v-btn @click="cancelEdit">Cancel</v-btn>
          <v-btn
            color="primary"
            :disabled="!formValid || itemsFromText.length < 25"
            @click="saveTheme"
            :loading="saving"
          >
            {{ editingTheme ? 'Update' : 'Create' }} Theme
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="showDeleteDialog" max-width="400px">
      <v-card>
        <v-card-title>Delete Theme</v-card-title>
        <v-card-text>
          Are you sure you want to delete the theme "{{ themeToDelete?.name }}"?
          This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="showDeleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="deleteTheme" :loading="deleting">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

// Dialog states
const showCreateDialog = ref(false)
const showDeleteDialog = ref(false)
const editingTheme = ref(null)
const themeToDelete = ref(null)

// Form states
const formValid = ref(false)
const saving = ref(false)
const deleting = ref(false)

// Form data
const themeForm = ref({
  name: '',
  description: '',
  items: []
})

// Items as text (for easier editing)
const itemsText = ref('')

// Computed property to get items from text without circular updates
const itemsFromText = computed(() => {
  return itemsText.value
    .split('\n')
    .map(item => item.trim())
    .filter(item => item.length > 0)
})

// Watch itemsText changes to update items array (but don't watch the reverse)
watch(itemsText, () => {
  themeForm.value.items = itemsFromText.value
})

// Validation rules
const validateItems = (value) => {
  const items = value.split('\n').map(item => item.trim()).filter(item => item.length > 0)
  if (items.length < 25) {
    return `Need at least 25 items (currently have ${items.length})`
  }
  return true
}

// Actions
async function setActiveTheme(themeId) {
  try {
    await store.setActiveTheme(themeId)
  } catch (error) {
    console.error('Failed to set active theme:', error)
  }
}

function editTheme(theme) {
  editingTheme.value = theme
  themeForm.value = {
    name: theme.name,
    description: theme.description || '',
    items: [...theme.items]
  }
  itemsText.value = theme.items.join('\n')
  showCreateDialog.value = true
}

function confirmDeleteTheme(theme) {
  themeToDelete.value = theme
  showDeleteDialog.value = true
}

async function saveTheme() {
  if (!formValid.value || itemsFromText.value.length < 25) return
  
  // Make sure the theme form has the latest items from the text
  themeForm.value.items = itemsFromText.value
  
  saving.value = true
  try {
    if (editingTheme.value) {
      await store.updateTheme(editingTheme.value.id, themeForm.value)
    } else {
      await store.createTheme(themeForm.value)
    }
    
    cancelEdit()
    await store.fetchThemes()
  } catch (error) {
    console.error('Failed to save theme:', error)
  } finally {
    saving.value = false
  }
}

async function deleteTheme() {
  if (!themeToDelete.value) return
  
  deleting.value = true
  try {
    await store.deleteTheme(themeToDelete.value.id)
    showDeleteDialog.value = false
    themeToDelete.value = null
    await store.fetchThemes()
  } catch (error) {
    console.error('Failed to delete theme:', error)
  } finally {
    deleting.value = false
  }
}

async function toggleThemeComplete(theme) {
    console.log('Toggling theme complete status for', theme)

  try {
    if (theme.is_complete) {
      // Reopen theme - update the theme to set is_complete to false
      const updatedTheme = { ...theme, is_complete: false }
      await store.updateTheme(theme.id, updatedTheme)
    } else {
      await store.markThemeComplete(theme.id)
    }
  } catch (error) {
    console.error(`Failed to ${theme.is_complete ? 'reopen' : 'complete'} theme:`, error)
  }
}

function cancelEdit() {
  showCreateDialog.value = false
  editingTheme.value = null
  themeForm.value = {
    name: '',
    description: '',
    items: []
  }
  itemsText.value = ''
}
</script>

<style scoped>
.theme-card {
  height: 100%;
}
</style>
