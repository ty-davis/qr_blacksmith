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
    class="bg-white rounded-xl shadow-sm border border-gray-100 p-4 cursor-pointer hover:shadow-md hover:border-blue-200 transition-all flex gap-4 items-start"
  >
    <img
      :src="getQRCodeImageUrl(qrcode.id, 80)"
      :alt="qrcode.label ?? qrcode.hash"
      class="w-16 h-16 rounded-lg border border-gray-200 object-contain flex-shrink-0"
      loading="lazy"
    />
    <div class="flex-1 min-w-0">
      <p class="font-medium text-gray-900 text-sm truncate">{{ qrcode.label ?? qrcode.hash }}</p>
      <p class="text-xs text-gray-400 truncate mt-0.5">{{ qrcode.effective_url }}</p>
      <div class="flex items-center gap-3 mt-2 text-xs text-gray-500">
        <span>{{ qrcode.scan_count.toLocaleString() }} scans</span>
        <span v-if="qrcode.last_scanned_at" class="text-gray-400">
          Last: {{ new Date(qrcode.last_scanned_at).toLocaleDateString() }}
        </span>
      </div>
    </div>
  </div>
</template>
