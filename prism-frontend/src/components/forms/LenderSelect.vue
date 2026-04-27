<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { useMasterStore } from '@/stores/master.store'

const props = withDefaults(
  defineProps<{
    modelValue: string | string[] | null
    multiple?: boolean
    allowedIds?: string[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    multiple: false,
    allowedIds: undefined,
    placeholder: 'Pilih lender',
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
  masterStore.lenders.filter(
    (lender) => !props.allowedIds || props.allowedIds.includes(lender.id),
  ),
)

onMounted(() => {
  void masterStore.fetchLenders()
})
</script>

<template>
  <MultiSelect
    v-if="multiple"
    v-model="selectedValue"
    :options="options"
    option-label="name"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    display="chip"
    class="w-full"
  >
    <template #option="{ option }">
      <div class="flex w-full items-center justify-between gap-3">
        <span>{{ option.name }}</span>
        <Tag :value="option.type" severity="info" rounded />
      </div>
    </template>
  </MultiSelect>

  <Select
    v-else
    v-model="selectedValue"
    :options="options"
    option-label="name"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    show-clear
    class="w-full"
  >
    <template #option="{ option }">
      <div class="flex w-full items-center justify-between gap-3">
        <span>{{ option.name }}</span>
        <Tag :value="option.type" severity="info" rounded />
      </div>
    </template>
  </Select>
</template>
