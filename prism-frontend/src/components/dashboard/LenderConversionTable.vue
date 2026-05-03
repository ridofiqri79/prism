<script setup lang="ts">
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Tag from 'primevue/tag'
import type { LenderConversionItem } from '@/types/dashboard.types'

defineProps<{
  items: LenderConversionItem[]
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
      <h2 class="text-lg font-semibold text-surface-950">Lender Conversion</h2>
      <p class="text-sm text-surface-500">Conversion dari lender indication ke Loan Agreement.</p>
    </div>
    <DataTable :value="items" :loading="loading" size="small" scrollable scroll-height="28rem">
      <Column field="lender_name" header="Lender" frozen>
        <template #body="{ data }">
          <div class="space-y-1">
            <p class="font-medium text-surface-900">{{ data.lender_name }}</p>
            <Tag :value="data.lender_type" severity="secondary" />
          </div>
        </template>
      </Column>
      <Column field="indication_count" header="Indication" class="w-28" />
      <Column field="loi_count" header="LoI" class="w-20" />
      <Column field="gb_count" header="GB" class="w-20" />
      <Column field="dk_count" header="DK" class="w-20" />
      <Column field="la_count" header="LA" class="w-20" />
      <Column field="la_conversion_pct" header="Conversion" class="w-32">
        <template #body="{ data }"> {{ data.la_conversion_pct.toFixed(2) }}% </template>
      </Column>
      <Column field="indication_usd" header="Indication USD" class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.indication_usd) }}</template>
      </Column>
      <Column field="la_usd" header="LA USD" class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.la_usd) }}</template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada conversion untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
