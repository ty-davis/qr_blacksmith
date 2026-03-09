<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { Batch } from '@/types'

const props = defineProps<{ batch: Batch }>()
const router = useRouter()

function formatDate(date: string) {
  return new Date(date).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })
}
</script>

<template>
  <div
    @click="router.push(`/batches/${batch.id}`)"
    class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 cursor-pointer hover:shadow-md hover:border-blue-200 transition-all"
  >
    <div class="flex items-start justify-between mb-3">
      <h3 class="font-semibold text-gray-900 text-lg leading-tight">{{ batch.name }}</h3>
    </div>
    <p v-if="batch.description" class="text-sm text-gray-500 mb-4 line-clamp-2">{{ batch.description }}</p>
    <p class="text-xs text-gray-400 truncate mb-4">→ {{ batch.redirect_url }}</p>
    <div class="flex items-center gap-4 text-sm">
      <div class="flex items-center gap-1 text-gray-600">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8H2a11 11 0 0014 10.9M15 8h-2a11 11 0 00-8.9 5.1M9 8a11 11 0 0114 0" />
        </svg>
        <span>{{ batch.qr_code_count }} codes</span>
      </div>
      <div class="flex items-center gap-1 text-gray-600">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
        </svg>
        <span>{{ batch.total_scans.toLocaleString() }} scans</span>
      </div>
    </div>
    <p class="text-xs text-gray-400 mt-3">Created {{ formatDate(batch.created_at) }}</p>
  </div>
</template>
