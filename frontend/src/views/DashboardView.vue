<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getOverviewStats } from '@/api/analytics'
import type { OverviewStats } from '@/types'
import StatCard from '@/components/analytics/StatCard.vue'
import ScanLineChart from '@/components/analytics/ScanLineChart.vue'

const stats = ref<OverviewStats | null>(null)
const loading = ref(true)
const error = ref('')
const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

onMounted(async () => {
  try {
    stats.value = await getOverviewStats()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load stats'
  } finally {
    loading.value = false
  }
})

function formatDate(d: string) {
  return new Date(d).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="loading" class="flex justify-center py-16">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
    <div v-else-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
      {{ error }}
    </div>
    <template v-else-if="stats">
      <!-- Stat cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Total Batches" :value="stats.total_batches" />
        <StatCard label="Total QR Codes" :value="stats.total_qr_codes" />
        <StatCard label="Total Scans" :value="stats.total_scans" />
        <StatCard label="Scans Today" :value="stats.scans_today" />
      </div>

      <!-- Scan chart -->
      <ScanLineChart
        v-if="stats.recent_scans.length || stats.total_scans > 0"
        :data="[]"
        title="Scan Activity"
      />

      <!-- Most scanned code -->
      <div v-if="stats.most_scanned_code" class="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        <h3 class="text-sm font-medium text-gray-500 mb-3">Most Scanned QR Code</h3>
        <div class="flex items-center gap-4">
          <img
            :src="`${apiBase}/api/qrcodes/${stats.most_scanned_code.id}/image?size=64`"
            class="w-16 h-16 rounded-lg border border-gray-200"
            alt="QR code"
          />
          <div>
            <p class="font-medium text-gray-900">{{ stats.most_scanned_code.label ?? stats.most_scanned_code.hash }}</p>
            <p class="text-sm text-gray-500">{{ stats.most_scanned_code.scan_count.toLocaleString() }} scans</p>
            <p class="text-xs text-gray-400 truncate max-w-sm">{{ stats.most_scanned_code.effective_url }}</p>
          </div>
        </div>
      </div>

      <!-- Recent scans -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div class="px-6 py-4 border-b border-gray-100">
          <h3 class="text-sm font-medium text-gray-700">Recent Scans</h3>
        </div>
        <div v-if="stats.recent_scans.length" class="divide-y divide-gray-100">
          <div
            v-for="scan in stats.recent_scans.slice(0, 10)"
            :key="scan.scanned_at + scan.qr_code_id"
            class="px-6 py-3 flex items-center justify-between text-sm"
          >
            <div>
              <span class="font-medium text-gray-800">{{ scan.qr_code_label ?? scan.qr_code_id }}</span>
              <span class="text-gray-400 mx-2">·</span>
              <span class="text-gray-500">{{ scan.batch_name }}</span>
            </div>
            <div class="flex items-center gap-3 text-gray-400 text-xs">
              <span>{{ scan.device_type }}</span>
              <span v-if="scan.country">{{ scan.country }}</span>
              <span>{{ formatDate(scan.scanned_at) }}</span>
            </div>
          </div>
        </div>
        <p v-else class="px-6 py-8 text-sm text-gray-400 text-center">No recent scans</p>
      </div>
    </template>
  </div>
</template>
