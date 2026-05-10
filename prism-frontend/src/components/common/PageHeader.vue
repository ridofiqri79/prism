<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <header class="flex flex-col gap-3 border-b border-surface-200 pb-4 sm:flex-row sm:items-start sm:justify-between">
    <div class="min-w-0">
      <h1 class="break-words text-2xl font-semibold text-surface-950">{{ title }}</h1>
      <p v-if="subtitle" class="mt-1 max-w-4xl text-sm leading-6 text-surface-500">{{ subtitle }}</p>
    </div>

    <div class="flex w-full flex-wrap items-center gap-2 sm:w-auto sm:shrink-0 sm:justify-end">
      <slot name="actions" />
    </div>
  </header>
</template>
