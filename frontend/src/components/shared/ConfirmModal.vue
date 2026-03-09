<script setup lang="ts">
defineProps<{ title: string; message?: string; confirmLabel?: string }>()
const emit = defineEmits<{ confirm: []; cancel: [] }>()
</script>

<template>
  <div class="fixed inset-0 bg-black/50 flex items-end sm:items-center justify-center z-50 p-0 sm:p-4">
    <div class="bg-white rounded-t-2xl sm:rounded-xl shadow-xl w-full sm:max-w-sm p-6">
      <h2 class="text-lg font-semibold text-gray-900 mb-2">{{ title }}</h2>
      <p v-if="message" class="text-sm text-gray-600 mb-6">{{ message }}</p>
      <div class="flex gap-3">
        <button
          @click="emit('confirm')"
          class="flex-1 bg-red-600 text-white py-3 px-4 rounded-lg text-sm font-medium hover:bg-red-700 transition-colors"
        >
          {{ confirmLabel ?? 'Confirm' }}
        </button>
        <button
          @click="emit('cancel')"
          class="flex-1 bg-gray-100 text-gray-700 py-3 px-4 rounded-lg text-sm font-medium hover:bg-gray-200 transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>
