import { ref } from 'vue'
import { useRouter, type LocationQueryRaw, type RouteLocationRaw } from 'vue-router'
import type { DashboardDrilldownQuery } from '@/types/dashboard.types'

type DrilldownTarget = 'projects' | 'monitoring' | 'loan_agreements' | 'spatial_distribution'

const targetRoutes: Record<DrilldownTarget, string> = {
  projects: 'project-master',
  monitoring: 'monitoring-overview',
  loan_agreements: 'loan-agreements',
  spatial_distribution: 'spatial-distribution',
}

const supportedQueryKeys: Record<DrilldownTarget, Set<string>> = {
  projects: new Set([
    'loan_types',
    'indication_lender_ids',
    'executing_agency_ids',
    'fixed_lender_ids',
    'project_statuses',
    'pipeline_statuses',
    'program_title_ids',
    'region_ids',
    'foreign_loan_min',
    'foreign_loan_max',
    'include_history',
    'search',
  ]),
  monitoring: new Set(['search', 'is_effective']),
  loan_agreements: new Set(['search', 'lender_id', 'is_extended', 'closing_date_before']),
  spatial_distribution: new Set(['region_ids', 'provinceCode', 'sectorId', 'mode']),
}

function isDrilldownTarget(value: string): value is DrilldownTarget {
  return (
    value === 'projects' ||
    value === 'monitoring' ||
    value === 'loan_agreements' ||
    value === 'spatial_distribution'
  )
}

function normalizeTarget(value: string): DrilldownTarget {
  return isDrilldownTarget(value) ? value : 'projects'
}

function firstValue(query: Record<string, string[]>, key: string) {
  return query[key]?.[0]
}

function addQueryValue(query: LocationQueryRaw, key: string, values?: string[]) {
  const clean = (values ?? []).filter(Boolean)
  if (clean.length === 0) return
  query[key] = clean.length === 1 ? clean[0] : clean
}

function mappedQuery(target: DrilldownTarget, sourceQuery: Record<string, string[]>) {
  const supported = supportedQueryKeys[target]
  const query: LocationQueryRaw = {}
  const ignored: string[] = []

  Object.entries(sourceQuery).forEach(([key, values]) => {
    if (supported.has(key)) {
      addQueryValue(query, key, values)
      return
    }

    if (target === 'projects' && key === 'institution_ids') {
      addQueryValue(query, 'executing_agency_ids', values)
      return
    }

    if (target === 'projects' && key === 'lender_ids') {
      addQueryValue(query, 'fixed_lender_ids', values)
      return
    }

    if (target === 'loan_agreements' && (key === 'lender_ids' || key === 'fixed_lender_ids')) {
      addQueryValue(query, 'lender_id', values.slice(0, 1))
      return
    }

    ignored.push(key)
  })

  return { query, ignored }
}

export function useAnalyticsDrilldown() {
  const router = useRouter()
  const ignoredQueryKeys = ref<string[]>([])

  function toRoute(drilldown: DashboardDrilldownQuery): RouteLocationRaw {
    const target = normalizeTarget(drilldown.target)
    const loanAgreementID = firstValue(drilldown.query, 'loan_agreement_id')

    if (target === 'monitoring' && loanAgreementID) {
      const { query, ignored } = mappedQuery(target, drilldown.query)
      ignoredQueryKeys.value = ignored.filter((key) => key !== 'loan_agreement_id')
      return { name: 'monitoring-list', params: { laId: loanAgreementID }, query }
    }

    const { query, ignored } = mappedQuery(target, drilldown.query)
    ignoredQueryKeys.value = ignored

    return {
      name: targetRoutes[target],
      query,
    }
  }

  async function openDrilldown(drilldown: DashboardDrilldownQuery) {
    await router.push(toRoute(drilldown))
  }

  return {
    ignoredQueryKeys,
    openDrilldown,
    toRoute,
  }
}
