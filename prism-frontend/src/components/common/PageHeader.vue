<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <header class="flex items-center justify-between border-b border-surface-200 pb-4">
    <div class="min-w-0">
      <h1 class="truncate text-2xl font-semibold text-surface-950">{{ title }}</h1>
      <p v-if="subtitle" class="mt-1 text-sm text-surface-500">{{ subtitle }}</p>
    </div>

    <div class="flex shrink-0 items-center gap-2">
      <slot name="actions" />
    </div>
  </header>
</template>
