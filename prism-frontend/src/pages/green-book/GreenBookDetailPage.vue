<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { greenBookSchema, importGBProjectsFromGreenBookSchema } from '@/schemas/green-book.schema'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useMasterStore } from '@/stores/master.store'
import type {
  GBProject,
  GBProjectListParams,
  GBProjectRevisionSourceOption,
  GreenBook,
  GreenBookPayload,
  GreenBookStatus,
  ImportGBProjectsFromGreenBookPayload,
} from '@/types/green-book.types'
import type { Institution, Region } from '@/types/master.types'
import { formatApiError } from '@/utils/api-error'
import {
  formatGBRevision,
  formatGreenBookStatus,
  joinNames,
  toFormErrors,
  type FormErrors,
} from './green-book-page-utils'

type GreenBookField = keyof GreenBookPayload
type ImportFromGreenBookField = keyof ImportGBProjectsFromGreenBookPayload

const route = useRoute()
const router = useRouter()
const greenBookStore = useGreenBookStore()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
interface GBProjectFilterState {
  bb_project_ids: string[]
  executing_agency_ids: string[]
  location_ids: string[]
}

const greenBookId = computed(() => String(route.params.id ?? ''))
const projectControls = useListControls<GBProjectFilterState>({
  initialFilters: {
    bb_project_ids: [],
    executing_agency_ids: [],
    location_ids: [],
  },
  filterLabels: {
    bb_project_ids: 'Proyek Blue Book',
    executing_agency_ids: 'Executing Agency',
    location_ids: 'Location',
  },
  formatFilterValue: (key, value) => {
    if (key === 'bb_project_ids' && Array.isArray(value)) return selectedBBProjectSummary(value)
    if (key === 'executing_agency_ids' && Array.isArray(value)) return selectedInstitutionSummary(value)
    if (key === 'location_ids' && Array.isArray(value)) return selectedRegionSummary(value)
    return Array.isArray(value) ? selectedLabelSummary(value) : String(value)
  },
})
const dialogVisible = ref(false)
const importDialogVisible = ref(false)
const form = reactive<GreenBookPayload>({
  publish_year: new Date().getFullYear(),
  revision_number: 0,
  status: 'active',
})
const errors = ref<FormErrors<GreenBookField>>({})
const importErrors = ref<FormErrors<ImportFromGreenBookField>>({})
const importSourceGreenBookOptions = ref<GreenBook[]>([])
const importSourceGreenBookLoading = ref(false)
const importProjectOptions = ref<GBProjectRevisionSourceOption[]>([])
const importProjectSearchQuery = ref('')
const importProjectLoading = ref(false)
const importForm = reactive<ImportGBProjectsFromGreenBookPayload>({
  source_green_book_id: '',
  project_ids: [],
})
const columns: ColumnDef[] = [
  { field: 'gb_code', header: 'Kode Green Book', sortable: true },
  { field: 'project_name', header: 'Nama Proyek', sortable: true },
  { field: 'bb_projects', header: 'Proyek Blue Book', sortable: true },
  { field: 'status', header: 'Status', sortable: true },
  { field: 'actions', header: 'Aksi', align: 'right' },
]
const statusOptions: Array<{ label: string; value: GreenBookStatus }> = [
  { label: 'Berlaku', value: 'active' },
  { label: 'Tidak Berlaku', value: 'superseded' },
]
const selectedCountryCodes = computed(() => {
  const selected = new Set(projectControls.draftFilters.location_ids)

  return masterStore.regions
    .filter((region) => region.type === 'COUNTRY' && selected.has(region.id))
    .map((region) => region.code)
})
const bbProjectOptions = computed(() =>
  blueBookStore.projectOptions.map((project) => ({
    ...project,
    label: `${project.bb_code} - ${project.project_name}`,
    value: project.id,
  })),
)
const institutionFilterOptions = computed(() =>
  masterStore.institutions.map((institution) => ({
    ...institution,
    label: formatInstitution(institution),
    value: institution.id,
  })),
)
const locationFilterOptions = computed(() =>
  masterStore.regions.map((region) => ({
    ...region,
    label: formatRegion(region),
    value: region.id,
    disabled:
      selectedCountryCodes.value.length > 0 &&
      region.type !== 'COUNTRY' &&
      isCoveredBySelectedCountry(region),
  })),
)
const canDeleteCurrentGreenBook = computed(
  () => can('green_book', 'delete') && (greenBookStore.currentGreenBook?.project_count ?? 0) === 0,
)
const canImportProjectsFromGreenBook = computed(
  () => can('gb_project', 'create') && Boolean(greenBookStore.currentGreenBook),
)
const importSourceGreenBookSelectOptions = computed(() =>
  importSourceGreenBookOptions.value.map((greenBook) => ({
    ...greenBook,
    label: sourceGreenBookLabel(greenBook),
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
    : '0 proyek bisa ditambahkan',
)
const allFilteredImportProjectsSelected = computed(
  () =>
    hasFilteredImportableProjectOptions.value &&
    filteredImportableProjectOptions.value.every((project) => importForm.project_ids.includes(project.id)),
)
const importSubmitLabel = computed(() =>
  selectedImportProjectCount.value > 0
    ? `Tambahkan ${selectedImportProjectCount.value} Proyek`
    : 'Tambahkan Proyek',
)

function buildProjectParams(): GBProjectListParams {
  const params = projectControls.buildParams() as GBProjectListParams
  const locationIDs = expandLocationFilterIds(projectControls.appliedFilters.location_ids)
  if (locationIDs.length > 0) {
    params.location_ids = locationIDs
  }
  return params
}

async function loadData() {
  await Promise.all([
    greenBookStore.fetchGreenBook(greenBookId.value),
    blueBookStore.fetchProjectOptions(),
    masterStore.fetchInstitutions(true, { limit: 1000, sort: 'name', order: 'asc' }),
    masterStore.fetchAllRegionLevels(true),
    loadProjects(),
  ])
}

async function loadProjects() {
  await greenBookStore.fetchProjects(greenBookId.value, buildProjectParams())
}

function clearImportProjects() {
  importProjectOptions.value = []
  importForm.project_ids = []
  importProjectSearchQuery.value = ''
}

function clearImportSource() {
  importSourceGreenBookOptions.value = []
  importForm.source_green_book_id = ''
  clearImportProjects()
}

async function loadImportSourceGreenBooks() {
  const current = greenBookStore.currentGreenBook
  if (!current) {
    clearImportSource()
    return
  }

  importSourceGreenBookLoading.value = true
  try {
    const allGreenBooks = await greenBookStore.getGreenBooksForImport()
    importSourceGreenBookOptions.value = allGreenBooks
      .filter((greenBook) => isSourceGreenBook(greenBook, current))
      .sort(compareGreenBookVersionDesc)

    const currentSourceStillAvailable = importSourceGreenBookOptions.value.some(
      (greenBook) => greenBook.id === importForm.source_green_book_id,
    )
    if (!currentSourceStillAvailable) {
      importForm.source_green_book_id = importSourceGreenBookOptions.value[0]?.id ?? ''
      if (!importForm.source_green_book_id) {
        clearImportProjects()
        return
      }
    }

    await loadImportProjects(importForm.source_green_book_id)
  } finally {
    importSourceGreenBookLoading.value = false
  }
}

async function loadImportProjects(sourceGreenBookId?: string) {
  importProjectSearchQuery.value = ''

  if (!sourceGreenBookId) {
    clearImportProjects()
    return
  }

  const current = greenBookStore.currentGreenBook
  if (!current) {
    clearImportProjects()
    return
  }

  importProjectLoading.value = true
  try {
    const [sourceProjects, currentProjects] = await Promise.all([
      greenBookStore.getProjectsByGreenBook(sourceGreenBookId),
      greenBookStore.getProjectsByGreenBook(current.id),
    ])
    const sourceGreenBook = importSourceGreenBookOptions.value.find(
      (greenBook) => greenBook.id === sourceGreenBookId,
    )
    const usedIdentityIds = new Set(currentProjects.map((project) => project.gb_project_identity_id))
    const usedCodes = new Set(currentProjects.map((project) => normalizeProjectCode(project.gb_code)))

    importProjectOptions.value = sourceProjects.map((project) => {
      const unavailableReason = importUnavailableReason(project, usedIdentityIds, usedCodes)

      return {
        ...project,
        source_green_book_id: sourceGreenBookId,
        source_green_book_label: sourceGreenBook ? sourceGreenBookLabel(sourceGreenBook) : '',
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

function importUnavailableReason(
  project: GBProject,
  usedIdentityIds: Set<string>,
  usedCodes: Set<string>,
) {
  if (usedIdentityIds.has(project.gb_project_identity_id)) {
    return 'Sudah ada di Green Book tujuan'
  }

  if (usedCodes.has(normalizeProjectCode(project.gb_code))) {
    return 'Kode Green Book sudah ada di Green Book tujuan'
  }

  return null
}

function openEdit() {
  const current = greenBookStore.currentGreenBook
  if (!current) return
  Object.assign(form, {
    publish_year: current.publish_year,
    revision_number: current.revision_number,
    status: current.status,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = greenBookSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['publish_year', 'revision_number', 'status'])
    return
  }

  await greenBookStore.updateGreenBook(greenBookId.value, parsed.data)
  toast.success('Berhasil', 'Green Book berhasil diperbarui')
  dialogVisible.value = false
  await loadData()
}

function deleteGreenBook() {
  const current = greenBookStore.currentGreenBook
  if (!current || current.project_count > 0) {
    toast.warn('Tidak Bisa Menghapus', 'Green Book masih memiliki Project Green Book.')
    return
  }

  confirm.confirmDelete(`Green Book ${current.publish_year}`, async () => {
    try {
      await greenBookStore.deleteGreenBook(current.id)
      toast.success('Berhasil', 'Green Book berhasil dihapus permanen')
      await router.push({ name: 'green-books' })
    } catch (error) {
      toast.warn(
        'Tidak Bisa Menghapus',
        formatApiError(error, 'Green Book masih memiliki Project Green Book'),
        12000,
      )
    }
  })
}

function deleteProject(project: GBProject) {
  confirm.confirmDelete(`Proyek Green Book ${project.gb_code}`, async () => {
    try {
      await greenBookStore.deleteProject(greenBookId.value, project.id)
      toast.success('Berhasil', 'Proyek Green Book berhasil dihapus permanen')
      if (greenBookStore.projects.length === 1 && projectControls.page.value > 1) {
        projectControls.page.value -= 1
      } else {
        await loadProjects()
      }
    } catch (error) {
      toast.warn(
        'Tidak Bisa Menghapus',
        formatApiError(error, 'Proyek Green Book masih memiliki relasi turunan'),
        12000,
      )
    }
  })
}

function openImportFromGreenBookDialog() {
  importErrors.value = {}
  importForm.source_green_book_id = ''
  importForm.project_ids = []
  importProjectSearchQuery.value = ''
  importDialogVisible.value = true
  void loadImportSourceGreenBooks()
}

async function saveImportFromGreenBook() {
  const parsed = importGBProjectsFromGreenBookSchema.safeParse(importForm)
  if (!parsed.success) {
    importErrors.value = toFormErrors(parsed.error, ['source_green_book_id', 'project_ids'])
    return
  }

  const importableProjectIds = new Set(
    importProjectOptions.value.filter((project) => !project.disabled).map((project) => project.id),
  )
  const projectIds = parsed.data.project_ids.filter((projectId) => importableProjectIds.has(projectId))
  if (projectIds.length === 0) {
    importErrors.value = {
      project_ids: 'Pilih minimal satu Project Green Book yang belum ada di Green Book tujuan',
    }
    return
  }

  await greenBookStore.importProjectsFromGreenBook(greenBookId.value, {
    ...parsed.data,
    project_ids: projectIds,
  })
  toast.success('Berhasil', 'Proyek Green Book berhasil ditambahkan')
  importDialogVisible.value = false
  await Promise.all([greenBookStore.fetchGreenBook(greenBookId.value), loadProjects()])
}

function formatInstitution(institution: Institution) {
  return institution.short_name ? `${institution.name} (${institution.short_name})` : institution.name
}

function formatRegion(region: Region) {
  if (region.type === 'COUNTRY') return `${region.name} (Nasional)`
  if (region.type === 'CITY') return `-- ${region.name}`
  return `- ${region.name}`
}

function isCoveredBySelectedCountry(region: Region) {
  if (!region.parent_code) return false
  if (selectedCountryCodes.value.includes(region.parent_code)) return true
  const parent = masterStore.regions.find((item) => item.code === region.parent_code)
  return parent?.parent_code ? selectedCountryCodes.value.includes(parent.parent_code) : false
}

function expandLocationFilterIds(locationIDs: string[]) {
  if (locationIDs.length === 0) return []

  const expanded = new Set(locationIDs)
  const selectedRegions = masterStore.regions.filter((region) => expanded.has(region.id))

  selectedRegions.forEach((selectedRegion) => {
    if (selectedRegion.type === 'COUNTRY') {
      masterStore.regions.forEach((region) => {
        const parent = region.parent_code
          ? masterStore.regions.find((item) => item.code === region.parent_code)
          : undefined

        if (region.parent_code === selectedRegion.code || parent?.parent_code === selectedRegion.code) {
          expanded.add(region.id)
        }
      })
    }

    if (selectedRegion.type === 'PROVINCE') {
      masterStore.regions
        .filter((region) => region.parent_code === selectedRegion.code)
        .forEach((region) => expanded.add(region.id))
    }
  })

  return [...expanded]
}

function selectedLabelSummary(labels: string[]) {
  if (labels.length <= 2) return labels.join(', ')
  return `${labels.slice(0, 2).join(', ')} +${labels.length - 2}`
}

function selectedBBProjectSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    blueBookStore.projectOptions
      .filter((project) => selected.has(project.id))
      .map((project) => project.bb_code),
  )
}

function selectedInstitutionSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    masterStore.institutions
      .filter((institution) => selected.has(institution.id))
      .map((institution) => institution.short_name || institution.name),
  )
}

function selectedRegionSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    masterStore.regions.filter((region) => selected.has(region.id)).map((region) => region.name),
  )
}

function compareGreenBookVersionDesc(left: GreenBook, right: GreenBook) {
  if (left.publish_year !== right.publish_year) {
    return right.publish_year - left.publish_year
  }

  if (left.revision_number !== right.revision_number) {
    return right.revision_number - left.revision_number
  }

  return String(right.created_at ?? '').localeCompare(String(left.created_at ?? ''))
}

function isSourceGreenBook(greenBook: GreenBook, targetGreenBook: GreenBook) {
  return greenBook.id !== targetGreenBook.id
}

function sourceGreenBookLabel(greenBook: GreenBook) {
  return `Green Book ${greenBook.publish_year} - ${formatGBRevision(greenBook.revision_number)} (${formatGreenBookStatus(greenBook.status)})`
}

function normalizeProjectCode(code: string) {
  return code.trim().toLowerCase()
}

function normalizeSearchText(value: string) {
  return value.trim().toLowerCase().replace(/\s+/g, ' ')
}

function projectMatchesImportSearch(project: GBProjectRevisionSourceOption, query: string) {
  const searchable = [
    project.gb_code,
    project.project_name,
    project.source_green_book_label,
    project.unavailable_reason ?? '',
  ]
    .join(' ')
    .toLowerCase()

  return searchable.includes(query)
}

function isImportProjectSelected(projectId: string) {
  return importForm.project_ids.includes(projectId)
}

function setImportProjectSelected(project: GBProjectRevisionSourceOption, selected: boolean) {
  if (project.disabled) {
    return
  }

  const ids = new Set(importForm.project_ids)
  if (selected) {
    ids.add(project.id)
  } else {
    ids.delete(project.id)
  }

  importForm.project_ids = [...ids]
}

function setAllFilteredImportProjectsSelected(selected: boolean) {
  const ids = new Set(importForm.project_ids)
  filteredImportableProjectOptions.value.forEach((project) => {
    if (selected) {
      ids.add(project.id)
    } else {
      ids.delete(project.id)
    }
  })

  importForm.project_ids = [...ids]
}

function selectAllImportProjects() {
  importForm.project_ids = importableProjectOptions.value.map((project) => project.id)
}

function clearAllImportProjects() {
  importForm.project_ids = []
}

onMounted(() => {
  void loadData()
})

watch(
  [
    projectControls.page,
    projectControls.limit,
    projectControls.sort,
    projectControls.order,
    projectControls.debouncedSearch,
    () => JSON.stringify(projectControls.appliedFilters),
  ],
  () => {
    void loadProjects()
  },
)
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="greenBookStore.currentGreenBook ? `Green Book ${greenBookStore.currentGreenBook.publish_year}` : 'Detail Green Book'"
      :subtitle="greenBookStore.currentGreenBook ? formatGBRevision(greenBookStore.currentGreenBook.revision_number) : undefined"
    >
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'green-books' })" />
        <Button v-if="can('green_book', 'update')" label="Edit Green Book" icon="pi pi-pencil" outlined @click="openEdit" />
        <Button
          v-if="canDeleteCurrentGreenBook"
          icon="pi pi-trash"
          label="Hapus Green Book"
          severity="danger"
          outlined
          @click="deleteGreenBook"
        />
        <Button
          v-if="canImportProjectsFromGreenBook"
          icon="pi pi-file-import"
          label="Tambahkan Proyek dari Green Book Lain"
          outlined
          @click="openImportFromGreenBookDialog"
        />
        <Button
          v-if="can('gb_project', 'create')"
          as="router-link"
          :to="{ name: 'gb-project-create', params: { gbId: greenBookId } }"
          label="Tambah Proyek"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <div v-if="greenBookStore.currentGreenBook" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-3">
      <div>
        <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Tahun Terbit</p>
        <p class="font-semibold text-surface-950">{{ greenBookStore.currentGreenBook.publish_year }}</p>
      </div>
      <div>
        <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Revisi</p>
        <p class="font-semibold text-surface-950">{{ formatGBRevision(greenBookStore.currentGreenBook.revision_number) }}</p>
      </div>
      <div>
        <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Status</p>
        <StatusBadge
          :status="greenBookStore.currentGreenBook.status"
          :label="formatGreenBookStatus(greenBookStore.currentGreenBook.status)"
        />
      </div>
    </div>

    <SearchFilterBar
      v-model:search="projectControls.search.value"
      search-placeholder="Cari kode, nama proyek, atau relasi Blue Book"
      :active-filters="projectControls.activeFilterPills.value"
      :filter-count="projectControls.activeFilterCount.value"
      @apply="projectControls.applyFilters"
      @reset="projectControls.resetFilters"
      @remove="projectControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Proyek Blue Book</span>
          <MultiSelect
            v-model="projectControls.draftFilters.bb_project_ids"
            :options="bbProjectOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua proyek Blue Book"
            filter
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Executing Agency</span>
          <MultiSelect
            v-model="projectControls.draftFilters.executing_agency_ids"
            :options="institutionFilterOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua executing agency"
            filter
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-1">
          <span class="text-sm font-medium text-surface-700">Lokasi</span>
          <MultiSelect
            v-model="projectControls.draftFilters.location_ids"
            :options="locationFilterOptions"
            option-label="label"
            option-value="value"
            option-disabled="disabled"
            placeholder="Semua lokasi"
            filter
            display="chip"
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

    <DataTable
      v-model:page="projectControls.page.value"
      v-model:limit="projectControls.limit.value"
      :data="greenBookStore.projects"
      :columns="columns"
      :loading="greenBookStore.loading"
      :total="greenBookStore.projectTotal"
      :sort-field="projectControls.sort.value"
      :sort-order="projectControls.order.value"
      @sort="projectControls.setSort"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'bb_projects'">{{ joinNames((row as GBProject).bb_projects) }}</span>
        <StatusBadge
          v-else-if="column.field === 'status'"
          :status="String(row.status)"
          :label="formatGreenBookStatus(String(row.status))"
        />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap justify-end gap-1.5">
          <Button
            v-tooltip.top="'Lihat proyek'"
            as="router-link"
            :to="{ name: 'gb-project-detail', params: { gbId: greenBookId, id: row.id } }"
            icon="pi pi-eye"
            size="small"
            outlined
            rounded
            aria-label="Lihat proyek"
          />
          <Button
            v-if="can('gb_project', 'update')"
            v-tooltip.top="'Edit proyek'"
            as="router-link"
            :to="{ name: 'gb-project-edit', params: { gbId: greenBookId, id: row.id } }"
            icon="pi pi-pencil"
            size="small"
            outlined
            rounded
            aria-label="Edit proyek"
          />
          <Button
            v-if="can('gb_project', 'delete')"
            v-tooltip.top="'Hapus proyek'"
            icon="pi pi-trash"
            size="small"
            severity="danger"
            outlined
            rounded
            aria-label="Hapus proyek"
            @click="deleteProject(row as GBProject)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog
      v-model:visible="importDialogVisible"
      modal
      header="Tambahkan Proyek dari Green Book Lain"
      class="w-[64rem] max-w-[calc(100vw-2rem)]"
    >
      <form class="flex flex-col gap-5" @submit.prevent="saveImportFromGreenBook">
        <label class="grid gap-2">
          <span class="text-sm font-medium text-surface-700">Green Book Sumber</span>
          <Select
            v-model="importForm.source_green_book_id"
            :options="importSourceGreenBookSelectOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih Green Book sumber"
            filter
            filter-placeholder="Cari Green Book"
            append-to="body"
            class="w-full"
            :loading="importSourceGreenBookLoading"
            :invalid="Boolean(importErrors.source_green_book_id)"
            @change="loadImportProjects(importForm.source_green_book_id)"
          />
          <small v-if="importErrors.source_green_book_id" class="text-red-600">
            {{ importErrors.source_green_book_id }}
          </small>
          <small
            v-else-if="!importSourceGreenBookLoading && importSourceGreenBookOptions.length === 0"
            class="text-surface-500"
          >
            Belum ada Green Book lain yang bisa dijadikan sumber.
          </small>
        </label>

        <section class="grid gap-3">
          <div class="flex flex-col gap-3 md:flex-row md:items-center">
            <label class="relative min-w-0 flex-1">
              <span class="sr-only">Cari Project Green Book</span>
              <i
                class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400"
              />
              <InputText
                v-model="importProjectSearchQuery"
                placeholder="Cari kode atau nama proyek..."
                class="w-full pl-10"
                :disabled="!importForm.source_green_book_id || importProjectLoading"
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
                      Memuat Project Green Book...
                    </td>
                  </tr>
                  <tr v-else-if="!importForm.source_green_book_id">
                    <td colspan="3" class="px-4 py-8 text-center text-sm text-surface-500">
                      Pilih Green Book sumber untuk melihat Project Green Book.
                    </td>
                  </tr>
                  <tr v-else-if="importProjectOptions.length === 0">
                    <td colspan="3" class="px-4 py-8 text-center text-sm text-surface-500">
                      Green Book sumber ini belum memiliki Project Green Book.
                    </td>
                  </tr>
                  <tr v-else-if="filteredImportProjectOptions.length === 0">
                    <td colspan="3" class="px-4 py-8 text-center text-sm text-surface-500">
                      Tidak ada Project Green Book yang cocok dengan pencarian.
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
                          <span
                            class="rounded border border-surface-200 bg-surface-0 px-2 py-0.5 font-mono text-xs font-semibold text-surface-700"
                          >
                            {{ option.gb_code }}
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
                        <p class="font-medium leading-snug">
                          {{ option.project_name }}
                        </p>
                        <p class="mt-1 text-xs text-surface-500">{{ option.source_green_book_label }}</p>
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

          <p v-if="importErrors.project_ids" class="text-sm text-red-600">
            {{ importErrors.project_ids }}
          </p>
          <p
            v-else-if="
              importForm.source_green_book_id &&
              !importProjectLoading &&
              importProjectOptions.length > 0 &&
              !hasImportableProjectOptions
            "
            class="rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-700"
          >
            Semua Project Green Book dari sumber ini sudah ada di Green Book tujuan.
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
            <Button type="button" label="Batal" severity="secondary" outlined @click="importDialogVisible = false" />
            <Button
              type="submit"
              :label="importSubmitLabel"
              icon="pi pi-file-import"
              :disabled="
                !importForm.source_green_book_id ||
                selectedImportProjectCount === 0 ||
                !hasImportableProjectOptions
              "
            />
          </div>
        </div>
      </form>
    </Dialog>

    <Dialog v-model:visible="dialogVisible" modal header="Edit Green Book" class="w-[32rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tahun Terbit</span>
          <InputNumber v-model="form.publish_year" :use-grouping="false" class="w-full" />
          <small v-if="errors.publish_year" class="text-red-600">{{ errors.publish_year }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nomor Revisi</span>
          <InputNumber v-model="form.revision_number" :min="0" class="w-full" />
          <small v-if="errors.revision_number" class="text-red-600">{{ errors.revision_number }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Status</span>
          <Select
            v-model="form.status"
            :options="statusOptions"
            option-label="label"
            option-value="value"
            placeholder="Pilih status"
            class="w-full"
            :invalid="Boolean(errors.status)"
          />
          <small v-if="errors.status" class="text-red-600">{{ errors.status }}</small>
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
