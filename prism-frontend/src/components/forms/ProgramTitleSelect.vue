<script setup lang="ts">
import { computed, onMounted } from 'vue'
import Select from 'primevue/select'
import { useMasterStore } from '@/stores/master.store'
import type { ProgramTitle } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    modelValue: string | null
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    placeholder: 'Pilih program title',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string | null]
}>()

const masterStore = useMasterStore()

const selectedValue = computed({
  get: () => props.modelValue,
  set: (value: string | null) => emit('update:modelValue', value),
})

const groupedOptions = computed(() => {
  const parents = masterStore.programTitles.filter((item) => !item.parent_id)
  const children = masterStore.programTitles.filter((item) => item.parent_id)

  return parents.map((parent) => ({
    label: parent.title,
    items: children
      .filter((child) => child.parent_id === parent.id)
      .map((child) => formatProgramOption(child)),
  }))
})

function formatProgramOption(programTitle: ProgramTitle) {
  return {
    label: programTitle.title,
    value: programTitle.id,
  }
}

onMounted(() => {
  void masterStore.fetchProgramTitles()
})
</script>

<template>
  <Select
    v-model="selectedValue"
    :options="groupedOptions"
    option-group-label="label"
    option-group-children="items"
    option-label="label"
    option-value="value"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    show-clear
    class="w-full"
  />
</template>
