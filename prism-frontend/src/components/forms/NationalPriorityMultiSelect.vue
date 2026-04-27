<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MultiSelect from 'primevue/multiselect'
import { useMasterStore } from '@/stores/master.store'

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    periodId?: string
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    periodId: undefined,
    placeholder: 'Pilih prioritas nasional',
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

const options = computed(() =>
  masterStore.nationalPriorities.filter(
    (priority) => !props.periodId || priority.period_id === props.periodId,
  ),
)

onMounted(() => {
  void masterStore.fetchNationalPriorities()
})
</script>

<template>
  <MultiSelect
    v-model="selectedValues"
    :options="options"
    option-label="title"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    display="chip"
    class="w-full"
  />
</template>
