<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ total: number; page: number; perPage: number }>()
const emit = defineEmits<{ 'update:page': [page: number] }>()

const totalPages = computed(() => Math.ceil(props.total / props.perPage))
</script>

<template>
  <div class="flex items-center justify-between pt-4">
    <p class="text-sm text-gray-500">
      {{ (page - 1) * perPage + 1 }}–{{ Math.min(page * perPage, total) }} of {{ total.toLocaleString() }}
    </p>
    <div class="flex gap-2">
      <button
        :disabled="page <= 1"
        @click="emit('update:page', page - 1)"
        class="px-3 py-1.5 text-sm border border-gray-200 rounded-lg disabled:opacity-40 hover:bg-gray-50 transition-colors"
      >
        ← Prev
      </button>
      <button
        :disabled="page >= totalPages"
        @click="emit('update:page', page + 1)"
        class="px-3 py-1.5 text-sm border border-gray-200 rounded-lg disabled:opacity-40 hover:bg-gray-50 transition-colors"
      >
        Next →
      </button>
    </div>
  </div>
</template>
