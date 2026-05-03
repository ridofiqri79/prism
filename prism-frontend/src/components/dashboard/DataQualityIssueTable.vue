<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Tag from 'primevue/tag'
import IssueSeverityBadge from '@/components/dashboard/IssueSeverityBadge.vue'
import type { DataQualityIssueItem } from '@/types/dashboard.types'

defineProps<{
  items: DataQualityIssueItem[]
  loading?: boolean
}>()

const emit = defineEmits<{
  open: [item: DataQualityIssueItem]
}>()
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3 flex flex-col gap-1 md:flex-row md:items-end md:justify-between">
      <div>
        <h2 class="text-lg font-semibold text-surface-950">Data Quality Issues</h2>
        <p class="text-sm text-surface-500">Business-rule consistency, integrity, and monitoring compliance.</p>
      </div>
      <p class="text-sm text-surface-500">{{ items.length }} issue</p>
    </div>

    <DataTable :value="items" :loading="loading" size="small" scrollable scroll-height="34rem">
      <Column field="severity" header="Severity" sortable class="w-32">
        <template #body="{ data }">
          <IssueSeverityBadge :severity="data.severity" />
        </template>
      </Column>

      <Column field="record_label" header="Record" sortable class="min-w-80">
        <template #body="{ data }">
          <div class="space-y-1">
            <button
              type="button"
              class="text-left font-medium text-primary hover:underline"
              @click="emit('open', data)"
            >
              {{ data.record_label }}
            </button>
            <p class="text-xs text-surface-500">{{ data.module }} - {{ data.record_id }}</p>
          </div>
        </template>
      </Column>

      <Column field="issue_type" header="Issue Type" sortable class="min-w-72">
        <template #body="{ data }">
          <Tag :value="data.issue_type" severity="secondary" />
        </template>
      </Column>

      <Column field="message" header="Issue" class="min-w-96" />
      <Column field="recommended_action" header="Recommended Action" class="min-w-96" />

      <Column field="is_resolved" header="Status" sortable class="w-32">
        <template #body="{ data }">
          <Tag
            :value="data.is_resolved ? 'Resolved' : 'Open'"
            :severity="data.is_resolved ? 'success' : 'warn'"
          />
        </template>
      </Column>

      <Column header="Open" class="w-24">
        <template #body="{ data }">
          <Button icon="pi pi-arrow-right" text rounded aria-label="Open record" @click="emit('open', data)" />
        </template>
      </Column>

      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada issue untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
