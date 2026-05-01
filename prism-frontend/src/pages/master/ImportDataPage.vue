<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { isAxiosError } from 'axios'
import Button from 'primevue/button'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import PageHeader from '@/components/common/PageHeader.vue'
import { useToast } from '@/composables/useToast'
import {
  blueBookImportFileSchema,
  daftarKegiatanImportFileSchema,
  greenBookImportFileSchema,
  loanAgreementImportFileSchema,
  masterImportFileSchema,
} from '@/schemas/master.schema'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useDaftarKegiatanStore } from '@/stores/daftar-kegiatan.store'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useLoanAgreementStore } from '@/stores/loan-agreement.store'
import { useMasterStore } from '@/stores/master.store'
import type { ApiErrorResponse } from '@/types/api.types'
import type { BlueBook } from '@/types/blue-book.types'
import type { GreenBook } from '@/types/green-book.types'
import type {
  MasterImportRowResult,
  MasterImportRowStatus,
  MasterImportSheetResult,
  MasterImportSummary,
} from '@/types/master.types'

type ImportStatusFilter = MasterImportRowStatus | 'all'
type ImportKind = 'master' | 'blue_book' | 'green_book' | 'daftar_kegiatan' | 'loan_agreement'
type ImportRowDisplay = MasterImportRowResult & { sheet: string }
type ParsedImportInput = { file: File; blueBookId?: string; greenBookId?: string }
type ImportKindOption = { value: ImportKind; label: string; description: string }

interface ImportPageEvent {
  page: number
  rows: number
}

const masterStore = useMasterStore()
const blueBookStore = useBlueBookStore()
const greenBookStore = useGreenBookStore()
const daftarKegiatanStore = useDaftarKegiatanStore()
const loanAgreementStore = useLoanAgreementStore()
const toast = useToast()

const fileInput = ref<HTMLInputElement | null>(null)
const activeImportKind = ref<ImportKind>('master')
const selectedBlueBookId = ref<string | null>(null)
const selectedGreenBookId = ref<string | null>(null)
const selectedFile = ref<File | null>(null)
const summary = ref<MasterImportSummary | null>(null)
const executed = ref(false)
const errorMessage = ref('')
const activeRowStatus = ref<ImportStatusFilter>('all')
const activeSheetFilter = ref('all')
const previewPage = ref(1)
const previewRowsPerPage = ref(10)

const importKindOptions: ImportKindOption[] = [
  {
    value: 'master',
    label: 'Master Data',
    description: 'Program title, instansi, wilayah, periode, prioritas, lender',
  },
  {
    value: 'blue_book',
    label: 'Blue Book',
    description: 'Proyek Blue Book beserta Executing Agency, Implementing Agency, lokasi, biaya, dan indikasi lender',
  },
  {
    value: 'green_book',
    label: 'Green Book',
    description: 'Proyek Green Book beserta Proyek Blue Book, kegiatan, funding, dan alokasi',
  },
  {
    value: 'daftar_kegiatan',
    label: 'Daftar Kegiatan',
    description: 'Header Daftar Kegiatan baru beserta proyek, Proyek Green Book, pembiayaan, alokasi, dan aktivitas',
  },
  {
    value: 'loan_agreement',
    label: 'Loan Agreement',
    description: 'Create-only Loan Agreement dari Proyek Daftar Kegiatan yang eligible',
  },
]

const masterWorkbookSheets = [
  'Program Titles',
  'Bappenas Partners',
  'Institutions',
  'Regions',
  'Periods',
  'National Priorities',
  'Lenders',
]

const blueBookWorkbookSheets = [
  'Input Data',
  'Relasi - EA',
  'Relasi - IA',
  'Relasi - Locations',
  'Relasi - National Priority',
  'Relasi - Project Cost',
  'Relasi - Lender Indication',
]

const greenBookWorkbookSheets = [
  'Input Data',
  'Relasi - BB Project',
  'Relasi - EA',
  'Relasi - IA',
  'Relasi - Locations',
  'Relasi - Activities',
  'Relasi - Funding Source',
  'Relasi - Disbursement Plan',
  'Relasi - Funding Allocation',
]

const daftarKegiatanWorkbookSheets = [
  'Daftar Kegiatan',
  'Input Data',
  'Relasi - GB Project',
  'Relasi - Locations',
  'Relasi - Financing Detail',
  'Relasi - Loan Allocation',
  'Relasi - Activity Detail',
]

const loanAgreementWorkbookSheets = ['Loan Agreement']

const sheetDisplayLabels: Record<string, string> = {
  'Relasi - BB Project': 'Relasi - Proyek Blue Book',
  'Relasi - GB Project': 'Relasi - Proyek Green Book',
}

const workbookSheets = computed(() => {
  if (activeImportKind.value === 'master') return masterWorkbookSheets
  if (activeImportKind.value === 'blue_book') return blueBookWorkbookSheets
  if (activeImportKind.value === 'green_book') return greenBookWorkbookSheets
  if (activeImportKind.value === 'loan_agreement') return loanAgreementWorkbookSheets

  return daftarKegiatanWorkbookSheets
})

const blueBookOptions = computed(() =>
  blueBookStore.blueBooks.map((blueBook) => ({
    id: blueBook.id,
    label: blueBookOptionLabel(blueBook),
  })),
)

const greenBookOptions = computed(() =>
  greenBookStore.greenBooks.map((greenBook) => ({
    id: greenBook.id,
    label: greenBookOptionLabel(greenBook),
  })),
)

const selectedImportKind = computed<ImportKindOption>(
  () =>
    importKindOptions.find((item) => item.value === activeImportKind.value) ??
    (importKindOptions[0] as ImportKindOption),
)

const importBusy = computed(() =>
  activeImportKind.value === 'master'
    ? masterStore.previewing || masterStore.importing
    : activeImportKind.value === 'blue_book'
      ? blueBookStore.importPreviewing || blueBookStore.importExecuting
      : activeImportKind.value === 'green_book'
        ? greenBookStore.importPreviewing || greenBookStore.importExecuting
        : activeImportKind.value === 'loan_agreement'
          ? loanAgreementStore.importPreviewing || loanAgreementStore.importExecuting
          : daftarKegiatanStore.importPreviewing || daftarKegiatanStore.importExecuting,
)

const previewLoading = computed(() =>
  activeImportKind.value === 'master'
    ? masterStore.previewing
    : activeImportKind.value === 'blue_book'
      ? blueBookStore.importPreviewing
      : activeImportKind.value === 'green_book'
        ? greenBookStore.importPreviewing
        : activeImportKind.value === 'loan_agreement'
          ? loanAgreementStore.importPreviewing
          : daftarKegiatanStore.importPreviewing,
)

const executeLoading = computed(() =>
  activeImportKind.value === 'master'
    ? masterStore.importing
    : activeImportKind.value === 'blue_book'
      ? blueBookStore.importExecuting
      : activeImportKind.value === 'green_book'
        ? greenBookStore.importExecuting
        : activeImportKind.value === 'loan_agreement'
          ? loanAgreementStore.importExecuting
          : daftarKegiatanStore.importExecuting,
)

const templateLoading = computed(() =>
  activeImportKind.value === 'master'
    ? masterStore.downloadingTemplate
    : activeImportKind.value === 'blue_book'
      ? blueBookStore.templateDownloading
      : activeImportKind.value === 'green_book'
        ? greenBookStore.templateDownloading
        : activeImportKind.value === 'loan_agreement'
          ? loanAgreementStore.templateDownloading
          : daftarKegiatanStore.templateDownloading,
)

const targetMissing = computed(
  () =>
    (activeImportKind.value === 'blue_book' && !selectedBlueBookId.value) ||
    (activeImportKind.value === 'green_book' && !selectedGreenBookId.value),
)

const selectedFileMeta = computed(() => {
  if (!selectedFile.value) return ''
  const sizeKb = Math.max(1, Math.round(selectedFile.value.size / 1024))

  return `${selectedFile.value.name} - ${sizeKb} KB`
})

const importRows = computed<ImportRowDisplay[]>(() => {
  const rows: ImportRowDisplay[] = []

  summary.value?.sheets.forEach((sheet) => {
    sheet.rows?.forEach((row) => {
      rows.push({ ...row, sheet: sheet.sheet })
    })
  })

  return rows
})

const filteredImportRows = computed(() =>
  importRows.value.filter(
    (row) =>
      rowMatchesStatus(row, activeRowStatus.value) &&
      (activeSheetFilter.value === 'all' || row.sheet === activeSheetFilter.value),
  ),
)

const previewFirst = computed(() => (previewPage.value - 1) * previewRowsPerPage.value)
const paginatedImportRows = computed(() =>
  filteredImportRows.value.slice(previewFirst.value, previewFirst.value + previewRowsPerPage.value),
)
const previewPageStart = computed(() =>
  filteredImportRows.value.length === 0 ? 0 : previewFirst.value + 1,
)
const previewPageEnd = computed(() =>
  Math.min(previewFirst.value + previewRowsPerPage.value, filteredImportRows.value.length),
)

const rowStatusFilters = computed<
  Array<{ value: ImportStatusFilter; label: string; count: number }>
>(() => [
  { value: 'all', label: 'Semua', count: importRows.value.length },
  { value: 'create', label: 'Create', count: countRowsByStatus('create') },
  { value: 'skip', label: 'Skip', count: countRowsByStatus('skip') },
  { value: 'failed', label: 'Failed', count: countRowsByStatus('failed') },
])

const sheetFilterOptions = computed(() => {
  const statusRows = importRows.value.filter((row) => rowMatchesStatus(row, activeRowStatus.value))

  return [
    { value: 'all', label: 'Semua sheet', count: statusRows.length },
    ...(summary.value?.sheets.map((sheet) => ({
      value: sheet.sheet,
      label: sheetLabel(sheet.sheet),
      count: statusRows.filter((row) => row.sheet === sheet.sheet).length,
    })) ?? []),
  ]
})

watch([activeRowStatus, activeSheetFilter], () => {
  resetPreviewPagination()
})

watch(activeImportKind, () => {
  clearFile()
})

watch(selectedBlueBookId, () => {
  if (activeImportKind.value === 'blue_book') {
    clearPreviewResult()
  }
})

watch(selectedGreenBookId, () => {
  if (activeImportKind.value === 'green_book') {
    clearPreviewResult()
  }
})

onMounted(() => {
  void Promise.all([
    blueBookStore.fetchBlueBooks({ limit: 1000 }),
    greenBookStore.fetchGreenBooks({ limit: 1000 }),
  ])
})

function openFilePicker() {
  fileInput.value?.click()
}

function handleFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  selectedFile.value = target.files?.[0] ?? null
  clearPreviewResult()
}

function clearPreviewResult() {
  summary.value = null
  executed.value = false
  errorMessage.value = ''
  activeRowStatus.value = 'all'
  activeSheetFilter.value = 'all'
  resetPreviewPagination()
}

function clearFile() {
  selectedFile.value = null
  clearPreviewResult()

  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function cancelImport() {
  const hadPreview = Boolean(summary.value)
  clearFile()

  toast.info(
    'Import dibatalkan',
    hadPreview ? 'Preview dihapus dan data tidak dieksekusi' : 'File import dilepas',
  )
}

async function previewFile() {
  const input = getImportInput()
  if (!input) return

  errorMessage.value = ''
  executed.value = false

  try {
    if (activeImportKind.value === 'master') {
      summary.value = await masterStore.previewMasterData(input.file)
    } else if (activeImportKind.value === 'blue_book') {
      summary.value = await blueBookStore.previewProjectImport(input.blueBookId ?? '', input.file)
    } else if (activeImportKind.value === 'green_book') {
      summary.value = await greenBookStore.previewProjectImport(input.greenBookId ?? '', input.file)
    } else if (activeImportKind.value === 'loan_agreement') {
      summary.value = await loanAgreementStore.previewImport(input.file)
    } else {
      summary.value = await daftarKegiatanStore.previewImport(input.file)
    }
    setDefaultResultFilter(summary.value)
    toast.success('Preview selesai', `${summary.value.total_inserted} baris siap ditambahkan`)

    if (summary.value.total_failed > 0) {
      toast.warn('Preview menemukan error', `${summary.value.total_failed} baris perlu diperiksa`)
    }
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
}

async function executeFile() {
  if (!summary.value || summary.value.total_failed > 0) {
    errorMessage.value = 'Preview harus selesai tanpa error sebelum eksekusi import'
    return
  }

  const input = getImportInput()
  if (!input) return

  errorMessage.value = ''

  try {
    if (activeImportKind.value === 'master') {
      summary.value = await masterStore.importMasterData(input.file)
    } else if (activeImportKind.value === 'blue_book') {
      summary.value = await blueBookStore.importProjects(input.blueBookId ?? '', input.file)
    } else if (activeImportKind.value === 'green_book') {
      summary.value = await greenBookStore.importProjects(input.greenBookId ?? '', input.file)
    } else if (activeImportKind.value === 'loan_agreement') {
      summary.value = await loanAgreementStore.executeImport(input.file)
    } else {
      summary.value = await daftarKegiatanStore.executeImport(input.file)
    }
    executed.value = true
    setDefaultResultFilter(summary.value)
    toast.success('Import selesai', `${summary.value.total_inserted} baris ditambahkan`)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
}

async function downloadTemplate() {
  errorMessage.value = ''

  try {
    if (activeImportKind.value === 'master') {
      const blob = await masterStore.downloadImportTemplate()
      saveBlob(blob, 'master_data_import_template.xlsx')
      toast.success('Template diunduh', 'Template Master Data sudah dibuat dari snapshot terbaru')
      return
    }

    if (activeImportKind.value === 'blue_book') {
      if (!selectedBlueBookId.value) {
        errorMessage.value = 'Pilih target Blue Book sebelum download template'
        return
      }

      const blob = await blueBookStore.downloadProjectImportTemplate(selectedBlueBookId.value)
      saveBlob(blob, 'blue_book_import_template.xlsx')
      toast.success('Template diunduh', 'Template Blue Book sudah dibuat dari snapshot master data')
      return
    }

    if (activeImportKind.value === 'green_book') {
      if (!selectedGreenBookId.value) {
        errorMessage.value = 'Pilih target Green Book sebelum download template'
        return
      }

      const blob = await greenBookStore.downloadProjectImportTemplate(selectedGreenBookId.value)
      saveBlob(blob, 'green_book_import_template.xlsx')
      toast.success('Template diunduh', 'Template Green Book sudah dibuat dari snapshot master data')
      return
    }

    if (activeImportKind.value === 'loan_agreement') {
      const laBlob = await loanAgreementStore.downloadImportTemplate()
      saveBlob(laBlob, 'loan_agreement_import_template.xlsx')
      toast.success('Template diunduh', 'Template Loan Agreement sudah dibuat dari snapshot master data')
      return
    }

    const dkBlob = await daftarKegiatanStore.downloadImportTemplate()
    saveBlob(dkBlob, 'daftar_kegiatan_import_template.xlsx')
    toast.success('Template diunduh', 'Template Daftar Kegiatan sudah dibuat dari snapshot master data')
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
}

function getImportInput(): ParsedImportInput | null {
  if (activeImportKind.value === 'master') {
    const parsed = masterImportFileSchema.safeParse({ file: selectedFile.value })
    if (!parsed.success) {
      errorMessage.value = parsed.error.issues[0]?.message ?? 'File tidak valid'
      return null
    }

    return { file: parsed.data.file }
  }

  if (activeImportKind.value === 'blue_book') {
    const parsed = blueBookImportFileSchema.safeParse({
      file: selectedFile.value,
      blue_book_id: selectedBlueBookId.value ?? '',
    })
    if (!parsed.success) {
      errorMessage.value = parsed.error.issues[0]?.message ?? 'File tidak valid'
      return null
    }

    return { file: parsed.data.file, blueBookId: parsed.data.blue_book_id }
  }

  if (activeImportKind.value === 'green_book') {
    const parsed = greenBookImportFileSchema.safeParse({
      file: selectedFile.value,
      green_book_id: selectedGreenBookId.value ?? '',
    })
    if (!parsed.success) {
      errorMessage.value = parsed.error.issues[0]?.message ?? 'File tidak valid'
      return null
    }

    return { file: parsed.data.file, greenBookId: parsed.data.green_book_id }
  }

  if (activeImportKind.value === 'loan_agreement') {
    const parsed = loanAgreementImportFileSchema.safeParse({ file: selectedFile.value })
    if (!parsed.success) {
      errorMessage.value = parsed.error.issues[0]?.message ?? 'File tidak valid'
      return null
    }

    return { file: parsed.data.file }
  }

  const parsed = daftarKegiatanImportFileSchema.safeParse({ file: selectedFile.value })
  if (!parsed.success) {
    errorMessage.value = parsed.error.issues[0]?.message ?? 'File tidak valid'
    return null
  }

  return { file: parsed.data.file }
}

function getErrorMessage(error: unknown) {
  if (isAxiosError<ApiErrorResponse>(error)) {
    return error.response?.data?.error?.message ?? 'Import gagal diproses'
  }

  return 'Import gagal diproses'
}

function sheetSeverity(sheet: MasterImportSheetResult) {
  if (sheet.failed > 0) return 'danger'
  if (sheet.inserted > 0) return 'success'

  return 'secondary'
}

function rowMatchesStatus(row: MasterImportRowResult, status: ImportStatusFilter) {
  return status === 'all' || row.status === status
}

function countRowsByStatus(status: ImportStatusFilter) {
  return importRows.value.filter((row) => rowMatchesStatus(row, status)).length
}

function setDefaultResultFilter(importSummary: MasterImportSummary) {
  activeSheetFilter.value = 'all'

  if (importSummary.total_failed > 0) {
    activeRowStatus.value = 'failed'
    return
  }
  if (importSummary.total_inserted > 0) {
    activeRowStatus.value = 'create'
    return
  }
  if (importSummary.total_skipped > 0) {
    activeRowStatus.value = 'skip'
    return
  }

  activeRowStatus.value = 'all'
}

function handlePreviewPage(event: ImportPageEvent) {
  previewPage.value = event.page + 1
  previewRowsPerPage.value = event.rows
}

function resetPreviewPagination() {
  previewPage.value = 1
}

function rowStatusLabel(status: MasterImportRowStatus) {
  if (status === 'create') return 'Create'
  if (status === 'skip') return 'Skip'

  return 'Failed'
}

function rowStatusSeverity(status: MasterImportRowStatus) {
  if (status === 'create') return 'success'
  if (status === 'skip') return 'secondary'

  return 'danger'
}

function rowNumberLabel(row: number) {
  return row > 0 ? row.toString() : 'Otomatis'
}

function sheetLabel(sheet: string) {
  return sheetDisplayLabels[sheet] ?? sheet
}

function blueBookOptionLabel(blueBook: BlueBook) {
  const revision =
    blueBook.revision_number > 0
      ? `Revisi ${blueBook.revision_number}${blueBook.revision_year ? `/${blueBook.revision_year}` : ''}`
      : 'Awal'

  return `${blueBook.period.name} - ${revision} - ${blueBook.publish_date}`
}

function greenBookOptionLabel(greenBook: GreenBook) {
  const revision = greenBook.revision_number > 0 ? `Revisi ${greenBook.revision_number}` : 'Awal'

  return `Green Book ${greenBook.publish_year} - ${revision}`
}

function saveBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

function filterButtonClass(active: boolean) {
  return [
    'rounded-md border px-3 py-2 text-left text-sm transition-colors',
    active
      ? 'border-primary bg-primary text-white'
      : 'border-surface-200 bg-white text-surface-700 hover:border-primary hover:text-primary',
  ]
}
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Import Data"
      subtitle="Preview workbook Excel sebelum eksekusi import data PRISM"
    />

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_24rem]">
      <div class="space-y-5 rounded-lg border border-surface-200 bg-white p-5">
        <div class="grid gap-3 md:grid-cols-2">
          <button
            v-for="kind in importKindOptions"
            :key="kind.value"
            type="button"
            :class="filterButtonClass(activeImportKind === kind.value)"
            @click="activeImportKind = kind.value"
          >
            <span class="block font-semibold">{{ kind.label }}</span>
            <span class="text-xs opacity-80">{{ kind.description }}</span>
          </button>
        </div>

        <label v-if="activeImportKind === 'blue_book'" class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Target Blue Book</span>
          <Select
            v-model="selectedBlueBookId"
            :options="blueBookOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih Blue Book"
            class="w-full"
            :loading="blueBookStore.loading"
          />
        </label>

        <label v-if="activeImportKind === 'green_book'" class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Target Green Book</span>
          <Select
            v-model="selectedGreenBookId"
            :options="greenBookOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih Green Book"
            class="w-full"
            :loading="greenBookStore.loading"
          />
        </label>

        <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div class="min-w-0">
            <p class="text-sm font-semibold text-surface-900">
              Workbook Excel - {{ selectedImportKind.label }}
            </p>
            <p class="mt-1 truncate text-sm text-surface-500">
              {{ selectedFileMeta || 'Belum ada file dipilih' }}
            </p>
          </div>

          <div class="flex flex-wrap gap-2">
            <input
              ref="fileInput"
              type="file"
              accept=".xlsx"
              class="hidden"
              @change="handleFileChange"
            />
            <Button
              label="Template"
              icon="pi pi-download"
              outlined
              :disabled="
                importBusy ||
                templateLoading ||
                targetMissing
              "
              :loading="templateLoading"
              @click="downloadTemplate"
            />
            <Button label="Pilih File" icon="pi pi-folder-open" outlined @click="openFilePicker" />
            <Button
              label="Preview"
              icon="pi pi-eye"
              :disabled="
                !selectedFile ||
                importBusy ||
                targetMissing
              "
              :loading="previewLoading"
              outlined
              @click="previewFile"
            />
            <Button
              label="Eksekusi"
              icon="pi pi-cloud-upload"
              :disabled="
                !selectedFile ||
                !summary ||
                summary.total_failed > 0 ||
                executed ||
                importBusy ||
                targetMissing
              "
              :loading="executeLoading"
              @click="executeFile"
            />
            <Button
              v-if="selectedFile && !executed"
              label="Batal"
              icon="pi pi-times"
              severity="secondary"
              outlined
              :disabled="importBusy"
              @click="cancelImport"
            />
          </div>
        </div>

        <Message v-if="errorMessage" severity="error" class="mt-4" :closable="false">
          {{ errorMessage }}
        </Message>
      </div>

      <aside class="rounded-lg border border-surface-200 bg-white p-5">
        <p class="text-sm font-semibold text-surface-900">Sheet yang dibaca</p>
        <p class="mt-1 text-sm text-surface-500">{{ selectedImportKind.description }}</p>
        <div class="mt-3 flex flex-wrap gap-2">
          <Tag
            v-for="sheet in workbookSheets"
            :key="sheet"
            :value="sheetLabel(sheet)"
            severity="info"
            rounded
          />
        </div>
      </aside>
    </div>

    <div v-if="summary" class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">
            {{ executed ? 'Hasil Eksekusi Import' : 'Preview Import' }}
          </h2>
          <p class="mt-1 text-sm text-surface-500">{{ summary.file_name }}</p>
        </div>

        <div class="grid grid-cols-3 gap-2 text-center">
          <div class="rounded-md bg-prism-green/10 px-4 py-2 text-prism-green-dark">
            <p class="text-lg font-semibold">{{ summary.total_inserted }}</p>
            <p class="text-xs">Ditambah</p>
          </div>
          <div class="rounded-md bg-surface-100 px-4 py-2 text-surface-700">
            <p class="text-lg font-semibold">{{ summary.total_skipped }}</p>
            <p class="text-xs">Skip</p>
          </div>
          <div class="rounded-md bg-red-50 px-4 py-2 text-red-700">
            <p class="text-lg font-semibold">{{ summary.total_failed }}</p>
            <p class="text-xs">Gagal</p>
          </div>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-surface-200 text-sm">
          <thead class="bg-surface-50 text-left text-xs uppercase text-surface-500">
            <tr>
              <th class="px-3 py-2 font-semibold">Sheet</th>
              <th class="px-3 py-2 font-semibold">Status</th>
              <th class="px-3 py-2 font-semibold">Ditambah</th>
              <th class="px-3 py-2 font-semibold">Skip</th>
              <th class="px-3 py-2 font-semibold">Gagal</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-surface-100">
            <tr v-for="sheet in summary.sheets" :key="sheet.sheet">
              <td class="px-3 py-2 font-medium text-surface-900">{{ sheetLabel(sheet.sheet) }}</td>
              <td class="px-3 py-2">
                <Tag
                  :value="sheet.failed > 0 ? 'Perlu cek' : 'Selesai'"
                  :severity="sheetSeverity(sheet)"
                  rounded
                />
              </td>
              <td class="px-3 py-2">{{ sheet.inserted }}</td>
              <td class="px-3 py-2">{{ sheet.skipped }}</td>
              <td class="px-3 py-2">{{ sheet.failed }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="space-y-3 border-t border-surface-100 pt-4">
        <div class="flex flex-col gap-1">
          <h3 class="text-sm font-semibold text-surface-900">Detail baris preview</h3>
          <p class="text-sm text-surface-500">
            Rincian baris berdasarkan status create, skip, dan failed.
          </p>
        </div>

        <div class="flex flex-col gap-3 xl:flex-row xl:items-start xl:justify-between">
          <div class="flex flex-wrap gap-2" role="tablist" aria-label="Filter status import">
            <button
              v-for="filter in rowStatusFilters"
              :key="filter.value"
              type="button"
              role="tab"
              :aria-selected="activeRowStatus === filter.value"
              :class="filterButtonClass(activeRowStatus === filter.value)"
              @click="activeRowStatus = filter.value"
            >
              <span class="block font-medium">{{ filter.label }}</span>
              <span class="text-xs opacity-80">{{ filter.count }} baris</span>
            </button>
          </div>

          <div class="flex max-w-full flex-wrap gap-2 xl:justify-end">
            <button
              v-for="sheet in sheetFilterOptions"
              :key="sheet.value"
              type="button"
              :class="filterButtonClass(activeSheetFilter === sheet.value)"
              @click="activeSheetFilter = sheet.value"
            >
              <span class="block font-medium">{{ sheet.label }}</span>
              <span class="text-xs opacity-80">{{ sheet.count }} baris</span>
            </button>
          </div>
        </div>

        <div class="overflow-x-auto rounded-lg border border-surface-200">
          <table class="min-w-full divide-y divide-surface-200 text-sm">
            <thead class="bg-surface-50 text-left text-xs uppercase text-surface-500">
              <tr>
                <th class="px-3 py-2 font-semibold">Sheet</th>
                <th class="px-3 py-2 font-semibold">Baris</th>
                <th class="px-3 py-2 font-semibold">Status</th>
                <th class="px-3 py-2 font-semibold">Data</th>
                <th class="px-3 py-2 font-semibold">Keterangan</th>
              </tr>
            </thead>
            <tbody v-if="filteredImportRows.length" class="divide-y divide-surface-100">
              <tr
                v-for="row in paginatedImportRows"
                :key="`${row.sheet}-${row.row}-${row.status}-${row.label}`"
              >
                <td class="px-3 py-2 font-medium text-surface-900">{{ sheetLabel(row.sheet) }}</td>
                <td class="px-3 py-2 text-surface-600">{{ rowNumberLabel(row.row) }}</td>
                <td class="px-3 py-2">
                  <Tag
                    :value="rowStatusLabel(row.status)"
                    :severity="rowStatusSeverity(row.status)"
                    rounded
                  />
                </td>
                <td class="max-w-[32rem] px-3 py-2 text-surface-800">{{ row.label }}</td>
                <td class="max-w-[32rem] px-3 py-2 text-surface-600">
                  {{ row.message || '-' }}
                </td>
              </tr>
            </tbody>
            <tbody v-else>
              <tr>
                <td colspan="5" class="px-3 py-6 text-center text-sm text-surface-500">
                  Tidak ada baris untuk filter ini.
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          v-if="filteredImportRows.length"
          class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
        >
          <p class="text-sm text-surface-500">
            Menampilkan {{ previewPageStart }}-{{ previewPageEnd }} dari
            {{ filteredImportRows.length }} baris
          </p>
          <Paginator
            :first="previewFirst"
            :rows="previewRowsPerPage"
            :total-records="filteredImportRows.length"
            :rows-per-page-options="[10, 25, 50, 100]"
            @page="handlePreviewPage"
          />
        </div>
      </div>

      <div v-if="summary.total_failed > 0" class="space-y-3">
        <div
          v-for="sheet in summary.sheets.filter((item) => item.errors?.length)"
          :key="`${sheet.sheet}-errors`"
          class="rounded-lg border border-red-100 bg-red-50 p-4"
        >
          <p class="text-sm font-semibold text-red-800">{{ sheetLabel(sheet.sheet) }}</p>
          <ul class="mt-2 space-y-1 text-sm text-red-700">
            <li v-for="error in sheet.errors" :key="`${sheet.sheet}-${error.row}-${error.message}`">
              Baris {{ error.row }}: {{ error.message }}
            </li>
          </ul>
        </div>
      </div>
    </div>
  </section>
</template>
