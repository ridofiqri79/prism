<script setup lang="ts">
import Button from 'primevue/button'

withDefaults(
  defineProps<{
    loading?: boolean
    gridClass?: string
  }>(),
  {
    loading: false,
    gridClass: 'grid gap-4 md:grid-cols-2 xl:grid-cols-4',
  },
)

const emit = defineEmits<{
  apply: []
  reset: []
}>()
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div :class="gridClass">
      <slot />
    </div>
    <div class="mt-4 flex flex-wrap justify-end gap-2">
      <Button icon="pi pi-filter" label="Terapkan" :loading="loading" @click="emit('apply')" />
      <Button icon="pi pi-times" label="Reset" severity="secondary" outlined @click="emit('reset')" />
    </div>
  </section>
</template>
