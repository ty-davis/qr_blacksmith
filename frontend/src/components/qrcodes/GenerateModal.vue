<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  generated: [data: { count: number; labelPrefix: string }]
  cancel: []
}>()

const count = ref(1)
const labelPrefix = ref('')

function submit() {
  if (count.value < 1 || count.value > 500) return
  emit('generated', { count: count.value, labelPrefix: labelPrefix.value.trim() })
}
</script>

<template>
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
    <div class="bg-white rounded-xl shadow-xl w-full max-w-md p-6">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">Generate QR Codes</h2>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Number of codes (1–500)</label>
          <input
            v-model.number="count"
            type="number"
            min="1"
            max="500"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Label prefix (optional)</label>
          <input
            v-model="labelPrefix"
            type="text"
            placeholder="e.g. Poster"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <p class="text-xs text-gray-400 mt-1">Codes will be labeled "Prefix 1", "Prefix 2", etc.</p>
        </div>
      </div>
      <div class="flex gap-3 mt-6">
        <button
          @click="submit"
          class="flex-1 bg-blue-600 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
        >
          Generate
        </button>
        <button
          @click="emit('cancel')"
          class="flex-1 bg-gray-100 text-gray-700 py-2 px-4 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>
