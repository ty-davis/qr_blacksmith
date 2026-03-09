import client from './client'
import type { User } from '@/types'

interface AuthResponse {
  access_token: string
  user: User
}

export const register = (email: string, password: string) =>
  client.post<AuthResponse>('/api/auth/register', { email, password }).then(r => r.data)

export const login = (email: string, password: string) =>
  client.post<AuthResponse>('/api/auth/login', { email, password }).then(r => r.data)

export const refresh = () =>
  client.post<{ access_token: string }>('/api/auth/refresh').then(r => r.data)

export const logout = () =>
  client.post('/api/auth/logout')

export const getMe = () =>
  client.get<User>('/api/auth/me').then(r => r.data)

export const updateMe = (data: { base_url?: string | null }) =>
  client.patch<User>('/api/auth/me', data).then(r => r.data)
