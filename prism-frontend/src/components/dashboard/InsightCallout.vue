<script setup lang="ts">
import Message from 'primevue/message'

defineProps<{
  insights: string[]
}>()
</script>

<template>
  <div class="grid gap-3 md:grid-cols-3">
    <Message
      v-for="insight in insights"
      :key="insight"
      severity="info"
      icon="pi pi-info-circle"
      class="m-0"
    >
      {{ insight }}
    </Message>
  </div>
</template>
