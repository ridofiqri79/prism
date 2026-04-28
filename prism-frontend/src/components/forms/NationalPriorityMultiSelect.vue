<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MultiSelect from 'primevue/multiselect'
import { useMasterStore } from '@/stores/master.store'
import type { NationalPriority } from '@/types/master.types'

type NationalPriorityOption = NationalPriority & {
  display_label: string
}

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

const options = computed<NationalPriorityOption[]>(() =>
  masterStore.nationalPriorities
    .filter((priority) => !props.periodId || priority.period_id === props.periodId)
    .map((priority) => ({
      ...priority,
      display_label: priority.period?.name ? `${priority.title} (${priority.period.name})` : priority.title,
    })),
)

onMounted(() => {
  void masterStore.fetchNationalPriorities()
})
</script>

<template>
  <MultiSelect
    v-model="selectedValues"
    :options="options"
    option-label="display_label"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    display="chip"
    class="w-full"
  />
</template>
