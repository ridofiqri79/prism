<script setup lang="ts">
import { computed, onMounted } from 'vue'
import Select from 'primevue/select'
import { useMasterStore } from '@/stores/master.store'
import type { ProgramTitle } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    modelValue: string | null
    extraOptions?: ProgramTitle[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    extraOptions: () => [],
    placeholder: 'Pilih program title',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string | null]
}>()

const masterStore = useMasterStore()

const selectedValue = computed({
  get: () => props.modelValue,
  set: (value: string | null) => emit('update:modelValue', value),
})

const mergedOptions = computed(() => {
  const byId = new Map<string, ProgramTitle>()

  for (const programTitle of [...masterStore.programTitles, ...props.extraOptions]) {
    byId.set(programTitle.id, programTitle)
  }

  return [...byId.values()]
})

const options = computed(() => mergedOptions.value.map((programTitle) => formatProgramOption(programTitle)))

function formatProgramOption(programTitle: ProgramTitle) {
  const parent = mergedOptions.value.find((item) => item.id === programTitle.parent_id)

  return {
    label: parent ? `${parent.title} / ${programTitle.title}` : programTitle.title,
    value: programTitle.id,
  }
}

function lookupParams(search?: string) {
  return {
    limit: 50,
    search: search?.trim() || undefined,
    sort: 'title',
    order: 'asc' as const,
  }
}

function loadOptions(search?: string, force = true) {
  void masterStore.fetchProgramTitles(force, lookupParams(search))
}

onMounted(() => {
  void masterStore.fetchProgramTitles(false, lookupParams())
})
</script>

<template>
  <Select
    v-model="selectedValue"
    :options="options"
    option-label="label"
    option-value="value"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    show-clear
    append-to="self"
    :overlay-style="{ minWidth: '100%' }"
    class="w-full"
    @filter="loadOptions($event.value)"
  />
</template>
