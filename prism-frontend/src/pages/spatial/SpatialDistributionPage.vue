<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import ToggleSwitch from 'primevue/toggleswitch'
import { useRouter } from 'vue-router'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import ListPaginationFooter from '@/components/common/ListPaginationFooter.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar, { type ActiveFilterPill } from '@/components/common/SearchFilterBar.vue'
import SpatialChoroplethMap from '@/components/spatial/SpatialChoroplethMap.vue'
import { useSpatialDistributionStore } from '@/stores/spatial-distribution.store'
import type { LenderType } from '@/types/master.types'
import type {
  ProjectMasterSortField,
  ProjectMasterSortOrder,
  ProjectPipelineStatus,
  ProjectStatus,
} from '@/types/project.types'
import type {
  SpatialDistributionLevel,
  SpatialDistributionMetric,
  SpatialDistributionParams,
  SpatialDistributionRegionMetric,
} from '@/types/spatial-distribution.types'
import { getPipelineStatusLabel, getPipelineStatusSeverity } from '@/utils/status-labels'

type FilterOption<T extends string> = {
  label: string
  value: T
}

interface SpatialFilterState {
  pipelineStatuses: ProjectPipelineStatus[]
  projectStatuses: ProjectStatus[]
  loanTypes: LenderType[]
  includeHistory: boolean
}

const store = useSpatialDistributionStore()
const router = useRouter()
const { choropleth, error, loadingMap, loadingProjects, projectError, projectList } = storeToRefs(store)

const level = ref<SpatialDistributionLevel>('province')
const provinceCode = ref<string | undefined>()
const provinceName = ref<string | undefined>()
const selectedRegion = ref<SpatialDistributionRegionMetric | null>(null)
const metric = ref<SpatialDistributionMetric>('count')
const projectPage = ref(1)
const projectLimit = ref(10)
const projectSortField = ref<ProjectMasterSortField>('project_name')
const projectSortOrder = ref<ProjectMasterSortOrder>('asc')
let searchTimer: ReturnType<typeof window.setTimeout> | undefined
let projectPaginationTimer: ReturnType<typeof window.setTimeout> | undefined
let searchWatcherPaused = false

const pipelineStatusValues: ProjectPipelineStatus[] = ['BB', 'GB', 'DK', 'LA', 'Monitoring']
const projectStatusValues: ProjectStatus[] = ['Pipeline', 'Ongoing']
const loanTypeValues: LenderType[] = ['Bilateral', 'Multilateral', 'KSA']
const indonesiaRegionCode = 'ID'

const filters = reactive<SpatialFilterState & {
  search: string
}>({
  pipelineStatuses: [...pipelineStatusValues],
  projectStatuses: [...projectStatusValues],
  loanTypes: [...loanTypeValues],
  search: '',
  includeHistory: false,
})

const appliedFilters = reactive<SpatialFilterState>({
  pipelineStatuses: [...pipelineStatusValues],
  projectStatuses: [...projectStatusValues],
  loanTypes: [...loanTypeValues],
  includeHistory: false,
})

const metricOptions: Array<FilterOption<SpatialDistributionMetric>> = [
  { label: 'Jumlah Proyek', value: 'count' },
  { label: 'Nilai Pinjaman', value: 'value' },
]

const pipelineOptions: Array<FilterOption<ProjectPipelineStatus>> = [
  { label: 'Blue Book', value: 'BB' },
  { label: 'Green Book', value: 'GB' },
  { label: 'Daftar Kegiatan', value: 'DK' },
  { label: 'Loan Agreement', value: 'LA' },
  { label: 'Monitoring', value: 'Monitoring' },
]

const projectStatusOptions: Array<FilterOption<ProjectStatus>> = [
  { label: 'Pipeline', value: 'Pipeline' },
  { label: 'Ongoing', value: 'Ongoing' },
]

const loanTypeOptions: Array<FilterOption<LenderType>> = [
  { label: 'Bilateral', value: 'Bilateral' },
  { label: 'Multilateral', value: 'Multilateral' },
  { label: 'KSA', value: 'KSA' },
]

const numberFormatter = new Intl.NumberFormat('id-ID')
const compactUsdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  notation: 'compact',
  maximumFractionDigits: 2,
})
const decimalFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 1,
})

const mapTitle = computed(() =>
  level.value === 'province'
    ? 'Peta Choropleth Indonesia'
    : `Peta Choropleth Kab/Kota ${provinceName.value ?? ''}`.trim(),
)

const topRegion = computed(() => {
  const regions = choropleth.value.regions.filter((region) =>
    metric.value === 'count' ? region.project_count > 0 : region.total_loan_usd > 0,
  )
  if (metric.value === 'count') {
    return regions.sort((left, right) => right.project_count - left.project_count)[0] ?? null
  }
  return regions.sort((left, right) => right.total_loan_usd - left.total_loan_usd)[0] ?? null
})

const indonesiaFocusRegion = computed<SpatialDistributionRegionMetric>(() => ({
  region_id: projectList.value?.region_code === indonesiaRegionCode ? projectList.value.region_id : indonesiaRegionCode,
  region_code: indonesiaRegionCode,
  region_name: 'Indonesia',
  region_type: 'COUNTRY',
  project_count: 0,
  total_loan_usd: 0,
}))

const activeFocusRegion = computed(() => {
  if (selectedRegion.value) return selectedRegion.value
  if (level.value === 'province') return indonesiaFocusRegion.value
  return null
})

const focusRegionName = computed(() =>
  activeFocusRegion.value?.region_name ?? provinceName.value ?? 'Wilayah',
)

const projectListMatchesFocus = computed(() =>
  Boolean(activeFocusRegion.value && projectList.value?.region_code === activeFocusRegion.value.region_code),
)

const focusProjectCount = computed(() => {
  if (projectListMatchesFocus.value) return projectList.value?.meta.total ?? 0
  return activeFocusRegion.value?.project_count ?? 0
})

const focusLoanUsd = computed(() => {
  if (projectListMatchesFocus.value) return projectList.value?.summary.total_loan_usd ?? 0
  return activeFocusRegion.value?.total_loan_usd ?? 0
})

const comparisonRegionCount = computed(() =>
  choropleth.value.summary.total_regions || choropleth.value.regions.length,
)

const averageProjectCount = computed(() => averagePerRegion(choropleth.value.summary.total_project_count))
const averageLoanUsd = computed(() => averagePerRegion(choropleth.value.summary.total_loan_usd))

const comparisonScopeLabel = computed(() =>
  level.value === 'province'
    ? 'rata-rata nasional'
    : `rata-rata kab/kota ${provinceName.value ?? 'provinsi'}`,
)

const activeComparisonBadge = computed(() => {
  const current = metric.value === 'count' ? focusProjectCount.value : focusLoanUsd.value
  const average = metric.value === 'count' ? averageProjectCount.value : averageLoanUsd.value
  return comparisonBadge(current, average)
})

const pipelineSelectionLabel = computed(() =>
  selectionControlLabel(filters.pipelineStatuses, pipelineOptions, 'Semua Tahap', 'tahap'),
)

const projectStatusSelectionLabel = computed(() =>
  selectionControlLabel(filters.projectStatuses, projectStatusOptions, 'Semua Status', 'status'),
)

const loanTypeSelectionLabel = computed(() =>
  selectionControlLabel(filters.loanTypes, loanTypeOptions, 'Semua Tipe Pinjaman', 'tipe'),
)

const activeFilterPills = computed<ActiveFilterPill[]>(() => {
  const pills: ActiveFilterPill[] = []
  if (metric.value !== 'count') {
    pills.push({ key: 'metric', label: 'Metrik', value: 'Nilai Pinjaman' })
  }
  if (isFilteredSelection(appliedFilters.pipelineStatuses, pipelineStatusValues)) {
    pills.push({
      key: 'pipelineStatuses',
      label: 'Tahap',
      value: selectionPillValue(appliedFilters.pipelineStatuses, pipelineOptions),
    })
  }
  if (isFilteredSelection(appliedFilters.projectStatuses, projectStatusValues)) {
    pills.push({
      key: 'projectStatuses',
      label: 'Status',
      value: selectionPillValue(appliedFilters.projectStatuses, projectStatusOptions),
    })
  }
  if (isFilteredSelection(appliedFilters.loanTypes, loanTypeValues)) {
    pills.push({
      key: 'loanTypes',
      label: 'Tipe',
      value: selectionPillValue(appliedFilters.loanTypes, loanTypeOptions),
    })
  }
  if (appliedFilters.includeHistory) {
    pills.push({ key: 'includeHistory', label: 'Riwayat revisi', value: 'Ditampilkan' })
  }
  return pills
})

const activeFilterCount = computed(() => activeFilterPills.value.length)

function buildParams(): SpatialDistributionParams {
  return {
    level: level.value,
    province_code: level.value === 'city' ? provinceCode.value : undefined,
    pipeline_statuses: selectedFilterValues(appliedFilters.pipelineStatuses, pipelineStatusValues),
    project_statuses: selectedFilterValues(appliedFilters.projectStatuses, projectStatusValues),
    loan_types: selectedFilterValues(appliedFilters.loanTypes, loanTypeValues),
    search: filters.search.trim() || undefined,
    include_history: appliedFilters.includeHistory || undefined,
  }
}

async function loadMap() {
  await store.fetchChoropleth(buildParams())

  if (!selectedRegion.value) {
    if (level.value === 'province') {
      await loadProjects()
    } else {
      store.clearProjectList()
    }
    return
  }

  const stillVisible = choropleth.value.regions.find(
    (region) => region.region_code === selectedRegion.value?.region_code,
  )

  if (stillVisible) {
    selectedRegion.value = stillVisible
    await loadProjects()
  } else {
    selectedRegion.value = null
    if (level.value === 'province') {
      await loadProjects()
    } else {
      store.clearProjectList()
    }
  }
}

async function loadProjects() {
  const focusRegion = activeFocusRegion.value

  if (!focusRegion) {
    store.clearProjectList()
    return
  }

  const projectLevel: SpatialDistributionLevel = focusRegion.region_type === 'CITY' ? 'city' : 'province'

  await store.fetchRegionProjects({
    ...buildParams(),
    level: projectLevel,
    province_code: projectLevel === 'city' ? provinceCode.value : undefined,
    region_code: focusRegion.region_code,
    page: projectPage.value,
    limit: projectLimit.value,
    sort: projectSortField.value,
    order: projectSortOrder.value,
  })
}

async function selectRegion(region: SpatialDistributionRegionMetric) {
  selectedRegion.value = region
  projectPage.value = 1
  await loadProjects()
}

async function applyFilters() {
  clearSearchTimer()
  syncAppliedFiltersFromDraft()
  projectPage.value = 1
  await loadMap()
}

async function resetFilters() {
  searchWatcherPaused = true
  clearSearchTimer()

  try {
    filters.pipelineStatuses = [...pipelineStatusValues]
    filters.projectStatuses = [...projectStatusValues]
    filters.loanTypes = [...loanTypeValues]
    filters.search = ''
    filters.includeHistory = false
    syncAppliedFiltersFromDraft()
    metric.value = 'count'
    projectPage.value = 1
    await loadMap()
  } finally {
    searchWatcherPaused = false
  }
}

async function drillDownSelectedProvince() {
  if (!selectedRegion.value || selectedRegion.value.region_type !== 'PROVINCE') return

  level.value = 'city'
  provinceCode.value = selectedRegion.value.region_code
  provinceName.value = selectedRegion.value.region_name
  selectedRegion.value = null
  projectPage.value = 1
  store.clearProjectList()
  await loadMap()
}

async function backToIndonesia() {
  level.value = 'province'
  provinceCode.value = undefined
  provinceName.value = undefined
  selectedRegion.value = null
  projectPage.value = 1
  await loadMap()
}

function updateProjectPage(page: number) {
  if (page < 1 || page > (projectList.value?.meta.total_pages ?? 1)) return
  projectPage.value = page
  scheduleProjectListReload()
}

function updateProjectLimit(limit: number) {
  projectLimit.value = limit
  projectPage.value = 1
  scheduleProjectListReload()
}

function scheduleProjectListReload() {
  clearProjectPaginationTimer()
  projectPaginationTimer = window.setTimeout(() => {
    projectPaginationTimer = undefined
    void loadProjects()
  }, 0)
}

async function setProjectSort(field: ProjectMasterSortField) {
  clearProjectPaginationTimer()

  if (projectSortField.value === field) {
    projectSortOrder.value = projectSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    projectSortField.value = field
    projectSortOrder.value = field === 'foreign_loan_usd' ? 'desc' : 'asc'
  }

  projectPage.value = 1
  await loadProjects()
}

function sortIcon(field: ProjectMasterSortField) {
  if (projectSortField.value !== field) return 'pi pi-sort-alt'
  return projectSortOrder.value === 'asc' ? 'pi pi-sort-up' : 'pi pi-sort-down'
}

function sortAriaLabel(field: ProjectMasterSortField, label: string) {
  if (projectSortField.value !== field) return `Urutkan ${label}`
  return `Urutkan ${label} ${projectSortOrder.value === 'asc' ? 'menurun' : 'menaik'}`
}

function selectedFilterValues<T extends string>(selected: T[], allValues: T[]) {
  if (selected.length === 0 || selected.length === allValues.length) return undefined
  return [...selected]
}

function isFilteredSelection<T extends string>(selected: T[], allValues: T[]) {
  return selected.length > 0 && selected.length < allValues.length
}

function selectionControlLabel<T extends string>(
  selected: T[],
  options: Array<FilterOption<T>>,
  allLabel: string,
  itemLabel: string,
) {
  if (selected.length === 0 || selected.length === options.length) return allLabel
  if (selected.length === 1) {
    return options.find((option) => option.value === selected[0])?.label ?? allLabel
  }

  return `${selected.length} ${itemLabel} dipilih`
}

function selectionPillValue<T extends string>(
  selected: T[],
  options: Array<FilterOption<T>>,
) {
  const labels = selected
    .map((value) => options.find((option) => option.value === value)?.label)
    .filter((label): label is string => Boolean(label))

  if (labels.length <= 2) return labels.join(', ')
  return `${labels.length} dipilih`
}

function averagePerRegion(total: number) {
  if (comparisonRegionCount.value <= 0) return 0
  return total / comparisonRegionCount.value
}

function comparisonBadge(current: number, average: number) {
  if (average <= 0) {
    return current > 0 ? 'Di atas rata-rata' : 'Setara rata-rata'
  }

  const ratio = current / average
  if (ratio > 1.05) return 'Di atas rata-rata'
  if (ratio < 0.95) return 'Di bawah rata-rata'
  return 'Setara rata-rata'
}

function comparisonBadgeClass(label: string) {
  if (label === 'Di atas rata-rata') return 'bg-emerald-50 text-emerald-700 ring-emerald-200'
  if (label === 'Di bawah rata-rata') return 'bg-amber-50 text-amber-700 ring-amber-200'
  return 'bg-sky-50 text-sky-700 ring-sky-200'
}

function comparisonSubtitle(current: number, average: number) {
  if (average <= 0) return 'Belum ada pembanding'

  const ratio = current / average
  if (ratio > 1.05) return `${decimalFormatter.format(ratio)}x ${comparisonScopeLabel.value}`
  if (ratio < 0.95) return `${decimalFormatter.format(ratio * 100)}% ${comparisonScopeLabel.value}`
  return `Setara ${comparisonScopeLabel.value}`
}

function comparisonBarWidth(current: number, average: number) {
  const max = Math.max(current, average)
  if (max <= 0) return '0%'
  return `${Math.max((current / max) * 100, current > 0 ? 3 : 0)}%`
}

function comparisonMarkerPosition(current: number, average: number) {
  const max = Math.max(current, average)
  if (max <= 0) return '0%'
  return `${(average / max) * 100}%`
}

function formatComparisonNumber(value: number) {
  return Number.isInteger(value) ? numberFormatter.format(value) : decimalFormatter.format(value)
}

function formatPanelUsd(value: number) {
  return compactUsdFormatter.format(value).replace('$', 'US$')
}

function clearSearchTimer() {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
    searchTimer = undefined
  }
}

function clearProjectPaginationTimer() {
  if (projectPaginationTimer) {
    window.clearTimeout(projectPaginationTimer)
    projectPaginationTimer = undefined
  }
}

function syncAppliedFiltersFromDraft() {
  appliedFilters.pipelineStatuses = [...filters.pipelineStatuses]
  appliedFilters.projectStatuses = [...filters.projectStatuses]
  appliedFilters.loanTypes = [...filters.loanTypes]
  appliedFilters.includeHistory = filters.includeHistory
}

function ensurePipelineSelection() {
  if (filters.pipelineStatuses.length === 0) filters.pipelineStatuses = [...pipelineStatusValues]
}

function ensureProjectStatusSelection() {
  if (filters.projectStatuses.length === 0) filters.projectStatuses = [...projectStatusValues]
}

function ensureLoanTypeSelection() {
  if (filters.loanTypes.length === 0) filters.loanTypes = [...loanTypeValues]
}

function pipelineStatusLabel(status: ProjectPipelineStatus) {
  return getPipelineStatusLabel(status)
}

function statusSeverity(status: ProjectPipelineStatus) {
  return getPipelineStatusSeverity(status)
}

function goToProjectDetail(projectID: string) {
  void router.push(`/projects/${projectID}`)
}

async function focusIndonesia() {
  await backToIndonesia()
}

async function removeFilter(key: string) {
  clearSearchTimer()

  if (key === 'metric') metric.value = 'count'
  if (key === 'pipelineStatuses') {
    filters.pipelineStatuses = [...pipelineStatusValues]
    appliedFilters.pipelineStatuses = [...pipelineStatusValues]
  }
  if (key === 'projectStatuses') {
    filters.projectStatuses = [...projectStatusValues]
    appliedFilters.projectStatuses = [...projectStatusValues]
  }
  if (key === 'loanTypes') {
    filters.loanTypes = [...loanTypeValues]
    appliedFilters.loanTypes = [...loanTypeValues]
  }
  if (key === 'includeHistory') {
    filters.includeHistory = false
    appliedFilters.includeHistory = false
  }

  projectPage.value = 1
  await loadMap()
}

watch(
  () => filters.search,
  () => {
    if (searchWatcherPaused) return

    clearSearchTimer()
    searchTimer = window.setTimeout(() => {
      projectPage.value = 1
      void loadMap()
    }, 500)
  },
)

onMounted(() => {
  void loadMap()
})

onUnmounted(() => {
  clearSearchTimer()
  clearProjectPaginationTimer()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Sebaran Wilayah"
      subtitle="Peta choropleth pinjaman luar negeri dengan drilldown provinsi sampai kabupaten/kota."
    />

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <article class="rounded-lg border border-surface-200 bg-white p-5 shadow-sm">
        <p class="text-sm text-surface-500">Total Proyek</p>
        <p class="mt-1 text-2xl font-semibold text-surface-950">
          {{ numberFormatter.format(choropleth.summary.total_project_count) }}
        </p>
      </article>
      <article class="rounded-lg border border-surface-200 bg-white p-5 shadow-sm">
        <p class="text-sm text-surface-500">Wilayah Aktif</p>
        <p class="mt-1 text-2xl font-semibold text-surface-950">
          {{ numberFormatter.format(choropleth.summary.active_regions) }}
          <span class="text-base font-medium text-surface-500">/ {{ numberFormatter.format(choropleth.summary.total_regions) }}</span>
        </p>
      </article>
      <article class="rounded-lg border border-surface-200 bg-white p-5 shadow-sm">
        <p class="text-sm text-surface-500">Total Nilai Pinjaman</p>
        <p class="mt-1 text-2xl font-semibold text-surface-950">
          {{ compactUsdFormatter.format(choropleth.summary.total_loan_usd) }}
        </p>
      </article>
      <article class="rounded-lg border border-surface-200 bg-white p-5 shadow-sm">
        <p class="text-sm text-surface-500">Wilayah Tertinggi</p>
        <p class="mt-1 truncate text-2xl font-semibold text-surface-950">
          {{ topRegion?.region_name ?? '-' }}
        </p>
      </article>
    </section>

    <SearchFilterBar
      v-model:search="filters.search"
      search-placeholder="Cari kode atau nama proyek"
      :active-filters="activeFilterPills"
      :filter-count="activeFilterCount"
      @apply="applyFilters"
      @reset="resetFilters"
      @remove="removeFilter"
    >
      <template #filters>
        <label class="flex items-center gap-3 rounded-lg border border-surface-200 px-3 py-2 xl:col-span-6">
          <ToggleSwitch v-model="filters.includeHistory" />
          <span class="text-sm font-medium text-surface-700">Tampilkan snapshot historis</span>
        </label>

        <label class="block min-w-0 space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Metrik</span>
          <Select
            v-model="metric"
            :options="metricOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>

        <label class="block min-w-0 space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Tahap</span>
          <MultiSelect
            v-model="filters.pipelineStatuses"
            :options="pipelineOptions"
            option-label="label"
            option-value="value"
            filter
            filter-placeholder="Cari tahap"
            :show-toggle-all="false"
            class="w-full"
            @change="ensurePipelineSelection"
          >
            <template #value>
              <span class="block truncate">{{ pipelineSelectionLabel }}</span>
            </template>
          </MultiSelect>
        </label>

        <label class="block min-w-0 space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Status Proyek</span>
          <MultiSelect
            v-model="filters.projectStatuses"
            :options="projectStatusOptions"
            option-label="label"
            option-value="value"
            filter
            filter-placeholder="Cari status"
            :show-toggle-all="false"
            class="w-full"
            @change="ensureProjectStatusSelection"
          >
            <template #value>
              <span class="block truncate">{{ projectStatusSelectionLabel }}</span>
            </template>
          </MultiSelect>
        </label>

        <label class="block min-w-0 space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Tipe Pinjaman</span>
          <MultiSelect
            v-model="filters.loanTypes"
            :options="loanTypeOptions"
            option-label="label"
            option-value="value"
            filter
            filter-placeholder="Cari tipe pinjaman"
            :show-toggle-all="false"
            class="w-full"
            @change="ensureLoanTypeSelection"
          >
            <template #value>
              <span class="block truncate">{{ loanTypeSelectionLabel }}</span>
            </template>
          </MultiSelect>
        </label>
      </template>
    </SearchFilterBar>

    <section class="space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-xl font-semibold text-surface-950">{{ mapTitle }}</h2>
          <p class="text-sm text-surface-500">
            Klik wilayah pada peta untuk melihat daftar proyek. Gunakan tombol drilldown untuk masuk ke kabupaten/kota.
          </p>
        </div>
        <Button
          v-if="level === 'city'"
          label="Kembali ke Indonesia"
          icon="pi pi-arrow-left"
          severity="secondary"
          outlined
          @click="backToIndonesia"
        />
        <Button
          v-else-if="selectedRegion"
          label="Fokus Indonesia"
          icon="pi pi-globe"
          severity="secondary"
          outlined
          @click="focusIndonesia"
        />
      </div>

      <p v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        {{ error }}
      </p>

      <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_22rem] xl:items-stretch">
        <SpatialChoroplethMap
          :level="level"
          :metric="metric"
          :province-code="provinceCode"
          :province-name="provinceName"
          :regions="choropleth.regions"
          :selected-region-code="selectedRegion?.region_code"
          :loading="loadingMap"
          @select="selectRegion"
        />

      <aside class="flex flex-col rounded-lg border border-surface-200 bg-white p-5 shadow-sm xl:h-full xl:min-h-[31rem]">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-prism-teal-dark">Fokus</p>
            <h3 class="mt-1 truncate text-2xl font-semibold text-surface-950">
              {{ focusRegionName }}
            </h3>
          </div>
        </div>

        <dl class="mt-5 grid grid-cols-2 gap-3">
          <div class="rounded-lg border border-surface-100 bg-surface-50 p-3">
            <dt class="text-xs font-medium text-surface-500">Project</dt>
            <dd class="mt-3 text-xl font-semibold text-surface-950">
              {{ numberFormatter.format(focusProjectCount) }}
            </dd>
            <p class="mt-2 text-xs font-semibold text-surface-500">
              {{ comparisonSubtitle(focusProjectCount, averageProjectCount) }}
            </p>
          </div>
          <div class="rounded-lg border border-surface-100 bg-surface-50 p-3">
            <dt class="text-xs font-medium text-surface-500">Nilai total</dt>
            <dd class="mt-3 text-xl font-semibold text-surface-950">
              {{ formatPanelUsd(focusLoanUsd) }}
            </dd>
            <p class="mt-2 text-xs font-semibold text-surface-500">
              {{ comparisonSubtitle(focusLoanUsd, averageLoanUsd) }}
            </p>
          </div>
        </dl>

        <div class="mt-5 rounded-lg border border-surface-200 p-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <p class="text-xs font-semibold text-surface-700">Vs {{ comparisonScopeLabel }}</p>
            <span
              class="rounded-lg px-2.5 py-1 text-xs font-semibold ring-1"
              :class="comparisonBadgeClass(activeComparisonBadge)"
            >
              {{ activeComparisonBadge }}
            </span>
          </div>

          <div class="mt-5 space-y-4">
            <div>
              <div class="flex items-center justify-between gap-3 text-xs">
                <span class="font-medium text-surface-500">Project</span>
                <span class="font-semibold text-surface-950">
                  {{ formatComparisonNumber(focusProjectCount) }} / {{ formatComparisonNumber(averageProjectCount) }}
                </span>
              </div>
              <div class="relative mt-2 h-1.5 rounded-full bg-surface-200">
                <div
                  class="h-full rounded-full bg-prism-teal"
                  :style="{ width: comparisonBarWidth(focusProjectCount, averageProjectCount) }"
                ></div>
                <span
                  v-if="averageProjectCount > 0"
                  class="absolute top-1/2 h-3 w-px -translate-y-1/2 bg-prism-gold"
                  :style="{ left: comparisonMarkerPosition(focusProjectCount, averageProjectCount) }"
                ></span>
              </div>
            </div>

            <div>
              <div class="flex items-center justify-between gap-3 text-xs">
                <span class="font-medium text-surface-500">Nilai</span>
                <span class="font-semibold text-surface-950">
                  {{ formatPanelUsd(focusLoanUsd) }} / {{ formatPanelUsd(averageLoanUsd) }}
                </span>
              </div>
              <div class="relative mt-2 h-1.5 rounded-full bg-surface-200">
                <div
                  class="h-full rounded-full bg-prism-teal"
                  :style="{ width: comparisonBarWidth(focusLoanUsd, averageLoanUsd) }"
                ></div>
                <span
                  v-if="averageLoanUsd > 0"
                  class="absolute top-1/2 h-3 w-px -translate-y-1/2 bg-prism-gold"
                  :style="{ left: comparisonMarkerPosition(focusLoanUsd, averageLoanUsd) }"
                ></span>
              </div>
            </div>
          </div>

          <p class="mt-4 text-xs text-surface-500">
            {{ focusRegionName }} / {{ comparisonScopeLabel }}
          </p>
        </div>

        <Button
          v-if="level === 'province' && selectedRegion?.region_type === 'PROVINCE'"
          class="mt-auto w-full"
          label="Lihat Kab/Kota"
          icon="pi pi-map-marker"
          @click="drillDownSelectedProvince"
        />
      </aside>
      </div>
    </section>

    <section class="overflow-hidden rounded-lg border border-surface-200 bg-white shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-surface-200 p-4">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">
            Proyek Wilayah {{ activeFocusRegion?.region_name ?? 'Terpilih' }}
          </h2>
          <p class="text-sm text-surface-500">
            Daftar ini mengikuti filter peta dan cakupan wilayah fokus.
          </p>
        </div>
        <div class="text-sm font-medium text-surface-600">
          {{ numberFormatter.format(projectList?.meta.total ?? 0) }} proyek
        </div>
      </div>

      <p v-if="projectError" class="m-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        {{ projectError }}
      </p>

      <div class="overflow-x-auto">
        <table class="w-full min-w-[62rem] text-left text-sm">
          <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
            <tr>
              <th class="px-4 py-3">Kode</th>
              <th class="px-4 py-3">
                <button
                  type="button"
                  class="inline-flex items-center gap-2 rounded-md text-left font-semibold text-surface-500 transition hover:text-prism-teal-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-prism-gold"
                  :aria-label="sortAriaLabel('project_name', 'nama proyek')"
                  @click="setProjectSort('project_name')"
                >
                  <span>Nama Proyek</span>
                  <i :class="[sortIcon('project_name'), 'text-[11px]']" />
                </button>
              </th>
              <th class="px-4 py-3">
                <span class="inline-flex items-center gap-1.5">
                  <span>Tahap</span>
                  <i
                    v-tooltip.top="'BB: Blue Book, GB: Green Book, DK: Daftar Kegiatan, LA: Loan Agreement'"
                    class="pi pi-info-circle text-[11px] text-surface-400"
                  />
                </span>
              </th>
              <th class="px-4 py-3">Executing Agency</th>
              <th class="px-4 py-3">Lender</th>
              <th class="px-4 py-3 text-right">
                <button
                  type="button"
                  class="ml-auto inline-flex items-center gap-2 rounded-md font-semibold text-surface-500 transition hover:text-prism-teal-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-prism-gold"
                  :aria-label="sortAriaLabel('foreign_loan_usd', 'pinjaman USD')"
                  @click="setProjectSort('foreign_loan_usd')"
                >
                  <span>Pinjaman USD</span>
                  <i :class="[sortIcon('foreign_loan_usd'), 'text-[11px]']" />
                </button>
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-surface-100">
            <tr v-if="!activeFocusRegion">
              <td colspan="6" class="px-4 py-8 text-center text-surface-500">
                Pilih provinsi atau kabupaten/kota pada peta untuk melihat proyek.
              </td>
            </tr>
            <tr v-else-if="loadingProjects">
              <td colspan="6" class="px-4 py-8 text-center text-surface-500">Memuat proyek...</td>
            </tr>
            <tr v-else-if="(projectList?.data.length ?? 0) === 0">
              <td colspan="6" class="px-4 py-8 text-center text-surface-500">Tidak ada proyek untuk wilayah dan filter ini.</td>
            </tr>
            <tr
              v-for="project in projectList?.data ?? []"
              :key="project.id"
              class="group cursor-pointer transition-colors hover:bg-teal-50/45 focus-visible:bg-teal-50/60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-inset focus-visible:outline-prism-gold"
              role="link"
              tabindex="0"
              @click="goToProjectDetail(project.id)"
              @keydown.enter.prevent="goToProjectDetail(project.id)"
              @keydown.space.prevent="goToProjectDetail(project.id)"
            >
              <td class="px-4 py-3 font-medium text-surface-700">{{ project.bb_code }}</td>
              <td class="px-4 py-3">
                <span class="font-semibold text-prism-teal-dark group-hover:underline">
                  {{ project.project_name }}
                </span>
                <p class="mt-1 text-xs text-surface-500">{{ project.program_title || 'Tanpa judul program' }}</p>
              </td>
              <td class="px-4 py-3">
                <Tag
                  v-tooltip.top="pipelineStatusLabel(project.pipeline_status)"
                  :value="project.pipeline_status"
                  :severity="statusSeverity(project.pipeline_status)"
                  rounded
                />
              </td>
              <td class="px-4 py-3 text-surface-600">
                {{ project.executing_agencies.join(', ') || '-' }}
              </td>
              <td class="px-4 py-3 text-surface-600">
                {{ [...project.fixed_lenders, ...project.indication_lenders].filter(Boolean).join(', ') || '-' }}
              </td>
              <td class="px-4 py-3 text-right font-medium text-surface-900">
                <CurrencyDisplay :amount="project.foreign_loan_usd" currency="USD" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="activeFocusRegion" class="border-t border-surface-200 p-3">
        <ListPaginationFooter
          :page="projectPage"
          :limit="projectLimit"
          :total="projectList?.meta.total ?? 0"
          @update:page="updateProjectPage"
          @update:limit="updateProjectLimit"
        />
      </div>
    </section>
  </section>
</template>
