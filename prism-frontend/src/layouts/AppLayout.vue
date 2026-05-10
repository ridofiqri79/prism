<template>
  <div class="prism-shell flex h-screen overflow-hidden">
    <!-- Mobile overlay backdrop -->
    <Transition name="prism-overlay">
      <div
        v-if="isMobileSidebarOpen"
        class="fixed inset-0 z-40 bg-black/40 lg:hidden"
        aria-hidden="true"
        @click="isMobileSidebarOpen = false"
      />
    </Transition>

    <AppSidebar :is-mobile-open="isMobileSidebarOpen" @close="isMobileSidebarOpen = false" />

    <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <AppTopbar @toggle-sidebar="isMobileSidebarOpen = !isMobileSidebarOpen" />
      <main class="prism-shell-content flex-1 overflow-y-auto p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppSidebar from '@/layouts/components/AppSidebar.vue'
import AppTopbar from '@/layouts/components/AppTopbar.vue'

const route = useRoute()
const isMobileSidebarOpen = ref(false)

// Close drawer automatically on route navigation
watch(
  () => route.path,
  () => {
    isMobileSidebarOpen.value = false
  },
)
</script>

<style>
.prism-overlay-enter-active,
.prism-overlay-leave-active {
  transition: opacity 0.2s ease;
}
.prism-overlay-enter-from,
.prism-overlay-leave-to {
  opacity: 0;
}
</style>
