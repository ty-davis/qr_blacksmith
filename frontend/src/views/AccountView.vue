<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { updateMe } from '@/api/auth'

const auth = useAuthStore()

const baseUrl = ref(auth.user?.base_url ?? '')
const saving = ref(false)
const success = ref(false)
const error = ref('')

async function save() {
  error.value = ''
  success.value = false
  saving.value = true
  try {
    const updated = await updateMe({ base_url: baseUrl.value.trim() || null })
    auth.user = updated
    auth.persist()
    success.value = true
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'Failed to save settings'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="max-w-xl mx-auto py-10 px-4">
    <h1 class="text-2xl font-bold text-gray-900 mb-6">Account Settings</h1>

    <div class="bg-white border border-gray-200 rounded-xl p-6 space-y-6">
      <div>
        <h2 class="text-base font-semibold text-gray-800 mb-1">Email</h2>
        <p class="text-sm text-gray-500">{{ auth.user?.email }}</p>
      </div>

      <hr class="border-gray-100" />

      <div>
        <h2 class="text-base font-semibold text-gray-800 mb-1">Account Base URL</h2>
        <p class="text-sm text-gray-500 mb-3">
          Sets the default domain for QR scan links across all your batches. Overridden by a batch-specific base URL.
          Leave blank to use the server default.
        </p>
        <input
          v-model="baseUrl"
          type="url"
          placeholder="https://your-domain.com"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>

      <div v-if="success" class="text-sm text-green-600 bg-green-50 border border-green-200 px-3 py-2 rounded-lg">
        Settings saved successfully.
      </div>
      <div v-if="error" class="text-sm text-red-600 bg-red-50 border border-red-200 px-3 py-2 rounded-lg">
        {{ error }}
      </div>

      <button
        @click="save"
        :disabled="saving"
        class="bg-blue-600 text-white px-5 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
      >
        {{ saving ? 'Saving…' : 'Save Changes' }}
      </button>
    </div>
  </div>
</template>
