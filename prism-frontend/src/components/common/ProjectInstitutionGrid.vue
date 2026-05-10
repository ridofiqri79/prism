<script setup lang="ts">
import ValueChipList from '@/components/common/ValueChipList.vue'

withDefaults(
  defineProps<{
    executingAgencies: string[]
    implementingAgencies: string[]
    bappenasPartners: string[]
    locations: string[]
    /** BB only — Prioritas Nasional (full-width row). Omit or pass [] to hide. */
    nationalPriorities?: string[]
    /** BB only — Durasi proyek dalam bulan. Pass null/undefined to hide. */
    duration?: number | null
  }>(),
  {
    nationalPriorities: () => [],
    duration: null,
  },
)
</script>

<template>
  <div class="grid gap-x-8 gap-y-5 p-5 md:grid-cols-2">
    <div class="min-w-0 space-y-2">
      <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
        Executing Agency
      </p>
      <ValueChipList :items="executingAgencies" />
    </div>
    <div class="min-w-0 space-y-2">
      <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
        Implementing Agency
      </p>
      <ValueChipList :items="implementingAgencies" />
    </div>
    <div class="min-w-0 space-y-2">
      <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
        Mitra Kerja Bappenas
      </p>
      <ValueChipList :items="bappenasPartners" />
    </div>
    <div class="min-w-0 space-y-2">
      <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Lokasi</p>
      <ValueChipList :items="locations" />
    </div>
    <div v-if="duration != null" class="min-w-0 space-y-2">
      <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Durasi</p>
      <p class="text-sm font-semibold text-surface-950">
        {{ duration ? `${duration} bulan` : '-' }}
      </p>
    </div>
    <div v-if="nationalPriorities?.length" class="min-w-0 space-y-2 md:col-span-2">
      <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
        Prioritas Nasional
      </p>
      <ValueChipList :items="nationalPriorities" />
    </div>
  </div>
</template>
