<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    description?: string
    compact?: boolean
  }>(),
  {
    title: 'Belum ada data',
    description: undefined,
    compact: false,
  },
)
</script>

<template>
  <div
    :class="[
      'flex flex-col items-center justify-center rounded-lg border border-dashed border-surface-300 bg-surface-50 text-center',
      compact ? 'min-h-32 p-5' : 'min-h-48 p-8',
    ]"
  >
    <i :class="['pi pi-inbox text-surface-400', compact ? 'text-3xl' : 'text-4xl']" />
    <h2 :class="['text-base font-semibold text-surface-800', compact ? 'mt-3' : 'mt-4']">{{ title }}</h2>
    <p v-if="description" class="mt-1 max-w-md text-sm text-surface-500">{{ description }}</p>
  </div>
</template>
