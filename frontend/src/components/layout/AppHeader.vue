<script setup lang="ts">
import { useRoute } from 'vue-router'

const emit = defineEmits<{ 'toggle-sidebar': [] }>()
const route = useRoute()

const titles: Record<string, string> = {
  dashboard: 'Dashboard',
  batches: 'Batches',
  'batch-detail': 'Batch Details',
  'qrcode-detail': 'QR Code Details',
  account: 'Account Settings',
}
</script>

<template>
  <header class="bg-white border-b border-gray-200 px-4 py-4 flex items-center gap-3 flex-shrink-0">
    <button
      @click="emit('toggle-sidebar')"
      class="lg:hidden p-1.5 rounded-md text-gray-500 hover:text-gray-700 hover:bg-gray-100 transition-colors"
      aria-label="Toggle menu"
    >
      <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
      </svg>
    </button>
    <h2 class="text-lg font-semibold text-gray-900">
      {{ titles[route.name as string] ?? 'QR Blacksmith' }}
    </h2>
  </header>
</template>
