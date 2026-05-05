<script setup lang="ts">
import { computed, ref } from 'vue'
import InputText from 'primevue/inputtext'
import { useDropdownOverlay, useOptionResolver } from '@/composables/useDropdownOverlay'

type SingleSelectOption = Record<string, unknown>
type SelectValue = unknown

type FilterEvent = {
  originalEvent: Event
  value: string
}

type ChangeEvent = {
  originalEvent: Event | null
  value: SelectValue | null
}

const props = withDefaults(
  defineProps<{
    modelValue?: SelectValue | null
    options?: SelectValue[]
    optionLabel?: string
    optionValue?: string
    optionDisabled?: string
    placeholder?: string
    disabled?: boolean
    invalid?: boolean
    filter?: boolean
    filterPlaceholder?: string
    showClear?: boolean
    loading?: boolean
    emptyMessage?: string
    emptyFilterMessage?: string
    appendTo?: string | HTMLElement
    overlayStyle?: Record<string, string | number> | null
    overlayClass?: string | null
    panelClass?: string | null
    inputId?: string | null
    scrollHeight?: string
    dataKey?: string | null
    ariaLabel?: string | null
    ariaLabelledby?: string | null
  }>(),
  {
    options: () => [],
    optionLabel: 'label',
    optionValue: 'value',
    optionDisabled: undefined,
    placeholder: 'Pilih opsi',
    disabled: false,
    invalid: false,
    filter: false,
    filterPlaceholder: undefined,
    showClear: true,
    loading: false,
    emptyMessage: 'Tidak ada opsi',
    emptyFilterMessage: 'Tidak ada hasil',
    appendTo: 'body',
    overlayStyle: null,
    overlayClass: null,
    panelClass: null,
    inputId: null,
    scrollHeight: '18rem',
    dataKey: null,
    ariaLabel: null,
    ariaLabelledby: null,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: SelectValue | null]
  change: [event: ChangeEvent]
  filter: [event: FilterEvent]
  show: []
  hide: []
}>()

defineSlots<{
  value?(props: {
    value: SelectValue | null
    placeholder: string
    selectedOption: SelectedOption | null
  }): unknown
  option?(props: { option: SingleSelectOption; selected: boolean; index: number }): unknown
}>()

type SelectedOption = {
  value: SelectValue
  label: string
}

const rootRef = ref<HTMLElement | null>(null)
const panelRef = ref<HTMLElement | null>(null)
const filterInputRef = ref<HTMLInputElement | null>(null)
const filterText = ref('')

// --------------- shared composables ---------------
const { toOptionRecord, getOptionLabel, getOptionValue, getOptionDisabled, stableKey, isSameValue } =
  useOptionResolver(props)

const overlay = useDropdownOverlay(rootRef, panelRef, filterInputRef, props, {
  show: () => emit('show'),
  hide: () => emit('hide'),
})
// --------------------------------------------------

const normalizedValue = computed(() => (props.modelValue === undefined ? null : props.modelValue))

const hasSelectedValue = computed(
  () =>
    normalizedValue.value !== null &&
    normalizedValue.value !== undefined &&
    normalizedValue.value !== '',
)

const selectedOption = computed<SelectedOption | null>(() => {
  if (!hasSelectedValue.value) return null

  const option = props.options.find((item) =>
    isSameValue(getOptionValue(item), normalizedValue.value),
  )

  return option
    ? { value: normalizedValue.value, label: getOptionLabel(option) }
    : { value: normalizedValue.value, label: String(normalizedValue.value ?? '') }
})

const searchPlaceholder = computed(() => props.filterPlaceholder ?? `Cari ${props.placeholder}`)

const visibleOptions = computed(() => {
  const search = filterText.value.trim().toLowerCase()
  if (!props.filter || !search) return props.options
  return props.options.filter((option) =>
    [getOptionLabel(option), String(getOptionValue(option) ?? '')].some((v) =>
      v.toLowerCase().includes(search),
    ),
  )
})

const controlClass = computed(() =>
  [
    'flex min-h-10 w-full cursor-pointer items-center justify-between gap-2 rounded-md border bg-white px-3 py-2 text-left text-sm text-surface-800 transition',
    props.invalid
      ? 'border-red-400 hover:border-red-500 focus:border-red-500 focus:outline-none focus:ring-2 focus:ring-red-100'
      : 'border-surface-300 hover:border-primary-300 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100',
    props.disabled ? 'cursor-not-allowed bg-surface-100 text-surface-400' : '',
  ].join(' '),
)

function isSelected(option: SelectValue) {
  if (!hasSelectedValue.value) return false
  return isSameValue(getOptionValue(option), normalizedValue.value)
}

function updateSelected(value: SelectValue | null, originalEvent: Event | null) {
  emit('update:modelValue', value)
  emit('change', { originalEvent, value })
}

function chooseOption(option: SelectValue, event: Event) {
  if (props.disabled || getOptionDisabled(option)) return

  const value = getOptionValue(option)

  if (isSelected(option)) {
    overlay.closePanel()
    return
  }

  updateSelected(value, event)
  overlay.closePanel()
}

function clearSelection(event: Event) {
  if (props.disabled) return
  updateSelected(null, event)
}

function handleFilterInput(event: Event) {
  const value = (event.target as HTMLInputElement).value
  filterText.value = value
  emit('filter', { originalEvent: event, value })
}
</script>

<template>
  <div ref="rootRef" class="relative w-full">
    <div
      role="combobox"
      aria-haspopup="listbox"
      :aria-expanded="overlay.isOpen.value"
      :aria-invalid="invalid || undefined"
      :aria-label="ariaLabel ?? undefined"
      :aria-labelledby="ariaLabelledby ?? undefined"
      :tabindex="disabled ? -1 : 0"
      :class="controlClass"
      @click.prevent="overlay.togglePanel"
      @keydown.enter.prevent="overlay.togglePanel"
      @keydown.space.prevent="overlay.togglePanel"
      @keydown.down.prevent="overlay.openPanel"
    >
      <slot name="value" :value="normalizedValue" :placeholder="placeholder" :selected-option="selectedOption">
        <div v-if="selectedOption" class="min-w-0 flex-1 pr-1">
          <span
            class="inline-flex max-w-full items-center rounded-md border border-primary-200 bg-primary-50 px-2 py-1 text-xs font-medium text-primary-700"
            :title="selectedOption.label"
          >
            <span class="max-w-72 truncate">{{ selectedOption.label }}</span>
          </span>
        </div>
        <span v-else class="min-w-0 flex-1 truncate text-surface-400">{{ placeholder }}</span>
      </slot>

      <div class="flex shrink-0 items-center gap-1 pl-1">
        <button
          v-if="showClear && selectedOption && !disabled"
          type="button"
          class="inline-flex h-6 w-6 items-center justify-center rounded-full text-surface-500 transition hover:bg-surface-100 hover:text-surface-700"
          :aria-label="`Hapus ${selectedOption.label}`"
          @click.stop.prevent="clearSelection"
        >
          <span class="pi pi-times text-xs" aria-hidden="true" />
        </button>
        <span
          class="pi shrink-0 text-xs text-surface-500"
          :class="overlay.isOpen.value ? 'pi-chevron-up' : 'pi-chevron-down'"
          aria-hidden="true"
        />
      </div>
    </div>

    <!-- Teleported (floating) panel -->
    <Teleport v-if="overlay.isOpen.value && overlay.shouldTeleportPanel.value" :to="overlay.teleportTarget.value">
      <div
        ref="panelRef"
        class="rounded-lg border border-surface-200 bg-white p-3 shadow-lg"
        :class="[panelClass, overlayClass]"
        :style="overlay.floatingOverlayStyle.value"
      >
        <label v-if="filter" class="relative block w-full">
          <span class="sr-only">{{ searchPlaceholder }}</span>
          <i class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400" aria-hidden="true" />
          <InputText
            :id="inputId ?? undefined"
            ref="filterInputRef"
            :model-value="filterText"
            class="w-full pl-10"
            :placeholder="searchPlaceholder"
            :aria-label="searchPlaceholder"
            :disabled="disabled"
            @input="handleFilterInput"
            @keydown.escape.stop.prevent="overlay.closePanel"
          />
        </label>

        <div role="listbox" class="space-y-1 overflow-y-auto pr-1" :class="{ 'mt-3': filter }" :style="{ maxHeight: scrollHeight }">
          <button
            v-for="(option, index) in visibleOptions"
            :key="stableKey(getOptionValue(option))"
            type="button"
            role="option"
            :aria-selected="isSelected(option)"
            :disabled="disabled || getOptionDisabled(option)"
            class="flex w-full items-center gap-2 rounded-md border px-3 py-2 text-left text-sm transition"
            :class="isSelected(option) ? 'border-primary-200 bg-primary-50 text-primary-800' : getOptionDisabled(option) ? 'cursor-not-allowed border-surface-100 bg-surface-50 text-surface-400' : 'border-surface-200 bg-white text-surface-700 hover:border-primary-200 hover:bg-surface-50'"
            @click="chooseOption(option, $event)"
          >
            <div class="min-w-0 flex-1">
              <slot name="option" :option="toOptionRecord(option)" :selected="isSelected(option)" :index="index">
                <span class="block truncate font-medium">{{ getOptionLabel(option) }}</span>
              </slot>
            </div>
            <span v-if="isSelected(option)" class="pi pi-check shrink-0 text-sm text-primary-600" aria-hidden="true" />
          </button>

          <p v-if="loading" class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500">Memuat data</p>
          <p v-else-if="visibleOptions.length === 0" class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500">
            {{ filterText ? emptyFilterMessage : emptyMessage }}
          </p>
        </div>
      </div>
    </Teleport>

    <!-- Inline (non-teleported) panel -->
    <div
      v-else-if="overlay.isOpen.value"
      ref="panelRef"
      class="absolute left-0 right-0 top-full z-50 mt-2 rounded-lg border border-surface-200 bg-white p-3 shadow-lg"
      :class="[panelClass, overlayClass]"
      :style="overlay.overlayInlineStyle.value"
    >
      <label v-if="filter" class="relative block w-full">
        <span class="sr-only">{{ searchPlaceholder }}</span>
        <i class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400" aria-hidden="true" />
        <InputText
          :id="inputId ?? undefined"
          ref="filterInputRef"
          :model-value="filterText"
          class="w-full pl-10"
          :placeholder="searchPlaceholder"
          :aria-label="searchPlaceholder"
          :disabled="disabled"
          @input="handleFilterInput"
          @keydown.escape.stop.prevent="overlay.closePanel"
        />
      </label>

      <div role="listbox" class="space-y-1 overflow-y-auto pr-1" :class="{ 'mt-3': filter }" :style="{ maxHeight: scrollHeight }">
        <button
          v-for="(option, index) in visibleOptions"
          :key="stableKey(getOptionValue(option))"
          type="button"
          role="option"
          :aria-selected="isSelected(option)"
          :disabled="disabled || getOptionDisabled(option)"
          class="flex w-full items-center gap-2 rounded-md border px-3 py-2 text-left text-sm transition"
          :class="isSelected(option) ? 'border-primary-200 bg-primary-50 text-primary-800' : getOptionDisabled(option) ? 'cursor-not-allowed border-surface-100 bg-surface-50 text-surface-400' : 'border-surface-200 bg-white text-surface-700 hover:border-primary-200 hover:bg-surface-50'"
          @click="chooseOption(option, $event)"
        >
          <div class="min-w-0 flex-1">
            <slot name="option" :option="toOptionRecord(option)" :selected="isSelected(option)" :index="index">
              <span class="block truncate font-medium">{{ getOptionLabel(option) }}</span>
            </slot>
          </div>
          <span v-if="isSelected(option)" class="pi pi-check shrink-0 text-sm text-primary-600" aria-hidden="true" />
        </button>

        <p v-if="loading" class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500">Memuat data</p>
        <p v-else-if="visibleOptions.length === 0" class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500">
          {{ filterText ? emptyFilterMessage : emptyMessage }}
        </p>
      </div>
    </div>
  </div>
</template>
