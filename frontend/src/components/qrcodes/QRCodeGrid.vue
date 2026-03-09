<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useQRCodeStore } from '@/stores/qrcodes'
import QRCodeCard from './QRCodeCard.vue'
import Pagination from '@/components/shared/Pagination.vue'
import EmptyState from '@/components/shared/EmptyState.vue'

const props = defineProps<{ batchId: string }>()

const store = useQRCodeStore()
const page = ref(1)
const perPage = 12

onMounted(() => store.fetchQRCodes(props.batchId, page.value, perPage))

watch(page, (p) => store.fetchQRCodes(props.batchId, p, perPage))

function refresh() {
  store.fetchQRCodes(props.batchId, page.value, perPage)
}

defineExpose({ refresh })
</script>

<template>
  <div>
    <div v-if="store.loading" class="flex justify-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>
    <EmptyState
      v-else-if="!store.qrcodes.length"
      title="No QR codes yet"
      message="Generate QR codes to get started"
    />
    <div v-else class="space-y-4">
      <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
        <QRCodeCard v-for="qr in store.qrcodes" :key="qr.id" :qrcode="qr" />
      </div>
      <Pagination
        v-if="store.total > perPage"
        :total="store.total"
        :page="page"
        :per-page="perPage"
        @update:page="p => { page = p }"
      />
    </div>
  </div>
</template>
