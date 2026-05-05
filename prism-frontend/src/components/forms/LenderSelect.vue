<script setup lang="ts">
import { computed, onMounted } from 'vue'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { useMasterStore } from '@/stores/master.store'
import type { Lender } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    modelValue: string | string[] | null
    multiple?: boolean
    allowedIds?: string[]
    extraOptions?: Lender[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    multiple: false,
    allowedIds: undefined,
    extraOptions: () => [],
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

const multiSelectedValue = computed({
  get: () => (Array.isArray(props.modelValue) ? props.modelValue : []),
  set: (value: string[]) => emit('update:modelValue', value),
})

const mergedOptions = computed(() => {
  const byId = new Map<string, Lender>()

  for (const lender of [...masterStore.lenders, ...props.extraOptions]) {
    byId.set(lender.id, lender)
  }

  return [...byId.values()]
})

const options = computed(() =>
  mergedOptions.value.filter(
    (lender) => !props.allowedIds || props.allowedIds.includes(lender.id),
  ),
)

function lookupParams(search?: string) {
  return {
    limit: 50,
    search: search?.trim() || undefined,
    sort: 'name',
    order: 'asc' as const,
  }
}

function loadOptions(search?: string, force = true) {
  void masterStore.fetchLenders(force, lookupParams(search))
}

onMounted(() => {
  void masterStore.fetchLenders(false, lookupParams())
})
</script>

<template>
  <MultiSelect
    v-if="multiple"
    v-model="multiSelectedValue"
    :options="options"
    option-label="name"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    display="chip"
    append-to="body"
    class="w-full"
    @filter="loadOptions($event.value)"
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
    append-to="body"
    class="w-full"
    @filter="loadOptions($event.value)"
  >
    <template #option="{ option }">
      <div class="flex w-full items-center justify-between gap-3">
        <span>{{ option.name }}</span>
        <Tag :value="option.type" severity="info" rounded />
      </div>
    </template>
  </Select>
</template>
