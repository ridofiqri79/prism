<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    value: number
    unit?: string
    compact?: boolean
    maximumFractionDigits?: number
  }>(),
  {
    unit: '',
    compact: false,
    maximumFractionDigits: 0,
  },
)

const formattedValue = computed(() => {
  if (props.unit === 'USD') {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      notation: props.compact ? 'compact' : 'standard',
      maximumFractionDigits: props.maximumFractionDigits,
    }).format(props.value)
  }

  if (props.unit === 'percent') {
    return `${props.value.toFixed(props.maximumFractionDigits)}%`
  }

  return new Intl.NumberFormat('en-US', {
    notation: props.compact ? 'compact' : 'standard',
    maximumFractionDigits: props.maximumFractionDigits,
  }).format(props.value)
})
</script>

<template>
  <span>{{ formattedValue }}</span>
</template>
