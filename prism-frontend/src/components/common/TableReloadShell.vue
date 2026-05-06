<script setup lang="ts">
withDefaults(
  defineProps<{
    refreshing?: boolean
    label?: string
    contentClass?: string
  }>(),
  {
    refreshing: false,
    label: 'Memuat ulang data...',
    contentClass: '',
  },
)
</script>

<template>
  <div class="prism-table-reload-shell" :class="{ 'is-refreshing': refreshing }">
    <Transition name="prism-table-reload">
      <div v-if="refreshing" class="prism-table-reload-indicator-wrap" aria-live="polite">
        <div class="prism-table-reload-indicator">
          <i class="pi pi-spin pi-spinner" aria-hidden="true" />
          <span>{{ label }}</span>
        </div>
      </div>
    </Transition>

    <div class="prism-table-reload-content" :class="contentClass">
      <slot />
    </div>
  </div>
</template>
