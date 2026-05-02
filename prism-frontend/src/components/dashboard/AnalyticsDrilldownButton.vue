<script setup lang="ts">
import Button from 'primevue/button'
import type { DashboardDrilldownQuery } from '@/types/dashboard.types'

withDefaults(
  defineProps<{
    drilldown?: DashboardDrilldownQuery
    label?: string
    disabled?: boolean
  }>(),
  {
    drilldown: undefined,
    label: 'Buka',
    disabled: false,
  },
)

const emit = defineEmits<{
  open: [DashboardDrilldownQuery]
}>()
</script>

<template>
  <Button
    :label="label"
    icon="pi pi-arrow-up-right"
    size="small"
    severity="secondary"
    outlined
    :disabled="disabled || !drilldown"
    @click="drilldown && emit('open', drilldown)"
  />
</template>
