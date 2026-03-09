<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import AppHeader from '@/components/layout/AppHeader.vue'

const route = useRoute()
const isPublic = computed(() => !!route.meta.public)
const sidebarOpen = ref(false)
</script>

<template>
  <div v-if="isPublic">
    <RouterView />
  </div>
  <div v-else class="flex h-screen bg-gray-50">
    <AppSidebar :open="sidebarOpen" @close="sidebarOpen = false" />
    <div class="flex-1 flex flex-col overflow-hidden min-w-0">
      <AppHeader @toggle-sidebar="sidebarOpen = !sidebarOpen" />
      <main class="flex-1 overflow-y-auto p-4 sm:p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>
