<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Tag from 'primevue/tag'
import RiskBadge from '@/components/dashboard/RiskBadge.vue'
import type { LAUnderDisbursementRiskItem } from '@/types/dashboard.types'

defineProps<{
  items: LAUnderDisbursementRiskItem[]
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
      <h2 class="text-lg font-semibold text-surface-950">Under-Disbursement Risks</h2>
      <p class="text-sm text-surface-500">Effective LA dengan elapsed time lebih tinggi dari serapan.</p>
    </div>
    <DataTable :value="items" :loading="loading" size="small" scrollable scroll-height="30rem">
      <Column field="loan_code" header="Loan Code" frozen class="min-w-40">
        <template #body="{ data }">
          <div class="space-y-1">
            <Button
              :label="data.loan_code"
              link
              class="p-0 text-left font-medium"
              @click="emit('open', data.loan_agreement_id)"
            />
            <Tag v-if="data.is_extended" value="extended" severity="secondary" />
          </div>
        </template>
      </Column>
      <Column field="project_name" header="Project" class="min-w-80" />
      <Column field="lender_name" header="Lender" class="min-w-40" />
      <Column field="risk_level" header="Risk" class="w-28">
        <template #body="{ data }">
          <RiskBadge :level="data.risk_level" />
        </template>
      </Column>
      <Column field="risk_type" header="Type" class="min-w-56" />
      <Column field="time_elapsed_pct" header="Elapsed" sortable class="w-28">
        <template #body="{ data }">{{ data.time_elapsed_pct.toFixed(2) }}%</template>
      </Column>
      <Column field="la_absorption_pct" header="Absorption" sortable class="w-32">
        <template #body="{ data }">{{ data.la_absorption_pct.toFixed(2) }}%</template>
      </Column>
      <Column field="absorption_gap_pct" header="Gap" sortable class="w-28">
        <template #body="{ data }">{{ data.absorption_gap_pct.toFixed(2) }}%</template>
      </Column>
      <Column field="undisbursed_usd" header="Undisbursed" sortable class="w-40">
        <template #body="{ data }">{{ usdFormatter.format(data.undisbursed_usd) }}</template>
      </Column>
      <Column field="required_monthly_disbursement_usd" header="Required Monthly" sortable class="w-44">
        <template #body="{ data }">
          {{ usdFormatter.format(data.required_monthly_disbursement_usd) }}
        </template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">
          Tidak ada under-disbursement risk untuk filter ini.
        </div>
      </template>
    </DataTable>
  </section>
</template>
