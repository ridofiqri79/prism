<script setup lang="ts">
import Tag from 'primevue/tag'

defineProps<{
  severity: string
}>()

function severityTone(value: string) {
  if (value === 'error') return 'danger'
  if (value === 'warning') return 'warn'
  return 'info'
}
</script>

<template>
  <Tag :value="severity" :severity="severityTone(severity)" />
</template>
