import { computed, reactive, ref } from 'vue'
import { DashboardService } from '@/services/dashboard.service'
import { useAnalyticsDrilldown } from '@/composables/useAnalyticsDrilldown'
import { useMasterStore } from '@/stores/master.store'
import type {
  DashboardAbsorptionAnalytics,
  DashboardAnalyticsFilterParams,
  DashboardAnalyticsFilterState,
  DashboardAnalyticsOverview,
  DashboardInstitutionAnalytics,
  DashboardLenderAnalytics,
  DashboardLenderProportionAnalytics,
  DashboardRiskAnalytics,
  DashboardYearlyAnalytics,
} from '@/types/dashboard.types'

export type DashboardAnalyticsSectionKey =
  | 'overview'
  | 'institutions'
  | 'lenders'
  | 'absorption'
  | 'yearly'
  | 'lenderProportion'
  | 'risks'

const sectionKeys: DashboardAnalyticsSectionKey[] = [
  'overview',
  'institutions',
  'lenders',
  'absorption',
  'yearly',
  'lenderProportion',
  'risks',
]

function createDefaultFilters(): DashboardAnalyticsFilterState {
  return {
    budget_year: new Date().getFullYear(),
    quarter: null,
    lender_ids: [],
    lender_types: [],
    institution_ids: [],
    pipeline_statuses: [],
    project_statuses: [],
    region_ids: [],
    program_title_ids: [],
    foreign_loan_min: null,
    foreign_loan_max: null,
    include_history: false,
  }
}

function cloneFilters(filters: DashboardAnalyticsFilterState): DashboardAnalyticsFilterState {
  return {
    budget_year: filters.budget_year,
    quarter: filters.quarter,
    lender_ids: [...filters.lender_ids],
    lender_types: [...filters.lender_types],
    institution_ids: [...filters.institution_ids],
    pipeline_statuses: [...filters.pipeline_statuses],
    project_statuses: [...filters.project_statuses],
    region_ids: [...filters.region_ids],
    program_title_ids: [...filters.program_title_ids],
    foreign_loan_min: filters.foreign_loan_min,
    foreign_loan_max: filters.foreign_loan_max,
    include_history: filters.include_history,
  }
}

function assignFilters(
  target: DashboardAnalyticsFilterState,
  source: DashboardAnalyticsFilterState,
) {
  target.budget_year = source.budget_year
  target.quarter = source.quarter
  target.lender_ids = [...source.lender_ids]
  target.lender_types = [...source.lender_types]
  target.institution_ids = [...source.institution_ids]
  target.pipeline_statuses = [...source.pipeline_statuses]
  target.project_statuses = [...source.project_statuses]
  target.region_ids = [...source.region_ids]
  target.program_title_ids = [...source.program_title_ids]
  target.foreign_loan_min = source.foreign_loan_min
  target.foreign_loan_max = source.foreign_loan_max
  target.include_history = source.include_history
}

function buildParams(filters: DashboardAnalyticsFilterState): DashboardAnalyticsFilterParams {
  const params: DashboardAnalyticsFilterParams = {}

  if (filters.budget_year !== null) params.budget_year = filters.budget_year
  if (filters.quarter !== null) params.quarter = filters.quarter
  if (filters.lender_ids.length > 0) params.lender_ids = [...filters.lender_ids]
  if (filters.lender_types.length > 0) params.lender_types = [...filters.lender_types]
  if (filters.institution_ids.length > 0) params.institution_ids = [...filters.institution_ids]
  if (filters.pipeline_statuses.length > 0) {
    params.pipeline_statuses = [...filters.pipeline_statuses]
  }
  if (filters.project_statuses.length > 0) params.project_statuses = [...filters.project_statuses]
  if (filters.region_ids.length > 0) params.region_ids = [...filters.region_ids]
  if (filters.program_title_ids.length > 0)
    params.program_title_ids = [...filters.program_title_ids]
  if (filters.foreign_loan_min !== null) params.foreign_loan_min = filters.foreign_loan_min
  if (filters.foreign_loan_max !== null) params.foreign_loan_max = filters.foreign_loan_max
  if (filters.include_history) params.include_history = true

  return params
}

function errorMessage(error: unknown) {
  if (error instanceof Error && error.message) return error.message

  return 'Data analytics tidak dapat dimuat saat ini.'
}

export function useDashboardAnalytics() {
  const masterStore = useMasterStore()
  const drilldown = useAnalyticsDrilldown()
  const defaultFilters = createDefaultFilters()
  const draftFilters = reactive<DashboardAnalyticsFilterState>(cloneFilters(defaultFilters))
  const appliedFilters = reactive<DashboardAnalyticsFilterState>(cloneFilters(defaultFilters))

  const overview = ref<DashboardAnalyticsOverview | null>(null)
  const institutions = ref<DashboardInstitutionAnalytics | null>(null)
  const lenders = ref<DashboardLenderAnalytics | null>(null)
  const absorption = ref<DashboardAbsorptionAnalytics | null>(null)
  const yearly = ref<DashboardYearlyAnalytics | null>(null)
  const lenderProportion = ref<DashboardLenderProportionAnalytics | null>(null)
  const risks = ref<DashboardRiskAnalytics | null>(null)
  const loading = reactive<Record<DashboardAnalyticsSectionKey, boolean>>({
    overview: false,
    institutions: false,
    lenders: false,
    absorption: false,
    yearly: false,
    lenderProportion: false,
    risks: false,
  })
  const errors = reactive<Record<DashboardAnalyticsSectionKey, string | null>>({
    overview: null,
    institutions: null,
    lenders: null,
    absorption: null,
    yearly: null,
    lenderProportion: null,
    risks: null,
  })
  const loadingMasterData = ref(false)

  const params = computed(() => buildParams(appliedFilters))
  const anyLoading = computed(() => sectionKeys.some((key) => loading[key]))

  async function fetchSection(section: DashboardAnalyticsSectionKey) {
    loading[section] = true
    errors[section] = null

    try {
      if (section === 'overview') {
        overview.value = await DashboardService.getAnalyticsOverview(params.value)
      } else if (section === 'institutions') {
        institutions.value = await DashboardService.getAnalyticsInstitutions(params.value)
      } else if (section === 'lenders') {
        lenders.value = await DashboardService.getAnalyticsLenders(params.value)
      } else if (section === 'absorption') {
        absorption.value = await DashboardService.getAnalyticsAbsorption(params.value)
      } else if (section === 'yearly') {
        yearly.value = await DashboardService.getAnalyticsYearly(params.value)
      } else if (section === 'lenderProportion') {
        lenderProportion.value = await DashboardService.getAnalyticsLenderProportion(params.value)
      } else {
        risks.value = await DashboardService.getAnalyticsRisks(params.value)
      }
    } catch (error) {
      errors[section] = errorMessage(error)
    } finally {
      loading[section] = false
    }
  }

  async function refreshAll() {
    await Promise.all(sectionKeys.map((section) => fetchSection(section)))
  }

  async function loadFilterOptions() {
    loadingMasterData.value = true
    try {
      await Promise.all([
        masterStore.fetchLenders(false, { limit: 1000, sort: 'name', order: 'asc' }),
        masterStore.fetchInstitutions(false, { limit: 1000, sort: 'name', order: 'asc' }),
        masterStore.fetchAllRegionLevels(false),
        masterStore.fetchProgramTitles(false, { limit: 1000, sort: 'title', order: 'asc' }),
      ])
    } finally {
      loadingMasterData.value = false
    }
  }

  async function initialize() {
    await Promise.all([loadFilterOptions(), refreshAll()])
  }

  async function applyFilters() {
    assignFilters(appliedFilters, draftFilters)
    await refreshAll()
  }

  async function resetFilters() {
    assignFilters(draftFilters, defaultFilters)
    assignFilters(appliedFilters, defaultFilters)
    await refreshAll()
  }

  function updateDraftFilters(filters: DashboardAnalyticsFilterState) {
    assignFilters(draftFilters, filters)
  }

  return {
    draftFilters,
    appliedFilters,
    overview,
    institutions,
    lenders,
    absorption,
    yearly,
    lenderProportion,
    risks,
    loading,
    errors,
    loadingMasterData,
    anyLoading,
    ignoredDrilldownQueryKeys: drilldown.ignoredQueryKeys,
    initialize,
    fetchSection,
    refreshAll,
    applyFilters,
    resetFilters,
    updateDraftFilters,
    openDrilldown: drilldown.openDrilldown,
  }
}
