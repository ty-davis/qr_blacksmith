<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'

ChartJS.register(ArcElement, Tooltip, Legend)

const props = defineProps<{
  data: { device_type: string; count: number }[]
}>()

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

const chartData = computed(() => ({
  labels: props.data.map(d => d.device_type),
  datasets: [
    {
      data: props.data.map(d => d.count),
      backgroundColor: COLORS.slice(0, props.data.length),
      borderWidth: 2,
      borderColor: '#fff',
    },
  ],
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { position: 'bottom' as const },
  },
}
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
    <h3 class="text-sm font-medium text-gray-500 mb-4">Devices</h3>
    <div class="h-48">
      <Doughnut v-if="data.length" :data="chartData" :options="options" />
      <div v-else class="flex items-center justify-center h-full text-gray-400 text-sm">No device data yet</div>
    </div>
  </div>
</template>
