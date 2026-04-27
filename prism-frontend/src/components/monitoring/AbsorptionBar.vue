<script setup lang="ts">
import { computed } from 'vue'
import ProgressBar from 'primevue/progressbar'

const props = defineProps<{
  pct: number
}>()

const safePct = computed(() => {
  if (!Number.isFinite(props.pct)) return 0
  return Math.max(0, props.pct)
})

const normalizedPct = computed(() => Math.min(100, safePct.value))
const label = computed(() => `${safePct.value.toFixed(1)}%`)
const tone = computed(() => {
  if (safePct.value < 50) return 'danger'
  if (safePct.value < 80) return 'warn'
  return 'success'
})

const valueClass = computed(() => ({
  'bg-red-500': tone.value === 'danger',
  'bg-amber-500': tone.value === 'warn',
  'bg-emerald-500': tone.value === 'success',
}))
</script>

<template>
  <div class="space-y-1">
    <ProgressBar
      :value="normalizedPct"
      :show-value="false"
      class="h-3"
      :pt="{ value: { class: valueClass } }"
    />
    <span
      class="text-xs font-semibold"
      :class="{
        'text-red-700': tone === 'danger',
        'text-amber-700': tone === 'warn',
        'text-emerald-700': tone === 'success',
      }"
    >
      {{ label }}
    </span>
  </div>
</template>
