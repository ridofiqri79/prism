<script setup lang="ts">
import { computed } from 'vue'
import Tag from 'primevue/tag'

const props = defineProps<{
  riskCount: number
  category?: string
}>()

const severity = computed(() => {
  if (props.riskCount >= 5 || props.category === 'High Risk') return 'danger'
  if (props.riskCount >= 2 || props.category === 'Watch') return 'warn'
  return 'success'
})

const label = computed(() => {
  if (props.riskCount === 0) return props.category ?? 'Good'
  return `${props.riskCount} risk`
})
</script>

<template>
  <Tag :value="label" :severity="severity" />
</template>
