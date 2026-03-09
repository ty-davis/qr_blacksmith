<script setup lang="ts">
import { useRoute } from 'vue-router'

const route = useRoute()

const titles: Record<string, string> = {
  dashboard: 'Dashboard',
  batches: 'Batches',
  'batch-detail': 'Batch Details',
  'qrcode-detail': 'QR Code Details',
}
</script>

<template>
  <header class="bg-white border-b border-gray-200 px-6 py-4">
    <h2 class="text-lg font-semibold text-gray-900">
      {{ titles[route.name as string] ?? 'QR Blacksmith' }}
    </h2>
  </header>
</template>
