<script setup lang="ts">
import Tag from 'primevue/tag'

defineProps<{
  level: string
}>()

function riskTone(level: string) {
  const normalized = level.toLowerCase()
  if (normalized === 'high' || normalized === 'error') return 'danger'
  if (normalized === 'medium' || normalized === 'warning') return 'warn'
  if (normalized === 'low' || normalized === 'info') return 'info'
  return 'secondary'
}
</script>

<template>
  <Tag :value="level" :severity="riskTone(level)" />
</template>
