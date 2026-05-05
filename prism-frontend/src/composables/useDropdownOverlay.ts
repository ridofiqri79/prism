/**
 * useDropdownOverlay
 *
 * Shared composable for custom dropdown components (MultiSelectDropdown, SingleSelectDropdown).
 * Handles:
 *   - open/close/toggle panel state
 *   - floating overlay positioning (fixed, teleported to body)
 *   - outside-click (pointerdown) to close
 *   - Escape key to close
 *   - resize/scroll repositioning
 *
 * Usage:
 *   const overlay = useDropdownOverlay(rootRef, panelRef, props, emit)
 *   overlay.isOpen, overlay.openPanel(), overlay.closePanel(), etc.
 */

import { computed, nextTick, onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

export interface DropdownOverlayProps {
  disabled?: boolean
  filter?: boolean
  appendTo?: string | HTMLElement
  overlayStyle?: Record<string, string | number> | null
}

export interface DropdownOverlayEmit {
  show: () => void
  hide: () => void
}

export function useDropdownOverlay(
  rootRef: Ref<HTMLElement | null>,
  panelRef: Ref<HTMLElement | null>,
  filterInputRef: Ref<HTMLInputElement | null>,
  props: DropdownOverlayProps,
  emit: DropdownOverlayEmit,
) {
  const isOpen = ref(false)
  const floatingOverlayStyle = ref<Record<string, string | number>>({})
  let hasFloatingListeners = false

  const shouldTeleportPanel = computed(() => props.appendTo !== 'self')

  const teleportTarget = computed(() =>
    props.appendTo === 'self' ? 'body' : (props.appendTo ?? 'body'),
  )

  const overlayInlineStyle = computed(() => ({
    minWidth: '100%',
    ...props.overlayStyle,
  }))

  function updateFloatingOverlayStyle() {
    if (!shouldTeleportPanel.value || !rootRef.value) return

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
    if (!shouldTeleportPanel.value || hasFloatingListeners) return

    window.addEventListener('resize', updateFloatingOverlayStyle)
    window.addEventListener('scroll', updateFloatingOverlayStyle, true)
    hasFloatingListeners = true
  }

  function removeFloatingListeners() {
    if (!hasFloatingListeners) return

    window.removeEventListener('resize', updateFloatingOverlayStyle)
    window.removeEventListener('scroll', updateFloatingOverlayStyle, true)
    hasFloatingListeners = false
  }

  function openPanel() {
    if (props.disabled || isOpen.value) return

    isOpen.value = true
    emit.show()

    void nextTick(() => {
      updateFloatingOverlayStyle()
      addFloatingListeners()

      if (props.filter) {
        filterInputRef.value?.focus()
      }
    })
  }

  function closePanel() {
    if (!isOpen.value) return

    isOpen.value = false
    removeFloatingListeners()
    emit.hide()
  }

  function togglePanel() {
    if (isOpen.value) {
      closePanel()
      return
    }
    openPanel()
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

  return {
    isOpen,
    floatingOverlayStyle,
    overlayInlineStyle,
    shouldTeleportPanel,
    teleportTarget,
    openPanel,
    closePanel,
    togglePanel,
  }
}

// ---------------------------------------------------------------------------
// useOptionResolver — shared option field resolution (label, value, disabled)
// ---------------------------------------------------------------------------

type SelectValue = unknown
type OptionRecord = Record<string, unknown>

export interface OptionResolverProps {
  optionLabel?: string
  optionValue?: string
  optionDisabled?: string
}

export function useOptionResolver(props: OptionResolverProps) {
  function toOptionRecord(option: SelectValue): OptionRecord {
    return option && typeof option === 'object' ? (option as OptionRecord) : { value: option }
  }

  function resolveField(option: SelectValue, field?: string) {
    if (!field) return undefined

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
    if (value === null || value === undefined) return ''
    if (typeof value === 'object') return JSON.stringify(value)
    return String(value)
  }

  function isSameValue(left: SelectValue, right: SelectValue) {
    return stableKey(left) === stableKey(right)
  }

  return {
    toOptionRecord,
    resolveField,
    getOptionLabel,
    getOptionValue,
    getOptionDisabled,
    stableKey,
    isSameValue,
  }
}
