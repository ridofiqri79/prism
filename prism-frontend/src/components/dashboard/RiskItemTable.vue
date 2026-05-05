<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Tag from 'primevue/tag'
import type { RiskItem } from '@/types/dashboard.types'

defineProps<{
  items: RiskItem[]
}>()

const router = useRouter()
const hasJourneyRoute = computed(() => router.hasRoute('project-journey'))

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

function severityLabel(severity: string) {
  if (severity === 'high') return 'High'
  if (severity === 'medium') return 'Medium'
  if (severity === 'low') return 'Low'
  return severity
}

function severityTone(severity: string) {
  if (severity === 'high') return 'danger'
  if (severity === 'medium') return 'warn'
  return 'secondary'
}
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3">
      <h2 class="text-lg font-semibold text-surface-950">Risk Items</h2>
      <p class="text-sm text-surface-500">Item prioritas untuk closing dan kelanjutan pipeline.</p>
    </div>
    <DataTable :value="items" size="small" scrollable scroll-height="32rem">
      <Column field="severity" header="Severity" class="w-28">
        <template #body="{ data }">
          <Tag :value="severityLabel(data.severity)" :severity="severityTone(data.severity)" />
        </template>
      </Column>
      <Column field="title" header="Project">
        <template #body="{ data }">
          <div class="min-w-0">
            <p class="font-medium text-surface-950">{{ data.title }}</p>
            <p class="mt-1 text-xs text-surface-500">{{ data.description }}</p>
          </div>
        </template>
      </Column>
      <Column field="code" header="Code" class="w-36" />
      <Column field="amount_usd" header="Amount" class="w-40">
        <template #body="{ data }">
          {{ usdFormatter.format(data.amount_usd ?? 0) }}
        </template>
      </Column>
      <Column header="Open" class="w-28">
        <template #body="{ data }">
          <RouterLink
            v-if="data.journey_bb_project_id && hasJourneyRoute"
            :to="{ name: 'project-journey', params: { bbProjectId: data.journey_bb_project_id } }"
          >
            <Button icon="pi pi-arrow-right" text rounded aria-label="Open journey" />
          </RouterLink>
          <RouterLink
            v-else-if="data.reference_type === 'loan_agreement' && data.reference_id"
            :to="{ name: 'loan-agreement-detail', params: { id: data.reference_id } }"
          >
            <Button icon="pi pi-arrow-right" text rounded aria-label="Open detail" />
          </RouterLink>
          <Button v-else icon="pi pi-minus" text rounded disabled aria-label="No route" />
        </template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada risk item untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
