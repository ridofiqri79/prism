<script setup lang="ts">
import { computed } from 'vue'
import Tag from 'primevue/tag'
import type { MetricCard } from '@/types/dashboard.types'

const props = defineProps<{
  card: MetricCard
}>()

const formattedValue = computed(() => {
  if (props.card.unit === 'USD') {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(props.card.value)
  }
  if (props.card.unit === 'percent') {
    return `${props.card.value.toFixed(2)}%`
  }
  return new Intl.NumberFormat('en-US', { maximumFractionDigits: 0 }).format(props.card.value)
})

const unitLabel = computed(() => {
  if (props.card.unit === 'project') return 'proyek'
  if (props.card.unit === 'USD' || props.card.unit === 'percent') return ''
  return props.card.unit ?? ''
})
</script>

<template>
  <article class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="flex items-start justify-between gap-3">
      <p class="text-sm font-medium text-surface-600">{{ card.label }}</p>
      <Tag v-if="card.category" :value="card.category" severity="secondary" />
    </div>
    <p class="mt-3 break-words text-2xl font-semibold text-surface-950">
      {{ formattedValue }}
    </p>
    <p v-if="unitLabel" class="mt-1 text-sm text-surface-500">{{ unitLabel }}</p>
  </article>
</template>
