<script setup lang="ts">
import { computed } from 'vue'
import Tag from 'primevue/tag'

const props = defineProps<{
  label: string
  value: number
  unit?: 'USD' | 'project' | 'score'
  tone?: 'ready' | 'partial' | 'incomplete' | 'cofinancing' | 'neutral'
}>()

const formattedValue = computed(() => {
  if (props.unit === 'USD') {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(props.value)
  }
  return new Intl.NumberFormat('en-US', { maximumFractionDigits: 0 }).format(props.value)
})

const tagValue = computed(() => {
  if (props.unit === 'project') return 'project'
  if (props.unit === 'score') return 'score'
  if (props.unit === 'USD') return 'USD'
  return ''
})

const severity = computed(() => {
  if (props.tone === 'ready') return 'success'
  if (props.tone === 'partial') return 'warn'
  if (props.tone === 'incomplete') return 'danger'
  if (props.tone === 'cofinancing') return 'info'
  return 'secondary'
})
</script>

<template>
  <article class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="flex items-start justify-between gap-3">
      <p class="text-sm font-medium text-surface-600">{{ label }}</p>
      <Tag v-if="tagValue" :value="tagValue" :severity="severity" />
    </div>
    <p class="mt-3 break-words text-2xl font-semibold text-surface-950">
      {{ formattedValue }}
    </p>
  </article>
</template>
