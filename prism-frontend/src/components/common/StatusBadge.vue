<script setup lang="ts">
import { computed } from 'vue'
import Tag from 'primevue/tag'

const props = defineProps<{
  status: string
}>()

const severity = computed(() => {
  const normalized = props.status.toLowerCase()

  if (normalized === 'active') return 'success'
  if (normalized === 'deleted') return 'danger'
  if (normalized === 'superseded') return 'secondary'
  if (normalized === 'extended') return 'warn'
  if (['tw1', 'tw2', 'tw3', 'tw4'].includes(normalized)) return 'info'

  return 'secondary'
})
</script>

<template>
  <Tag :severity="severity" :value="status" rounded />
</template>
