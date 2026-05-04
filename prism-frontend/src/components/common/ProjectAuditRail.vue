<script setup lang="ts">
import { ref } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import type { ProjectAuditEntry } from '@/types/audit.types'

interface ProjectAuditRailItem extends ProjectAuditEntry {
  snapshot_label: string
}

defineProps<{
  items: ProjectAuditRailItem[]
}>()

const isAuditTrailOpen = ref(false)

function actionSeverity(action: string) {
  if (action === 'INSERT') return 'success'
  if (action === 'DELETE') return 'danger'
  return 'info'
}

function formatDateTime(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}
</script>

<template>
  <aside v-if="items.length" class="rounded-lg border border-surface-200 bg-white p-5">
    <div class="flex items-center justify-between gap-3">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-lg font-semibold text-surface-950">Audit Trail</h2>
        <Tag :value="`${items.length} event`" severity="secondary" rounded />
      </div>
      <Button
        :label="isAuditTrailOpen ? 'Tutup' : 'Detail'"
        :icon="isAuditTrailOpen ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
        severity="secondary"
        size="small"
        outlined
        @click="isAuditTrailOpen = !isAuditTrailOpen"
      />
    </div>

    <ol v-if="isAuditTrailOpen" class="relative mt-4 space-y-4 border-l border-surface-200 pl-4">
      <li v-for="item in items" :key="item.id" class="relative">
        <span class="absolute -left-[1.35rem] top-1.5 h-2.5 w-2.5 rounded-full bg-primary" />
        <div class="space-y-2">
          <div class="flex flex-wrap items-center gap-2">
            <Tag :value="item.action_label" :severity="actionSeverity(item.action)" rounded />
            <span class="text-xs text-surface-500">{{ formatDateTime(item.changed_at) }}</span>
          </div>
          <div>
            <p class="text-sm font-semibold text-surface-950">{{ item.summary }}</p>
            <p class="text-xs text-surface-500">
              {{ item.snapshot_label }} oleh {{ item.changed_by_username }}
            </p>
          </div>
          <div
            v-if="(item.changed_field_labels?.length ?? 0) > 0"
            class="flex flex-wrap gap-1.5"
          >
            <span
              v-for="field in item.changed_field_labels ?? []"
              :key="`${item.id}-${field}`"
              class="rounded-md bg-surface-100 px-2 py-1 text-xs font-medium text-surface-700"
            >
              {{ field }}
            </span>
          </div>
          <p v-else class="text-xs text-surface-500">Tidak ada field yang berubah.</p>
        </div>
      </li>
    </ol>
  </aside>
</template>
