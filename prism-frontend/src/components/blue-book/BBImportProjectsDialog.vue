<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import type {
  BBProject,
  BBProjectRevisionSourceOption,
  BlueBook,
  ImportBBProjectsFromBlueBookPayload,
} from '@/types/blue-book.types'
import { formatBlueBookStatus, formatRevision } from '@/pages/blue-book/blue-book-page-utils'

const props = defineProps<{
  visible: boolean
  /** The Blue Book we are importing INTO */
  targetBlueBook: BlueBook | null
  getBlueBooksByPeriod: (periodId: string) => Promise<BlueBook[]>
  getProjectsByBlueBook: (blueBookId: string) => Promise<BBProject[]>
  importProjectsFromBlueBook: (
    blueBookId: string,
    payload: ImportBBProjectsFromBlueBookPayload,
  ) => Promise<void>
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  imported: []
}>()

const importSourceBlueBookOptions = ref<BlueBook[]>([])
const importSourceBlueBookLoading = ref(false)
const importProjectOptions = ref<BBProjectRevisionSourceOption[]>([])
const importProjectSearchQuery = ref('')
const importProjectLoading = ref(false)
const importError = ref<{ source_blue_book_id?: string; project_ids?: string }>({})

const importForm = reactive<ImportBBProjectsFromBlueBookPayload>({
  source_blue_book_id: '',
  project_ids: [],
})

const isRevisionBlueBook = computed(
  () =>
    Boolean(props.targetBlueBook?.replaces_blue_book_id) ||
    Number(props.targetBlueBook?.revision_number ?? 0) > 0,
)

const importSourceBlueBookSelectOptions = computed(() =>
  importSourceBlueBookOptions.value.map((blueBook) => ({
    ...blueBook,
    label: sourceBlueBookLabel(blueBook),
  })),
)

const importableProjectOptions = computed(() =>
  importProjectOptions.value.filter((project) => !project.disabled),
)

const hasImportableProjectOptions = computed(() => importableProjectOptions.value.length > 0)

const filteredImportProjectOptions = computed(() => {
  const query = normalizeSearchText(importProjectSearchQuery.value)
  if (!query) return importProjectOptions.value
  return importProjectOptions.value.filter((project) => projectMatchesImportSearch(project, query))
})

const filteredImportableProjectOptions = computed(() =>
  filteredImportProjectOptions.value.filter((project) => !project.disabled),
)

const hasFilteredImportableProjectOptions = computed(
  () => filteredImportableProjectOptions.value.length > 0,
)

const selectedImportProjectCount = computed(() => {
  const importableProjectIds = new Set(importableProjectOptions.value.map((project) => project.id))
  return importForm.project_ids.filter((projectId) => importableProjectIds.has(projectId)).length
})

const importProjectSelectionSummary = computed(() =>
  importableProjectOptions.value.length > 0
    ? `${selectedImportProjectCount.value} / ${importableProjectOptions.value.length} dipilih`
    : '0 proyek bisa diimpor',
)

const allFilteredImportProjectsSelected = computed(
  () =>
    hasFilteredImportableProjectOptions.value &&
    filteredImportableProjectOptions.value.every((project) =>
      importForm.project_ids.includes(project.id),
    ),
)

const importSubmitLabel = computed(() =>
  selectedImportProjectCount.value > 0
    ? `Impor ${selectedImportProjectCount.value} Proyek`
    : 'Impor Proyek',
)

function sourceBlueBookLabel(blueBook: BlueBook) {
  return `${blueBook.period.name} - ${formatRevision(blueBook.revision_number, blueBook.revision_year)} (${formatBlueBookStatus(blueBook.status)})`
}

function normalizeProjectCode(code: string) {
  return code.trim().toLowerCase()
}

function normalizeSearchText(value: string) {
  return value.trim().toLowerCase().replace(/\s+/g, ' ')
}

function projectMatchesImportSearch(project: BBProjectRevisionSourceOption, query: string) {
  const searchable = [
    project.bb_code,
    project.project_name,
    project.source_blue_book_label,
    project.unavailable_reason ?? '',
  ]
    .join(' ')
    .toLowerCase()

  return searchable.includes(query)
}

function isImportProjectSelected(projectId: string) {
  return importForm.project_ids.includes(projectId)
}

function setImportProjectSelected(project: BBProjectRevisionSourceOption, selected: boolean) {
  if (project.disabled) return
  const ids = new Set(importForm.project_ids)
  if (selected) ids.add(project.id)
  else ids.delete(project.id)
  importForm.project_ids = [...ids]
}

function setAllFilteredImportProjectsSelected(selected: boolean) {
  const ids = new Set(importForm.project_ids)
  filteredImportableProjectOptions.value.forEach((project) => {
    if (selected) ids.add(project.id)
    else ids.delete(project.id)
  })
  importForm.project_ids = [...ids]
}

function selectAllImportProjects() {
  importForm.project_ids = importableProjectOptions.value.map((project) => project.id)
}

function clearAllImportProjects() {
  importForm.project_ids = []
}

function clearImportProjects() {
  importProjectOptions.value = []
  importForm.project_ids = []
  importProjectSearchQuery.value = ''
}

function compareBlueBookVersionDesc(left: BlueBook, right: BlueBook) {
  if (left.revision_number !== right.revision_number) {
    return right.revision_number - left.revision_number
  }
  const leftYear = left.revision_year ?? 0
  const rightYear = right.revision_year ?? 0
  if (leftYear !== rightYear) return rightYear - leftYear
  return String(right.created_at ?? '').localeCompare(String(left.created_at ?? ''))
}

function isSourceBlueBook(blueBook: BlueBook, targetBlueBook: BlueBook) {
  if (blueBook.id === targetBlueBook.id || blueBook.period.id !== targetBlueBook.period.id) {
    return false
  }
  if (blueBook.id === targetBlueBook.replaces_blue_book_id) return true
  if (blueBook.revision_number < targetBlueBook.revision_number) return true
  return (
    blueBook.revision_number === targetBlueBook.revision_number &&
    Boolean(blueBook.created_at) &&
    Boolean(targetBlueBook.created_at) &&
    String(blueBook.created_at) < String(targetBlueBook.created_at)
  )
}

async function loadImportSourceBlueBooks() {
  const current = props.targetBlueBook
  if (!current || !isRevisionBlueBook.value) {
    importSourceBlueBookOptions.value = []
    importForm.source_blue_book_id = ''
    clearImportProjects()
    return
  }

  importSourceBlueBookLoading.value = true
  try {
    const samePeriodBlueBooks = await props.getBlueBooksByPeriod(current.period.id)
    importSourceBlueBookOptions.value = samePeriodBlueBooks
      .filter((blueBook) => isSourceBlueBook(blueBook, current))
      .sort(compareBlueBookVersionDesc)

    const currentSourceStillAvailable = importSourceBlueBookOptions.value.some(
      (blueBook) => blueBook.id === importForm.source_blue_book_id,
    )
    if (!currentSourceStillAvailable) {
      importForm.source_blue_book_id = importSourceBlueBookOptions.value[0]?.id ?? ''
      if (!importForm.source_blue_book_id) {
        clearImportProjects()
        return
      }
    }

    await loadImportProjects(importForm.source_blue_book_id)
  } finally {
    importSourceBlueBookLoading.value = false
  }
}

async function loadImportProjects(sourceBlueBookId?: string) {
  importProjectSearchQuery.value = ''
  if (!sourceBlueBookId) {
    clearImportProjects()
    return
  }

  const current = props.targetBlueBook
  if (!current) {
    clearImportProjects()
    return
  }

  importProjectLoading.value = true
  try {
    const [sourceProjects, currentProjects] = await Promise.all([
      props.getProjectsByBlueBook(sourceBlueBookId),
      props.getProjectsByBlueBook(current.id),
    ])
    const sourceBlueBook = importSourceBlueBookOptions.value.find(
      (blueBook) => blueBook.id === sourceBlueBookId,
    )
    const usedIdentityIds = new Set(currentProjects.map((project) => project.project_identity_id))
    const usedCodes = new Set(currentProjects.map((project) => normalizeProjectCode(project.bb_code)))

    importProjectOptions.value = sourceProjects.map((project) => {
      let unavailableReason: string | null = null
      if (usedIdentityIds.has(project.project_identity_id)) {
        unavailableReason = 'Sudah ada di Blue Book tujuan'
      } else if (usedCodes.has(normalizeProjectCode(project.bb_code))) {
        unavailableReason = 'Kode Blue Book sudah ada di Blue Book tujuan'
      }

      return {
        ...project,
        source_blue_book_id: sourceBlueBookId,
        source_blue_book_label: sourceBlueBook ? sourceBlueBookLabel(sourceBlueBook) : '',
        disabled: Boolean(unavailableReason),
        unavailable_reason: unavailableReason ?? undefined,
      }
    })

    importForm.project_ids = importProjectOptions.value
      .filter((project) => !project.disabled)
      .map((project) => project.id)
  } finally {
    importProjectLoading.value = false
  }
}

async function handleSubmit() {
  importError.value = {}

  if (!importForm.source_blue_book_id) {
    importError.value = { source_blue_book_id: 'Pilih Blue Book sumber terlebih dahulu' }
    return
  }

  const importableProjectIds = new Set(
    importProjectOptions.value.filter((project) => !project.disabled).map((project) => project.id),
  )
  const projectIds = importForm.project_ids.filter((projectId) => importableProjectIds.has(projectId))
  if (projectIds.length === 0) {
    importError.value = {
      project_ids: 'Pilih minimal satu Project Blue Book yang belum ada di Blue Book tujuan',
    }
    return
  }

  await props.importProjectsFromBlueBook(props.targetBlueBook!.id, {
    source_blue_book_id: importForm.source_blue_book_id,
    project_ids: projectIds,
  })
  emit('update:visible', false)
  emit('imported')
}

function handleShow() {
  importError.value = {}
  importForm.source_blue_book_id = ''
  importForm.project_ids = []
  importProjectSearchQuery.value = ''
  void loadImportSourceBlueBooks()
}
</script>

<template>
  <Dialog
    :visible="visible"
    modal
    header="Impor Proyek dari Blue Book Lain"
    class="w-[64rem] max-w-[calc(100vw-2rem)]"
    @update:visible="emit('update:visible', $event)"
    @show="handleShow"
  >
    <form class="flex flex-col gap-5" @submit.prevent="handleSubmit">
      <label class="grid gap-2">
        <span class="text-sm font-medium text-surface-700">Blue Book Sumber</span>
        <Select
          v-model="importForm.source_blue_book_id"
          :options="importSourceBlueBookSelectOptions"
          option-label="label"
          option-value="id"
          placeholder="Pilih Blue Book sumber"
          filter
          append-to="body"
          class="w-full"
          :loading="importSourceBlueBookLoading"
          :invalid="Boolean(importError.source_blue_book_id)"
          @change="loadImportProjects(importForm.source_blue_book_id)"
        />
        <small v-if="importError.source_blue_book_id" class="text-red-600">
          {{ importError.source_blue_book_id }}
        </small>
        <small
          v-else-if="!importSourceBlueBookLoading && importSourceBlueBookOptions.length === 0"
          class="text-surface-500"
        >
          Belum ada Blue Book sumber pada periode ini.
        </small>
      </label>

      <section class="grid gap-3">
        <div class="flex flex-col gap-3 md:flex-row md:items-center">
          <label class="relative min-w-0 flex-1">
            <span class="sr-only">Cari Project Blue Book</span>
            <i class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400" />
            <InputText
              v-model="importProjectSearchQuery"
              placeholder="Cari kode atau nama proyek..."
              class="w-full pl-10"
              :disabled="!importForm.source_blue_book_id || importProjectLoading"
            />
          </label>
          <p class="shrink-0 text-sm font-semibold text-surface-600">
            {{ importProjectSelectionSummary }}
          </p>
        </div>

        <div class="overflow-hidden rounded-lg border border-surface-200 bg-surface-0">
          <div class="max-h-[22rem] overflow-auto">
            <table class="min-w-full table-fixed text-sm">
              <thead class="sticky top-0 z-10 bg-surface-50 text-left text-xs font-semibold uppercase tracking-wide text-surface-500">
                <tr class="border-b border-surface-200">
                  <th class="w-12 px-4 py-3">
                    <Checkbox
                      binary
                      :model-value="allFilteredImportProjectsSelected"
                      :disabled="!hasFilteredImportableProjectOptions"
                      @update:model-value="setAllFilteredImportProjectsSelected(Boolean($event))"
                    />
                  </th>
                  <th class="w-44 px-3 py-3">Kode Proyek</th>
                  <th class="px-3 py-3">Nama Proyek</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="importProjectLoading">
                  <td colspan="3" class="px-4 py-8 text-center text-sm text-surface-500">
                    Memuat Project Blue Book...
                  </td>
                </tr>
                <tr v-else-if="!importForm.source_blue_book_id">
                  <td colspan="3" class="px-4 py-8 text-center text-sm text-surface-500">
                    Pilih Blue Book sumber untuk melihat Project Blue Book.
                  </td>
                </tr>
                <tr v-else-if="importProjectOptions.length === 0">
                  <td colspan="3" class="px-4 py-8 text-center text-sm text-surface-500">
                    Blue Book sumber ini belum memiliki Project Blue Book.
                  </td>
                </tr>
                <tr v-else-if="filteredImportProjectOptions.length === 0">
                  <td colspan="3" class="px-4 py-8 text-center text-sm text-surface-500">
                    Tidak ada Project Blue Book yang cocok dengan pencarian.
                  </td>
                </tr>
                <template v-else>
                  <tr
                    v-for="option in filteredImportProjectOptions"
                    :key="option.id"
                    class="border-b border-surface-100 last:border-b-0"
                    :class="
                      option.disabled
                        ? 'bg-surface-50 text-surface-400'
                        : isImportProjectSelected(option.id)
                          ? 'bg-primary-50/60 text-surface-950'
                          : 'bg-surface-0 text-surface-800 hover:bg-surface-50'
                    "
                  >
                    <td class="px-4 py-3 align-top">
                      <Checkbox
                        binary
                        :model-value="isImportProjectSelected(option.id)"
                        :disabled="option.disabled"
                        @update:model-value="setImportProjectSelected(option, Boolean($event))"
                      />
                    </td>
                    <td class="px-3 py-3 align-top">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="rounded border border-surface-200 bg-surface-0 px-2 py-0.5 font-mono text-xs font-semibold text-surface-700">
                          {{ option.bb_code }}
                        </span>
                        <span
                          v-if="isImportProjectSelected(option.id) && !option.disabled"
                          class="rounded bg-primary-100 px-2 py-0.5 text-xs font-medium text-primary-700"
                        >
                          Dipilih
                        </span>
                        <span
                          v-if="option.unavailable_reason"
                          class="rounded bg-amber-50 px-2 py-0.5 text-xs font-medium text-amber-700"
                        >
                          Sudah ada
                        </span>
                      </div>
                    </td>
                    <td class="px-3 py-3 align-top">
                      <p class="font-medium leading-snug">{{ option.project_name }}</p>
                      <p class="mt-1 text-xs text-surface-500">{{ option.source_blue_book_label }}</p>
                      <p v-if="option.unavailable_reason" class="mt-1 text-xs text-amber-700">
                        {{ option.unavailable_reason }}
                      </p>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>

        <p v-if="importError.project_ids" class="text-sm text-red-600">
          {{ importError.project_ids }}
        </p>
        <p
          v-else-if="
            importForm.source_blue_book_id &&
            !importProjectLoading &&
            importProjectOptions.length > 0 &&
            !hasImportableProjectOptions
          "
          class="rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-700"
        >
          Semua Project Blue Book dari sumber ini sudah ada di Blue Book tujuan.
        </p>
      </section>

      <div class="flex flex-col gap-3 border-t border-surface-200 pt-5 md:flex-row md:items-center md:justify-between">
        <div class="flex flex-wrap gap-2">
          <Button
            type="button"
            label="Pilih Semua"
            severity="secondary"
            outlined
            :disabled="!hasImportableProjectOptions || selectedImportProjectCount === importableProjectOptions.length"
            @click="selectAllImportProjects"
          />
          <Button
            type="button"
            label="Hapus Semua"
            severity="secondary"
            outlined
            :disabled="selectedImportProjectCount === 0"
            @click="clearAllImportProjects"
          />
        </div>
        <div class="flex justify-end gap-2">
          <Button
            type="button"
            label="Batal"
            severity="secondary"
            outlined
            @click="emit('update:visible', false)"
          />
          <Button
            type="submit"
            :label="importSubmitLabel"
            icon="pi pi-file-import"
            :disabled="
              !importForm.source_blue_book_id ||
              selectedImportProjectCount === 0 ||
              !hasImportableProjectOptions
            "
          />
        </div>
      </div>
    </form>
  </Dialog>
</template>
