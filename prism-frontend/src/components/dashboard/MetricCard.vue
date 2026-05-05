<script setup lang="ts">
import { computed } from 'vue'
import Tag from 'primevue/tag'
import AmountDisplay from '@/components/dashboard/AmountDisplay.vue'
import type { MetricCard } from '@/types/dashboard.types'

const props = defineProps<{
  card: MetricCard
}>()

const categoryLabel = computed(() => {
  const labels: Record<string, string> = {
    pipeline: 'Alur',
    commitment: 'Komitmen',
    project: 'Proyek',
  }

  return props.card.category ? (labels[props.card.category] ?? props.card.category) : ''
})
</script>

<template>
  <article class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="flex items-start justify-between gap-3">
      <p class="text-sm font-medium text-surface-600">{{ card.label }}</p>
      <Tag v-if="categoryLabel" :value="categoryLabel" severity="secondary" />
    </div>
    <p class="mt-3 break-words text-2xl font-semibold text-surface-950">
      <AmountDisplay
        :value="card.value"
        :unit="card.unit"
        :maximum-fraction-digits="card.unit === 'percent' ? 2 : 0"
      />
    </p>
    <p v-if="card.unit && !['USD', 'percent'].includes(card.unit)" class="mt-1 text-sm text-surface-500">
      {{ card.unit === 'project' ? 'proyek' : card.unit }}
    </p>
    <p v-if="card.hint" class="mt-2 text-xs leading-5 text-surface-500">
      {{ card.hint }}
    </p>
  </article>
</template>
