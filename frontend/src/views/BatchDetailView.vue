<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBatchStore } from '@/stores/batches'
import * as batchApi from '@/api/batches'
import type { ScanAnalytics } from '@/types'
import ScanLineChart from '@/components/analytics/ScanLineChart.vue'
import DevicePieChart from '@/components/analytics/DevicePieChart.vue'
import CountryTable from '@/components/analytics/CountryTable.vue'
import GenerateModal from '@/components/qrcodes/GenerateModal.vue'
import QRCodeGrid from '@/components/qrcodes/QRCodeGrid.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import { useQRCodeStore } from '@/stores/qrcodes'

const route = useRoute()
const router = useRouter()
const batchStore = useBatchStore()
const qrStore = useQRCodeStore()
const batchId = route.params.id as string

const analytics = ref<ScanAnalytics | null>(null)
const showGenerateModal = ref(false)
const showDeleteConfirm = ref(false)
const editing = ref(false)
const editName = ref('')
const editDescription = ref('')
const editRedirectUrl = ref('')
const bulkUrl = ref('')
const overrideIndividuals = ref(false)
const gridRef = ref<InstanceType<typeof QRCodeGrid> | null>(null)
const saving = ref(false)
const error = ref('')

onMounted(async () => {
  await batchStore.fetchBatch(batchId)
  if (batchStore.currentBatch) {
    editName.value = batchStore.currentBatch.name
    editDescription.value = batchStore.currentBatch.description ?? ''
    editRedirectUrl.value = batchStore.currentBatch.redirect_url
    bulkUrl.value = batchStore.currentBatch.redirect_url
  }
  try {
    analytics.value = await batchApi.getBatchAnalytics(batchId)
  } catch {}
})

async function saveEdit() {
  saving.value = true
  try {
    await batchStore.updateBatch(batchId, {
      name: editName.value,
      description: editDescription.value,
      redirect_url: editRedirectUrl.value,
    })
    editing.value = false
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function handleBulkRedirect() {
  saving.value = true
  try {
    await batchApi.updateBatchRedirect(batchId, bulkUrl.value, overrideIndividuals.value)
    await batchStore.fetchBatch(batchId)
    if (overrideIndividuals.value) gridRef.value?.refresh()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update redirect'
  } finally {
    saving.value = false
  }
}

async function handleGenerated(data: { count: number; labelPrefix: string }) {
  showGenerateModal.value = false
  await qrStore.generateQRCodes(batchId, data.count, data.labelPrefix || undefined)
  gridRef.value?.refresh()
}

async function handleDelete() {
  await batchStore.deleteBatch(batchId)
  router.push('/batches')
}
</script>

<template>
  <div class="space-y-6">
    <div v-if="batchStore.loading" class="flex justify-center py-16">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
    <template v-else-if="batchStore.currentBatch">
      <!-- Header -->
      <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3">
        <div class="min-w-0">
          <button @click="router.back()" class="text-sm text-gray-400 hover:text-gray-600 mb-2">← Back</button>
          <h1 class="text-xl font-semibold text-gray-900 truncate">{{ batchStore.currentBatch.name }}</h1>
          <p v-if="batchStore.currentBatch.description" class="text-sm text-gray-500 mt-1">{{ batchStore.currentBatch.description }}</p>
        </div>
        <div class="flex flex-wrap gap-2 flex-shrink-0">
          <button
            @click="editing = !editing"
            class="px-3 py-2 text-sm border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
          >
            {{ editing ? 'Cancel' : 'Edit' }}
          </button>
          <button
            @click="showGenerateModal = true"
            class="px-3 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            + Generate QR Codes
          </button>
          <button
            @click="showDeleteConfirm = true"
            class="px-3 py-2 text-sm border border-red-200 text-red-600 rounded-lg hover:bg-red-50 transition-colors"
          >
            Delete
          </button>
        </div>
      </div>

      <!-- Error -->
      <div v-if="error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">{{ error }}</div>

      <!-- Edit form -->
      <div v-if="editing" class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-4">
        <h3 class="font-medium text-gray-900">Edit Batch</h3>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
          <input v-model="editName" type="text" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
          <textarea v-model="editDescription" rows="2" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Default Redirect URL</label>
          <input v-model="editRedirectUrl" type="url" class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <button @click="saveEdit" :disabled="saving" class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors">
          Save Changes
        </button>
      </div>

      <!-- Bulk redirect -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6 space-y-3">
        <h3 class="font-medium text-gray-900">Bulk Redirect Update</h3>
        <div class="flex flex-col sm:flex-row gap-3">
          <input
            v-model="bulkUrl"
            type="url"
            placeholder="https://new-destination.com"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <button @click="handleBulkRedirect" :disabled="saving" class="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors">
            Update
          </button>
        </div>
        <label class="flex items-center gap-2 text-sm text-gray-600 cursor-pointer">
          <input v-model="overrideIndividuals" type="checkbox" class="rounded border-gray-300" />
          Override individual QR code URLs
        </label>
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

      <!-- QR Code Grid -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
        <h3 class="font-medium text-gray-900 mb-4">QR Codes ({{ batchStore.currentBatch.qr_code_count }})</h3>
        <QRCodeGrid ref="gridRef" :batch-id="batchId" />
      </div>
    </template>

    <!-- Modals -->
    <GenerateModal v-if="showGenerateModal" @generated="handleGenerated" @cancel="showGenerateModal = false" />
    <ConfirmModal
      v-if="showDeleteConfirm"
      title="Delete batch?"
      message="This will permanently delete the batch and all its QR codes. This cannot be undone."
      confirm-label="Delete"
      @confirm="handleDelete"
      @cancel="showDeleteConfirm = false"
    />
  </div>
</template>
