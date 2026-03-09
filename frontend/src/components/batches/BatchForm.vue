<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Batch } from '@/types'

const props = defineProps<{ batch?: Batch }>()
const emit = defineEmits<{
  saved: [data: { name: string; description: string; redirect_url: string; base_url: string }]
  cancel: []
}>()

const name = ref(props.batch?.name ?? '')
const description = ref(props.batch?.description ?? '')
const redirectUrl = ref(props.batch?.redirect_url ?? '')
const baseUrl = ref(props.batch?.base_url ?? '')
const loading = ref(false)
const error = ref('')

watch(() => props.batch, (b) => {
  name.value = b?.name ?? ''
  description.value = b?.description ?? ''
  redirectUrl.value = b?.redirect_url ?? ''
  baseUrl.value = b?.base_url ?? ''
})

function submit() {
  if (!name.value.trim() || !redirectUrl.value.trim()) {
    error.value = 'Name and redirect URL are required'
    return
  }
  emit('saved', { name: name.value.trim(), description: description.value.trim(), redirect_url: redirectUrl.value.trim(), base_url: baseUrl.value.trim() })
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">{{ error }}</div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Name *</label>
      <input
        v-model="name"
        type="text"
        placeholder="My QR Campaign"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
      <textarea
        v-model="description"
        rows="2"
        placeholder="Optional description"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Default Redirect URL *</label>
      <input
        v-model="redirectUrl"
        type="url"
        placeholder="https://example.com"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700 mb-1">Base URL <span class="text-gray-400 font-normal">(optional override)</span></label>
      <input
        v-model="baseUrl"
        type="url"
        placeholder="https://your-domain.com"
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
      />
      <p class="text-xs text-gray-500 mt-1">Overrides the account and server base URL for QR scan links in this batch.</p>
    </div>
    <div class="flex gap-3 pt-2">
      <button
        @click="submit"
        :disabled="loading"
        class="flex-1 bg-blue-600 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
      >
        {{ batch ? 'Save Changes' : 'Create Batch' }}
      </button>
      <button
        @click="emit('cancel')"
        class="flex-1 bg-gray-100 text-gray-700 py-2 px-4 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
      >
        Cancel
      </button>
    </div>
  </div>
</template>
