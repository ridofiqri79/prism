<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Popover from 'primevue/popover'
import Tag from 'primevue/tag'
import ToggleSwitch from 'primevue/toggleswitch'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ListPaginationFooter from '@/components/common/ListPaginationFooter.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import SummaryCard from '@/components/common/SummaryCard.vue'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { useMasterStore } from '@/stores/master.store'
import { useProjectStore } from '@/stores/project.store'
import type { Institution, Lender, LenderType, ProgramTitle, Region } from '@/types/master.types'
import type {
  ProjectMasterColumnConfig,
  ProjectMasterColumnKey,
  ProjectMasterFilterState,
  ProjectMasterListParams,
  ProjectMasterRow,
  ProjectMasterSortField,
  ProjectMasterSortOrder,
  ProjectPipelineStatus,
  ProjectStatus,
} from '@/types/project.types'
import { getPipelineStatusLabel } from '@/utils/status-labels'

const projectStore = useProjectStore()
const masterStore = useMasterStore()
const { can } = usePermission()
const toast = useToast()

const listControls = useListControls<ProjectMasterFilterState>({
  initialFilters: createDefaultFilters(),
  initialSort: 'project_name',
  initialOrder: 'asc',
  searchDebounceMs: 500,
  filterLabels: {
    loan_types: 'Jenis Pinjaman',
    indication_lender_ids: 'Indikasi Lender',
    executing_agency_ids: 'Executing Agency',
    fixed_lender_ids: 'Fixed Lender',
    project_statuses: 'Status Project',
    pipeline_statuses: 'Status Pipeline',
    program_title_ids: 'Program Title',
    region_ids: 'Region/Location',
    foreign_loan_min: 'Foreign Loan Min',
    foreign_loan_max: 'Foreign Loan Max',
    dk_date_from: 'Tanggal DK dari',
    dk_date_to: 'Tanggal DK sampai',
    include_history: 'Snapshot historis',
  },
  formatFilterValue: (key, value) => formatProjectFilterValue(key, value),
})
const page = listControls.page
const limit = listControls.limit
const sortField = listControls.sort as Ref<ProjectMasterSortField>
const sortOrder = listControls.order as Ref<ProjectMasterSortOrder>
const filters = listControls.draftFilters
const appliedFilters = listControls.appliedFilters
const columnConfigs: ProjectMasterColumnConfig[] = [
  { key: 'loan_types', label: 'Jenis Pinjaman', sortField: 'loan_types', defaultVisible: false },
  { key: 'indication_lenders', label: 'Indikasi Lender', sortField: 'indication_lenders', defaultVisible: false },
  { key: 'executing_agencies', label: 'Executing Agency', sortField: 'executing_agencies', defaultVisible: true },
  { key: 'fixed_lenders', label: 'Fixed Lender', sortField: 'fixed_lenders', defaultVisible: false },
  { key: 'status', label: 'Status', sortField: 'project_status', defaultVisible: true },
  { key: 'program_title', label: 'Program Title', sortField: 'program_title', defaultVisible: false },
  { key: 'locations', label: 'Region/Location', sortField: 'locations', defaultVisible: false },
  { key: 'foreign_loan_usd', label: 'Nilai Pinjaman', sortField: 'foreign_loan_usd', defaultVisible: true },
  { key: 'dk_dates', label: 'Tanggal Daftar Kegiatan', sortField: 'dk_dates', defaultVisible: false },
  { key: 'bb_book_ref', label: 'Kode Blue Book', sortField: 'bb_code', defaultVisible: false },
  { key: 'gb_book_ref', label: 'Kode Green Book', sortField: 'gb_codes', defaultVisible: false },
]
const visibleColumnKeys = ref<ProjectMasterColumnKey[]>(
  columnConfigs.filter((column) => column.defaultVisible).map((column) => column.key),
)

const visibleColumns = computed(() =>
  columnConfigs.filter((column) => visibleColumnKeys.value.includes(column.key)),
)
const columnVisibleCount = computed(() => visibleColumns.value.length + 1) // +1 for fixed name col
const columnPopover = ref()
const tableSortOrder = computed(() => (sortOrder.value === 'asc' ? 1 : -1))
const tablePt = {
  thead: { class: 'bg-surface-50 text-left text-xs font-semibold uppercase tracking-wide text-surface-500' },
  headerCell: { class: 'px-4 py-3 text-xs font-semibold uppercase tracking-wide text-surface-500' },
  columnHeaderContent: { class: 'gap-2' },
  bodyCell: { class: 'px-4 py-2.5 text-sm text-surface-800' },
}
const programTitleOptions = computed(() =>
  masterStore.programTitles.map((programTitle) => ({
    label: formatProgramTitle(programTitle),
    value: programTitle.id,
  })),
)
const lenderOptions = computed(() =>
  masterStore.lenders.map((lender) => ({
    ...lender,
    label: formatLender(lender),
    value: lender.id,
  })),
)
const institutionOptions = computed(() =>
  masterStore.institutions.map((institution) => ({
    ...institution,
    label: formatInstitution(institution),
    value: institution.id,
  })),
)
const selectedCountryCodes = computed(() => {
  const selected = new Set(filters.region_ids)

  return masterStore.regions
    .filter((region) => region.type === 'COUNTRY' && selected.has(region.id))
    .map((region) => region.code)
})
const regionOptions = computed(() =>
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

const loanTypeOptions: Array<{ label: string; value: LenderType }> = [
  { label: 'Bilateral', value: 'Bilateral' },
  { label: 'Multilateral', value: 'Multilateral' },
  { label: 'KSA', value: 'KSA' },
]
const projectStatusOptions: Array<{ label: string; value: ProjectStatus }> = [
  { label: 'Pipeline (Blue Book-Daftar Kegiatan)', value: 'Pipeline' },
  { label: 'Ongoing (Loan Agreement-Monitoring)', value: 'Ongoing' },
]
const pipelineStatusOptions: Array<{ label: string; value: ProjectPipelineStatus }> = [
  { label: 'Blue Book', value: 'BB' },
  { label: 'Green Book', value: 'GB' },
  { label: 'Daftar Kegiatan', value: 'DK' },
  { label: 'Loan Agreement', value: 'LA' },
  { label: 'Monitoring', value: 'Monitoring' },
]

const fundingSummaryCards = computed(() => [
  {
    label: 'Total Pinjaman',
    value: projectStore.fundingSummary.total_loan_usd,
  },
  {
    label: 'Total Hibah',
    value: projectStore.fundingSummary.total_grant_usd,
  },
  {
    label: 'Total Dana Pendamping',
    value: projectStore.fundingSummary.total_counterpart_usd,
  },
])

function createDefaultFilters(): ProjectMasterFilterState {
  return {
    loan_types: [],
    indication_lender_ids: [],
    executing_agency_ids: [],
    fixed_lender_ids: [],
    project_statuses: [],
    pipeline_statuses: [],
    program_title_ids: [],
    region_ids: [],
    foreign_loan_min: null,
    foreign_loan_max: null,
    dk_date_from: '',
    dk_date_to: '',
    search: '',
    include_history: false,
  }
}

function textParam(value: string) {
  const normalized = value.trim()
  return normalized.length > 0 ? normalized : undefined
}

function buildParams(): ProjectMasterListParams {
  const params: ProjectMasterListParams = {
    page: page.value,
    limit: limit.value,
    sort: sortField.value,
    order: sortOrder.value,
  }

  if (appliedFilters.loan_types.length > 0) params.loan_types = [...appliedFilters.loan_types]
  if (appliedFilters.indication_lender_ids.length > 0) {
    params.indication_lender_ids = [...appliedFilters.indication_lender_ids]
  }
  if (appliedFilters.executing_agency_ids.length > 0) {
    params.executing_agency_ids = [...appliedFilters.executing_agency_ids]
  }
  if (appliedFilters.fixed_lender_ids.length > 0) {
    params.fixed_lender_ids = [...appliedFilters.fixed_lender_ids]
  }
  if (appliedFilters.project_statuses.length > 0) {
    params.project_statuses = [...appliedFilters.project_statuses]
  }
  if (appliedFilters.pipeline_statuses.length > 0) {
    params.pipeline_statuses = [...appliedFilters.pipeline_statuses]
  }
  if (appliedFilters.program_title_ids.length > 0) {
    params.program_title_ids = [...appliedFilters.program_title_ids]
  }
  const expandedRegionIds = expandRegionFilterIds(appliedFilters.region_ids)
  if (expandedRegionIds.length > 0) params.region_ids = expandedRegionIds
  if (appliedFilters.foreign_loan_min !== null) {
    params.foreign_loan_min = appliedFilters.foreign_loan_min
  }
  if (appliedFilters.foreign_loan_max !== null) {
    params.foreign_loan_max = appliedFilters.foreign_loan_max
  }
  if (appliedFilters.dk_date_from) params.dk_date_from = appliedFilters.dk_date_from
  if (appliedFilters.dk_date_to) params.dk_date_to = appliedFilters.dk_date_to
  if (appliedFilters.include_history) params.include_history = true

  params.search = textParam(listControls.debouncedSearch.value)

  return params
}

async function loadProjectMaster() {
  await projectStore.fetchProjectMaster(buildParams())
}

function buildExportParams(): ProjectMasterListParams {
  const params = { ...buildParams() }
  delete params.page
  delete params.limit

  return params
}

async function exportFilteredProjects() {
  try {
    const blob = await projectStore.downloadProjectMasterExport(buildExportParams())
    saveBlob(blob, projectExportFileName())
    toast.success('Export selesai', 'File Excel dibuat dari filter aktif')
  } catch {
    toast.error('Export gagal', 'Data project tidak dapat diexport saat ini')
  }
}

async function refreshFromFirstPage() {
  if (page.value !== 1) {
    page.value = 1
    return
  }

  await loadProjectMaster()
}

function handleSort(event: { sortField?: unknown; sortOrder?: unknown }) {
  if (typeof event.sortField !== 'string' || event.sortOrder === 0) return
  sortField.value = event.sortField as ProjectMasterSortField
  sortOrder.value = event.sortOrder === 1 ? 'asc' : 'desc'
  page.value = 1
}

function listLabel(values: string[]) {
  return values.length > 0 ? values.join(', ') : '-'
}

function statusLabel(project: ProjectMasterRow) {
  return `${project.project_status} - ${getPipelineStatusLabel(project.pipeline_status)}`
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

function projectExportFileName() {
  return `projects_filtered_${new Date().toISOString().slice(0, 10).replace(/-/g, '')}.xlsx`
}

function formatProgramTitle(programTitle: ProgramTitle) {
  const parent = masterStore.programTitles.find((item) => item.id === programTitle.parent_id)

  return parent ? `${parent.title} / ${programTitle.title}` : programTitle.title
}

function formatLender(lender: Lender) {
  return lender.short_name ? `${lender.name} (${lender.short_name})` : lender.name
}

function formatInstitution(institution: Institution) {
  return institution.short_name ? `${institution.name} (${institution.short_name})` : institution.name
}

function formatRegion(region: Region) {
  const levelLabel: Record<Region['type'], string> = {
    COUNTRY: 'Region',
    PROVINCE: 'Provinsi',
    CITY: 'Kab/Kota',
  }

  if (region.type === 'COUNTRY') {
    return `${region.name} (${levelLabel[region.type]})`
  }

  if (region.type === 'CITY') {
    return `-- ${region.name} (${levelLabel[region.type]})`
  }

  return `- ${region.name} (${levelLabel[region.type]})`
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

function loanTypeSeverity(type: LenderType) {
  if (type === 'Bilateral') return 'info'
  if (type === 'KSA') return 'warn'
  return 'secondary'
}

function projectStatusSeverity(status: ProjectStatus) {
  return status === 'Ongoing' ? 'success' : 'info'
}

function selectedLabelSummary(labels: string[]) {
  if (labels.length <= 2) {
    return labels.join(', ')
  }

  return `${labels.slice(0, 2).join(', ')} +${labels.length - 2}`
}

function formatProjectFilterValue(key: string, value: unknown) {
  if (Array.isArray(value)) {
    if (key === 'indication_lender_ids' || key === 'fixed_lender_ids') {
      const selected = new Set(value)
      return selectedLabelSummary(
        masterStore.lenders
          .filter((lender) => selected.has(lender.id))
          .map((lender) => lender.short_name || lender.name),
      )
    }

    if (key === 'executing_agency_ids') {
      const selected = new Set(value)
      return selectedLabelSummary(
        masterStore.institutions
          .filter((institution) => selected.has(institution.id))
          .map((institution) => institution.short_name || institution.name),
      )
    }

    if (key === 'program_title_ids') {
      const selected = new Set(value)
      return selectedLabelSummary(
        masterStore.programTitles
          .filter((programTitle) => selected.has(programTitle.id))
          .map(formatProgramTitle),
      )
    }

    if (key === 'region_ids') {
      const selected = new Set(value)
      return selectedLabelSummary(
        masterStore.regions.filter((region) => selected.has(region.id)).map((region) => region.name),
      )
    }

    if (key === 'pipeline_statuses') {
      return selectedLabelSummary(value.map((item) => getPipelineStatusLabel(item as ProjectPipelineStatus)))
    }

    return selectedLabelSummary(value.map(String))
  }

  if (typeof value === 'boolean') {
    return value ? 'Ditampilkan' : 'Tidak'
  }

  return String(value)
}

function expandRegionFilterIds(regionIds: string[]) {
  if (regionIds.length === 0) {
    return []
  }

  const expanded = new Set(regionIds)
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

watch([page, limit], () => {
  void loadProjectMaster()
})

watch(
  [listControls.debouncedSearch, () => JSON.stringify(appliedFilters), sortField, sortOrder],
  () => {
    void refreshFromFirstPage()
  },
)

onMounted(() => {
  void Promise.all([
    masterStore.fetchLenders(true, { limit: 1000, sort: 'name', order: 'asc' }),
    masterStore.fetchInstitutions(true, { limit: 1000, sort: 'name', order: 'asc' }),
    masterStore.fetchAllRegionLevels(true),
    masterStore.fetchProgramTitles(true, { limit: 1000, sort: 'title', order: 'asc' }),
    loadProjectMaster(),
  ])
})

onUnmounted(() => {
  listControls.dispose()
})
</script>

<template>
  <section class="space-y-5">
    <PageHeader
      title="Project"
      subtitle="Master table seluruh Proyek Blue Book beserta status pipeline, lender, instansi, lokasi, dan nilai pinjaman"
    >
      <template #actions>
        <Button
          v-tooltip.top="'Export semua project sesuai filter aktif'"
          label="Export Excel"
          icon="pi pi-download"
          severity="secondary"
          outlined
          :loading="projectStore.exporting"
          :disabled="projectStore.total === 0 || projectStore.exporting"
          @click="exportFilteredProjects"
        />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari nama proyek, indikasi lender, fixed lender, atau executing agency"
      :active-filters="listControls.activeFilterPills.value"
      :filter-count="listControls.activeFilterCount.value"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
      <template #actions>
        <Button
          v-tooltip.bottom="'Atur kolom yang ditampilkan'"
          type="button"
          icon="pi pi-table"
          severity="secondary"
          outlined
          class="h-12 shrink-0 gap-2"
          :badge="String(columnVisibleCount)"
          badge-severity="secondary"
          @click="(e) => columnPopover.toggle(e)"
        />
        <Popover ref="columnPopover">
          <div class="w-52">
            <p class="mb-2 text-xs font-semibold uppercase tracking-wide text-surface-400">Kolom Tampil</p>
            <div class="space-y-0.5">
              <label
                v-for="column in columnConfigs"
                :key="column.key"
                class="flex cursor-pointer items-center gap-3 rounded-md px-2 py-2 transition-colors hover:bg-surface-50"
              >
                <Checkbox v-model="visibleColumnKeys" :value="column.key" />
                <span class="text-sm text-surface-700">{{ column.label }}</span>
              </label>
            </div>
          </div>
        </Popover>
      </template>
      <template #filters>
        <label class="flex items-center gap-3 rounded-lg border border-surface-200 px-3 py-2 xl:col-span-6">
          <ToggleSwitch v-model="filters.include_history" />
          <span class="text-sm font-medium text-surface-700">Tampilkan snapshot historis</span>
        </label>

        <div class="contents">
          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Jenis Pinjaman</span>
            <MultiSelect
              v-model="filters.loan_types"
              :options="loanTypeOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua jenis"
              filter
              filter-placeholder="Cari jenis pinjaman"
              display="chip"
              class="w-full"
            />
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Indikasi Lender</span>
            <MultiSelect
              v-model="filters.indication_lender_ids"
              :options="lenderOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua indikasi lender"
              filter
              filter-placeholder="Cari indikasi lender"
              display="chip"
              class="w-full"
            >
              <template #option="{ option }">
                <div class="flex w-full items-center justify-between gap-3">
                  <span>{{ option.label }}</span>
                  <Tag :value="option.type" severity="info" rounded />
                </div>
              </template>
            </MultiSelect>
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Executing Agency</span>
            <MultiSelect
              v-model="filters.executing_agency_ids"
              :options="institutionOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua executing agency"
              filter
              filter-placeholder="Cari executing agency"
              display="chip"
              class="w-full"
            />
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Fixed Lender (Green Book)</span>
            <MultiSelect
              v-model="filters.fixed_lender_ids"
              :options="lenderOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua fixed lender"
              filter
              filter-placeholder="Cari fixed lender"
              display="chip"
              class="w-full"
            >
              <template #option="{ option }">
                <div class="flex w-full items-center justify-between gap-3">
                  <span>{{ option.label }}</span>
                  <Tag :value="option.type" severity="info" rounded />
                </div>
              </template>
            </MultiSelect>
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Status Project</span>
            <MultiSelect
              v-model="filters.project_statuses"
              :options="projectStatusOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua status"
              filter
              filter-placeholder="Cari status project"
              display="chip"
              class="w-full"
            />
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Status Pipeline</span>
            <MultiSelect
              v-model="filters.pipeline_statuses"
              :options="pipelineStatusOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua step"
              filter
              filter-placeholder="Cari status pipeline"
              display="chip"
              class="w-full"
            />
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Program Title</span>
            <MultiSelect
              v-model="filters.program_title_ids"
              :options="programTitleOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua program title"
              filter
              filter-placeholder="Cari program title"
              display="chip"
              class="w-full"
            />
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Region/Location</span>
            <MultiSelect
              v-model="filters.region_ids"
              :options="regionOptions"
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

          <div class="grid gap-4 sm:grid-cols-2 xl:col-span-2">
            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-700">Foreign Loan Min</span>
              <InputNumber
                v-model="filters.foreign_loan_min"
                mode="decimal"
                :min="0"
                :min-fraction-digits="0"
                :max-fraction-digits="2"
                placeholder="USD minimum"
                class="w-full"
              />
            </label>
            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-700">Foreign Loan Max</span>
              <InputNumber
                v-model="filters.foreign_loan_max"
                mode="decimal"
                :min="0"
                :min-fraction-digits="0"
                :max-fraction-digits="2"
                placeholder="USD maksimum"
                class="w-full"
              />
            </label>
          </div>

          <div class="grid gap-4 sm:grid-cols-2 xl:col-span-2">
            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-700">Tanggal Daftar Kegiatan Dari</span>
              <InputText v-model="filters.dk_date_from" type="date" class="w-full" />
            </label>
            <label class="block space-y-2">
              <span class="text-sm font-medium text-surface-700">Tanggal Daftar Kegiatan Sampai</span>
              <InputText v-model="filters.dk_date_to" type="date" class="w-full" />
            </label>
          </div>
        </div>
      </template>
    </SearchFilterBar>

    <section class="grid gap-4 md:grid-cols-3">
      <SummaryCard
        v-for="card in fundingSummaryCards"
        :key="card.label"
        :label="card.label"
        :value="card.value"
        unit="USD"
        format="currency"
        :compact="false"
      />
    </section>

    <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
      <div class="overflow-x-auto">
        <DataTable
          :value="projectStore.projects"
          :loading="projectStore.loading"
          lazy
          striped-rows
          removable-sort
          data-key="id"
          :sort-field="sortField"
          :sort-order="tableSortOrder"
          :table-style="{ minWidth: '68rem', width: '100%', tableLayout: 'auto' }"
          :pt="tablePt"
          class="w-full"
          @sort="handleSort"
        >
          <template #empty>
            <EmptyState title="Tidak ada project" description="Ubah filter atau kata kunci pencarian." />
          </template>

          <!-- Kolom Nama Proyek (fixed) -->
          <Column field="project_name" header="Nama Proyek" sortable :style="{ minWidth: '18rem', width: '18rem' }">
            <template #body="{ data: project }">
              <RouterLink
                :to="{ name: 'bb-project-detail', params: { bbId: project.blue_book_id, id: project.id } }"
                class="block whitespace-normal font-semibold leading-relaxed text-surface-950 hover:text-primary-600"
              >
                {{ project.project_name }}
              </RouterLink>
              <p v-if="project.program_title" class="mt-0.5 whitespace-normal text-xs leading-relaxed text-surface-500">
                {{ project.program_title }}
              </p>
            </template>
          </Column>

          <!-- Kolom dinamis -->
          <Column
            v-for="column in visibleColumns"
            :key="column.key"
            :field="column.sortField"
            :header="column.label"
            sortable
            :style="column.key === 'foreign_loan_usd'
              ? { minWidth: '11rem', width: '11rem', textAlign: 'right' }
              : { minWidth: '8rem', width: '8rem' }"
            :header-style="column.key === 'foreign_loan_usd' ? { textAlign: 'right' } : {}"
            :body-style="column.key === 'foreign_loan_usd'
              ? { textAlign: 'right', fontWeight: '500', color: 'var(--p-surface-900)' }
              : {}"
          >
            <template #body="{ data: project }">
              <div v-if="column.key === 'loan_types'">
                <div v-if="project.loan_types.length > 0" class="flex flex-wrap gap-1.5">
                  <Tag
                    v-for="type in project.loan_types"
                    :key="type"
                    :value="type"
                    :severity="loanTypeSeverity(type)"
                    rounded
                  />
                </div>
                <span v-else class="text-surface-400">-</span>
              </div>
              <span
                v-else-if="column.key === 'indication_lenders'"
                class="block whitespace-normal leading-relaxed"
                :title="listLabel(project.indication_lenders)"
              >{{ listLabel(project.indication_lenders) }}</span>
              <span
                v-else-if="column.key === 'executing_agencies'"
                class="block whitespace-normal leading-relaxed"
                :title="listLabel(project.executing_agencies)"
              >{{ listLabel(project.executing_agencies) }}</span>
              <span
                v-else-if="column.key === 'fixed_lenders'"
                class="block whitespace-normal leading-relaxed"
                :title="listLabel(project.fixed_lenders)"
              >{{ listLabel(project.fixed_lenders) }}</span>
              <Tag
                v-else-if="column.key === 'status'"
                :value="statusLabel(project)"
                :severity="projectStatusSeverity(project.project_status)"
                rounded
              />
              <span v-else-if="column.key === 'program_title'" class="block whitespace-normal leading-relaxed">
                {{ project.program_title || '-' }}
              </span>
              <span
                v-else-if="column.key === 'locations'"
                class="block whitespace-normal leading-relaxed"
                :title="listLabel(project.locations)"
              >{{ listLabel(project.locations) }}</span>
              <CurrencyDisplay
                v-else-if="column.key === 'foreign_loan_usd'"
                :amount="project.foreign_loan_usd"
                currency="USD"
              />
              <span
                v-else-if="column.key === 'dk_dates'"
                class="block whitespace-normal leading-relaxed"
                :title="listLabel(project.dk_dates)"
              >{{ listLabel(project.dk_dates) }}</span>
              <div v-else-if="column.key === 'bb_book_ref'">
                <div class="font-semibold">{{ project.bb_code }}</div>
                <div class="mt-0.5 text-xs text-surface-500">{{ project.blue_book_revision_label }}</div>
              </div>
              <div v-else-if="column.key === 'gb_book_ref'">
                <template v-if="project.gb_codes.length > 0">
                  <div
                    v-for="(gbCode, i) in project.gb_codes"
                    :key="gbCode"
                    :class="i > 0 ? 'mt-1.5 border-t border-surface-100 pt-1.5' : ''"
                  >
                    <div class="font-semibold">{{ gbCode }}</div>
                    <div class="mt-0.5 text-xs text-surface-500">{{ project.green_book_revision_labels[i] ?? '-' }}</div>
                  </div>
                </template>
                <span v-else class="text-surface-400">-</span>
              </div>
            </template>
          </Column>
        </DataTable>
      </div>

      <div class="border-t border-surface-200 p-3">
        <ListPaginationFooter
          v-model:page="page"
          v-model:limit="limit"
          :total="projectStore.total"
        />
      </div>
    </div>
  </section>
</template>
