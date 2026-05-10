<script setup lang="ts">
import Tag from 'primevue/tag'
import StatusBadge from '@/components/common/StatusBadge.vue'

defineProps<{
  /** Judul program, ditampilkan sebagai judul utama card */
  programTitleName: string
  /** Nilai status mentah dari DB (mis. 'active', 'superseded') */
  status: string
  /** Label status yang sudah diformat untuk ditampilkan ke user */
  statusLabel: string
  /** True jika snapshot ini adalah versi terbaru */
  isLatest?: boolean
  /** True jika ada revisi yang lebih baru dari snapshot ini */
  hasNewerRevision?: boolean
  /** Slot tambahan: konten informasi di bawah Judul Program (opsional) */
}>()
</script>

<template>
  <div class="rounded-lg border border-surface-200 bg-white p-5">
    <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
      <div class="min-w-0 space-y-1.5">
        <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Judul Program</p>
        <p class="truncate text-base font-semibold text-surface-950">{{ programTitleName }}</p>
        <!-- Slot untuk info tambahan di bawah Judul Program (mis. durasi) -->
        <slot />
      </div>
      <div class="flex flex-col gap-2 lg:items-end">
        <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Status Snapshot</p>
        <div class="flex flex-wrap items-center gap-2 lg:justify-end">
          <StatusBadge :status="status" :label="statusLabel" />
          <Tag v-if="isLatest" value="Versi terbaru" severity="success" rounded />
          <Tag
            v-else-if="hasNewerRevision"
            value="Ada revisi lebih baru"
            severity="warn"
            rounded
          />
        </div>
      </div>
    </div>
  </div>
</template>
