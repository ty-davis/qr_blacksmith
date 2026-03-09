<script setup lang="ts">
defineProps<{
  data: { country: string; country_code: string; count: number }[]
}>()
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
    <h3 class="text-sm font-medium text-gray-500 mb-4">Top Countries</h3>
    <div v-if="data.length" class="space-y-3">
      <div
        v-for="(item, idx) in data.slice(0, 10)"
        :key="item.country_code"
        class="flex items-center justify-between"
      >
        <div class="flex items-center gap-2">
          <span class="text-xs text-gray-400 w-5">{{ idx + 1 }}</span>
          <span class="text-sm text-gray-700">{{ item.country || 'Unknown' }}</span>
        </div>
        <span class="text-sm font-medium text-gray-900">{{ item.count.toLocaleString() }}</span>
      </div>
    </div>
    <p v-else class="text-sm text-gray-400 text-center py-4">No country data yet</p>
  </div>
</template>
