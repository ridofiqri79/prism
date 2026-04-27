<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import ConfirmDialog from 'primevue/confirmdialog'
import Toast from 'primevue/toast'
import { useToast } from 'primevue/usetoast'
import type { AppToastMessage } from '@/utils/app-events'
import { TOAST_EVENT_NAME } from '@/utils/app-events'

const toast = useToast()

function handleToastEvent(event: Event) {
  const { detail } = event as CustomEvent<AppToastMessage>

  if (detail) {
    toast.add(detail)
  }
}

onMounted(() => {
  window.addEventListener(TOAST_EVENT_NAME, handleToastEvent as EventListener)
})

onBeforeUnmount(() => {
  window.removeEventListener(TOAST_EVENT_NAME, handleToastEvent as EventListener)
})
</script>

<template>
  <Toast position="top-right" />
  <ConfirmDialog />
  <RouterView />
</template>
