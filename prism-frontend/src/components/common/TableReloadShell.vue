<script setup lang="ts">
withDefaults(
  defineProps<{
    refreshing?: boolean
    contentClass?: string
  }>(),
  {
    refreshing: false,
    contentClass: '',
  },
)
</script>

<template>
  <div class="prism-table-reload-shell" :class="{ 'is-refreshing': refreshing }">
    <div class="prism-table-reload-content" :class="contentClass">
      <slot />
    </div>
  </div>
</template>
