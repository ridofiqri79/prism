<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import ProgressBar from 'primevue/progressbar'
import Tag from 'primevue/tag'
import type { GreenBookReadinessItem } from '@/types/dashboard.types'

defineProps<{
  items: GreenBookReadinessItem[]
  loading: boolean
}>()

const router = useRouter()

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

function statusTone(status: string) {
  if (status === 'READY') return 'success'
  if (status === 'PARTIAL') return 'warn'
  return 'danger'
}
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3 flex flex-col gap-1 md:flex-row md:items-end md:justify-between">
      <div>
        <h2 class="text-lg font-semibold text-surface-950">Readiness Worklist</h2>
        <p class="text-sm text-surface-500">Skor readiness dan field yang masih perlu dilengkapi.</p>
      </div>
      <p class="text-sm text-surface-500">{{ items.length }} project</p>
    </div>

    <DataTable :value="items" :loading="loading" size="small" scrollable scroll-height="34rem">
      <Column field="project_name" header="Project" sortable>
        <template #body="{ data }">
          <div class="min-w-72">
            <p class="font-medium text-surface-950">{{ data.project_name }}</p>
            <p class="mt-1 text-xs text-surface-500">{{ data.gb_code }} - GB {{ data.publish_year }}</p>
          </div>
        </template>
      </Column>

      <Column field="readiness_score" header="Score" sortable class="w-52">
        <template #body="{ data }">
          <div class="space-y-2">
            <div class="flex items-center justify-between gap-3">
              <Tag :value="data.readiness_status" :severity="statusTone(data.readiness_status)" />
              <span class="text-sm font-semibold text-surface-900">{{ data.readiness_score }}/100</span>
            </div>
            <ProgressBar :value="data.readiness_score" :show-value="false" class="h-2" />
          </div>
        </template>
      </Column>

      <Column field="is_cofinancing" header="Cofinancing" sortable class="w-36">
        <template #body="{ data }">
          <Tag
            :value="data.is_cofinancing ? 'Yes' : 'No'"
            :severity="data.is_cofinancing ? 'info' : 'secondary'"
          />
        </template>
      </Column>

      <Column field="total_funding_usd" header="Funding" sortable class="w-44">
        <template #body="{ data }">
          {{ usdFormatter.format(data.total_funding_usd ?? 0) }}
        </template>
      </Column>

      <Column field="institution_name" header="K/L" class="w-64">
        <template #body="{ data }">
          {{ data.institution_name || '-' }}
        </template>
      </Column>

      <Column header="Lender" class="w-56">
        <template #body="{ data }">
          <div class="flex flex-wrap gap-1">
            <Tag
              v-for="lender in data.lender_names"
              :key="lender"
              :value="lender"
              severity="info"
            />
            <span v-if="!data.lender_names.length" class="text-sm text-surface-400">-</span>
          </div>
        </template>
      </Column>

      <Column header="Missing fields" class="w-96">
        <template #body="{ data }">
          <div class="flex flex-wrap gap-1">
            <Tag
              v-for="field in data.missing_fields"
              :key="field"
              :value="field"
              severity="danger"
            />
            <Tag v-if="!data.missing_fields.length" value="Complete" severity="success" />
          </div>
        </template>
      </Column>

      <Column header="Open" class="w-24">
        <template #body="{ data }">
          <RouterLink
            v-if="router.hasRoute('gb-project-detail')"
            :to="{ name: 'gb-project-detail', params: { gbId: data.green_book_id, id: data.project_id } }"
          >
            <Button icon="pi pi-arrow-right" text rounded aria-label="Open Green Book project" />
          </RouterLink>
          <Button v-else icon="pi pi-minus" text rounded disabled aria-label="No route" />
        </template>
      </Column>

      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada item readiness untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
