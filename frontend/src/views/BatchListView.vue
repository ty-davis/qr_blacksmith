<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useBatchStore } from '@/stores/batches'
import BatchCard from '@/components/batches/BatchCard.vue'
import BatchForm from '@/components/batches/BatchForm.vue'
import Pagination from '@/components/shared/Pagination.vue'
import EmptyState from '@/components/shared/EmptyState.vue'

const store = useBatchStore()
const showForm = ref(false)
const page = ref(1)
const perPage = 20

onMounted(() => store.fetchBatches(page.value, perPage))

async function handleSaved(data: { name: string; description: string; redirect_url: string }) {
  await store.createBatch(data)
  showForm.value = false
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold text-gray-900">Batches</h1>
      <button
        @click="showForm = !showForm"
        class="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition-colors"
      >
        + New Batch
      </button>
    </div>

    <!-- Inline form -->
    <div v-if="showForm" class="bg-white rounded-xl shadow-sm border border-gray-100 p-6">
      <h2 class="text-base font-semibold text-gray-900 mb-4">New Batch</h2>
      <BatchForm @saved="handleSaved" @cancel="showForm = false" />
    </div>

    <!-- Error -->
    <div v-if="store.error" class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
      {{ store.error }}
    </div>

    <!-- Loading -->
    <div v-if="store.loading" class="flex justify-center py-16">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <!-- Empty state -->
    <EmptyState
      v-else-if="!store.batches.length"
      title="No batches yet"
      message="Create your first batch to start generating QR codes"
      actionLabel="Create Batch"
      @action="showForm = true"
    />

    <!-- Batch grid -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      <BatchCard v-for="batch in store.batches" :key="batch.id" :batch="batch" />
    </div>

    <!-- Pagination -->
    <Pagination
      v-if="store.total > perPage"
      :total="store.total"
      :page="page"
      :per-page="perPage"
      @update:page="p => { page = p; store.fetchBatches(p, perPage) }"
    />
  </div>
</template>
