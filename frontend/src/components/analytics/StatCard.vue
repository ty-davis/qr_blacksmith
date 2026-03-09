<script setup lang="ts">
defineProps<{ label: string; value: number | string; icon?: string; color?: string }>()
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
    <p class="text-sm font-medium text-gray-500">{{ label }}</p>
    <p class="text-3xl font-bold text-gray-900 mt-2">{{ value.toLocaleString() }}</p>
  </div>
</template>
