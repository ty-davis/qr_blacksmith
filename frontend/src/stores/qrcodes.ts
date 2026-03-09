import { defineStore } from 'pinia'
import * as qrApi from '@/api/qrcodes'
import type { QRCode } from '@/types'

export const useQRCodeStore = defineStore('qrcodes', {
  state: () => ({
    qrcodes: [] as QRCode[],
    total: 0,
    currentQRCode: null as QRCode | null,
    loading: false,
    error: null as string | null,
  }),
  actions: {
    async fetchQRCodes(batchId: string, page = 1, perPage = 20) {
      this.loading = true
      this.error = null
      try {
        const res = await qrApi.listQRCodes(batchId, page, perPage)
        this.qrcodes = res.data
        this.total = res.total
      } catch (e: unknown) {
        this.error = e instanceof Error ? e.message : 'Unknown error'
      } finally {
        this.loading = false
      }
    },
    async generateQRCodes(batchId: string, count: number, labelPrefix?: string) {
      const res = await qrApi.generateQRCodes(batchId, count, labelPrefix)
      this.qrcodes.push(...res.qr_codes)
      this.total += res.created
      return res
    },
    async fetchQRCode(id: string) {
      this.currentQRCode = await qrApi.getQRCode(id)
    },
    async updateQRCode(id: string, data: { redirect_url?: string | null; label?: string }) {
      const updated = await qrApi.updateQRCode(id, data)
      const idx = this.qrcodes.findIndex(q => q.id === id)
      if (idx !== -1) this.qrcodes[idx] = updated
      if (this.currentQRCode?.id === id) this.currentQRCode = updated
      return updated
    },
    async deleteQRCode(id: string) {
      await qrApi.deleteQRCode(id)
      this.qrcodes = this.qrcodes.filter(q => q.id !== id)
      this.total--
    },
  },
})
