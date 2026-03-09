<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { QRCode } from '@/types'
import { getQRCodeImageUrl } from '@/api/qrcodes'

const props = defineProps<{ qrcode: QRCode }>()
const router = useRouter()
</script>

<template>
  <div
    @click="router.push(`/batches/${qrcode.batch_id}/qrcodes/${qrcode.id}`)"
    class="bg-white rounded-xl shadow-sm border border-gray-100 p-3 cursor-pointer hover:shadow-md hover:border-blue-200 transition-all flex flex-col items-center gap-2 text-center"
  >
    <img
      :src="getQRCodeImageUrl(qrcode.id, 80)"
      :alt="qrcode.label ?? qrcode.hash"
      class="w-14 h-14 rounded-lg border border-gray-200 object-contain flex-shrink-0"
      loading="lazy"
    />
    <div class="w-full min-w-0">
      <p class="font-medium text-gray-900 text-xs truncate">{{ qrcode.label ?? qrcode.hash }}</p>
      <p class="text-xs text-gray-400 mt-0.5">{{ qrcode.scan_count.toLocaleString() }} scans</p>
    </div>
  </div>
</template>
