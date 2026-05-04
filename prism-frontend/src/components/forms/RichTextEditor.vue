<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import Button from 'primevue/button'
import { isRichTextEmpty, sanitizeRichText } from '@/utils/rich-text'

const props = withDefaults(
  defineProps<{
    modelValue?: string | null
    placeholder?: string
    minHeight?: string
    maxHeight?: string
    resizable?: boolean
    disabled?: boolean
  }>(),
  {
    modelValue: '',
    placeholder: '',
    minHeight: '9rem',
    maxHeight: 'none',
    resizable: true,
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editorRef = ref<HTMLDivElement | null>(null)

function renderValue(value?: string | null) {
  if (!editorRef.value) return

  const html = sanitizeRichText(value)
  if (editorRef.value.innerHTML !== html) {
    editorRef.value.innerHTML = html
  }
}

function emitValue() {
  if (!editorRef.value) return

  const html = sanitizeRichText(editorRef.value.innerHTML)
  emit('update:modelValue', isRichTextEmpty(html) ? '' : html)
}

function runCommand(
  command: 'bold' | 'italic' | 'underline' | 'insertUnorderedList' | 'insertOrderedList',
) {
  if (props.disabled) return

  editorRef.value?.focus()
  document.execCommand(command)
  emitValue()
}

function handlePaste(event: ClipboardEvent) {
  event.preventDefault()
  const text = event.clipboardData?.getData('text/plain') ?? ''
  document.execCommand('insertText', false, text)
  emitValue()
}

watch(
  () => props.modelValue,
  (value) => renderValue(value),
)

onMounted(() => {
  renderValue(props.modelValue)
})
</script>

<template>
  <div class="rounded-lg border border-surface-200 bg-white">
    <div class="flex items-center gap-1 rounded-t-lg border-b border-surface-200 bg-surface-50 px-2 py-1.5">
      <Button
        type="button"
        label="B"
        text
        rounded
        size="small"
        :disabled="disabled"
        aria-label="Bold"
        title="Bold"
        class="rich-text-editor__bold"
        @click="runCommand('bold')"
      />
      <Button
        type="button"
        label="I"
        text
        rounded
        size="small"
        :disabled="disabled"
        aria-label="Italic"
        title="Italic"
        class="rich-text-editor__italic"
        @click="runCommand('italic')"
      />
      <Button
        type="button"
        label="U"
        text
        rounded
        size="small"
        :disabled="disabled"
        aria-label="Underline"
        title="Underline"
        class="rich-text-editor__underline"
        @click="runCommand('underline')"
      />
      <span class="mx-1 h-5 w-px bg-surface-200" />
      <Button
        type="button"
        icon="pi pi-list"
        text
        rounded
        size="small"
        :disabled="disabled"
        aria-label="Bullet list"
        title="Bullet list"
        @click="runCommand('insertUnorderedList')"
      />
      <Button
        type="button"
        icon="pi pi-list-check"
        text
        rounded
        size="small"
        :disabled="disabled"
        aria-label="Numbered list"
        title="Numbered list"
        @click="runCommand('insertOrderedList')"
      />
    </div>
    <div
      ref="editorRef"
      class="rich-text-editor__content rounded-b-lg px-3 py-2 text-sm leading-6 text-surface-950 outline-none"
      :class="{
        'bg-surface-100 text-surface-500': disabled,
        'rich-text-editor__content--resizable': resizable && !disabled,
      }"
      :contenteditable="!disabled"
      :data-placeholder="placeholder"
      :style="{ minHeight, maxHeight }"
      role="textbox"
      aria-multiline="true"
      @input="emitValue"
      @blur="emitValue"
      @paste="handlePaste"
    />
  </div>
</template>

<style scoped>
.rich-text-editor__bold {
  font-weight: 700;
}

.rich-text-editor__italic {
  font-style: italic;
}

.rich-text-editor__underline {
  text-decoration: underline;
}

.rich-text-editor__content:empty::before {
  color: var(--p-surface-400);
  content: attr(data-placeholder);
}

.rich-text-editor__content {
  overflow-y: auto;
  white-space: pre-wrap;
}

.rich-text-editor__content--resizable {
  resize: vertical;
}

.rich-text-editor__content :deep(p) {
  margin: 0 0 0.5rem;
}

.rich-text-editor__content :deep(p:last-child) {
  margin-bottom: 0;
}

.rich-text-editor__content :deep(ol),
.rich-text-editor__content :deep(ul) {
  margin: 0.25rem 0;
  padding-left: 1.25rem;
}
</style>
