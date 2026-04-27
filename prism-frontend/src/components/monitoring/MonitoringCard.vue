<script setup lang="ts">
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import AbsorptionBar from '@/components/monitoring/AbsorptionBar.vue'
import type { MonitoringDisbursement } from '@/types/monitoring.types'

defineProps<{
  monitoring: MonitoringDisbursement
}>()
</script>

<template>
  <article class="space-y-4 rounded-lg border border-surface-200 bg-white p-4">
    <div class="flex items-center justify-between gap-3">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">{{ monitoring.budget_year }}</p>
        <h3 class="font-semibold text-surface-950">Monitoring {{ monitoring.quarter }}</h3>
      </div>
      <StatusBadge :status="monitoring.quarter" />
    </div>

    <dl class="grid gap-3 text-sm">
      <div class="flex items-center justify-between gap-3">
        <dt class="text-surface-500">Planned USD</dt>
        <dd class="font-semibold text-surface-900">
          <CurrencyDisplay :amount="monitoring.planned_usd" currency="USD" />
        </dd>
      </div>
      <div class="flex items-center justify-between gap-3">
        <dt class="text-surface-500">Realized USD</dt>
        <dd class="font-semibold text-surface-900">
          <CurrencyDisplay :amount="monitoring.realized_usd" currency="USD" />
        </dd>
      </div>
    </dl>

    <AbsorptionBar :pct="monitoring.absorption_pct" />
  </article>
</template>
