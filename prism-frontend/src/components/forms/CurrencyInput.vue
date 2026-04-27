<script setup lang="ts">
import { computed } from 'vue'
import InputNumber from 'primevue/inputnumber'

const props = withDefaults(
  defineProps<{
    modelValue: number
    currency?: string
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    currency: 'USD',
    placeholder: '0.00',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

const inputValue = computed({
  get: () => props.modelValue,
  set: (value: number | null) => emit('update:modelValue', value ?? 0),
})
</script>

<template>
  <InputNumber
    v-model="inputValue"
    :prefix="`${currency} `"
    :placeholder="placeholder"
    :disabled="disabled"
    :min-fraction-digits="2"
    :max-fraction-digits="2"
    class="w-full"
  />
</template>
