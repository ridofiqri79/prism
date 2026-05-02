<script setup lang="ts">
import { computed } from 'vue'
import Tag from 'primevue/tag'

const props = withDefaults(
  defineProps<{
    value: string
    severity?: string
  }>(),
  {
    severity: 'secondary',
  },
)

const primeSeverity = computed(() => {
  if (props.severity === 'warning') return 'warn'
  if (props.severity === 'danger') return 'danger'
  if (props.severity === 'success') return 'success'
  if (props.severity === 'info') return 'info'
  return 'secondary'
})
</script>

<template>
  <Tag :value="value" :severity="primeSeverity" rounded />
</template>
