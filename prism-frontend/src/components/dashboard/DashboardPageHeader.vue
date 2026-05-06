<script setup lang="ts">
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import type { DashboardHeaderAction } from '@/types/dashboard-flow.types'

withDefaults(
  defineProps<{
    eyebrow: string
    title: string
    subtitle: string
    statusLabel?: string
    actions?: DashboardHeaderAction[]
  }>(),
  {
    statusLabel: undefined,
    actions: () => [],
  },
)

const emit = defineEmits<{
  action: [key: string]
}>()
</script>

<template>
  <header class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
    <div class="min-w-0">
      <p class="text-xs font-semibold uppercase text-surface-400">{{ eyebrow }}</p>
      <h1 class="mt-2 text-2xl font-semibold leading-tight text-surface-950">
        {{ title }}
      </h1>
      <p class="mt-1 max-w-3xl text-sm leading-6 text-surface-600">
        {{ subtitle }}
      </p>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <Tag v-if="statusLabel" severity="success" rounded>
        <span class="inline-flex items-center gap-2">
          <span class="h-1.5 w-1.5 rounded-full bg-prism-green-dark" />
          {{ statusLabel }}
        </span>
      </Tag>
      <Button
        v-for="action in actions"
        :key="action.key"
        :label="action.label"
        :icon="action.icon"
        :severity="action.severity"
        :outlined="action.outlined"
        size="small"
        @click="emit('action', action.key)"
      />
    </div>
  </header>
</template>
