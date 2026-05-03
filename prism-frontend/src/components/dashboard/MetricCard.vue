<script setup lang="ts">
import Tag from 'primevue/tag'
import AmountDisplay from '@/components/dashboard/AmountDisplay.vue'
import type { MetricCard } from '@/types/dashboard.types'

defineProps<{
  card: MetricCard
}>()
</script>

<template>
  <article class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="flex items-start justify-between gap-3">
      <p class="text-sm font-medium text-surface-600">{{ card.label }}</p>
      <Tag v-if="card.category" :value="card.category" severity="secondary" />
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
  </article>
</template>
