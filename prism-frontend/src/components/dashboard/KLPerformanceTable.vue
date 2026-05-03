<script setup lang="ts">
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import KLRiskBadge from '@/components/dashboard/KLRiskBadge.vue'
import type { KLPortfolioPerformanceItem } from '@/types/dashboard.types'

defineProps<{
  items: KLPortfolioPerformanceItem[]
  loading?: boolean
}>()

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3">
      <h2 class="text-lg font-semibold text-surface-950">K/L Ranking</h2>
      <p class="text-sm text-surface-500">Project count per stage, exposure, absorption, and risk by institution.</p>
    </div>
    <DataTable :value="items" :loading="loading" size="small" scrollable scroll-height="32rem">
      <Column field="institution_name" header="K/L" frozen sortable class="min-w-64">
        <template #body="{ data }">
          <div class="space-y-1">
            <p class="font-medium text-surface-900">{{ data.institution_name }}</p>
            <p class="text-xs text-surface-500">Score {{ data.performance_score.toFixed(1) }}</p>
          </div>
        </template>
      </Column>
      <Column field="bb_project_count" header="BB" sortable class="w-20" />
      <Column field="gb_project_count" header="GB" sortable class="w-20" />
      <Column field="dk_project_count" header="DK" sortable class="w-20" />
      <Column field="la_count" header="LA" sortable class="w-20" />
      <Column field="pipeline_usd" header="Pipeline" sortable class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.pipeline_usd) }}</template>
      </Column>
      <Column field="la_commitment_usd" header="LA Commitment" sortable class="w-44">
        <template #body="{ data }">{{ usdFormatter.format(data.la_commitment_usd) }}</template>
      </Column>
      <Column field="planned_usd" header="Planned" sortable class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.planned_usd) }}</template>
      </Column>
      <Column field="realized_usd" header="Realized" sortable class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.realized_usd) }}</template>
      </Column>
      <Column field="absorption_pct" header="Absorption" sortable class="w-32">
        <template #body="{ data }">{{ data.absorption_pct.toFixed(2) }}%</template>
      </Column>
      <Column field="risk_count" header="Risk" sortable class="w-28">
        <template #body="{ data }">
          <KLRiskBadge :risk-count="data.risk_count" :category="data.performance_category" />
        </template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada data K/L untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
