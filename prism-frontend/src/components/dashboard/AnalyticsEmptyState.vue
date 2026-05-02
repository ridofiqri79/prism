<script setup lang="ts">
import EmptyState from '@/components/common/EmptyState.vue'

withDefaults(
  defineProps<{
    title?: string
    description?: string
    compact?: boolean
  }>(),
  {
    title: 'Tidak ada data analytics',
    description: 'Data kosong untuk filter aktif.',
    compact: true,
  },
)
</script>

<template>
  <EmptyState :title="title" :description="description" :compact="compact">
    <slot />
  </EmptyState>
</template>
