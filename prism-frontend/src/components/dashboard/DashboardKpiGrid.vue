<script setup lang="ts">
import DashboardKpiCard from '@/components/dashboard/DashboardKpiCard.vue'
import type { DashboardKpi } from '@/types/dashboard-flow.types'

defineProps<{
  items: DashboardKpi[]
}>()
</script>

<template>
  <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
    <DashboardKpiCard v-for="item in items" :key="item.label" :item="item" />
  </div>
</template>
