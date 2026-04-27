<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    description?: string
  }>(),
  {
    title: 'Belum ada data',
    description: undefined,
  },
)
</script>

<template>
  <div class="flex min-h-48 flex-col items-center justify-center rounded-lg border border-dashed border-surface-300 bg-surface-50 p-8 text-center">
    <i class="pi pi-inbox text-4xl text-surface-400" />
    <h2 class="mt-4 text-base font-semibold text-surface-800">{{ title }}</h2>
    <p v-if="description" class="mt-1 max-w-md text-sm text-surface-500">{{ description }}</p>
  </div>
</template>
