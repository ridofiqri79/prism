<script setup lang="ts">
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
    <span
      v-for="(item, index) in items"
      :key="`${item}-${index}`"
      class="inline-flex min-h-7 items-center rounded-md border border-primary-200 bg-primary-50 px-2.5 py-1 text-xs font-medium leading-snug text-primary-700"
      :class="maxWidthClass"
      :title="item"
    >
      <span class="min-w-0 truncate">{{ item }}</span>
    </span>
  </div>
  <span v-else class="text-sm text-surface-400">{{ emptyLabel }}</span>
</template>
