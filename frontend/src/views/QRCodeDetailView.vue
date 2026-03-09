<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useQRCodeStore } from '@/stores/qrcodes'
import * as qrApi from '@/api/qrcodes'
import type { ScanAnalytics } from '@/types'
import ScanLineChart from '@/components/analytics/ScanLineChart.vue'
import DevicePieChart from '@/components/analytics/DevicePieChart.vue'
import CountryTable from '@/components/analytics/CountryTable.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'

const route = useRoute()
const router = useRouter()
const qrStore = useQRCodeStore()
const batchId = route.params.id as string
const qrId = route.params.qrId as string

const analytics = ref<ScanAnalytics | null>(null)
const showDeleteConfirm = ref(false)
const editLabel = ref('')
const editRedirectUrl = ref('')
const saving = ref(false)
const error = ref('')

onMounted(async () => {
  await qrStore.fetchQRCode(qrId)
  if (qrStore.currentQRCode) {
    editLabel.value = qrStore.currentQRCode.label ?? ''
    editRedirectUrl.value = qrStore.currentQRCode.redirect_url ?? ''
  }
  try {
    analytics.value = await qrApi.getQRCodeAnalytics(qrId)
  } catch {}
})

async function saveLabel() {
  saving.value = true
  try {
    await qrStore.updateQRCode(qrId, { label: editLabel.value })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function saveRedirect() {
  saving.value = true
  try {
    await qrStore.updateQRCode(qrId, { redirect_url: editRedirectUrl.value || null })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  await qrStore.deleteQRCode(qrId)
  router.push(`/batches/${batchId}`)
}

function downloadQR() {
  const url = qrApi.getQRCodeImageUrl(qrId, 512)
  const a = document.createElement('a')
  a.href = url
  a.download = `qr-${qrStore.currentQRCode?.label ?? qrId}.png`
  a.click()
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="!qrStore.currentQRCode" class="flex justify-center py-16">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
    <template v-else>
      <!-- Header -->
      <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
        <div class="min-w-0">
          <button @click="router.back()" class="text-sm text-gray-400 hover:text-gray-600 mb-2">← Back</button>
          <h1 class="text-xl font-semibold text-gray-900 truncate">{{ qrStore.currentQRCode.label ?? qrStore.currentQRCode.hash }}</h1>
        </div>
        <button @click="showDeleteConfirm = true" class="self-start px-3 py-2 text-sm border border-red-200 text-red-600 rounded-lg hover:bg-red-50 transition-colors flex-shrink-0">
          Delete
        </button>
      </div>

      <!-- Error -->
      <div v-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">{{ error }}</div>

      <!-- Main content -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- QR image card -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 flex flex-col items-center gap-4">
          <img
            :src="qrApi.getQRCodeImageUrl(qrId, 256)"
            :alt="qrStore.currentQRCode.label ?? qrStore.currentQRCode.hash"
            class="w-48 h-48 rounded-xl border border-gray-200"
          />
          <button @click="downloadQR" class="w-full bg-blue-600 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors">
            Download
          </button>
          <div class="w-full space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-gray-500">Scans</span>
              <span class="font-medium text-gray-900">{{ qrStore.currentQRCode.scan_count.toLocaleString() }}</span>
            </div>
            <div v-if="qrStore.currentQRCode.last_scanned_at" class="flex justify-between">
              <span class="text-gray-500">Last scan</span>
              <span class="font-medium text-gray-900">{{ new Date(qrStore.currentQRCode.last_scanned_at).toLocaleDateString() }}</span>
            </div>
          </div>
        </div>

        <!-- Edit fields -->
        <div class="lg:col-span-2 space-y-4">
          <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-4">
            <h3 class="font-medium text-gray-900">Details</h3>
            <!-- Label -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Label</label>
              <div class="flex gap-2">
                <input v-model="editLabel" type="text" placeholder="Optional label" class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                <button @click="saveLabel" :disabled="saving" class="px-3 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors">Save</button>
              </div>
            </div>
            <!-- Personal redirect URL -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Personal Redirect URL</label>
              <p class="text-xs text-gray-400 mb-1">Leave empty to use batch default: {{ qrStore.currentQRCode.effective_url }}</p>
              <div class="flex gap-2">
                <input v-model="editRedirectUrl" type="url" placeholder="https://custom-url.com (optional)" class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                <button @click="saveRedirect" :disabled="saving" class="px-3 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors">Save</button>
              </div>
            </div>
            <div>
              <span class="text-sm font-medium text-gray-700">Scan URL: </span>
              <a :href="qrStore.currentQRCode.effective_url" target="_blank" class="text-sm text-blue-600 hover:underline break-all">
                {{ qrStore.currentQRCode.effective_url }}
              </a>
            </div>
          </div>
        </div>
      </div>

      <!-- Analytics -->
      <div v-if="analytics" class="grid grid-cols-1 lg:grid-cols-3 gap-4">
        <div class="lg:col-span-2">
          <ScanLineChart :data="analytics.scans_by_day" title="Scans Over Time" />
        </div>
        <DevicePieChart :data="analytics.device_breakdown" />
        <div class="lg:col-span-3">
          <CountryTable :data="analytics.top_countries" />
        </div>
      </div>
    </template>

    <ConfirmModal
      v-if="showDeleteConfirm"
      title="Delete QR code?"
      message="This will permanently delete the QR code and all its scan data."
      confirm-label="Delete"
      @confirm="handleDelete"
      @cancel="showDeleteConfirm = false"
    />
  </div>
</template>
