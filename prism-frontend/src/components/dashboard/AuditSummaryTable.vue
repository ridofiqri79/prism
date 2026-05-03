<script setup lang="ts">
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Tag from 'primevue/tag'
import type { AuditRecentActivityItem, AuditSummaryItem } from '@/types/dashboard.types'

defineProps<{
  byUser: AuditSummaryItem[]
  byTable: AuditSummaryItem[]
  recentActivity: AuditRecentActivityItem[]
  loading?: boolean
}>()

function formatDate(value?: string) {
  if (!value) return '-'
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}
</script>

<template>
  <section class="space-y-4">
    <div class="grid gap-4 lg:grid-cols-2">
      <section class="rounded-lg border border-surface-200 bg-white p-4">
        <div class="mb-3">
          <h2 class="text-lg font-semibold text-surface-950">Audit by User</h2>
          <p class="text-sm text-surface-500">Ringkasan aktivitas perubahan per user.</p>
        </div>
        <DataTable :value="byUser" :loading="loading" size="small">
          <Column field="label" header="User" />
          <Column field="event_count" header="Events" sortable class="w-32" />
          <Column field="last_changed_at" header="Last Change" class="w-48">
            <template #body="{ data }">{{ formatDate(data.last_changed_at) }}</template>
          </Column>
          <template #empty>
            <div class="py-6 text-center text-sm text-surface-500">Tidak ada audit event.</div>
          </template>
        </DataTable>
      </section>

      <section class="rounded-lg border border-surface-200 bg-white p-4">
        <div class="mb-3">
          <h2 class="text-lg font-semibold text-surface-950">Audit by Table</h2>
          <p class="text-sm text-surface-500">Ringkasan aktivitas perubahan per tabel.</p>
        </div>
        <DataTable :value="byTable" :loading="loading" size="small">
          <Column field="label" header="Table" />
          <Column field="event_count" header="Events" sortable class="w-32" />
          <Column field="last_changed_at" header="Last Change" class="w-48">
            <template #body="{ data }">{{ formatDate(data.last_changed_at) }}</template>
          </Column>
          <template #empty>
            <div class="py-6 text-center text-sm text-surface-500">Tidak ada audit event.</div>
          </template>
        </DataTable>
      </section>
    </div>

    <section class="rounded-lg border border-surface-200 bg-white p-4">
      <div class="mb-3">
        <h2 class="text-lg font-semibold text-surface-950">Recent Audit Activity</h2>
        <p class="text-sm text-surface-500">Ringkasan aktivitas terbaru.</p>
      </div>
      <DataTable :value="recentActivity" :loading="loading" size="small" scrollable scroll-height="24rem">
        <Column field="changed_at" header="Changed At" sortable class="w-48">
          <template #body="{ data }">{{ formatDate(data.changed_at) }}</template>
        </Column>
        <Column field="username" header="User" class="w-44" />
        <Column field="action" header="Action" class="w-32">
          <template #body="{ data }">
            <Tag :value="data.action" severity="secondary" />
          </template>
        </Column>
        <Column field="table_name" header="Table" class="w-52" />
        <Column field="record_id" header="Record ID" class="min-w-72" />
        <template #empty>
          <div class="py-6 text-center text-sm text-surface-500">Tidak ada audit activity terbaru.</div>
        </template>
      </DataTable>
    </section>
  </section>
</template>
