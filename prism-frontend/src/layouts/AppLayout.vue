<template>
  <div class="prism-shell flex h-screen overflow-hidden">
    <AppSidebar />
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <AppTopbar />
      <main class="prism-shell-content flex-1 overflow-y-auto p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import AppSidebar from '@/layouts/components/AppSidebar.vue'
import AppTopbar from '@/layouts/components/AppTopbar.vue'
</script>
