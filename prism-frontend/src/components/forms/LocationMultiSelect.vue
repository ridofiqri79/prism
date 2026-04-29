<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MultiSelect from 'primevue/multiselect'
import { useMasterStore } from '@/stores/master.store'
import type { Region } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    placeholder: 'Pilih wilayah',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const masterStore = useMasterStore()

const selectedValues = computed({
  get: () => props.modelValue,
  set: (value: string[]) => emit('update:modelValue', value),
})

const selectedCountryCodes = computed(() => {
  const selected = new Set(props.modelValue)

  return masterStore.regions
    .filter((region) => region.type === 'COUNTRY' && selected.has(region.id))
    .map((region) => region.code)
})

const regionOptions = computed(() =>
  masterStore.regions.map((region) => ({
    ...region,
    label: formatRegionLabel(region),
    disabled:
      selectedCountryCodes.value.length > 0 &&
      region.type !== 'COUNTRY' &&
      isCoveredBySelectedCountry(region),
  })),
)

function formatRegionLabel(region: Region) {
  if (region.type === 'COUNTRY') {
    return `${region.name} (Nasional)`
  }

  if (region.type === 'CITY') {
    return `-- ${region.name}`
  }

  return `- ${region.name}`
}

function isCoveredBySelectedCountry(region: Region) {
  if (!region.parent_code) {
    return false
  }

  if (selectedCountryCodes.value.includes(region.parent_code)) {
    return true
  }

  const parent = masterStore.regions.find((item) => item.code === region.parent_code)

  return parent?.parent_code ? selectedCountryCodes.value.includes(parent.parent_code) : false
}

onMounted(() => {
  void masterStore.fetchAllRegionLevels()
})
</script>

<template>
  <MultiSelect
    v-model="selectedValues"
    :options="regionOptions"
    option-label="label"
    option-value="id"
    option-disabled="disabled"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    display="chip"
    class="w-full"
  />
</template>
