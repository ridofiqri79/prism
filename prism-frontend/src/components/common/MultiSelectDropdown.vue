<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import InputText from 'primevue/inputtext'

type MultiSelectOption = Record<string, unknown>
type SelectValue = unknown

type FilterEvent = {
  originalEvent: Event
  value: string
}

type ChangeEvent = {
  originalEvent: Event | null
  value: SelectValue[]
}

const props = withDefaults(
  defineProps<{
    modelValue?: SelectValue[] | null
    options?: SelectValue[]
    optionLabel?: string
    optionValue?: string
    optionDisabled?: string
    placeholder?: string
    disabled?: boolean
    filter?: boolean
    filterPlaceholder?: string
    display?: 'chip' | 'comma'
    showClear?: boolean
    maxSelectedLabels?: number | null
    selectionLimit?: number | null
    showToggleAll?: boolean
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
    modelValue: () => [],
    options: () => [],
    optionLabel: 'label',
    optionValue: 'value',
    optionDisabled: undefined,
    placeholder: 'Pilih opsi',
    disabled: false,
    filter: false,
    filterPlaceholder: undefined,
    display: 'chip',
    showClear: true,
    maxSelectedLabels: null,
    selectionLimit: null,
    showToggleAll: true,
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
  'update:modelValue': [value: SelectValue[]]
  change: [event: ChangeEvent]
  filter: [event: FilterEvent]
  show: []
  hide: []
}>()

defineSlots<{
  value?(props: { value: SelectValue[]; placeholder: string }): unknown
  option?(props: { option: MultiSelectOption; selected: boolean; index: number }): unknown
}>()

type SelectedOption = {
  value: SelectValue
  label: string
}

const rootRef = ref<HTMLElement | null>(null)
const panelRef = ref<HTMLElement | null>(null)
const filterInputRef = ref<HTMLInputElement | null>(null)
const isOpen = ref(false)
const filterText = ref('')
const floatingOverlayStyle = ref<Record<string, string | number>>({})
let hasFloatingListeners = false

const normalizedValue = computed(() => (Array.isArray(props.modelValue) ? props.modelValue : []))

const selectedIds = computed(() => new Set(normalizedValue.value.map((value) => stableKey(value))))

const selectedOptions = computed<SelectedOption[]>(() =>
  normalizedValue.value.map((value) => {
    const option = props.options.find((item) => isSameValue(getOptionValue(item), value))

    return {
      value,
      label: option ? getOptionLabel(option) : String(value ?? ''),
    }
  }),
)

const displayedSelectedOptions = computed(() => {
  if (!props.maxSelectedLabels || props.maxSelectedLabels <= 0) {
    return selectedOptions.value
  }

  return selectedOptions.value.slice(0, props.maxSelectedLabels)
})

const hiddenSelectedCount = computed(() =>
  Math.max(0, selectedOptions.value.length - displayedSelectedOptions.value.length),
)

const commaValueLabel = computed(() => selectedOptions.value.map((option) => option.label).join(', '))

const searchPlaceholder = computed(() => props.filterPlaceholder ?? `Cari ${props.placeholder}`)

const visibleOptions = computed(() => {
  const search = filterText.value.trim().toLowerCase()

  if (!props.filter || !search) {
    return props.options
  }

  return props.options.filter((option) =>
    [getOptionLabel(option), String(getOptionValue(option) ?? '')].some((value) =>
      value.toLowerCase().includes(search),
    ),
  )
})

const overlayInlineStyle = computed(() => ({
  minWidth: '100%',
  ...props.overlayStyle,
}))

const shouldTeleportPanel = computed(() => props.appendTo !== 'self')

const teleportTarget = computed(() =>
  props.appendTo === 'self' ? 'body' : props.appendTo,
)

function toOptionRecord(option: SelectValue): MultiSelectOption {
  return option && typeof option === 'object' ? (option as MultiSelectOption) : { value: option }
}

function resolveField(option: SelectValue, field?: string) {
  if (!field) {
    return undefined
  }

  return field.split('.').reduce<unknown>((value, key) => {
    if (value && typeof value === 'object' && key in value) {
      return (value as Record<string, unknown>)[key]
    }

    return undefined
  }, toOptionRecord(option))
}

function getOptionLabel(option: SelectValue) {
  const value = resolveField(option, props.optionLabel)

  return String(value ?? option ?? '')
}

function getOptionValue(option: SelectValue) {
  return props.optionValue ? resolveField(option, props.optionValue) : option
}

function getOptionDisabled(option: SelectValue) {
  const value = props.optionDisabled ? resolveField(option, props.optionDisabled) : false

  return Boolean(value)
}

function stableKey(value: SelectValue) {
  if (value === null || value === undefined) {
    return ''
  }

  if (typeof value === 'object') {
    return JSON.stringify(value)
  }

  return String(value)
}

function isSameValue(left: SelectValue, right: SelectValue) {
  return stableKey(left) === stableKey(right)
}

function isSelected(option: SelectValue) {
  return selectedIds.value.has(stableKey(getOptionValue(option)))
}

function isSelectionLimitReached(option: SelectValue) {
  return (
    props.selectionLimit !== null &&
    props.selectionLimit > 0 &&
    normalizedValue.value.length >= props.selectionLimit &&
    !isSelected(option)
  )
}

function isOptionDisabled(option: SelectValue) {
  return (getOptionDisabled(option) || isSelectionLimitReached(option)) && !isSelected(option)
}

function openPanel() {
  if (props.disabled || isOpen.value) {
    return
  }

  isOpen.value = true
  emit('show')

  void nextTick(() => {
    updateFloatingOverlayStyle()
    addFloatingListeners()

    if (props.filter) {
      filterInputRef.value?.focus()
    }
  })
}

function closePanel() {
  if (!isOpen.value) {
    return
  }

  isOpen.value = false
  removeFloatingListeners()
  emit('hide')
}

function togglePanel() {
  if (isOpen.value) {
    closePanel()
    return
  }

  openPanel()
}

function updateSelected(value: SelectValue[], originalEvent: Event | null) {
  const deduped = value.filter(
    (item, index, items) => items.findIndex((candidate) => isSameValue(candidate, item)) === index,
  )

  emit('update:modelValue', deduped)
  emit('change', { originalEvent, value: deduped })
}

function toggleOption(option: SelectValue, event: Event) {
  if (props.disabled || isOptionDisabled(option)) {
    return
  }

  const value = getOptionValue(option)

  if (isSelected(option)) {
    updateSelected(
      normalizedValue.value.filter((item) => !isSameValue(item, value)),
      event,
    )
    return
  }

  updateSelected([...normalizedValue.value, value], event)
}

function clearSelection(event: Event) {
  if (props.disabled) {
    return
  }

  updateSelected([], event)
}

function handleFilterInput(event: Event) {
  const value = (event.target as HTMLInputElement).value

  filterText.value = value
  emit('filter', { originalEvent: event, value })
}

function updateFloatingOverlayStyle() {
  if (!shouldTeleportPanel.value || !rootRef.value) {
    return
  }

  const rect = rootRef.value.getBoundingClientRect()

  floatingOverlayStyle.value = {
    ...props.overlayStyle,
    position: 'fixed',
    left: `${rect.left}px`,
    top: `${rect.bottom + 8}px`,
    width: `${rect.width}px`,
    minWidth: `${rect.width}px`,
    zIndex: 1200,
  }
}

function addFloatingListeners() {
  if (!shouldTeleportPanel.value || hasFloatingListeners) {
    return
  }

  window.addEventListener('resize', updateFloatingOverlayStyle)
  window.addEventListener('scroll', updateFloatingOverlayStyle, true)
  hasFloatingListeners = true
}

function removeFloatingListeners() {
  if (!hasFloatingListeners) {
    return
  }

  window.removeEventListener('resize', updateFloatingOverlayStyle)
  window.removeEventListener('scroll', updateFloatingOverlayStyle, true)
  hasFloatingListeners = false
}

function handleDocumentPointerDown(event: PointerEvent) {
  const target = event.target as Node

  if (!rootRef.value?.contains(target) && !panelRef.value?.contains(target)) {
    closePanel()
  }
}

function handleDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    closePanel()
  }
}

onMounted(() => {
  document.addEventListener('pointerdown', handleDocumentPointerDown)
  document.addEventListener('keydown', handleDocumentKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
  document.removeEventListener('keydown', handleDocumentKeydown)
  removeFloatingListeners()
})
</script>

<template>
  <div ref="rootRef" class="relative w-full">
    <div
      role="combobox"
      aria-haspopup="listbox"
      :aria-expanded="isOpen"
      :aria-label="ariaLabel ?? undefined"
      :aria-labelledby="ariaLabelledby ?? undefined"
      :tabindex="disabled ? -1 : 0"
      class="flex min-h-10 w-full cursor-pointer items-center justify-between gap-2 rounded-md border border-surface-300 bg-white px-3 py-2 text-left text-sm text-surface-800 transition hover:border-primary-300 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-100"
      :class="{ 'cursor-not-allowed bg-surface-100 text-surface-400': disabled }"
      @click.prevent="togglePanel"
      @keydown.enter.prevent="togglePanel"
      @keydown.space.prevent="togglePanel"
      @keydown.down.prevent="openPanel"
    >
      <slot name="value" :value="normalizedValue" :placeholder="placeholder">
        <div v-if="selectedOptions.length" class="min-w-0 flex-1 pr-1">
          <div v-if="display === 'chip'" class="flex min-w-0 flex-wrap gap-1.5">
            <span
              v-for="option in displayedSelectedOptions"
              :key="stableKey(option.value)"
              class="inline-flex max-w-full items-center rounded-md border border-primary-200 bg-primary-50 px-2 py-1 text-xs font-medium text-primary-700"
              :title="option.label"
            >
              <span class="max-w-72 truncate">{{ option.label }}</span>
            </span>
            <span
              v-if="hiddenSelectedCount"
              class="inline-flex items-center rounded-md border border-surface-200 bg-surface-50 px-2 py-1 text-xs font-medium text-surface-600"
            >
              +{{ hiddenSelectedCount }} lainnya
            </span>
          </div>
          <span v-else class="block truncate">{{ commaValueLabel }}</span>
        </div>
        <span v-else class="min-w-0 flex-1 truncate text-surface-400">{{ placeholder }}</span>
      </slot>
      <div class="flex shrink-0 items-center gap-1 pl-1">
        <button
          v-if="showClear && selectedOptions.length && !disabled"
          type="button"
          class="inline-flex h-6 w-6 items-center justify-center rounded-full text-surface-500 transition hover:bg-surface-100 hover:text-surface-700"
          aria-label="Hapus semua pilihan"
          @click.stop.prevent="clearSelection"
        >
          <span class="pi pi-times text-xs" aria-hidden="true" />
        </button>
        <span
          class="pi shrink-0 text-xs text-surface-500"
          :class="isOpen ? 'pi-chevron-up' : 'pi-chevron-down'"
          aria-hidden="true"
        />
      </div>
    </div>

    <Teleport v-if="isOpen && shouldTeleportPanel" :to="teleportTarget">
      <div
        ref="panelRef"
        class="rounded-lg border border-surface-200 bg-white p-3 shadow-lg"
        :class="[panelClass, overlayClass]"
        :style="floatingOverlayStyle"
      >
        <label v-if="filter" class="relative block w-full">
          <span class="sr-only">{{ searchPlaceholder }}</span>
          <i
            class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400"
            aria-hidden="true"
          />
          <InputText
            :id="inputId ?? undefined"
            ref="filterInputRef"
            :model-value="filterText"
            class="w-full pl-10"
            :placeholder="searchPlaceholder"
            :aria-label="searchPlaceholder"
            :disabled="disabled"
            @input="handleFilterInput"
            @keydown.escape.stop.prevent="closePanel"
          />
        </label>

        <div
          role="listbox"
          aria-multiselectable="true"
          class="space-y-1 overflow-y-auto pr-1"
          :class="{ 'mt-3': filter }"
          :style="{ maxHeight: scrollHeight }"
        >
          <button
            v-for="(option, index) in visibleOptions"
            :key="stableKey(getOptionValue(option))"
            type="button"
            role="option"
            :aria-selected="isSelected(option)"
            :disabled="isOptionDisabled(option)"
            class="flex w-full items-center gap-3 rounded-md border px-3 py-2 text-left text-sm transition"
            :class="
              isSelected(option)
                ? 'border-primary-200 bg-primary-50 text-primary-800'
                : isOptionDisabled(option)
                  ? 'cursor-not-allowed border-surface-100 bg-surface-50 text-surface-400'
                  : 'border-surface-200 bg-white text-surface-700 hover:border-primary-200 hover:bg-surface-50'
            "
            @click="toggleOption(option, $event)"
          >
            <span
              class="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded border"
              :class="
                isSelected(option)
                  ? 'border-primary-500 bg-primary-500 text-white'
                  : 'border-surface-300 bg-white text-transparent'
              "
            >
              <span class="pi pi-check text-[0.65rem]" aria-hidden="true" />
            </span>
            <div class="min-w-0 flex-1">
              <slot
                name="option"
                :option="toOptionRecord(option)"
                :selected="isSelected(option)"
                :index="index"
              >
                <span class="block truncate font-medium">{{ getOptionLabel(option) }}</span>
              </slot>
            </div>
          </button>

          <p
            v-if="loading"
            class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500"
          >
            Memuat data
          </p>
          <p
            v-else-if="visibleOptions.length === 0"
            class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500"
          >
            {{ filterText ? emptyFilterMessage : emptyMessage }}
          </p>
        </div>

        <div
          v-if="selectedOptions.length"
          class="mt-3 flex items-center justify-between border-t border-surface-100 pt-3 text-xs text-surface-500"
        >
          <span>{{ selectedOptions.length }} dipilih</span>
          <button
            type="button"
            class="cursor-pointer font-medium text-primary-600 hover:text-primary-700"
            @click="clearSelection"
          >
            Hapus semua
          </button>
        </div>
      </div>
    </Teleport>

    <div
      v-else-if="isOpen"
      ref="panelRef"
      class="absolute left-0 right-0 top-full z-50 mt-2 rounded-lg border border-surface-200 bg-white p-3 shadow-lg"
      :class="[panelClass, overlayClass]"
      :style="overlayInlineStyle"
    >
      <label v-if="filter" class="relative block w-full">
        <span class="sr-only">{{ searchPlaceholder }}</span>
        <i
          class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400"
          aria-hidden="true"
        />
        <InputText
          :id="inputId ?? undefined"
          ref="filterInputRef"
          :model-value="filterText"
          class="w-full pl-10"
          :placeholder="searchPlaceholder"
          :aria-label="searchPlaceholder"
          :disabled="disabled"
          @input="handleFilterInput"
          @keydown.escape.stop.prevent="closePanel"
        />
      </label>

      <div
        role="listbox"
        aria-multiselectable="true"
        class="space-y-1 overflow-y-auto pr-1"
        :class="{ 'mt-3': filter }"
        :style="{ maxHeight: scrollHeight }"
      >
        <button
          v-for="(option, index) in visibleOptions"
          :key="stableKey(getOptionValue(option))"
          type="button"
          role="option"
          :aria-selected="isSelected(option)"
          :disabled="isOptionDisabled(option)"
          class="flex w-full items-center gap-3 rounded-md border px-3 py-2 text-left text-sm transition"
          :class="
            isSelected(option)
              ? 'border-primary-200 bg-primary-50 text-primary-800'
              : isOptionDisabled(option)
                ? 'cursor-not-allowed border-surface-100 bg-surface-50 text-surface-400'
                : 'border-surface-200 bg-white text-surface-700 hover:border-primary-200 hover:bg-surface-50'
          "
          @click="toggleOption(option, $event)"
        >
          <span
            class="inline-flex h-5 w-5 shrink-0 items-center justify-center rounded border"
            :class="
              isSelected(option)
                ? 'border-primary-500 bg-primary-500 text-white'
                : 'border-surface-300 bg-white text-transparent'
            "
          >
            <span class="pi pi-check text-[0.65rem]" aria-hidden="true" />
          </span>
          <div class="min-w-0 flex-1">
            <slot
              name="option"
              :option="toOptionRecord(option)"
              :selected="isSelected(option)"
              :index="index"
            >
              <span class="block truncate font-medium">{{ getOptionLabel(option) }}</span>
            </slot>
          </div>
        </button>

        <p
          v-if="loading"
          class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500"
        >
          Memuat data
        </p>
        <p
          v-else-if="visibleOptions.length === 0"
          class="rounded-md bg-surface-50 px-3 py-2 text-sm text-surface-500"
        >
          {{ filterText ? emptyFilterMessage : emptyMessage }}
        </p>
      </div>

      <div
        v-if="selectedOptions.length"
        class="mt-3 flex items-center justify-between border-t border-surface-100 pt-3 text-xs text-surface-500"
      >
        <span>{{ selectedOptions.length }} dipilih</span>
        <button
          type="button"
          class="cursor-pointer font-medium text-primary-600 hover:text-primary-700"
          @click="clearSelection"
        >
          Hapus semua
        </button>
      </div>
    </div>
  </div>
</template>
