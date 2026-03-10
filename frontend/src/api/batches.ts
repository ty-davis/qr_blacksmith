import client from './client'
import type { Batch, ScanAnalytics, PaginatedResponse } from '@/types'

export const listBatches = (page = 1, perPage = 20) =>
  client.get<PaginatedResponse<Batch>>('/api/batches', { params: { page, per_page: perPage } }).then(r => r.data)

export const getBatch = (id: string) =>
  client.get<Batch>(`/api/batches/${id}`).then(r => r.data)

export const createBatch = (data: { name: string; description?: string; redirect_url: string; base_url?: string }) =>
  client.post<Batch>('/api/batches', data).then(r => r.data)

export const updateBatch = (id: string, data: { name?: string; description?: string; redirect_url?: string; base_url?: string | null }) =>
  client.put<Batch>(`/api/batches/${id}`, data).then(r => r.data)

export const deleteBatch = (id: string) =>
  client.delete(`/api/batches/${id}`)

export const updateBatchRedirect = (id: string, redirectUrl: string, overrideIndividuals: boolean) =>
  client.put(`/api/batches/${id}/redirect`, { redirect_url: redirectUrl, override_individuals: overrideIndividuals }).then(r => r.data)

export const getBatchAnalytics = (id: string, from?: string, to?: string) =>
  client.get<ScanAnalytics>(`/api/batches/${id}/analytics`, { params: { from, to } }).then(r => r.data)

export const exportBatchCSV = async (id: string, batchName: string): Promise<void> => {
  const response = await client.get(`/api/batches/${id}/export/csv`, { responseType: 'blob' })
  const url = URL.createObjectURL(response.data)
  const a = document.createElement('a')
  a.href = url
  a.download = `${batchName}.csv`
  a.click()
  URL.revokeObjectURL(url)
}
