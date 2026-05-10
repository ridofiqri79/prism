<script setup lang="ts">
import { ref } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import StatusBadge from '@/components/common/StatusBadge.vue'

export interface RevisionHistoryItem {
  id: string
  /** Label yang sudah diformat, misal: "BB 2025-2029 Revisi ke-1 - Rev 1 / 2026" */
  label: string
  /** Kode proyek, misal bb_code atau gb_code */
  code: string
  book_status: string
  /** Label status yang sudah diformat, misal "Berlaku" */
  status_label: string
  is_latest: boolean
  route: RouteLocationRaw
}

defineProps<{
  items: RevisionHistoryItem[]
}>()

const isOpen = ref(false)
</script>

<template>
  <section v-if="items.length" class="space-y-3 rounded-lg border border-surface-200 bg-white p-5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-wrap items-center gap-2">
        <h2 class="text-lg font-semibold text-surface-950">Histori Revisi</h2>
        <Tag :value="`${items.length} snapshot`" severity="secondary" rounded />
      </div>
      <Button
        :label="isOpen ? 'Tutup' : 'Detail'"
        :icon="isOpen ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
        severity="secondary"
        size="small"
        outlined
        @click="isOpen = !isOpen"
      />
    </div>

    <ol
      v-if="isOpen"
      class="mt-4 overflow-hidden rounded-lg border border-surface-200 bg-surface-0"
    >
      <li
        v-for="item in items"
        :key="item.id"
        class="flex items-center justify-between gap-3 border-b border-surface-100 px-4 py-3 last:border-b-0"
      >
        <div class="flex min-w-0 flex-wrap items-center gap-2">
          <p class="min-w-0 text-sm font-semibold text-surface-950">{{ item.label }}</p>
          <span
            class="rounded border border-surface-200 bg-surface-50 px-2 py-0.5 font-mono text-xs font-semibold text-surface-700"
          >
            {{ item.code }}
          </span>
          <StatusBadge :status="item.book_status" :label="item.status_label" />
          <Tag
            :value="item.is_latest ? 'Versi terbaru' : 'Historis'"
            :severity="item.is_latest ? 'success' : 'secondary'"
            rounded
          />
        </div>
        <Button
          v-tooltip.top="'Lihat detail snapshot'"
          as="router-link"
          :to="item.route"
          icon="pi pi-eye"
          severity="secondary"
          size="small"
          outlined
          rounded
          aria-label="Lihat detail snapshot"
        />
      </li>
    </ol>
  </section>
</template>
