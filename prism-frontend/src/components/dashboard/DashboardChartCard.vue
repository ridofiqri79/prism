<script setup lang="ts">
import ProgressSpinner from 'primevue/progressspinner'
import EmptyInsightState from '@/components/dashboard/EmptyInsightState.vue'

withDefaults(
  defineProps<{
    title: string
    subtitle?: string
    loading?: boolean
    empty?: boolean
    emptyTitle?: string
    emptyMessage?: string
  }>(),
  {
    subtitle: '',
    loading: false,
    empty: false,
    emptyTitle: 'Tidak ada data',
    emptyMessage: '',
  },
)
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3 flex flex-col gap-1 md:flex-row md:items-end md:justify-between">
      <div>
        <h2 class="text-lg font-semibold text-surface-950">{{ title }}</h2>
        <p v-if="subtitle" class="text-sm text-surface-500">{{ subtitle }}</p>
      </div>
      <slot name="actions" />
    </div>

    <div v-if="loading" class="flex min-h-56 items-center justify-center">
      <ProgressSpinner class="h-10 w-10" stroke-width="4" />
    </div>
    <EmptyInsightState v-else-if="empty" :title="emptyTitle" :message="emptyMessage" />
    <slot v-else />
  </section>
</template>
