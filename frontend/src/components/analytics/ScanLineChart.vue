<script setup lang="ts">
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const props = defineProps<{
  data: { date: string; count: number }[]
  title?: string
}>()

const chartData = computed(() => ({
  labels: props.data.map(d => d.date),
  datasets: [
    {
      label: 'Scans',
      data: props.data.map(d => d.count),
      borderColor: '#3b82f6',
      backgroundColor: 'rgba(59, 130, 246, 0.1)',
      fill: true,
      tension: 0.3,
    },
  ],
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    title: { display: false },
  },
  scales: {
    y: { beginAtZero: true, ticks: { stepSize: 1 } },
  },
}
</script>

<template>
  <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 w-full">
    <h3 v-if="title" class="text-sm font-medium text-gray-500 mb-4">{{ title }}</h3>
    <div class="h-48">
      <Line v-if="data.length" :data="chartData" :options="options" />
      <div v-else class="flex items-center justify-center h-full text-gray-400 text-sm">No scan data yet</div>
    </div>
  </div>
</template>
