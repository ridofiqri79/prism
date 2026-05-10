<script setup lang="ts">
import Chip from 'primevue/chip'

withDefaults(
  defineProps<{
    items?: string[]
    emptyLabel?: string
    maxWidthClass?: string
  }>(),
  {
    items: () => [],
    emptyLabel: '-',
    maxWidthClass: 'max-w-full',
  },
)
</script>

<template>
  <div v-if="items.length" class="flex min-w-0 flex-wrap gap-1.5">
    <Chip
      v-for="(item, index) in items"
      :key="`${item}-${index}`"
      :label="item"
      :class="[maxWidthClass, 'text-xs min-h-7 leading-snug']"
      :title="item"
    />
  </div>
  <span v-else class="text-sm text-surface-400">{{ emptyLabel }}</span>
</template>
