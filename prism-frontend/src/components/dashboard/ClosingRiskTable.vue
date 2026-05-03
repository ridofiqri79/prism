<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import RiskBadge from '@/components/dashboard/RiskBadge.vue'
import type { LAClosingRiskItem } from '@/types/dashboard.types'

defineProps<{
  items: LAClosingRiskItem[]
  loading?: boolean
}>()

const emit = defineEmits<{
  open: [id: string]
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
      <h2 class="text-lg font-semibold text-surface-950">Closing Risks</h2>
      <p class="text-sm text-surface-500">Loan Agreement yang mendekati closing date.</p>
    </div>
    <DataTable :value="items" :loading="loading" size="small" scrollable scroll-height="28rem">
      <Column field="loan_code" header="Loan Code" frozen class="min-w-40">
        <template #body="{ data }">
          <Button
            :label="data.loan_code"
            link
            class="p-0 text-left font-medium"
            @click="emit('open', data.loan_agreement_id)"
          />
        </template>
      </Column>
      <Column field="project_name" header="Project" class="min-w-80" />
      <Column field="lender_name" header="Lender" class="min-w-40" />
      <Column field="closing_date" header="Closing Date" class="w-36" />
      <Column field="days_until_closing" header="Days" sortable class="w-24" />
      <Column field="risk_level" header="Risk" class="w-28">
        <template #body="{ data }">
          <RiskBadge :level="data.risk_level" />
        </template>
      </Column>
      <Column field="commitment_usd" header="Commitment" sortable class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.commitment_usd) }}</template>
      </Column>
      <Column field="undisbursed_usd" header="Undisbursed" sortable class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.undisbursed_usd) }}</template>
      </Column>
      <Column field="la_absorption_pct" header="LA Absorption" sortable class="w-36">
        <template #body="{ data }">{{ data.la_absorption_pct.toFixed(2) }}%</template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada closing risk untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
