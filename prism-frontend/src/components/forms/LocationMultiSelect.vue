<script setup lang="ts">
import { computed } from 'vue'
import MultiSelectDropdown from '@/components/common/MultiSelectDropdown.vue'
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
    type_label: formatRegionType(region),
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

function formatRegionType(region: Region) {
  if (region.type === 'COUNTRY') {
    return 'Nasional'
  }

  if (region.type === 'PROVINCE') {
    return 'Provinsi'
  }

  return 'Kab/Kota'
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
</script>

<template>
  <MultiSelectDropdown
    v-model="selectedValues"
    :options="regionOptions"
    option-label="label"
    option-value="id"
    option-disabled="disabled"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    filter-placeholder="Cari wilayah"
    display="chip"
    scroll-height="18rem"
    @show="void masterStore.fetchAllRegionLevels()"
  >
    <template #option="{ option }">
      <span class="block min-w-0">
        <span class="block truncate font-medium">{{ option.label }}</span>
        <span class="mt-1 flex items-center gap-2 text-xs text-surface-500">
          <span class="rounded bg-surface-100 px-1.5 py-0.5 font-mono text-[0.65rem]">
            {{ option.code }}
          </span>
          <span>{{ option.type_label }}</span>
        </span>
      </span>
    </template>
  </MultiSelectDropdown>
</template>
