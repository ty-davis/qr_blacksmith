export interface User {
  id: string
  email: string
  base_url: string | null
  created_at: string
}

export interface Batch {
  id: string
  name: string
  description: string | null
  redirect_url: string
  base_url: string | null
  qr_code_count: number
  total_scans: number
  created_at: string
  updated_at: string
}

export interface QRCode {
  id: string
  hash: string
  batch_id: string
  redirect_url: string | null
  effective_url: string
  label: string | null
  scan_count: number
  last_scanned_at: string | null
  qr_image_url: string
  scan_url: string
  created_at: string
}

export interface ScanAnalytics {
  total_scans: number
  unique_codes_scanned?: number
  scans_by_day: { date: string; count: number }[]
  top_countries: { country: string; country_code: string; count: number }[]
  device_breakdown: { device_type: string; count: number }[]
}

export interface OverviewStats {
  total_batches: number
  total_qr_codes: number
  total_scans: number
  scans_today: number
  scans_this_week: number
  most_scanned_code: QRCode | null
  recent_scans: RecentScan[]
}

export interface RecentScan {
  qr_code_id: string
  qr_code_label: string | null
  batch_name: string
  scanned_at: string
  city: string | null
  country: string | null
  device_type: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  per_page: number
}
