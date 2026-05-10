<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import { useMasterStore } from '@/stores/master.store'
import type { Institution } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    modelValue: string | string[] | null
    multiple?: boolean
    levelFilter?: string[]
    extraOptions?: Institution[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    multiple: false,
    levelFilter: undefined,
    extraOptions: () => [],
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

const multiSelectedValue = computed({
  get: () => (Array.isArray(props.modelValue) ? props.modelValue : []),
  set: (value: string[]) => emit('update:modelValue', value),
})

const mergedOptions = computed(() => {
  const byId = new Map<string, Institution>()

  for (const institution of [...masterStore.institutions, ...props.extraOptions]) {
    byId.set(institution.id, institution)
  }

  return [...byId.values()]
})

const options = computed(() =>
  mergedOptions.value
    .filter((institution) => !props.levelFilter || props.levelFilter.includes(institution.level))
    .map((institution) => ({
      ...institution,
      label: formatInstitutionLabel(institution),
    })),
)

function lookupParams(search?: string) {
  return {
    limit: 50,
    search: search?.trim() || undefined,
    level: props.levelFilter,
    sort: 'name',
    order: 'asc' as const,
  }
}

function loadOptions(search?: string, force = true) {
  void masterStore.fetchInstitutions(force, lookupParams(search))
}

function formatInstitutionLabel(institution: Institution) {
  const baseLabel = institution.short_name
    ? `${institution.name} (${institution.short_name})`
    : institution.name
  const parent = mergedOptions.value.find((item) => item.id === institution.parent_id)

  if (!parent) {
    return baseLabel
  }

  const parentLabel = parent.short_name ? `${parent.name} (${parent.short_name})` : parent.name

  return `${parentLabel} / ${baseLabel}`
}

onMounted(() => {
  void masterStore.fetchInstitutions(false, lookupParams())
})
</script>

<template>
  <MultiSelect
    v-if="multiple"
    v-model="multiSelectedValue"
    :options="options"
    option-label="label"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    display="chip"
    append-to="body"
    class="w-full"
    @filter="loadOptions($event.value)"
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
    append-to="body"
    class="w-full"
    @filter="loadOptions($event.value)"
  />
</template>
