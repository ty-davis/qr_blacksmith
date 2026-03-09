import client from './client'
import type { QRCode, ScanAnalytics, PaginatedResponse } from '@/types'

export const listQRCodes = (batchId: string, page = 1, perPage = 20) =>
  client.get<PaginatedResponse<QRCode>>(`/api/batches/${batchId}/qrcodes`, { params: { page, per_page: perPage } }).then(r => r.data)

export const generateQRCodes = (batchId: string, count: number, labelPrefix?: string) =>
  client.post<{ created: number; qr_codes: QRCode[] }>(`/api/batches/${batchId}/qrcodes`, { count, label_prefix: labelPrefix }).then(r => r.data)

export const getQRCode = (id: string) =>
  client.get<QRCode>(`/api/qrcodes/${id}`).then(r => r.data)

export const updateQRCode = (id: string, data: { redirect_url?: string | null; label?: string }) =>
  client.put<QRCode>(`/api/qrcodes/${id}`, data).then(r => r.data)

export const deleteQRCode = (id: string) =>
  client.delete(`/api/qrcodes/${id}`)

export const getQRCodeAnalytics = (id: string, from?: string, to?: string) =>
  client.get<ScanAnalytics>(`/api/qrcodes/${id}/analytics`, { params: { from, to } }).then(r => r.data)

export const getQRCodeImageUrl = (id: string, size = 256) =>
  `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/api/qrcodes/${id}/image?size=${size}`
