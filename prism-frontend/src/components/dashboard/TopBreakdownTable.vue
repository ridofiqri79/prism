<script setup lang="ts">
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import type { BreakdownItem } from '@/types/dashboard.types'

defineProps<{
  title: string
  items: BreakdownItem[]
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
      <h2 class="text-lg font-semibold text-surface-950">{{ title }}</h2>
      <p class="text-sm text-surface-500">10 teratas berdasarkan eksposur USD.</p>
    </div>
    <DataTable :value="items" size="small" scrollable scroll-height="28rem">
      <Column field="label" header="Nama" />
      <Column field="item_count" header="Proyek" class="w-24">
        <template #body="{ data }">
          {{ data.item_count ?? 0 }}
        </template>
      </Column>
      <Column field="amount_usd" header="Komitmen" class="w-40">
        <template #body="{ data }">
          {{ usdFormatter.format(data.amount_usd ?? 0) }}
        </template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada data untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
