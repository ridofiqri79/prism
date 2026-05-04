<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
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
import { blueBookSchema, importBBProjectsFromBlueBookSchema } from '@/schemas/blue-book.schema'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useMasterStore } from '@/stores/master.store'
import type {
  BBProject,
  BBProjectListParams,
  BBProjectRevisionSourceOption,
  BlueBook,
  BlueBookPayload,
  BlueBookStatus,
  ImportBBProjectsFromBlueBookPayload,
} from '@/types/blue-book.types'
import type { Institution, Region } from '@/types/master.types'
import { formatApiError } from '@/utils/api-error'
import { formatBlueBookStatus, formatRevision, toFormErrors, type FormErrors } from './blue-book-page-utils'

type BlueBookField = keyof BlueBookPayload
type ImportFromBlueBookField = keyof ImportBBProjectsFromBlueBookPayload
interface ProjectFilterState {
  executing_agency_ids: string[]
  location_ids: string[]
}

const route = useRoute()
const router = useRouter()
const blueBookStore = useBlueBookStore()
const greenBookStore = useGreenBookStore()
const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const blueBookId = computed(() => String(route.params.id ?? ''))
const dialogVisible = ref(false)
const importDialogVisible = ref(false)
const gbCreateDialogVisible = ref(false)
const selectedBBProject = ref<BBProject | null>(null)
const projectControls = useListControls<ProjectFilterState>({
  initialFilters: {
    executing_agency_ids: [],
    location_ids: [],
  },
  filterLabels: {
    executing_agency_ids: 'Executing Agency',
    location_ids: 'Location',
  },
  formatFilterValue: (key, value) => {
    if (key === 'executing_agency_ids' && Array.isArray(value)) {
      return selectedInstitutionSummary(value)
    }
    if (key === 'location_ids' && Array.isArray(value)) {
      return selectedRegionSummary(value)
    }
    return Array.isArray(value) ? selectedLabelSummary(value) : String(value)
  },
})
const form = reactive<BlueBookPayload>({
  period_id: '',
  publish_date: '',
  revision_number: 0,
  revision_year: null,
  status: 'active',
})
const projectFilters = projectControls.draftFilters
const gbCreateForm = reactive({
  greenBookId: '',
  useBBData: false,
})
const importForm = reactive<ImportBBProjectsFromBlueBookPayload>({
  source_blue_book_id: '',
  project_ids: [],
})
const errors = ref<FormErrors<BlueBookField>>({})
const importErrors = ref<FormErrors<ImportFromBlueBookField>>({})
const importSourceBlueBookOptions = ref<BlueBook[]>([])
const importSourceBlueBookLoading = ref(false)
const importProjectOptions = ref<BBProjectRevisionSourceOption[]>([])
const importProjectLoading = ref(false)
const columns: ColumnDef[] = [
  { field: 'bb_code', header: 'Kode Blue Book', width: '11rem' },
  { field: 'project_name', header: 'Nama Proyek', width: '32%' },
  { field: 'executing_agency', header: 'Executing Agency', width: '24%' },
  { field: 'location', header: 'Location', width: '21%' },
  { field: 'actions', header: 'Aksi', align: 'right', width: '12rem' },
]
const statusOptions: Array<{ label: string; value: BlueBookStatus }> = [
  { label: 'Berlaku', value: 'active' },
  { label: 'Tidak Berlaku', value: 'superseded' },
]
const canDeleteCurrentBlueBook = computed(
  () => can('blue_book', 'delete') && (blueBookStore.currentBlueBook?.project_count ?? 0) === 0,
)
const isRevisionBlueBook = computed(() =>
  Boolean(blueBookStore.currentBlueBook?.replaces_blue_book_id) ||
  Number(blueBookStore.currentBlueBook?.revision_number ?? 0) > 0,
)
const canImportProjectsFromBlueBook = computed(
  () => can('bb_project', 'create') && isRevisionBlueBook.value,
)

const selectedCountryCodes = computed(() => {
  const selected = new Set(projectControls.draftFilters.location_ids)

  return masterStore.regions
    .filter((region) => region.type === 'COUNTRY' && selected.has(region.id))
    .map((region) => region.code)
})

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

async function loadData() {
  await Promise.all([
    masterStore.fetchPeriods(true, { limit: 1000 }),
    masterStore.fetchInstitutions(true, { limit: 1000, sort: 'name', order: 'asc' }),
    masterStore.fetchAllRegionLevels(true),
    blueBookStore.fetchBlueBook(blueBookId.value),
    loadProjects(),
    greenBookStore.fetchGreenBooks({ limit: 1000 }),
  ])
}

async function loadProjects() {
  await blueBookStore.fetchProjects(blueBookId.value, buildProjectParams())
}

function clearImportProjects() {
  importProjectOptions.value = []
  importForm.project_ids = []
}

function clearImportSource() {
  importSourceBlueBookOptions.value = []
  importForm.source_blue_book_id = ''
  clearImportProjects()
}

async function loadImportSourceBlueBooks() {
  const current = blueBookStore.currentBlueBook
  if (!current || !isRevisionBlueBook.value) {
    clearImportSource()
    return
  }

  importSourceBlueBookLoading.value = true
  try {
    const samePeriodBlueBooks = await blueBookStore.getBlueBooksByPeriod(current.period.id)
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
  if (!sourceBlueBookId) {
    clearImportProjects()
    return
  }

  const current = blueBookStore.currentBlueBook
  if (!current) {
    clearImportProjects()
    return
  }

  importProjectLoading.value = true
  try {
    const [sourceProjects, currentProjects] = await Promise.all([
      blueBookStore.getProjectsByBlueBook(sourceBlueBookId),
      blueBookStore.getProjectsByBlueBook(current.id),
    ])
    const sourceBlueBook = importSourceBlueBookOptions.value.find(
      (blueBook) => blueBook.id === sourceBlueBookId,
    )
    const usedIdentityIds = new Set(currentProjects.map((project) => project.project_identity_id))
    const usedCodes = new Set(currentProjects.map((project) => normalizeProjectCode(project.bb_code)))

    importProjectOptions.value = sourceProjects
      .filter((project) => !usedIdentityIds.has(project.project_identity_id))
      .filter((project) => !usedCodes.has(normalizeProjectCode(project.bb_code)))
      .map((project) => ({
        ...project,
        source_blue_book_id: sourceBlueBookId,
        source_blue_book_label: sourceBlueBook ? sourceBlueBookLabel(sourceBlueBook) : '',
      }))
    importForm.project_ids = importProjectOptions.value.map((project) => project.id)
  } finally {
    importProjectLoading.value = false
  }
}

function buildProjectParams(): BBProjectListParams {
  const params = projectControls.buildParams() as BBProjectListParams
  const locationIDs = expandLocationFilterIds(projectControls.appliedFilters.location_ids)
  if (locationIDs.length > 0) {
    params.location_ids = locationIDs
  }

  return params
}

const greenBookOptions = computed(() =>
  greenBookStore.greenBooks
    .filter((greenBook) => greenBook.status === 'active')
    .map((greenBook) => ({
      ...greenBook,
      label: `Green Book ${greenBook.publish_year} Rev ${greenBook.revision_number}`,
    })),
)
const importSourceBlueBookSelectOptions = computed(() =>
  importSourceBlueBookOptions.value.map((blueBook) => ({
    ...blueBook,
    label: sourceBlueBookLabel(blueBook),
  })),
)
const importProjectSelectOptions = computed(() =>
  importProjectOptions.value.map((project) => ({
    ...project,
    label: carryOverProjectLabel(project),
  })),
)

function compareBlueBookVersionDesc(left: BlueBook, right: BlueBook) {
  if (left.revision_number !== right.revision_number) {
    return right.revision_number - left.revision_number
  }

  const leftYear = left.revision_year ?? 0
  const rightYear = right.revision_year ?? 0
  if (leftYear !== rightYear) {
    return rightYear - leftYear
  }

  return String(right.created_at ?? '').localeCompare(String(left.created_at ?? ''))
}

function isSourceBlueBook(blueBook: BlueBook, targetBlueBook: BlueBook) {
  if (blueBook.id === targetBlueBook.id || blueBook.period.id !== targetBlueBook.period.id) {
    return false
  }

  if (blueBook.id === targetBlueBook.replaces_blue_book_id) {
    return true
  }

  if (blueBook.revision_number < targetBlueBook.revision_number) {
    return true
  }

  return (
    blueBook.revision_number === targetBlueBook.revision_number &&
    Boolean(blueBook.created_at) &&
    Boolean(targetBlueBook.created_at) &&
    String(blueBook.created_at) < String(targetBlueBook.created_at)
  )
}

function sourceBlueBookLabel(blueBook: BlueBook) {
  return `${blueBook.period.name} - ${formatRevision(blueBook.revision_number, blueBook.revision_year)} (${formatBlueBookStatus(blueBook.status)})`
}

function carryOverProjectLabel(project: { bb_code: string; project_name: string }) {
  return `${project.bb_code} - ${project.project_name}`
}

function normalizeProjectCode(code: string) {
  return code.trim().toLowerCase()
}

function formatInstitution(institution: Institution) {
  return institution.short_name ? `${institution.name} (${institution.short_name})` : institution.name
}

function formatRegion(region: Region) {
  if (region.type === 'COUNTRY') {
    return `${region.name} (Nasional)`
  }

  if (region.type === 'CITY') {
    return `-- ${region.name}`
  }

  return `- ${region.name}`
}

function isCoveredBySelectedCountry(region: Region) {
  if (!region.parent_code) {
    return false
  }

  if (selectedCountryCodes.value.includes(region.parent_code)) {
    return true
  }

  const parent = masterStore.regions.find((item) => item.code === region.parent_code)

  return parent?.parent_code ? selectedCountryCodes.value.includes(parent.parent_code) : false
}

function expandLocationFilterIds(locationIDs: string[]) {
  if (locationIDs.length === 0) {
    return []
  }

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

function listNames(items: Array<{ name: string }>) {
  return items.map((item) => item.name).join(', ') || '-'
}

function selectedLabelSummary(labels: string[]) {
  if (labels.length <= 2) {
    return labels.join(', ')
  }

  return `${labels.slice(0, 2).join(', ')} +${labels.length - 2}`
}

function selectedInstitutionSummary(ids: string[]) {
  const selected = new Set(ids)
  const labels = masterStore.institutions
    .filter((institution) => selected.has(institution.id))
    .map((institution) => institution.short_name || institution.name)

  return selectedLabelSummary(labels)
}

function selectedRegionSummary(ids: string[]) {
  const selected = new Set(ids)
  const labels = masterStore.regions
    .filter((region) => selected.has(region.id))
    .map((region) => region.name)

  return selectedLabelSummary(labels)
}

function openEdit() {
  const current = blueBookStore.currentBlueBook
  if (!current) return
  Object.assign(form, {
    period_id: current.period.id,
    publish_date: current.publish_date,
    revision_number: current.revision_number,
    revision_year: current.revision_year ?? null,
    status: current.status,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = blueBookSchema.safeParse({
    ...form,
    revision_year: form.revision_year ?? undefined,
  })
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, [
      'period_id',
      'publish_date',
      'revision_number',
      'revision_year',
      'status',
    ])
    return
  }

  await blueBookStore.updateBlueBook(blueBookId.value, {
    ...parsed.data,
    revision_year: parsed.data.revision_year ?? null,
  })
  toast.success('Berhasil', 'Blue Book berhasil diperbarui')
  dialogVisible.value = false
  await loadData()
}

function deleteBlueBook() {
  const current = blueBookStore.currentBlueBook
  if (!current) return
  if (current.project_count > 0) {
    toast.warn('Tidak Bisa Menghapus', 'Blue Book masih memiliki Project Blue Book.')
    return
  }

  confirm.confirmDelete(`Blue Book ${current.period.name}`, async () => {
    try {
      await blueBookStore.deleteBlueBook(current.id)
      toast.success('Berhasil', 'Blue Book berhasil dihapus permanen')
      await router.push({ name: 'blue-books' })
    } catch (error) {
      toast.warn(
        'Tidak Bisa Menghapus',
        formatApiError(error, 'Blue Book masih memiliki Project Blue Book'),
        12000,
      )
    }
  })
}

function deleteProject(project: BBProject) {
  confirm.confirmDelete(`Proyek Blue Book ${project.bb_code}`, async () => {
    try {
      await blueBookStore.deleteProject(blueBookId.value, project.id)
      toast.success('Berhasil', 'Proyek Blue Book berhasil dihapus permanen')
      if (blueBookStore.projects.length === 1 && projectControls.page.value > 1) {
        projectControls.page.value -= 1
      } else {
        await loadProjects()
      }
    } catch (error) {
      toast.warn(
        'Tidak Bisa Menghapus',
        formatApiError(error, 'Proyek Blue Book masih memiliki relasi turunan'),
        12000,
      )
    }
  })
}

function openGBCreateDialog(project: BBProject) {
  selectedBBProject.value = project
  gbCreateForm.greenBookId = greenBookOptions.value[0]?.id ?? ''
  gbCreateForm.useBBData = false
  gbCreateDialogVisible.value = true
}

function openImportFromBlueBookDialog() {
  importErrors.value = {}
  importForm.source_blue_book_id = ''
  importForm.project_ids = []
  importDialogVisible.value = true
  void loadImportSourceBlueBooks()
}

async function saveImportFromBlueBook() {
  const parsed = importBBProjectsFromBlueBookSchema.safeParse(importForm)
  if (!parsed.success) {
    importErrors.value = toFormErrors(parsed.error, ['source_blue_book_id', 'project_ids'])
    return
  }

  await blueBookStore.importProjectsFromBlueBook(blueBookId.value, parsed.data)
  toast.success('Berhasil', 'Proyek Blue Book berhasil diimpor')
  importDialogVisible.value = false
  await Promise.all([blueBookStore.fetchBlueBook(blueBookId.value), loadProjects()])
}

async function createGBProjectFromBB() {
  if (!selectedBBProject.value || !gbCreateForm.greenBookId) return

  const source = selectedBBProject.value
  gbCreateDialogVisible.value = false
  await router.push({
    name: 'gb-project-create',
    params: { gbId: gbCreateForm.greenBookId },
    query: {
      source_bb_project_id: source.id,
      source_mode: gbCreateForm.useBBData ? 'existing' : 'new',
    },
  })
}

onMounted(() => {
  void loadData()
})

onUnmounted(() => {
  projectControls.dispose()
})

watch(
  [
    projectControls.page,
    projectControls.limit,
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
      :title="blueBookStore.currentBlueBook?.period.name ?? 'Blue Book Detail'"
      :subtitle="blueBookStore.currentBlueBook ? `Terbit ${blueBookStore.currentBlueBook.publish_date}` : undefined"
    >
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'blue-books' })" />
        <Button v-if="can('blue_book', 'update')" label="Edit Blue Book" icon="pi pi-pencil" outlined @click="openEdit" />
        <Button
          v-if="canDeleteCurrentBlueBook"
          label="Hapus Blue Book"
          icon="pi pi-trash"
          severity="danger"
          outlined
          @click="deleteBlueBook"
        />
        <Button
          v-if="canImportProjectsFromBlueBook"
          label="Impor Proyek dari Blue Book Lain"
          icon="pi pi-file-import"
          outlined
          @click="openImportFromBlueBookDialog"
        />
        <Button
          v-if="can('bb_project', 'create')"
          as="router-link"
          :to="{ name: 'bb-project-create', params: { bbId: blueBookId } }"
          label="Tambah Proyek"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <div v-if="blueBookStore.currentBlueBook" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-4">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Periode</p>
        <p class="font-semibold text-surface-950">{{ blueBookStore.currentBlueBook.period.name }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Tanggal Terbit</p>
        <p class="font-semibold text-surface-950">{{ blueBookStore.currentBlueBook.publish_date }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Revisi</p>
        <p class="font-semibold text-surface-950">
          {{ formatRevision(blueBookStore.currentBlueBook.revision_number, blueBookStore.currentBlueBook.revision_year) }}
        </p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Status</p>
        <StatusBadge
          :status="blueBookStore.currentBlueBook.status"
          :label="formatBlueBookStatus(blueBookStore.currentBlueBook.status)"
        />
      </div>
    </div>

    <SearchFilterBar
      v-model:search="projectControls.search.value"
      search-placeholder="Nama proyek / executing agency"
      :active-filters="projectControls.activeFilterPills.value"
      :filter-count="projectControls.activeFilterCount.value"
      @apply="projectControls.applyFilters"
      @reset="projectControls.resetFilters"
      @remove="projectControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Executing Agency</span>
          <MultiSelect
            v-model="projectFilters.executing_agency_ids"
            :options="institutionFilterOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua executing agency"
            filter
            filter-placeholder="Cari executing agency"
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Location</span>
          <MultiSelect
            v-model="projectFilters.location_ids"
            :options="locationFilterOptions"
            option-label="label"
            option-value="value"
            option-disabled="disabled"
            placeholder="Semua lokasi"
            filter
            filter-placeholder="Cari lokasi"
            display="chip"
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

    <DataTable
      v-model:page="projectControls.page.value"
      v-model:limit="projectControls.limit.value"
      :data="blueBookStore.projects"
      :columns="columns"
      :loading="blueBookStore.loading"
      :total="blueBookStore.projectTotal"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'executing_agency'">
          {{ listNames((row as BBProject).executing_agencies) }}
        </span>
        <span v-else-if="column.field === 'location'">
          {{ listNames((row as BBProject).locations) }}
        </span>
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap justify-end gap-1.5">
          <Button
            v-tooltip.top="'Lihat proyek'"
            as="router-link"
            :to="{ name: 'bb-project-detail', params: { bbId: blueBookId, id: row.id } }"
            icon="pi pi-eye"
            size="small"
            outlined
            rounded
            aria-label="Lihat proyek"
          />
          <Button
            v-if="can('bb_project', 'update')"
            v-tooltip.top="'Edit proyek'"
            as="router-link"
            :to="{ name: 'bb-project-edit', params: { bbId: blueBookId, id: row.id } }"
            icon="pi pi-pencil"
            size="small"
            outlined
            rounded
            aria-label="Edit proyek"
          />
          <Button
            v-if="can('gb_project', 'create')"
            v-tooltip.top="'Tambah ke Green Book'"
            icon="pi pi-plus"
            size="small"
            severity="secondary"
            outlined
            rounded
            aria-label="Tambah proyek Green Book"
            @click="openGBCreateDialog(row as BBProject)"
          />
          <Button
            v-if="can('bb_project', 'delete')"
            v-tooltip.top="'Hapus proyek'"
            icon="pi pi-trash"
            size="small"
            severity="danger"
            outlined
            rounded
            aria-label="Hapus proyek"
            @click="deleteProject(row as BBProject)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Edit Blue Book" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Periode</span>
          <Select
            v-model="form.period_id"
            :options="masterStore.periods"
            option-label="name"
            option-value="id"
            placeholder="Pilih period"
            class="w-full"
            disabled
          />
          <small class="text-surface-500">Periode tidak diubah saat edit Blue Book.</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tanggal Terbit</span>
          <InputText v-model="form.publish_date" type="date" class="w-full" :invalid="Boolean(errors.publish_date)" />
          <small v-if="errors.publish_date" class="text-red-600">{{ errors.publish_date }}</small>
        </label>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Nomor Revisi</span>
            <InputNumber v-model="form.revision_number" :min="0" class="w-full" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Tahun Revisi</span>
            <InputNumber v-model="form.revision_year" :use-grouping="false" class="w-full" />
          </label>
        </div>
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

    <Dialog
      v-model:visible="importDialogVisible"
      modal
      header="Impor Proyek dari Blue Book Lain"
      class="w-[44rem] max-w-[95vw]"
    >
      <form class="space-y-4" @submit.prevent="saveImportFromBlueBook">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Blue Book Sumber</span>
          <Select
            v-model="importForm.source_blue_book_id"
            :options="importSourceBlueBookSelectOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih Blue Book sumber"
            filter
            class="w-full"
            :loading="importSourceBlueBookLoading"
            :invalid="Boolean(importErrors.source_blue_book_id)"
            @change="loadImportProjects(importForm.source_blue_book_id)"
          />
          <small v-if="importErrors.source_blue_book_id" class="text-red-600">
            {{ importErrors.source_blue_book_id }}
          </small>
          <small
            v-else-if="!importSourceBlueBookLoading && importSourceBlueBookOptions.length === 0"
            class="text-surface-500"
          >
            Belum ada Blue Book sumber pada periode ini.
          </small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Project Blue Book yang Diimpor</span>
          <MultiSelect
            v-model="importForm.project_ids"
            :options="importProjectSelectOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih Project Blue Book"
            filter
            class="w-full"
            :max-selected-labels="2"
            selected-items-label="{0} Project Blue Book dipilih"
            :loading="importProjectLoading"
            :invalid="Boolean(importErrors.project_ids)"
            :disabled="!importForm.source_blue_book_id"
          >
            <template #option="{ option }">
              <div class="space-y-1">
                <p class="font-medium text-surface-950">
                  {{ option.bb_code }} - {{ option.project_name }}
                </p>
                <p class="text-xs text-surface-500">{{ option.source_blue_book_label }}</p>
              </div>
            </template>
          </MultiSelect>
          <small v-if="importErrors.project_ids" class="text-red-600">
            {{ importErrors.project_ids }}
          </small>
          <small
            v-else-if="importForm.source_blue_book_id && !importProjectLoading && importProjectOptions.length === 0"
            class="text-surface-500"
          >
            Tidak ada Project Blue Book yang bisa diimpor dari sumber ini.
          </small>
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="importDialogVisible = false" />
          <Button
            type="submit"
            label="Impor"
            icon="pi pi-file-import"
            :disabled="!importForm.source_blue_book_id || importForm.project_ids.length === 0"
          />
        </div>
      </form>
    </Dialog>

    <Dialog v-model:visible="gbCreateDialogVisible" modal header="Tambah Proyek Green Book" class="w-[34rem] max-w-[95vw]">
      <div class="space-y-4">
        <div class="rounded-lg border border-surface-200 bg-surface-50 p-3">
          <p class="text-xs uppercase tracking-wide text-surface-500">Proyek Blue Book</p>
          <p class="mt-1 font-semibold text-surface-950">
            {{ selectedBBProject?.bb_code }} - {{ selectedBBProject?.project_name }}
          </p>
        </div>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Green Book Tujuan</span>
          <Select
            v-model="gbCreateForm.greenBookId"
            :options="greenBookOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih Green Book"
            class="w-full"
          />
        </label>
        <label class="flex items-start gap-3 rounded-lg border border-surface-200 bg-white p-3">
          <Checkbox v-model="gbCreateForm.useBBData" binary input-id="use-bb-data-for-gb" />
          <span class="text-sm font-medium text-surface-700">
            Gunakan data di Blue Book sebagai data Green Book
          </span>
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="gbCreateDialogVisible = false" />
          <Button label="Lanjut" icon="pi pi-arrow-right" :disabled="!gbCreateForm.greenBookId" @click="createGBProjectFromBB" />
        </div>
      </div>
    </Dialog>
  </section>
</template>
