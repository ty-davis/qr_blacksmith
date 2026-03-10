import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import * as authApi from '@/api/auth'
import type { User } from '@/types'

const TOKEN_KEY = 'qrbs_access_token'
const USER_KEY = 'qrbs_user'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(JSON.parse(sessionStorage.getItem(USER_KEY) ?? 'null'))
  const accessToken = ref<string | null>(sessionStorage.getItem(TOKEN_KEY))

  const isLoggedIn = computed(() => !!accessToken.value)

  function persist() {
    if (accessToken.value) sessionStorage.setItem(TOKEN_KEY, accessToken.value)
    else sessionStorage.removeItem(TOKEN_KEY)
    if (user.value) sessionStorage.setItem(USER_KEY, JSON.stringify(user.value))
    else sessionStorage.removeItem(USER_KEY)
  }

  async function login(email: string, password: string) {
    const data = await authApi.login(email, password)
    accessToken.value = data.access_token
    user.value = data.user
    persist()
  }

  async function register(email: string, password: string) {
    const data = await authApi.register(email, password)
    accessToken.value = data.access_token
    user.value = data.user
    persist()
  }

  async function refresh() {
    const data = await authApi.refresh()
    accessToken.value = data.access_token
    persist()
  }

  async function logout() {
    try { await authApi.logout() } catch {}
    clearAuth()
  }

  function clearAuth() {
    accessToken.value = null
    user.value = null
    sessionStorage.removeItem(TOKEN_KEY)
    sessionStorage.removeItem(USER_KEY)
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(USER_KEY)
  }

  return { user, accessToken, isLoggedIn, login, register, refresh, logout, clearAuth, persist }
})
