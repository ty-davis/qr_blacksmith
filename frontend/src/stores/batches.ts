import { defineStore } from 'pinia'
import * as batchApi from '@/api/batches'
import type { Batch } from '@/types'

export const useBatchStore = defineStore('batches', {
  state: () => ({
    batches: [] as Batch[],
    total: 0,
    currentBatch: null as Batch | null,
    loading: false,
    error: null as string | null,
  }),
  actions: {
    async fetchBatches(page = 1, perPage = 20) {
      this.loading = true
      this.error = null
      try {
        const res = await batchApi.listBatches(page, perPage)
        this.batches = res.data
        this.total = res.total
      } catch (e: unknown) {
        this.error = e instanceof Error ? e.message : 'Unknown error'
      } finally {
        this.loading = false
      }
    },
    async fetchBatch(id: string) {
      this.loading = true
      this.error = null
      try {
        this.currentBatch = await batchApi.getBatch(id)
      } catch (e: unknown) {
        this.error = e instanceof Error ? e.message : 'Unknown error'
      } finally {
        this.loading = false
      }
    },
    async createBatch(data: { name: string; description?: string; redirect_url: string }) {
      const batch = await batchApi.createBatch(data)
      this.batches.unshift(batch)
      this.total++
      return batch
    },
    async updateBatch(id: string, data: { name?: string; description?: string; redirect_url?: string }) {
      const updated = await batchApi.updateBatch(id, data)
      const idx = this.batches.findIndex(b => b.id === id)
      if (idx !== -1) this.batches[idx] = updated
      if (this.currentBatch?.id === id) this.currentBatch = updated
      return updated
    },
    async deleteBatch(id: string) {
      await batchApi.deleteBatch(id)
      this.batches = this.batches.filter(b => b.id !== id)
      this.total--
    },
  },
})
