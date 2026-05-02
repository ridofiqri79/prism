<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    value: string
    severity?: string
  }>(),
  {
    severity: 'secondary',
  },
)

const badgeClass = computed(() => {
  if (props.severity === 'warning') return 'border-amber-300 bg-amber-100 text-amber-900'
  if (props.severity === 'danger') return 'border-red-300 bg-red-100 text-red-800'
  if (props.severity === 'success') return 'border-emerald-300 bg-emerald-100 text-emerald-800'
  if (props.severity === 'info') return 'border-sky-300 bg-sky-100 text-sky-800'
  return 'border-surface-300 bg-surface-100 text-surface-700'
})
</script>

<template>
  <span
    class="inline-flex max-w-full items-center rounded-full border px-2 py-0.5 text-xs font-semibold leading-5"
    :class="badgeClass"
  >
    <span class="truncate">{{ value }}</span>
  </span>
</template>
