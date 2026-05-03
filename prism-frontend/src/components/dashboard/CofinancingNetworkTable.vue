<script setup lang="ts">
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Tag from 'primevue/tag'
import type { CofinancingItem } from '@/types/dashboard.types'

defineProps<{
  items: CofinancingItem[]
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
      <h2 class="text-lg font-semibold text-surface-950">Cofinancing Projects</h2>
      <p class="text-sm text-surface-500">Dihitung per proyek dengan lebih dari satu lender.</p>
    </div>
    <DataTable :value="items" :loading="loading" size="small" scrollable scroll-height="28rem">
      <Column field="project_name" header="Project">
        <template #body="{ data }">
          <div class="space-y-1">
            <p class="font-medium text-surface-900">{{ data.project_name }}</p>
            <p class="text-xs text-surface-500">{{ data.project_code || '-' }}</p>
          </div>
        </template>
      </Column>
      <Column field="reference_type" header="Source" class="w-28">
        <template #body="{ data }">
          <Tag :value="data.reference_type" severity="info" />
        </template>
      </Column>
      <Column field="lender_count" header="Lenders" class="w-24" />
      <Column field="lender_names" header="Lender Network">
        <template #body="{ data }">
          <div class="flex flex-wrap gap-2">
            <Tag v-for="name in data.lender_names" :key="name" :value="name" severity="secondary" />
          </div>
        </template>
      </Column>
      <Column field="amount_usd" header="Amount" class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.amount_usd) }}</template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada cofinancing project.</div>
      </template>
    </DataTable>
  </section>
</template>
