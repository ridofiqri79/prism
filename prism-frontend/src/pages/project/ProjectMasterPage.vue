<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'
import { useRoute } from 'vue-router'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Skeleton from 'primevue/skeleton'
import Tag from 'primevue/tag'
import ToggleSwitch from 'primevue/toggleswitch'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ListPaginationFooter from '@/components/common/ListPaginationFooter.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import SummaryCard from '@/components/common/SummaryCard.vue'
import TableReloadShell from '@/components/common/TableReloadShell.vue'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { useMasterStore } from '@/stores/master.store'
import { useProjectStore } from '@/stores/project.store'
import type { Institution, Lender, LenderType, ProgramTitle, Region } from '@/types/master.types'
import type {
  ProjectMasterColumnConfig,
  ProjectMasterColumnKey,
  ProjectDataQualityCode,
  ProjectDataQualityStage,
  ProjectMasterFilterState,
  ProjectMasterListParams,
  ProjectMasterRow,
  ProjectMasterSortField,
  ProjectMasterSortOrder,
  ProjectPipelineStatus,
  ProjectStatus,
} from '@/types/project.types'

const projectStore = useProjectStore()
const masterStore = useMasterStore()
const { can } = usePermission()
const toast = useToast()
const route = useRoute()
const hydratingRouteQuery = ref(false)

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
    data_quality_codes: 'Kelengkapan Data',
    data_quality_stages: 'Tahap Kelengkapan Data',
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
const activeFilterPills = computed(() =>
  listControls.activeFilterPills.value.filter(
    (pill) => pill.key !== 'include_history' || appliedFilters.include_history,
  ),
)
const activeFilterCount = computed(() => activeFilterPills.value.length)
const columnConfigs: ProjectMasterColumnConfig[] = [
  { key: 'loan_types', label: 'Jenis Pinjaman', sortField: 'loan_types', defaultVisible: true },
  { key: 'indication_lenders', label: 'Indikasi Lender', sortField: 'indication_lenders', defaultVisible: false },
  { key: 'executing_agencies', label: 'Executing Agency', sortField: 'executing_agencies', defaultVisible: true },
  { key: 'fixed_lenders', label: 'Fixed Lender', sortField: 'fixed_lenders', defaultVisible: true },
  { key: 'status', label: 'Status', sortField: 'project_status', defaultVisible: true },
  { key: 'program_title', label: 'Program Title', sortField: 'program_title', defaultVisible: false },
  { key: 'locations', label: 'Region/Location', sortField: 'locations', defaultVisible: false },
  { key: 'foreign_loan_usd', label: 'Nilai Pinjaman', sortField: 'foreign_loan_usd', defaultVisible: true },
  { key: 'dk_dates', label: 'Tanggal Daftar Kegiatan', sortField: 'dk_dates', defaultVisible: false },
]
const visibleColumnKeys = ref<ProjectMasterColumnKey[]>(
  columnConfigs.filter((column) => column.defaultVisible).map((column) => column.key),
)

const skeletonRows = computed(() => Array.from({ length: Math.min(limit.value, 10) }, (_, index) => index))
const visibleColumns = computed(() =>
  columnConfigs.filter((column) => visibleColumnKeys.value.includes(column.key)),
)
const initialTableLoading = computed(() => projectStore.loading && projectStore.projects.length === 0)
const refreshingExistingRows = computed(() => projectStore.loading && projectStore.projects.length > 0)
const columnSelectionLabel = computed(() => `${visibleColumns.value.length + 2} kolom tampil`)
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
const dataQualityCodeOptions: Array<{ label: string; value: ProjectDataQualityCode }> = [
  { label: 'Tanpa Kementerian/Lembaga', value: 'NO_EXECUTING_AGENCY' },
  { label: 'Tanpa lender', value: 'NO_LENDER' },
  { label: 'Tanpa lokasi', value: 'NO_REGION' },
  { label: 'Tanpa nilai pendanaan', value: 'NO_FUNDING_AMOUNT' },
]
const dataQualityStageOptions: Array<{ label: string; value: ProjectDataQualityStage }> = [
  { label: 'Blue Book', value: 'Blue Book' },
  { label: 'Green Book Funding Source', value: 'Green Book Funding Source' },
  { label: 'Daftar Kegiatan Financing', value: 'Daftar Kegiatan Financing' },
]

const pipelineStatusLabels: Record<ProjectPipelineStatus, string> = {
  BB: 'Blue Book',
  GB: 'Green Book',
  DK: 'Daftar Kegiatan',
  LA: 'Loan Agreement',
  Monitoring: 'Monitoring',
}
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
    data_quality_codes: [],
    data_quality_stages: [],
    search: '',
    include_history: false,
  }
}

function assignFilterState(target: ProjectMasterFilterState, source: ProjectMasterFilterState) {
  target.loan_types = [...source.loan_types]
  target.indication_lender_ids = [...source.indication_lender_ids]
  target.executing_agency_ids = [...source.executing_agency_ids]
  target.fixed_lender_ids = [...source.fixed_lender_ids]
  target.project_statuses = [...source.project_statuses]
  target.pipeline_statuses = [...source.pipeline_statuses]
  target.program_title_ids = [...source.program_title_ids]
  target.region_ids = [...source.region_ids]
  target.foreign_loan_min = source.foreign_loan_min
  target.foreign_loan_max = source.foreign_loan_max
  target.dk_date_from = source.dk_date_from
  target.dk_date_to = source.dk_date_to
  target.data_quality_codes = [...source.data_quality_codes]
  target.data_quality_stages = [...source.data_quality_stages]
  target.search = source.search
  target.include_history = source.include_history
}

function routeQueryValues(key: string) {
  const value = route.query[key]
  const rawValues = Array.isArray(value) ? value : [value]

  return rawValues
    .flatMap((item) => (typeof item === 'string' ? item.split(',') : []))
    .map((item) => item.trim())
    .filter(Boolean)
}

function routeQueryString(key: string) {
  return routeQueryValues(key)[0] ?? ''
}

function routeQueryNumber(key: string) {
  const raw = routeQueryString(key)
  if (!raw) return null

  const parsed = Number(raw)

  return Number.isFinite(parsed) ? parsed : null
}

function routeQueryBoolean(key: string) {
  const raw = routeQueryString(key).toLowerCase()

  if (raw === 'true' || raw === '1') return true
  if (raw === 'false' || raw === '0') return false

  return false
}

function routeLenderTypes(key: string): LenderType[] {
  return routeQueryValues(key).filter(
    (value): value is LenderType =>
      value === 'Bilateral' || value === 'Multilateral' || value === 'KSA',
  )
}

function routeProjectStatuses(key: string): ProjectStatus[] {
  return routeQueryValues(key).filter(
    (value): value is ProjectStatus => value === 'Pipeline' || value === 'Ongoing',
  )
}

function routePipelineStatuses(key: string): ProjectPipelineStatus[] {
  return routeQueryValues(key).filter(
    (value): value is ProjectPipelineStatus =>
      value === 'BB' ||
      value === 'GB' ||
      value === 'DK' ||
      value === 'LA' ||
      value === 'Monitoring',
  )
}

function routeDataQualityCodes(key: string): ProjectDataQualityCode[] {
  return routeQueryValues(key).filter(
    (value): value is ProjectDataQualityCode =>
      value === 'NO_EXECUTING_AGENCY' ||
      value === 'NO_LENDER' ||
      value === 'NO_REGION' ||
      value === 'NO_FUNDING_AMOUNT',
  )
}

function routeDataQualityStages(key: string): ProjectDataQualityStage[] {
  return routeQueryValues(key).filter(
    (value): value is ProjectDataQualityStage =>
      value === 'Blue Book' ||
      value === 'Green Book Funding Source' ||
      value === 'Daftar Kegiatan Financing',
  )
}

function hydrateFiltersFromRouteQuery() {
  const next = createDefaultFilters()

  next.loan_types = routeLenderTypes('loan_types')
  next.indication_lender_ids = routeQueryValues('indication_lender_ids')
  next.executing_agency_ids = routeQueryValues('executing_agency_ids')
  next.fixed_lender_ids = routeQueryValues('fixed_lender_ids')
  next.project_statuses = routeProjectStatuses('project_statuses')
  next.pipeline_statuses = routePipelineStatuses('pipeline_statuses')
  next.program_title_ids = routeQueryValues('program_title_ids')
  next.region_ids = routeQueryValues('region_ids')
  next.foreign_loan_min = routeQueryNumber('foreign_loan_min')
  next.foreign_loan_max = routeQueryNumber('foreign_loan_max')
  next.dk_date_from = routeQueryString('dk_date_from')
  next.dk_date_to = routeQueryString('dk_date_to')
  next.data_quality_codes = routeDataQualityCodes('data_quality_codes')
  next.data_quality_stages = routeDataQualityStages('data_quality_stages')
  next.search = routeQueryString('search')
  next.include_history = routeQueryBoolean('include_history')

  hydratingRouteQuery.value = true
  assignFilterState(filters, next)
  assignFilterState(appliedFilters, next)
  listControls.search.value = next.search
  listControls.debouncedSearch.value = next.search
  window.queueMicrotask(() => {
    hydratingRouteQuery.value = false
  })
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
  if (appliedFilters.data_quality_codes.length > 0) {
    params.data_quality_codes = [...appliedFilters.data_quality_codes]
  }
  if (appliedFilters.data_quality_stages.length > 0) {
    params.data_quality_stages = [...appliedFilters.data_quality_stages]
  }
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

function sortBy(field: ProjectMasterSortField) {
  if (sortField.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortOrder.value = 'asc'
  }

  page.value = 1
}

function listLabel(values: string[]) {
  return values.length > 0 ? values.join(', ') : '-'
}

function statusLabel(project: ProjectMasterRow) {
  return `${project.project_status} - ${pipelineStatusLabels[project.pipeline_status]}`
}

function sortIcon(field: ProjectMasterSortField) {
  if (sortField.value !== field) {
    return 'pi pi-sort-alt text-surface-400'
  }

  return sortOrder.value === 'asc' ? 'pi pi-sort-amount-up-alt' : 'pi pi-sort-amount-down'
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

function headerAlignClass(column?: ProjectMasterColumnConfig) {
  return column?.key === 'foreign_loan_usd' ? 'justify-end text-right' : 'justify-start text-left'
}

function bodyCellClass(column: ProjectMasterColumnConfig) {
  if (column.key === 'foreign_loan_usd') {
    return 'px-4 py-3 text-right font-medium text-surface-900'
  }

  return 'px-4 py-3 text-surface-700'
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
      return selectedLabelSummary(value.map((item) => pipelineStatusLabels[item as ProjectPipelineStatus] ?? item))
    }

    if (key === 'data_quality_codes') {
      return selectedLabelSummary(
        value.map(
          (item) =>
            dataQualityCodeOptions.find((option) => option.value === item)?.label ?? String(item),
        ),
      )
    }

    if (key === 'data_quality_stages') {
      return selectedLabelSummary(
        value.map(
          (item) =>
            dataQualityStageOptions.find((option) => option.value === item)?.label ?? String(item),
        ),
      )
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
    if (hydratingRouteQuery.value) return
    void refreshFromFirstPage()
  },
)

watch(
  () => route.query,
  () => {
    hydrateFiltersFromRouteQuery()
    void refreshFromFirstPage()
  },
)

onMounted(() => {
  hydrateFiltersFromRouteQuery()
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
    />

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari nama proyek, indikasi lender, fixed lender, atau executing agency"
      :active-filters="activeFilterPills"
      :filter-count="activeFilterCount"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
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
            <span class="text-sm font-medium text-surface-700">Kelengkapan Data</span>
            <MultiSelect
              v-model="filters.data_quality_codes"
              :options="dataQualityCodeOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua isu"
              filter
              filter-placeholder="Cari isu data"
              display="chip"
              class="w-full"
            />
          </label>

          <label class="block space-y-2 xl:col-span-2">
            <span class="text-sm font-medium text-surface-700">Tahap Kelengkapan Data</span>
            <MultiSelect
              v-model="filters.data_quality_stages"
              :options="dataQualityStageOptions"
              option-label="label"
              option-value="value"
              placeholder="Semua tahap"
              filter
              filter-placeholder="Cari tahap"
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

    <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-surface-200 p-4">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Master Table Project</h2>
          <p class="text-sm text-surface-500">{{ projectStore.total }} project ditemukan.</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
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
          <Tag :value="columnSelectionLabel" severity="secondary" rounded />
          <MultiSelect
            v-model="visibleColumnKeys"
            :options="columnConfigs"
            option-label="label"
            option-value="key"
            placeholder="Kolom tampil"
            filter
            filter-placeholder="Cari kolom"
            class="w-64 max-w-full"
          />
        </div>
      </div>

      <div v-if="initialTableLoading" class="overflow-x-auto">
        <table class="w-full min-w-[72rem] table-fixed text-left text-sm">
          <tbody class="divide-y divide-surface-100">
            <tr v-for="row in skeletonRows" :key="row">
              <td class="w-[28rem] px-4 py-3">
                <Skeleton height="2rem" />
              </td>
              <td v-for="column in visibleColumns" :key="column.key" class="px-4 py-3">
                <Skeleton height="1.5rem" />
              </td>
              <td class="w-40 px-4 py-3">
                <Skeleton height="1.5rem" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else-if="projectStore.projects.length === 0" class="p-8">
        <EmptyState title="Tidak ada project" description="Ubah filter atau kata kunci pencarian." />
      </div>

      <TableReloadShell v-else :refreshing="refreshingExistingRows" content-class="overflow-x-auto">
        <table class="w-full min-w-[72rem] table-fixed text-left text-sm">
          <thead class="bg-surface-50 text-xs uppercase text-surface-500">
            <tr>
              <th class="w-[28rem] px-4 py-3">
                <button
                  type="button"
                  class="inline-flex w-full items-center gap-2 text-left font-semibold"
                  @click="sortBy('project_name')"
                >
                  <span>Nama Proyek</span>
                  <i :class="sortIcon('project_name')" aria-hidden="true" />
                </button>
              </th>
              <th
                v-for="column in visibleColumns"
                :key="column.key"
                class="px-4 py-3"
                :class="column.key === 'foreign_loan_usd' ? 'w-48 text-right' : 'w-60'"
              >
                <button
                  type="button"
                  class="inline-flex w-full items-center gap-2 font-semibold"
                  :class="headerAlignClass(column)"
                  @click="sortBy(column.sortField)"
                >
                  <span>{{ column.label }}</span>
                  <i :class="sortIcon(column.sortField)" aria-hidden="true" />
                </button>
              </th>
              <th class="w-40 px-4 py-3 text-right">Aksi</th>
            </tr>
          </thead>
          <TransitionGroup tag="tbody" name="prism-table-row-fade" class="divide-y divide-surface-100">
            <tr v-for="project in projectStore.projects" :key="project.id" class="align-top hover:bg-surface-50/70">
              <td class="w-[28rem] px-4 py-3">
                <RouterLink
                  :to="{ name: 'bb-project-detail', params: { bbId: project.blue_book_id, id: project.id } }"
                  class="block whitespace-normal text-sm font-semibold leading-relaxed text-surface-950 hover:text-primary-600"
                >
                  {{ project.project_name }}
                </RouterLink>
                <div class="mt-1 flex flex-wrap items-center gap-2">
                  <span class="text-xs font-medium text-surface-500">{{ project.bb_code }}</span>
                  <Tag
                    :value="project.blue_book_revision_label"
                    :severity="project.is_latest ? 'success' : 'secondary'"
                    rounded
                  />
                  <Tag
                    v-if="project.has_newer_revision"
                    value="Ada revisi lebih baru"
                    severity="warn"
                    rounded
                  />
                </div>
              </td>
              <td v-for="column in visibleColumns" :key="column.key" :class="bodyCellClass(column)">
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
                >
                  {{ listLabel(project.indication_lenders) }}
                </span>
                <span
                  v-else-if="column.key === 'executing_agencies'"
                  class="block whitespace-normal leading-relaxed"
                  :title="listLabel(project.executing_agencies)"
                >
                  {{ listLabel(project.executing_agencies) }}
                </span>
                <span
                  v-else-if="column.key === 'fixed_lenders'"
                  class="block whitespace-normal leading-relaxed"
                  :title="listLabel(project.fixed_lenders)"
                >
                  {{ listLabel(project.fixed_lenders) }}
                </span>
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
                >
                  {{ listLabel(project.locations) }}
                </span>
                <CurrencyDisplay
                  v-else-if="column.key === 'foreign_loan_usd'"
                  :amount="project.foreign_loan_usd"
                  currency="USD"
                />
                <span
                  v-else-if="column.key === 'dk_dates'"
                  class="block whitespace-normal leading-relaxed"
                  :title="listLabel(project.dk_dates)"
                >
                  {{ listLabel(project.dk_dates) }}
                </span>
              </td>
              <td class="w-40 px-4 py-3">
                <div class="flex justify-end gap-1.5">
                  <Button
                    v-tooltip.top="'Detail proyek'"
                    as="router-link"
                    :to="{ name: 'bb-project-detail', params: { bbId: project.blue_book_id, id: project.id } }"
                    icon="pi pi-eye"
                    severity="secondary"
                    size="small"
                    outlined
                    rounded
                    aria-label="Detail proyek"
                  />
                  <Button
                    v-tooltip.top="'Lihat perjalanan proyek'"
                    as="router-link"
                    :to="{ name: 'project-journey', params: { bbProjectId: project.id } }"
                    icon="pi pi-sitemap"
                    severity="secondary"
                    size="small"
                    outlined
                    rounded
                    aria-label="Lihat perjalanan proyek"
                  />
                  <Button
                    v-if="can('bb_project', 'update')"
                    v-tooltip.top="'Edit proyek'"
                    as="router-link"
                    :to="{ name: 'bb-project-edit', params: { bbId: project.blue_book_id, id: project.id } }"
                    icon="pi pi-pencil"
                    size="small"
                    outlined
                    rounded
                    aria-label="Edit proyek"
                  />
                </div>
              </td>
            </tr>
          </TransitionGroup>
        </table>
      </TableReloadShell>

      <div class="border-t border-surface-200 p-3">
        <ListPaginationFooter
          v-model:page="page"
          v-model:limit="limit"
          :total="projectStore.total"
        />
      </div>
    </section>
  </section>
</template>
