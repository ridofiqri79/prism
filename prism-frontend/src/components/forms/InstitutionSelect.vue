<script setup lang="ts">
import { computed, onMounted } from 'vue'
import Select from 'primevue/select'
import MultiSelect from 'primevue/multiselect'
import { useMasterStore } from '@/stores/master.store'
import type { Institution } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    modelValue: string | string[] | null
    multiple?: boolean
    levelFilter?: string[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    multiple: false,
    levelFilter: undefined,
    placeholder: 'Pilih instansi',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string | string[] | null]
}>()

const masterStore = useMasterStore()

const selectedValue = computed({
  get: () => props.modelValue,
  set: (value: string | string[] | null) => emit('update:modelValue', value),
})

const options = computed(() =>
  masterStore.institutions
    .filter((institution) => !props.levelFilter || props.levelFilter.includes(institution.level))
    .map((institution) => ({
      ...institution,
      label: formatInstitutionLabel(institution),
    })),
)

function formatInstitutionLabel(institution: Institution) {
  return institution.short_name
    ? `${institution.name} (${institution.short_name})`
    : institution.name
}

onMounted(() => {
  void masterStore.fetchInstitutions()
})
</script>

<template>
  <MultiSelect
    v-if="multiple"
    v-model="selectedValue"
    :options="options"
    option-label="label"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    display="chip"
    class="w-full"
  />
  <Select
    v-else
    v-model="selectedValue"
    :options="options"
    option-label="label"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    show-clear
    class="w-full"
  />
</template>
