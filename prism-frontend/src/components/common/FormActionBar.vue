<script setup lang="ts">
import Button from 'primevue/button'

defineProps<{
  loading?: boolean
  cancelLabel?: string
  submitLabel?: string
  submitIcon?: string
}>()

const emit = defineEmits<{
  cancel: []
}>()
</script>

<template>
  <div
    class="sticky bottom-0 flex justify-end gap-2 border-t border-surface-200 bg-surface-50/95 py-4 backdrop-blur"
  >
    <Button
      :label="cancelLabel ?? 'Batal'"
      severity="secondary"
      outlined
      @click="emit('cancel')"
    />
    <Button
      type="submit"
      :label="submitLabel ?? 'Simpan'"
      :icon="submitIcon ?? 'pi pi-save'"
      :loading="loading"
    />
  </div>
</template>
