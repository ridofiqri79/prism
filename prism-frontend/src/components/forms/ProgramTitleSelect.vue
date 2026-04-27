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

const options = computed(() =>
  masterStore.programTitles.map((programTitle) => formatProgramOption(programTitle)),
)

function formatProgramOption(programTitle: ProgramTitle) {
  const parent = masterStore.programTitles.find((item) => item.id === programTitle.parent_id)

  return {
    label: parent ? `${parent.title} / ${programTitle.title}` : programTitle.title,
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
    :options="options"
    option-label="label"
    option-value="value"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    show-clear
    class="w-full"
  />
</template>
