<script setup lang="ts">
import { computed, onMounted } from 'vue'
import SingleSelectDropdown from '@/components/common/SingleSelectDropdown.vue'
import { useMasterStore } from '@/stores/master.store'

const props = withDefaults(
  defineProps<{
    modelValue?: string | null
    placeholder?: string
    disabled?: boolean
    invalid?: boolean
  }>(),
  {
    modelValue: null,
    placeholder: 'Pilih mata uang',
    disabled: false,
    invalid: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const masterStore = useMasterStore()

const options = computed(() =>
  masterStore.currencies
    .filter((currency) => currency.is_active || currency.code === props.modelValue?.toUpperCase())
    .map((currency) => ({
      code: currency.code,
      label: `${currency.code} - ${currency.name}`,
    })),
)

const selectedValue = computed<string | null>({
  get: () => props.modelValue?.toUpperCase() || null,
  set: (value) => emit('update:modelValue', value?.toUpperCase() ?? ''),
})

onMounted(() => {
  void masterStore.fetchCurrencies(false, { limit: 1000, sort: 'sort_order', order: 'asc' })
})
</script>

<template>
  <SingleSelectDropdown
    v-model="selectedValue"
    :options="options"
    option-label="label"
    option-value="code"
    :placeholder="placeholder"
    :disabled="disabled"
    :invalid="invalid"
    filter
    :show-clear="false"
    append-to="body"
    :overlay-style="{ minWidth: '100%' }"
    class="w-full"
  />
</template>
