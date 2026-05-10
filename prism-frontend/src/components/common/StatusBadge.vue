<script setup lang="ts">
import { computed } from 'vue'
import Tag from 'primevue/tag'
import { getStatusLabel, getStatusSeverity, type StatusDomain } from '@/utils/status-labels'

const props = defineProps<{
  status: string
  /** Override label — if provided, skips the status-labels lookup */
  label?: string
  /** Domain context for domain-specific label/severity resolution */
  domain?: StatusDomain
}>()

const resolvedSeverity = computed(() => getStatusSeverity(props.status, props.domain ?? 'default'))
const resolvedLabel = computed(() => props.label ?? getStatusLabel(props.status, props.domain ?? 'default'))
</script>

<template>
  <Tag :severity="resolvedSeverity" :value="resolvedLabel" rounded />
</template>
