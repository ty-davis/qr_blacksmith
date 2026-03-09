import client from './client'
import type { OverviewStats } from '@/types'

export const getOverviewStats = () =>
  client.get<OverviewStats>('/api/analytics/overview').then(r => r.data)
