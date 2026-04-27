<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    amount: number
    currency?: string
    compact?: boolean
  }>(),
  {
    currency: 'USD',
    compact: false,
  },
)

const formattedAmount = computed(() => {
  const formatter = new Intl.NumberFormat('en-US', {
    minimumFractionDigits: props.compact ? 0 : 2,
    maximumFractionDigits: props.compact ? 2 : 2,
    notation: props.compact ? 'compact' : 'standard',
    compactDisplay: 'short',
  })

  return `${props.currency} ${formatter.format(props.amount)}`
})
</script>

<template>
  <span class="tabular-nums">{{ formattedAmount }}</span>
</template>
