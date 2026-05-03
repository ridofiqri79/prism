<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    message?: string
    icon?: string
  }>(),
  {
    message: '',
    icon: 'pi pi-inbox',
  },
)
</script>

<template>
  <div class="flex min-h-40 flex-col items-center justify-center rounded-lg border border-dashed border-surface-200 bg-surface-50 p-6 text-center">
    <i :class="[icon, 'text-2xl text-surface-400']" />
    <p class="mt-3 text-sm font-semibold text-surface-900">{{ title }}</p>
    <p v-if="message" class="mt-1 max-w-xl text-sm text-surface-500">{{ message }}</p>
    <div v-if="$slots.action" class="mt-4">
      <slot name="action" />
    </div>
  </div>
</template>
