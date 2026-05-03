<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    label: string
    value: number | string
    unit?: string
    format?: 'number' | 'currency' | 'percent'
    compact?: boolean
  }>(),
  {
    unit: '',
    format: 'number',
    compact: true,
  },
)

const formattedValue = computed(() => {
  if (typeof props.value === 'string') {
    return props.value
  }

  if (props.format === 'currency') {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      notation: props.compact ? 'compact' : 'standard',
      maximumFractionDigits: 2,
    }).format(props.value)
  }

  if (props.format === 'percent') {
    return `${props.value.toFixed(1)}%`
  }

  return new Intl.NumberFormat('id-ID').format(props.value)
})
</script>

<template>
  <article class="rounded-lg border border-surface-200 bg-white p-5 shadow-sm">
    <p class="text-sm text-surface-500">{{ label }}</p>
    <p class="mt-1 text-2xl font-semibold text-surface-950">{{ formattedValue }}</p>
    <p v-if="unit" class="mt-0.5 text-xs text-surface-500">{{ unit }}</p>
  </article>
</template>
