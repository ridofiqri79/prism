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
  if (severity === 'high') return 'Tinggi'
  if (severity === 'medium') return 'Sedang'
  if (severity === 'low') return 'Rendah'
  return severity
}

function severityTone(severity: string) {
  if (severity === 'high') return 'danger'
  if (severity === 'medium') return 'warn'
  if (severity === 'low') return 'success'
  return 'secondary'
}
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3">
      <h2 class="text-lg font-semibold text-surface-950">Daftar Risiko</h2>
      <p class="text-sm text-surface-500">Item prioritas untuk penutupan dan kelanjutan pipeline.</p>
    </div>
    <DataTable :value="items" size="small" scrollable scroll-height="32rem">
      <Column field="severity" header="Tingkat Risiko" class="w-36">
        <template #body="{ data }">
          <Tag :value="severityLabel(data.severity)" :severity="severityTone(data.severity)" />
        </template>
      </Column>
      <Column field="title" header="Proyek">
        <template #body="{ data }">
          <div class="min-w-0">
            <p class="font-medium text-surface-950">{{ data.title }}</p>
            <p class="mt-1 text-xs text-surface-500">{{ data.description }}</p>
          </div>
        </template>
      </Column>
      <Column field="code" header="Kode" class="w-36" />
      <Column field="amount_usd" header="Nilai" class="w-40">
        <template #body="{ data }">
          {{ usdFormatter.format(data.amount_usd ?? 0) }}
        </template>
      </Column>
      <Column header="Buka" class="w-28">
        <template #body="{ data }">
          <RouterLink
            v-if="data.journey_bb_project_id && hasJourneyRoute"
            :to="{ name: 'project-journey', params: { bbProjectId: data.journey_bb_project_id } }"
          >
            <Button icon="pi pi-arrow-right" text rounded aria-label="Buka journey" />
          </RouterLink>
          <RouterLink
            v-else-if="data.reference_type === 'loan_agreement' && data.reference_id"
            :to="{ name: 'loan-agreement-detail', params: { id: data.reference_id } }"
          >
            <Button icon="pi pi-arrow-right" text rounded aria-label="Buka detail" />
          </RouterLink>
          <Button v-else icon="pi pi-minus" text rounded disabled aria-label="Tidak ada rute" />
        </template>
      </Column>
      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada risiko untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
