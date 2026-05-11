<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import MultiSelect from 'primevue/multiselect'
import PlanningFunnelFlow from '@/components/dashboard/PlanningFunnelFlow.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SpatialChoroplethMap from '@/components/spatial/SpatialChoroplethMap.vue'
import { DashboardService } from '@/services/dashboard.service'
import { LoanAgreementService } from '@/services/loan-agreement.service'
import { ProjectService } from '@/services/project.service'
import { SpatialDistributionService } from '@/services/spatial-distribution.service'
import { useDashboardStore } from '@/stores/dashboard.store'
import { useMasterStore } from '@/stores/master.store'
import type {
  DashboardDatum,
  DashboardInsightTarget,
  DashboardStage,
  DashboardStageKey,
} from '@/types/dashboard-flow.types'
import type { DashboardDistributionItem } from '@/types/dashboard.types'
import type { ProjectMasterListParams, ProjectMasterRow } from '@/types/project.types'
import type {
  SpatialDistributionChoroplethResponse,
  SpatialDistributionLevel,
  SpatialDistributionMetric,
  SpatialDistributionRegionMetric,
} from '@/types/spatial-distribution.types'

const dashboardStore = useDashboardStore()
const masterStore = useMasterStore()
const {
  blueBookDistribution,
  greenBookDistribution,
  daftarKegiatanDistribution,
  loanAgreementDistribution,
} = storeToRefs(dashboardStore)
const { periods, regions: masterRegions } = storeToRefs(masterStore)
const selectedPeriodIds = ref<string[]>([])

const countFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 0,
})

const compactUsdFormatter = new Intl.NumberFormat('en-US', {
  maximumFractionDigits: 1,
  notation: 'compact',
})

const percentFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 1,
  minimumFractionDigits: 1,
})

const agencyColors = ['#0b6f73', '#1fb5b2', '#1fa06f', '#fdb813', '#64748b', '#7c3aed']

type DashboardStageSummary = {
  count: number
  totalLoanUsd: number
}

type DashboardStageRegionPanels = Record<DashboardStageKey, DashboardDatum[] | null>

type DashboardMapStageOption = {
  key: DashboardStageKey
  label: string
  color: string
}

const dashboardStageKeys: DashboardStageKey[] = ['BB', 'GB', 'DK', 'LA']

const dashboardMapStageOptions: DashboardMapStageOption[] = [
  { key: 'BB', label: 'Blue Book', color: '#2563eb' },
  { key: 'GB', label: 'Green Book', color: '#16a34a' },
  { key: 'DK', label: 'Daftar Kegiatan', color: '#d97706' },
  { key: 'LA', label: 'Loan Agreement', color: '#7c3aed' },
]

function emptyDashboardChoropleth(): SpatialDistributionChoroplethResponse {
  return {
    level: 'province',
    regions: [],
    summary: {
      total_regions: 0,
      active_regions: 0,
      total_project_count: 0,
      total_loan_usd: 0,
      max_project_count: 0,
      max_loan_usd: 0,
    },
  }
}

function emptyDashboardStageMaps(): Record<
  DashboardStageKey,
  SpatialDistributionChoroplethResponse
> {
  return {
    BB: emptyDashboardChoropleth(),
    GB: emptyDashboardChoropleth(),
    DK: emptyDashboardChoropleth(),
    LA: emptyDashboardChoropleth(),
  }
}

const defaultStageSummaries: Record<DashboardStageKey, DashboardStageSummary> = {
  BB: { count: 96, totalLoanUsd: 0 },
  GB: { count: 36, totalLoanUsd: 0 },
  DK: { count: 35, totalLoanUsd: 0 },
  LA: { count: 0, totalLoanUsd: 0 },
}

const stageSummaries = ref<Record<DashboardStageKey, DashboardStageSummary>>({
  ...defaultStageSummaries,
})

const stageRegionData = ref<DashboardStageRegionPanels>({
  BB: null,
  GB: null,
  DK: null,
  LA: null,
})
const blueBookPipelineCards = ref<DashboardDatum[] | null>(null)
const blueBookLoiCards = ref<DashboardDatum[] | null>(null)
const blueBookIndicationCards = ref<DashboardDatum[] | null>(null)
const loanAgreementStatusCards = ref<DashboardDatum[] | null>(null)
const dashboardMapLevel = ref<SpatialDistributionLevel>('province')
const dashboardMapProvinceCode = ref<string | undefined>()
const dashboardMapProvinceName = ref<string | undefined>()
const dashboardMapMetric = ref<SpatialDistributionMetric>('count')
const dashboardMap = ref<SpatialDistributionChoroplethResponse>(emptyDashboardChoropleth())
const dashboardStageMaps =
  ref<Record<DashboardStageKey, SpatialDistributionChoroplethResponse>>(emptyDashboardStageMaps())
const dashboardMapLoading = ref(false)
const dashboardMapError = ref<string | null>(null)
const selectedDashboardMapRegion = ref<SpatialDistributionRegionMetric | null>(null)
const dashboardMapRegionFilter = ref<SpatialDistributionRegionMetric | null>(null)
const dashboardMapProvinceFilter = ref<SpatialDistributionRegionMetric | null>(null)
const dashboardPeriodsLoading = ref(true)
const dashboardPeriodsLoaded = ref(false)

const periodOptions = computed(() => periods.value)
const availablePeriodIds = computed(() => periodOptions.value.map((period) => period.id))
const periodFilterReady = computed(
  () => dashboardPeriodsLoaded.value && periodOptions.value.length > 0,
)
const periodFilterKey = computed(() => periodOptions.value.map((period) => period.id).join('|'))
const periodFilterFallbackLabel = computed(() =>
  dashboardPeriodsLoading.value ? 'Memuat periode' : 'Belum ada periode',
)

const selectedPeriodLabel = computed(() => {
  const selectedIds = selectedPeriodIds.value.filter((periodId) =>
    availablePeriodIds.value.includes(periodId),
  )

  if (selectedIds.length === 0 || selectedIds.length === availablePeriodIds.value.length) {
    return 'Semua periode'
  }

  if (selectedIds.length === 1) {
    return (
      periodOptions.value.find((period) => period.id === selectedIds[0])?.name ?? 'Semua periode'
    )
  }

  return `${countFormatter.format(selectedIds.length)} periode`
})

const dashboardMapFocusRegion = computed(
  () =>
    selectedDashboardMapRegion.value?.region_name ??
    dashboardMapProvinceName.value ??
    'Seluruh Indonesia',
)

const dashboardMapScopeLabel = computed(() =>
  dashboardMapLevel.value === 'city'
    ? `Kabupaten/Kota ${dashboardMapProvinceName.value?.replace(/^Provinsi\s+/i, '') ?? ''}`.trim()
    : 'Indonesia',
)

const dashboardMapResetLabel = computed(() =>
  dashboardMapLevel.value === 'city' ? 'Kembali ke Indonesia' : 'Fokus Indonesia',
)

const dashboardMapResetIcon = computed(() =>
  dashboardMapLevel.value === 'city' ? 'pi pi-arrow-left' : 'pi pi-globe',
)

const showDashboardMapReset = computed(
  () => dashboardMapLevel.value === 'city' || Boolean(selectedDashboardMapRegion.value),
)

const canDrillDownDashboardMap = computed(
  () =>
    dashboardMapLevel.value === 'province' &&
    selectedDashboardMapRegion.value?.region_type === 'PROVINCE',
)

const dashboardMapDetailRoute = computed(() => ({
  name: 'spatial-distribution',
  query: dashboardMapRouteQuery(selectedDashboardMapRegion.value),
}))

const dashboardMapStageBreakdown = computed(() => {
  const selectedRegionCode = selectedDashboardMapRegion.value?.region_code

  return dashboardMapStageOptions.map((stage) => {
    const stageMap = dashboardStageMaps.value[stage.key]
    const region = selectedRegionCode
      ? stageMap.regions.find((item) => item.region_code === selectedRegionCode)
      : null
    const projectCount =
      region?.project_count ?? (selectedRegionCode ? 0 : stageMap.summary.total_project_count)

    return {
      ...stage,
      projectCount,
      totalLoanUsd:
        region?.total_loan_usd ?? (selectedRegionCode ? 0 : stageMap.summary.total_loan_usd),
    }
  })
})

let stageSummaryRequestId = 0
let loanAgreementStatusRequestId = 0
let dashboardMapRequestId = 0

function selectedPeriodFilterIds() {
  const selectedIds = selectedPeriodIds.value.filter((periodId) =>
    availablePeriodIds.value.includes(periodId),
  )

  if (selectedIds.length === 0 || selectedIds.length === availablePeriodIds.value.length) {
    return []
  }

  return selectedIds
}

function selectedPeriodParams(): Pick<ProjectMasterListParams, 'period_ids'> {
  const periodIds = selectedPeriodFilterIds()
  return periodIds.length ? { period_ids: periodIds } : {}
}

function selectedDashboardRegionFilterIds() {
  const region = dashboardMapRegionFilter.value
  if (!region) return []

  const expandedIds = new Set<string>([region.region_id])

  if (region.region_type === 'PROVINCE') {
    const province = masterRegions.value.find((item) => item.id === region.region_id)
    const parentCountry = province?.parent_code
      ? masterRegions.value.find((item) => item.code === province.parent_code)
      : null

    if (parentCountry) {
      expandedIds.add(parentCountry.id)
    }

    masterRegions.value
      .filter((item) => item.parent_code === region.region_code)
      .forEach((item) => expandedIds.add(item.id))
  }

  return [...expandedIds]
}

function selectedDashboardRegionParams(): Pick<ProjectMasterListParams, 'region_ids'> {
  const regionIds = selectedDashboardRegionFilterIds()
  return regionIds.length ? { region_ids: regionIds } : {}
}

function selectedDashboardProjectParams(): Pick<
  ProjectMasterListParams,
  'period_ids' | 'region_ids'
> {
  return {
    ...selectedPeriodParams(),
    ...selectedDashboardRegionParams(),
  }
}

function selectedDashboardStageParams() {
  const params = selectedDashboardProjectParams()
  return Object.keys(params).length ? params : undefined
}

const hasActivePeriodFilter = computed(() => selectedPeriodFilterIds().length > 0)

function dashboardMapRouteQuery(region?: SpatialDistributionRegionMetric | null) {
  const periodIds = selectedPeriodFilterIds()
  const query: Record<string, string | string[]> = {
    level: dashboardMapLevel.value,
    metric: dashboardMapMetric.value,
  }

  if (dashboardMapLevel.value === 'city' && dashboardMapProvinceCode.value) {
    query.province_code = dashboardMapProvinceCode.value
  }

  if (region) {
    query.region_code = region.region_code
  }

  if (periodIds.length) {
    query.period_ids = periodIds
  }

  return query
}

function selectDashboardMapRegion(region: SpatialDistributionRegionMetric) {
  const isSameRegion = selectedDashboardMapRegion.value?.region_code === region.region_code
  const nextRegion = isSameRegion ? null : region

  selectedDashboardMapRegion.value = nextRegion
  dashboardMapRegionFilter.value =
    nextRegion ?? (dashboardMapLevel.value === 'city' ? dashboardMapProvinceFilter.value : null)

  void fetchStageSummaries()
}

async function drillDownDashboardMapProvince() {
  const region = selectedDashboardMapRegion.value
  if (!region || region.region_type !== 'PROVINCE') return

  dashboardMapLevel.value = 'city'
  dashboardMapProvinceCode.value = region.region_code
  dashboardMapProvinceName.value = region.region_name
  dashboardMapProvinceFilter.value = region
  dashboardMapRegionFilter.value = region
  selectedDashboardMapRegion.value = null
  await Promise.allSettled([fetchDashboardMap(), fetchStageSummaries()])
}

async function resetDashboardMapFocus() {
  dashboardMapLevel.value = 'province'
  dashboardMapProvinceCode.value = undefined
  dashboardMapProvinceName.value = undefined
  dashboardMapProvinceFilter.value = null
  dashboardMapRegionFilter.value = null
  selectedDashboardMapRegion.value = null
  await Promise.allSettled([fetchDashboardMap(), fetchStageSummaries()])
}

function formatDashboardMapUsd(value: number) {
  if (value <= 0) return 'USD 0'
  return `USD ${compactUsdFormatter.format(value)}`
}

function withSelectedDashboardQuery(query: DashboardInsightTarget['query']) {
  const filterParams = selectedDashboardProjectParams()
  if (!Object.keys(filterParams).length) return query

  return {
    ...query,
    ...filterParams,
  }
}

function projectTarget(
  query: DashboardInsightTarget['query'],
  label: string,
): DashboardInsightTarget {
  return {
    name: 'project-master',
    query: withSelectedDashboardQuery(query),
    label,
    exact: true,
  }
}

function usdAmountLabel(value: number) {
  if (value <= 0) return 'USD 0'
  return `USD ${compactUsdFormatter.format(value)}`
}

function stageSummary(stageKey: DashboardStageKey) {
  return stageSummaries.value[stageKey]
}

function stagePipelineShare(stageKey: DashboardStageKey) {
  const blueBookCount = stageSummary('BB').count
  if (blueBookCount <= 0) return 0
  if (stageKey === 'BB') return 100
  return (stageSummary(stageKey).count / blueBookCount) * 100
}

function stageConversionLabel(currentKey: DashboardStageKey, nextKey: DashboardStageKey) {
  const currentCount = stageSummary(currentKey).count
  if (currentCount <= 0) return '0,0%'
  return `${percentFormatter.format((stageSummary(nextKey).count / currentCount) * 100)}%`
}

function stageGap(currentKey: DashboardStageKey, nextKey: DashboardStageKey) {
  return Math.max(stageSummary(currentKey).count - stageSummary(nextKey).count, 0)
}

function buildBlueBookPipelineStatusData(counts: {
  advancedToGreenBook: number
  loiOnly: number
  indicationOnly: number
  withoutIndication: number
}): DashboardDatum[] {
  return [
    {
      label: 'Sudah lanjut ke Green Book',
      value: counts.advancedToGreenBook,
      valueLabel: `${countFormatter.format(counts.advancedToGreenBook)} proyek`,
      description: 'Pipeline sudah melewati tahap indikasi lender dan masuk prioritas Green Book.',
      color: '#1fa06f',
      tone: 'good',
      target: projectTarget(
        { reached_stages: 'GB' },
        'Buka Project Master untuk proyek yang sudah lanjut ke Green Book',
      ),
    },
    {
      label: 'Sudah LoI, belum GB',
      value: counts.loiOnly,
      valueLabel: `${countFormatter.format(counts.loiOnly)} proyek`,
      description:
        'Ada minat tertulis dari lender, tetapi belum dipaketkan menjadi proyek Green Book.',
      color: '#fdb813',
      tone: 'warning',
      target: projectTarget(
        { reached_stages: 'BB', missing_stages: 'GB', has_loi: true },
        'Buka Project Master untuk proyek yang sudah LoI tetapi belum Green Book',
      ),
    },
    {
      label: 'Indikasi saja, tanpa LoI',
      value: counts.indicationOnly,
      valueLabel: `${countFormatter.format(counts.indicationOnly)} proyek`,
      description: 'Masih perlu dorongan konfirmasi lender sebelum bisa diprioritaskan lebih jauh.',
      color: '#f59e0b',
      tone: 'warning',
      target: projectTarget(
        {
          reached_stages: 'BB',
          missing_stages: 'GB',
          has_lender_indication: true,
          has_loi: false,
        },
        'Buka Project Master untuk proyek dengan indikasi lender tanpa LoI',
      ),
    },
    {
      label: 'Belum ada indikasi',
      value: counts.withoutIndication,
      valueLabel: `${countFormatter.format(counts.withoutIndication)} proyek`,
      description: 'Belum punya lender lead dan menjadi risiko utama pada konversi Blue Book.',
      color: '#dc2626',
      tone: 'danger',
      target: projectTarget(
        { reached_stages: 'BB', missing_stages: 'GB', has_lender_indication: false },
        'Buka Project Master untuk proyek yang belum memiliki indikasi lender',
      ),
    },
  ]
}

function buildBlueBookLenderMixData(
  kind: 'loi' | 'indication',
  counts: { bilateral: number; multilateral: number; ksa: number; other?: number },
): DashboardDatum[] {
  const label =
    kind === 'loi'
      ? {
          hintSuffix: 'LoI',
          bilateral: 'Buka Project Master untuk proyek LoI Bilateral',
          multilateral: 'Buka Project Master untuk proyek LoI Multilateral',
          ksa: 'Buka Project Master untuk proyek LoI KSA',
        }
      : {
          hintSuffix: 'indikasi',
          bilateral: 'Buka Project Master untuk indikasi lender Bilateral',
          multilateral: 'Buka Project Master untuk indikasi lender Multilateral',
          ksa: 'Buka Project Master untuk indikasi lender KSA',
        }

  const filters = kind === 'loi' ? { has_loi: true } : { has_lender_indication: true }
  const items: DashboardDatum[] = [
    {
      label: 'Bilateral',
      value: counts.bilateral,
      valueLabel: `${countFormatter.format(counts.bilateral)} ${label.hintSuffix}`,
      color: '#0b6f73',
      target: projectTarget(
        { reached_stages: 'BB', loan_types: 'Bilateral', ...filters },
        label.bilateral,
      ),
    },
    {
      label: 'Multilateral',
      value: counts.multilateral,
      valueLabel: `${countFormatter.format(counts.multilateral)} ${label.hintSuffix}`,
      color: '#1fb5b2',
      target: projectTarget(
        { reached_stages: 'BB', loan_types: 'Multilateral', ...filters },
        label.multilateral,
      ),
    },
    {
      label: 'KSA',
      value: counts.ksa,
      valueLabel: `${countFormatter.format(counts.ksa)} ${label.hintSuffix}`,
      color: '#fdb813',
      target: projectTarget({ reached_stages: 'BB', loan_types: 'KSA', ...filters }, label.ksa),
    },
  ]

  if ((counts.other ?? 0) > 0) {
    items.push({
      label: 'Lainnya',
      value: counts.other ?? 0,
      valueLabel: `${countFormatter.format(counts.other ?? 0)} ${label.hintSuffix}`,
      color: '#94a3b8',
    })
  }

  return items
}

function collectLoanTypeCounts(rows: ProjectMasterRow[]) {
  const counts = {
    bilateral: 0,
    multilateral: 0,
    ksa: 0,
    other: 0,
  }

  rows.forEach((row) => {
    const types = new Set(row.loan_types)
    let matched = false

    if (types.has('Bilateral')) {
      counts.bilateral += 1
      matched = true
    }

    if (types.has('Multilateral')) {
      counts.multilateral += 1
      matched = true
    }

    if (types.has('KSA')) {
      counts.ksa += 1
      matched = true
    }

    if (!matched) {
      counts.other += 1
    }
  })

  return counts
}

function buildLoanAgreementStatusData(counts: {
  onSchedule: { count: number; amountUsd: number }
  behindSchedule: { count: number; amountUsd: number }
  atRisk: { count: number; amountUsd: number }
}): DashboardDatum[] {
  return [
    {
      label: 'On Schedule',
      value: counts.onSchedule.count,
      valueLabel: countFormatter.format(counts.onSchedule.count),
      amountLabel: usdAmountLabel(counts.onSchedule.amountUsd),
      color: '#10a36b',
    },
    {
      label: 'Behind',
      value: counts.behindSchedule.count,
      valueLabel: countFormatter.format(counts.behindSchedule.count),
      amountLabel: usdAmountLabel(counts.behindSchedule.amountUsd),
      color: '#d97706',
    },
    {
      label: 'At Risk',
      value: counts.atRisk.count,
      valueLabel: countFormatter.format(counts.atRisk.count),
      amountLabel: usdAmountLabel(counts.atRisk.amountUsd),
      color: '#dc2626',
    },
  ]
}

function fallbackStageLocationData(stageKey: DashboardStageKey): DashboardDatum[] {
  switch (stageKey) {
    case 'BB':
      return [
        { label: 'Jawa', value: 28 },
        { label: 'Sumatera', value: 17 },
        { label: 'Sulawesi', value: 12 },
        { label: 'Kalimantan', value: 9 },
        { label: 'Papua', value: 6 },
        { label: 'Nasional', value: 24 },
      ]
    case 'GB':
      return [
        { label: 'Jawa', value: 10 },
        { label: 'Sumatera', value: 7 },
        { label: 'Sulawesi', value: 5 },
        { label: 'Kalimantan', value: 4 },
        { label: 'Papua', value: 3 },
        { label: 'Nasional', value: 7 },
      ]
    case 'DK':
    case 'LA':
      return [
        { label: 'Jawa', value: 11 },
        { label: 'Sumatera', value: 8 },
        { label: 'Sulawesi', value: 6 },
        { label: 'Kalimantan', value: 5 },
        { label: 'Papua', value: 3 },
        { label: 'Bali, NT, Maluku', value: 2 },
      ]
  }
}

const fallbackBlueBookAgencyGroups: DashboardDatum[] = [
  { label: 'Kementerian/Lembaga', value: 85, valueLabel: '85 proyek', color: '#0b6f73' },
  { label: 'BUMN', value: 10, valueLabel: '10 proyek', color: '#1fb5b2' },
  { label: 'Pemerintah Daerah', value: 2, valueLabel: '2 proyek', color: '#fdb813' },
]

const fallbackBlueBookTopLevelAgencies: DashboardDatum[] = [
  { label: 'Kemen PU', value: 27, valueLabel: '27 proyek', color: '#0b6f73' },
  { label: 'Kemenristekdikti', value: 15, valueLabel: '15 proyek', color: '#1fb5b2' },
  { label: 'Pertamina', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'Kemenhan', value: 5, valueLabel: '5 proyek', color: '#fdb813' },
  { label: 'Kemenkes', value: 4, valueLabel: '4 proyek', color: '#64748b' },
  { label: 'PLN', value: 4, valueLabel: '4 proyek', color: '#7c3aed' },
]

const fallbackBlueBookPrograms: DashboardDatum[] = [
  { label: 'Konektivitas & Transportasi', value: 38, valueLabel: '38 proyek', color: '#0b6f73' },
  { label: 'Sumber Daya Air & Pangan', value: 19, valueLabel: '19 proyek', color: '#1fb5b2' },
  { label: 'Ketahanan Energi', value: 16, valueLabel: '16 proyek', color: '#1fa06f' },
  { label: 'Transformasi Digital', value: 12, valueLabel: '12 proyek', color: '#fdb813' },
  { label: 'Kesehatan & Pendidikan', value: 11, valueLabel: '11 proyek', color: '#64748b' },
]

const fallbackGreenBookLenderTypes: DashboardDatum[] = [
  {
    label: 'Bilateral',
    value: 16,
    valueLabel: '16 proyek',
    color: '#0b6f73',
    target: projectTarget(
      { reached_stages: 'GB', loan_types: 'Bilateral' },
      'Buka Project Master untuk proyek Green Book Bilateral',
    ),
  },
  {
    label: 'Multilateral',
    value: 14,
    valueLabel: '14 proyek',
    color: '#1fb5b2',
    target: projectTarget(
      { reached_stages: 'GB', loan_types: 'Multilateral' },
      'Buka Project Master untuk proyek Green Book Multilateral',
    ),
  },
  {
    label: 'KSA',
    value: 6,
    valueLabel: '6 proyek',
    color: '#fdb813',
    target: projectTarget(
      { reached_stages: 'GB', loan_types: 'KSA' },
      'Buka Project Master untuk proyek Green Book KSA',
    ),
  },
]

const fallbackGreenBookTopLenders: DashboardDatum[] = [
  { label: 'JICA', value: 10, valueLabel: '10 proyek', color: '#0b6f73' },
  { label: 'World Bank', value: 7, valueLabel: '7 proyek', color: '#1fb5b2' },
  { label: 'ADB', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'KfW', value: 4, valueLabel: '4 proyek', color: '#fdb813' },
  { label: 'AIIB', value: 3, valueLabel: '3 proyek', color: '#64748b' },
]

const fallbackGreenBookTopLevelAgencies: DashboardDatum[] = [
  { label: 'PUPR', value: 11, valueLabel: '11 proyek', color: '#0b6f73' },
  { label: 'Kemenhub', value: 8, valueLabel: '8 proyek', color: '#1fb5b2' },
  { label: 'ESDM', value: 5, valueLabel: '5 proyek', color: '#1fa06f' },
  { label: 'Kementan', value: 6, valueLabel: '6 proyek', color: '#fdb813' },
  { label: 'Kominfo', value: 4, valueLabel: '4 proyek', color: '#64748b' },
]

const fallbackDaftarKegiatanLenderTypes: DashboardDatum[] = [
  {
    label: 'Bilateral',
    value: 19,
    valueLabel: '19 proyek',
    color: '#0b6f73',
    target: projectTarget(
      { reached_stages: 'DK', loan_types: 'Bilateral' },
      'Buka Project Master untuk proyek Daftar Kegiatan Bilateral',
    ),
  },
  {
    label: 'Multilateral',
    value: 11,
    valueLabel: '11 proyek',
    color: '#1fb5b2',
    target: projectTarget(
      { reached_stages: 'DK', loan_types: 'Multilateral' },
      'Buka Project Master untuk proyek Daftar Kegiatan Multilateral',
    ),
  },
  {
    label: 'KSA',
    value: 5,
    valueLabel: '5 proyek',
    color: '#fdb813',
    target: projectTarget(
      { reached_stages: 'DK', loan_types: 'KSA' },
      'Buka Project Master untuk proyek Daftar Kegiatan KSA',
    ),
  },
]

const fallbackDaftarKegiatanTopLenders: DashboardDatum[] = [
  { label: 'JICA', value: 9, valueLabel: '9 proyek', color: '#0b6f73' },
  { label: 'World Bank', value: 8, valueLabel: '8 proyek', color: '#1fb5b2' },
  { label: 'ADB', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'KfW', value: 4, valueLabel: '4 proyek', color: '#fdb813' },
  { label: 'AIIB', value: 3, valueLabel: '3 proyek', color: '#64748b' },
]

const fallbackDaftarKegiatanTopLevelAgencies: DashboardDatum[] = [
  { label: 'Kemenhub', value: 9, valueLabel: '9 proyek', color: '#0b6f73' },
  { label: 'PUPR', value: 8, valueLabel: '8 proyek', color: '#1fb5b2' },
  { label: 'ESDM', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'Kominfo', value: 4, valueLabel: '4 proyek', color: '#fdb813' },
  { label: 'Kemenkes', value: 3, valueLabel: '3 proyek', color: '#64748b' },
]

const fallbackDaftarKegiatanPrograms: DashboardDatum[] = [
  { label: 'Konektivitas & Transportasi', value: 10, valueLabel: '10 proyek', color: '#0b6f73' },
  { label: 'Ketahanan Energi', value: 7, valueLabel: '7 proyek', color: '#1fb5b2' },
  { label: 'Sumber Daya Air & Pangan', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'Transformasi Digital', value: 4, valueLabel: '4 proyek', color: '#fdb813' },
]

const fallbackLoanAgreementLenderTypes: DashboardDatum[] = [
  { label: 'Bilateral', value: 19, valueLabel: '19 proyek', color: '#0b6f73' },
  { label: 'Multilateral', value: 11, valueLabel: '11 proyek', color: '#1fb5b2' },
  { label: 'KSA', value: 5, valueLabel: '5 proyek', color: '#fdb813' },
]

const fallbackLoanAgreementTopLenders: DashboardDatum[] = [
  { label: 'JICA', value: 9, valueLabel: '9 proyek', color: '#0b6f73' },
  { label: 'World Bank', value: 8, valueLabel: '8 proyek', color: '#1fb5b2' },
  { label: 'ADB', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'KfW', value: 4, valueLabel: '4 proyek', color: '#fdb813' },
  { label: 'AIIB', value: 3, valueLabel: '3 proyek', color: '#64748b' },
]

const fallbackLoanAgreementTopLevelAgencies: DashboardDatum[] = [
  { label: 'Kemenhub', value: 9, valueLabel: '9 proyek', color: '#0b6f73' },
  { label: 'PUPR', value: 8, valueLabel: '8 proyek', color: '#1fb5b2' },
  { label: 'ESDM', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'Kominfo', value: 4, valueLabel: '4 proyek', color: '#fdb813' },
  { label: 'Kemenkes', value: 3, valueLabel: '3 proyek', color: '#64748b' },
]

const fallbackLoanAgreementPrograms: DashboardDatum[] = [
  { label: 'Konektivitas & Transportasi', value: 10, valueLabel: '10 proyek', color: '#0b6f73' },
  { label: 'Ketahanan Energi', value: 7, valueLabel: '7 proyek', color: '#1fb5b2' },
  { label: 'Sumber Daya Air & Pangan', value: 6, valueLabel: '6 proyek', color: '#1fa06f' },
  { label: 'Transformasi Digital', value: 4, valueLabel: '4 proyek', color: '#fdb813' },
]

function distributionValueLabel(item: DashboardDistributionItem) {
  const projectLabel = `${countFormatter.format(item.project_count)} proyek`
  if (item.foreign_loan_usd <= 0) return projectLabel
  return `${projectLabel} - USD ${compactUsdFormatter.format(item.foreign_loan_usd)}`
}

function distributionDatum(
  item: DashboardDistributionItem,
  index: number,
  linkToProjectMaster = false,
): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target:
      linkToProjectMaster && item.id
        ? projectTarget(
            { reached_stages: 'BB', executing_agency_ids: item.id },
            `Buka Project Master untuk proyek Blue Book dengan EA utama ${item.label}`,
          )
        : undefined,
  }
}

function regionGroupDatum(item: DashboardDistributionItem, index: number): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
  }
}

function programDatum(item: DashboardDistributionItem, index: number): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: item.id
      ? projectTarget(
          { reached_stages: 'BB', program_title_ids: item.id },
          `Buka Project Master untuk program ${item.label}`,
        )
      : undefined,
  }
}

function greenBookLenderTypeDatum(item: DashboardDistributionItem, index: number): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: projectTarget(
      { reached_stages: 'GB', loan_types: item.label },
      `Buka Project Master untuk proyek Green Book ${item.label}`,
    ),
  }
}

function greenBookLenderDatum(item: DashboardDistributionItem, index: number): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: item.id
      ? projectTarget(
          { reached_stages: 'GB', fixed_lender_ids: item.id },
          `Buka Project Master untuk proyek Green Book dengan lender ${item.label}`,
        )
      : undefined,
  }
}

function greenBookAgencyDatum(item: DashboardDistributionItem, index: number): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: item.id
      ? projectTarget(
          { reached_stages: 'GB', executing_agency_ids: item.id },
          `Buka Project Master untuk proyek Green Book dengan EA utama ${item.label}`,
        )
      : undefined,
  }
}

function stageLenderTypeDatum(
  stage: 'GB' | 'DK' | 'LA',
  stageLabel: string,
  item: DashboardDistributionItem,
  index: number,
): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: projectTarget(
      { reached_stages: stage, loan_types: item.label },
      `Buka Project Master untuk proyek ${stageLabel} ${item.label}`,
    ),
  }
}

function stageLenderDatum(
  stage: 'GB' | 'DK' | 'LA',
  stageLabel: string,
  item: DashboardDistributionItem,
  index: number,
): DashboardDatum {
  const lenderFilterKey =
    stage === 'GB'
      ? 'fixed_lender_ids'
      : stage === 'DK'
        ? 'dk_lender_ids'
        : 'loan_agreement_lender_ids'

  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: item.id
      ? projectTarget(
          { reached_stages: stage, [lenderFilterKey]: item.id },
          `Buka Project Master untuk proyek ${stageLabel} dengan lender ${item.label}`,
        )
      : undefined,
  }
}

function stageAgencyDatum(
  stage: 'GB' | 'DK' | 'LA',
  stageLabel: string,
  item: DashboardDistributionItem,
  index: number,
): DashboardDatum {
  const agencyFilterKey = stage === 'GB' ? 'executing_agency_ids' : 'dk_executing_agency_ids'

  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: item.id
      ? projectTarget(
          { reached_stages: stage, [agencyFilterKey]: item.id },
          `Buka Project Master untuk proyek ${stageLabel} dengan EA utama ${item.label}`,
        )
      : undefined,
  }
}

function stageProgramDatum(
  stage: 'DK' | 'LA',
  stageLabel: string,
  item: DashboardDistributionItem,
  index: number,
): DashboardDatum {
  return {
    label: item.label,
    value: item.project_count,
    valueLabel: distributionValueLabel(item),
    color: agencyColors[index % agencyColors.length],
    target: item.id
      ? projectTarget(
          { reached_stages: stage, program_title_ids: item.id },
          `Buka Project Master untuk program ${item.label} pada ${stageLabel}`,
        )
      : undefined,
  }
}

function totalDistributionItems(items: DashboardDatum[]) {
  return items.reduce((sum, item) => sum + item.value, 0)
}

function withDashboardFallback(items: DashboardDatum[], fallback: DashboardDatum[]) {
  if (items.length > 0 || hasActivePeriodFilter.value) {
    return items
  }

  return fallback
}

const blueBookAgencyGroups = computed(() => {
  const items = blueBookDistribution.value.agency_groups
  return withDashboardFallback(
    items.map((item, index) => distributionDatum(item, index)),
    fallbackBlueBookAgencyGroups,
  )
})

const blueBookTopLevelAgencies = computed(() => {
  const items = blueBookDistribution.value.top_agencies
  return withDashboardFallback(
    items.map((item, index) => distributionDatum(item, index, true)),
    fallbackBlueBookTopLevelAgencies,
  )
})

const blueBookPrograms = computed(() => {
  const items = blueBookDistribution.value.programs
  return withDashboardFallback(
    items.map((item, index) => programDatum(item, index)),
    fallbackBlueBookPrograms,
  )
})

const greenBookLenderTypes = computed(() => {
  const items = greenBookDistribution.value.lender_types
  return withDashboardFallback(
    items.map((item, index) => greenBookLenderTypeDatum(item, index)),
    fallbackGreenBookLenderTypes,
  )
})

const greenBookTopLenders = computed(() => {
  const items = greenBookDistribution.value.top_lenders
  return withDashboardFallback(
    items.map((item, index) => greenBookLenderDatum(item, index)),
    fallbackGreenBookTopLenders,
  )
})

const greenBookTopLevelAgencies = computed(() => {
  const items = greenBookDistribution.value.top_agencies
  return withDashboardFallback(
    items.map((item, index) => greenBookAgencyDatum(item, index)),
    fallbackGreenBookTopLevelAgencies,
  )
})

const daftarKegiatanLenderTypes = computed(() => {
  const items = daftarKegiatanDistribution.value.lender_types
  return withDashboardFallback(
    items.map((item, index) => stageLenderTypeDatum('DK', 'Daftar Kegiatan', item, index)),
    fallbackDaftarKegiatanLenderTypes,
  )
})

const daftarKegiatanTopLenders = computed(() => {
  const items = daftarKegiatanDistribution.value.top_lenders
  return withDashboardFallback(
    items.map((item, index) => stageLenderDatum('DK', 'Daftar Kegiatan', item, index)),
    fallbackDaftarKegiatanTopLenders,
  )
})

const daftarKegiatanTopLevelAgencies = computed(() => {
  const items = daftarKegiatanDistribution.value.top_agencies
  return withDashboardFallback(
    items.map((item, index) => stageAgencyDatum('DK', 'Daftar Kegiatan', item, index)),
    fallbackDaftarKegiatanTopLevelAgencies,
  )
})

const daftarKegiatanPrograms = computed(() => {
  const items = daftarKegiatanDistribution.value.programs ?? []
  return withDashboardFallback(
    items.map((item, index) => stageProgramDatum('DK', 'Daftar Kegiatan', item, index)),
    fallbackDaftarKegiatanPrograms,
  )
})

const loanAgreementLenderTypes = computed(() => {
  const items = loanAgreementDistribution.value.lender_types
  return withDashboardFallback(
    items.map((item, index) => stageLenderTypeDatum('LA', 'Loan Agreement', item, index)),
    fallbackLoanAgreementLenderTypes,
  )
})

const loanAgreementTopLenders = computed(() => {
  const items = loanAgreementDistribution.value.top_lenders
  return withDashboardFallback(
    items.map((item, index) => stageLenderDatum('LA', 'Loan Agreement', item, index)),
    fallbackLoanAgreementTopLenders,
  )
})

const loanAgreementTopLevelAgencies = computed(() => {
  const items = loanAgreementDistribution.value.top_agencies
  return withDashboardFallback(
    items.map((item, index) => stageAgencyDatum('LA', 'Loan Agreement', item, index)),
    fallbackLoanAgreementTopLevelAgencies,
  )
})

const loanAgreementPrograms = computed(() => {
  const items = loanAgreementDistribution.value.programs ?? []
  return withDashboardFallback(
    items.map((item, index) => stageProgramDatum('LA', 'Loan Agreement', item, index)),
    fallbackLoanAgreementPrograms,
  )
})

async function fetchStageSummaries() {
  const requestId = ++stageSummaryRequestId
  const stageParams = selectedDashboardStageParams()
  const [overviewResponse, blueBookResponse] = await Promise.allSettled([
    DashboardService.getStageOverview(stageParams),
    ProjectService.getProjectMaster({
      page: 1,
      limit: 1000,
      reached_stages: ['BB'],
      ...selectedDashboardProjectParams(),
    }),
  ])

  if (requestId !== stageSummaryRequestId) return

  const nextSummaries: Record<DashboardStageKey, DashboardStageSummary> =
    hasActivePeriodFilter.value
      ? {
          BB: { count: 0, totalLoanUsd: 0 },
          GB: { count: 0, totalLoanUsd: 0 },
          DK: { count: 0, totalLoanUsd: 0 },
          LA: { count: 0, totalLoanUsd: 0 },
        }
      : {
          ...defaultStageSummaries,
        }
  const nextStageRegions: DashboardStageRegionPanels = {
    BB: null,
    GB: null,
    DK: null,
    LA: null,
  }
  let nextBlueBookPipelineCards = blueBookPipelineCards.value
  let nextBlueBookLoiCards = blueBookLoiCards.value
  let nextBlueBookIndicationCards = blueBookIndicationCards.value

  if (overviewResponse.status === 'fulfilled') {
    overviewResponse.value.stages.forEach((stage) => {
      if (!dashboardStageKeys.includes(stage.stage)) return

      nextSummaries[stage.stage] = {
        count: stage.project_count,
        totalLoanUsd: stage.total_loan_usd,
      }
      nextStageRegions[stage.stage] = stage.regions.map((item, index) =>
        regionGroupDatum(item, index),
      )
    })
  }

  if (blueBookResponse.status === 'fulfilled') {
    const blueBookRows = blueBookResponse.value.data
    const remainingBeforeGreenBook = Math.max(nextSummaries.BB.count - nextSummaries.GB.count, 0)
    const candidateRows = blueBookRows.filter((row) => row.pipeline_status === 'BB')
    const loiOnlyRows = Math.min(
      candidateRows.filter((row) => row.has_loi).length,
      remainingBeforeGreenBook,
    )
    const indicationOnlyRows = Math.min(
      candidateRows.filter((row) => !row.has_loi && row.has_lender_indication).length,
      Math.max(remainingBeforeGreenBook - loiOnlyRows, 0),
    )
    const rowsWithoutIndication = Math.max(
      remainingBeforeGreenBook - loiOnlyRows - indicationOnlyRows,
      0,
    )

    nextBlueBookPipelineCards = buildBlueBookPipelineStatusData({
      advancedToGreenBook: nextSummaries.GB.count,
      loiOnly: loiOnlyRows,
      indicationOnly: indicationOnlyRows,
      withoutIndication: rowsWithoutIndication,
    })
    nextBlueBookLoiCards = buildBlueBookLenderMixData(
      'loi',
      collectLoanTypeCounts(blueBookRows.filter((row) => row.has_loi)),
    )
    nextBlueBookIndicationCards = buildBlueBookLenderMixData(
      'indication',
      collectLoanTypeCounts(blueBookRows.filter((row) => row.has_lender_indication)),
    )
  }

  stageSummaries.value = nextSummaries
  stageRegionData.value = nextStageRegions
  blueBookPipelineCards.value = nextBlueBookPipelineCards
  blueBookLoiCards.value = nextBlueBookLoiCards
  blueBookIndicationCards.value = nextBlueBookIndicationCards
}

async function refreshDashboardDistributions() {
  const periodIds = selectedPeriodFilterIds()

  await Promise.allSettled([
    dashboardStore.fetchBlueBookDistribution(periodIds),
    dashboardStore.fetchGreenBookDistribution(periodIds),
    dashboardStore.fetchDaftarKegiatanDistribution(periodIds),
    dashboardStore.fetchLoanAgreementDistribution(periodIds),
  ])
}

async function fetchDashboardMap() {
  const requestId = ++dashboardMapRequestId
  const periodIds = selectedPeriodFilterIds()
  const level = dashboardMapLevel.value
  const provinceCode = dashboardMapProvinceCode.value

  if (level === 'city' && !provinceCode) {
    dashboardMapLevel.value = 'province'
    dashboardMapProvinceName.value = undefined
    selectedDashboardMapRegion.value = null
    return fetchDashboardMap()
  }

  const baseParams = {
    level,
    province_code: level === 'city' ? provinceCode : undefined,
    period_ids: periodIds.length ? periodIds : undefined,
  }

  dashboardMapLoading.value = true
  dashboardMapError.value = null

  try {
    const [response, ...stageResponses] = await Promise.all([
      SpatialDistributionService.getChoropleth(baseParams),
      ...dashboardMapStageOptions.map((stage) =>
        SpatialDistributionService.getChoropleth({
          ...baseParams,
          reached_stages: [stage.key],
        }),
      ),
    ])

    if (requestId !== dashboardMapRequestId) return

    dashboardMap.value = response
    if (response.level === 'province') {
      dashboardMapProvinceCode.value = undefined
      dashboardMapProvinceName.value = undefined
    }
    dashboardStageMaps.value = dashboardMapStageOptions.reduce(
      (maps, stage, index) => ({
        ...maps,
        [stage.key]: stageResponses[index] ?? emptyDashboardChoropleth(),
      }),
      emptyDashboardStageMaps(),
    )

    const selectedRegionCode = selectedDashboardMapRegion.value?.region_code
    if (selectedRegionCode) {
      const nextSelectedRegion =
        response.regions.find((region) => region.region_code === selectedRegionCode) ?? null

      selectedDashboardMapRegion.value = nextSelectedRegion
      if (dashboardMapRegionFilter.value?.region_code === selectedRegionCode) {
        dashboardMapRegionFilter.value = nextSelectedRegion
      }
    }
  } catch {
    if (requestId !== dashboardMapRequestId) return
    dashboardMapError.value = 'Gagal memuat peta distribusi dashboard.'
  } finally {
    if (requestId === dashboardMapRequestId) {
      dashboardMapLoading.value = false
    }
  }
}

async function fetchLoanAgreementStatuses() {
  const requestId = ++loanAgreementStatusRequestId
  const periodIds = selectedPeriodFilterIds()

  const response = await LoanAgreementService.getLoanAgreements({
    page: 1,
    limit: 1000,
    period_ids: periodIds.length ? periodIds : undefined,
  })

  if (requestId !== loanAgreementStatusRequestId) return

  const counts = {
    onSchedule: { count: 0, amountUsd: 0 },
    behindSchedule: { count: 0, amountUsd: 0 },
    atRisk: { count: 0, amountUsd: 0 },
  }

  response.data.forEach((item) => {
    const status = item.performance_status?.toLowerCase()

    if (status === 'on schedule') {
      counts.onSchedule.count += 1
      counts.onSchedule.amountUsd += item.amount_usd
      return
    }

    if (status === 'behind schedule') {
      counts.behindSchedule.count += 1
      counts.behindSchedule.amountUsd += item.amount_usd
      return
    }

    if (status === 'at-risk') {
      counts.atRisk.count += 1
      counts.atRisk.amountUsd += item.amount_usd
    }
  })

  loanAgreementStatusCards.value = buildLoanAgreementStatusData(counts)
}

async function refreshDashboard() {
  await Promise.allSettled([
    refreshDashboardDistributions(),
    fetchDashboardMap(),
    fetchStageSummaries(),
    fetchLoanAgreementStatuses(),
  ])
}

async function loadDashboardPeriods() {
  dashboardPeriodsLoading.value = true

  try {
    await masterStore.fetchPeriods(true, { limit: 1000, sort: 'year_start', order: 'desc' })
  } finally {
    dashboardPeriodsLoaded.value = true
    dashboardPeriodsLoading.value = false
  }
}

onMounted(() => {
  void loadDashboardPeriods()
  void masterStore.fetchAllRegionLevels(false)
  void refreshDashboard()
})

watch(
  periodOptions,
  (options) => {
    const ids = options.map((period) => period.id)
    if (!ids.length) return

    const allowedIds = new Set(ids)
    const nextSelectedIds = selectedPeriodIds.value.filter((periodId) => allowedIds.has(periodId))

    if (nextSelectedIds.length !== selectedPeriodIds.value.length) {
      selectedPeriodIds.value = nextSelectedIds
    }
  },
  { immediate: true },
)

watch(selectedPeriodIds, () => {
  void refreshDashboard()
})

const stages = computed<DashboardStage[]>(() => {
  const blueBook = stageSummary('BB')
  const greenBook = stageSummary('GB')
  const daftarKegiatan = stageSummary('DK')
  const loanAgreement = stageSummary('LA')
  const blueBookToGreenBook = stageGap('BB', 'GB')
  const greenBookToDaftarKegiatan = stageGap('GB', 'DK')
  const daftarKegiatanToLoanAgreement = stageGap('DK', 'LA')
  const blueBookPipelineData =
    blueBookPipelineCards.value ??
    (hasActivePeriodFilter.value
      ? []
      : buildBlueBookPipelineStatusData({
          advancedToGreenBook: greenBook.count,
          loiOnly: 18,
          indicationOnly: 24,
          withoutIndication: 18,
        }))
  const blueBookLoiData =
    blueBookLoiCards.value ??
    (hasActivePeriodFilter.value
      ? []
      : buildBlueBookLenderMixData('loi', {
          bilateral: 16,
          multilateral: 13,
          ksa: 5,
          other: 2,
        }))
  const blueBookIndicationData =
    blueBookIndicationCards.value ??
    (hasActivePeriodFilter.value
      ? []
      : buildBlueBookLenderMixData('indication', {
          bilateral: 22,
          multilateral: 14,
          ksa: 6,
        }))
  const blueBookRegionItems =
    stageRegionData.value.BB ?? (hasActivePeriodFilter.value ? [] : fallbackStageLocationData('BB'))
  const greenBookRegionItems =
    stageRegionData.value.GB ?? (hasActivePeriodFilter.value ? [] : fallbackStageLocationData('GB'))
  const daftarKegiatanRegionItems =
    stageRegionData.value.DK ?? (hasActivePeriodFilter.value ? [] : fallbackStageLocationData('DK'))
  const loanAgreementRegionItems =
    stageRegionData.value.LA ?? (hasActivePeriodFilter.value ? [] : fallbackStageLocationData('LA'))
  const loanAgreementStatusData =
    loanAgreementStatusCards.value ??
    buildLoanAgreementStatusData({
      onSchedule: { count: 0, amountUsd: 0 },
      behindSchedule: { count: 0, amountUsd: 0 },
      atRisk: { count: 0, amountUsd: 0 },
    })

  return [
    {
      key: 'BB',
      stepLabel: 'Tahap 01',
      title: 'Blue Book',
      subtitle: 'Daftar usulan sebagai dasar penyusunan rencana pinjaman luar negeri.',
      count: blueBook.count,
      amountLabel: usdAmountLabel(blueBook.totalLoanUsd),
      pipelineShare: stagePipelineShare('BB'),
      color: '#2563eb',
      colorSoft: '#60a5fa',
      target: projectTarget(
        { reached_stages: 'BB' },
        'Buka Project Master untuk proyek yang sudah mencapai Blue Book',
      ),
      nextLabel: 'Green Book',
      conversionLabel: stageConversionLabel('BB', 'GB'),
      progressLabel: 'proyek sudah masuk Green Book',
      blockedLabel: `Masih ${countFormatter.format(blueBookToGreenBook)} proyek belum masuk Green Book`,
      details: {
        tabs: [
          { label: 'Status Pipeline' },
          { label: 'Distribusi K/L' },
          { label: 'Wilayah' },
          { label: 'Program' },
        ],
        counters: [],
        panels: [
          {
            tab: 'Status Pipeline',
            title: 'Status pipeline lender',
            hint: `${countFormatter.format(blueBook.count)} proyek BB`,
            kind: 'card',
            span: 'wide',
            description: `Dari ${countFormatter.format(blueBook.count)} proyek Blue Book, ${countFormatter.format(greenBook.count)} sudah masuk Green Book. Sisa pipeline masih bergantung pada tindak lanjut LoI, penguatan indikasi lender, dan pencarian lender lead.`,
            data: blueBookPipelineData,
          },
          {
            tab: 'Status Pipeline',
            title: 'Tipe lender yang sudah mengirim LoI',
            hint: `${countFormatter.format(totalDistributionItems(blueBookLoiData))} LoI masuk`,
            kind: 'donutbar',
            span: 'medium',
            data: blueBookLoiData,
          },
          {
            tab: 'Status Pipeline',
            title: 'Tipe lender terindikasi',
            hint: `${countFormatter.format(totalDistributionItems(blueBookIndicationData))} indikasi`,
            kind: 'donutbar',
            span: 'medium',
            data: blueBookIndicationData,
          },
          {
            tab: 'Distribusi K/L',
            title: 'Distribusi K/L berdasarkan EA tertinggi',
            hint: `${countFormatter.format(totalDistributionItems(blueBookAgencyGroups.value))} relasi EA`,
            kind: 'donutbar',
            span: 'medium',
            data: blueBookAgencyGroups.value,
          },
          {
            tab: 'Distribusi K/L',
            title: 'Top K/L berdasarkan EA tertinggi',
            hint: `Top ${blueBookTopLevelAgencies.value.length}`,
            kind: 'bars',
            span: 'medium',
            data: blueBookTopLevelAgencies.value,
          },
          {
            tab: 'Wilayah',
            title: 'Wilayah',
            hint: 'Jumlah proyek',
            kind: 'regions',
            span: 'wide',
            data: blueBookRegionItems,
          },
          {
            tab: 'Program',
            title: 'Program prioritas',
            hint: `${blueBookPrograms.value.length} program`,
            kind: 'bars',
            span: 'wide',
            data: blueBookPrograms.value,
          },
        ],
      },
    },
    {
      key: 'GB',
      stepLabel: 'Tahap 02',
      title: 'Green Book',
      subtitle: 'Proyek prioritas yang telah masuk rencana pendanaan dan penyiapan kegiatan.',
      count: greenBook.count,
      amountLabel: usdAmountLabel(greenBook.totalLoanUsd),
      pipelineShare: stagePipelineShare('GB'),
      color: '#16a34a',
      colorSoft: '#4ade80',
      target: projectTarget(
        { reached_stages: 'GB' },
        'Buka Project Master untuk proyek yang sudah mencapai Green Book',
      ),
      nextLabel: 'Daftar Kegiatan',
      conversionLabel: stageConversionLabel('GB', 'DK'),
      progressLabel: 'proyek sudah masuk tahap Daftar Kegiatan',
      blockedLabel: `Masih ${countFormatter.format(greenBookToDaftarKegiatan)} proyek belum masuk tahap Daftar Kegiatan`,
      details: {
        tabs: [{ label: 'Lender & Funding' }, { label: 'Distribusi K/L' }, { label: 'Wilayah' }],
        counters: [],

        panels: [
          {
            tab: 'Lender & Funding',
            title: 'Komposisi lender',
            hint: `${countFormatter.format(greenBook.count)} proyek`,
            kind: 'donutbar',
            span: 'medium',
            data: greenBookLenderTypes.value,
          },
          {
            tab: 'Lender & Funding',
            title: 'Top lender',
            hint: greenBookTopLenders.value.length
              ? `${greenBookTopLenders.value.length} lender`
              : 'Belum ada lender',
            kind: 'bars',
            span: 'medium',
            data: greenBookTopLenders.value,
          },

          {
            tab: 'Distribusi K/L',
            title: 'Distribusi K/L berdasarkan EA tertinggi',
            hint: `${countFormatter.format(totalDistributionItems(greenBookTopLevelAgencies.value))} relasi EA`,
            kind: 'bars',
            span: 'wide',
            data: greenBookTopLevelAgencies.value,
          },
          {
            tab: 'Wilayah',
            title: 'Wilayah Green Book',
            hint: 'Jumlah proyek',
            kind: 'regions',
            span: 'wide',
            data: greenBookRegionItems,
          },
        ],
      },
    },
    {
      key: 'DK',
      stepLabel: 'Tahap 03',
      title: 'Daftar Kegiatan',
      subtitle: 'Kegiatan yang telah ditetapkan dalam surat dan rincian pembiayaan.',
      count: daftarKegiatan.count,
      amountLabel: usdAmountLabel(daftarKegiatan.totalLoanUsd),
      pipelineShare: stagePipelineShare('DK'),
      color: '#d97706',
      colorSoft: '#f59e0b',
      target: projectTarget(
        { reached_stages: 'DK' },
        'Buka Project Master untuk proyek yang sudah mencapai Daftar Kegiatan',
      ),
      nextLabel: 'Loan Agreement',
      conversionLabel: stageConversionLabel('DK', 'LA'),
      progressLabel: 'proyek sudah Loan Agreement',
      blockedLabel: `Masih ${countFormatter.format(daftarKegiatanToLoanAgreement)} proyek belum Loan Agreement`,
      details: {
        tabs: [{ label: 'Lender & Funding' }, { label: 'Distribusi K/L' }, { label: 'Wilayah' }],
        counters: [],
        panels: [
          {
            tab: 'Lender & Funding',
            title: 'Komposisi lender',
            hint: `${countFormatter.format(totalDistributionItems(daftarKegiatanLenderTypes.value))} proyek DK`,
            kind: 'donutbar',
            span: 'medium',
            data: daftarKegiatanLenderTypes.value,
          },
          {
            tab: 'Lender & Funding',
            title: 'Top lender DK',
            hint: `Top ${daftarKegiatanTopLenders.value.length}`,
            kind: 'bars',
            span: 'medium',
            data: daftarKegiatanTopLenders.value,
          },
          {
            tab: 'Distribusi K/L',
            title: 'Distribusi K/L berdasarkan EA tertinggi',
            hint: `Top ${daftarKegiatanTopLevelAgencies.value.length}`,
            kind: 'bars',
            span: 'medium',
            data: daftarKegiatanTopLevelAgencies.value,
          },
          {
            tab: 'Distribusi K/L',
            title: 'Program prioritas',
            hint: `${daftarKegiatanPrograms.value.length} program`,
            kind: 'bars',
            span: 'medium',
            data: daftarKegiatanPrograms.value,
          },
          {
            tab: 'Wilayah',
            title: 'Sebaran wilayah DK',
            hint: 'Jumlah proyek',
            kind: 'regions',
            span: 'wide',
            data: daftarKegiatanRegionItems,
          },
        ],
      },
    },
    {
      key: 'LA',
      stepLabel: 'Tahap 04',
      title: 'Loan Agreement',
      subtitle: 'Perjanjian pinjaman yang telah memiliki dasar hukum pelaksanaan.',
      count: loanAgreement.count,
      amountLabel: usdAmountLabel(loanAgreement.totalLoanUsd),
      pipelineShare: stagePipelineShare('LA'),
      color: '#7c3aed',
      colorSoft: '#a78bfa',
      target: projectTarget(
        { reached_stages: 'LA' },
        'Buka Project Master untuk proyek yang sudah mencapai Loan Agreement',
      ),
      finalLabel: 'Tahap akhir perencanaan',
      details: {
        tabs: [
          { label: 'Kondisi Pinjaman' },
          { label: 'Lender & Funding' },
          { label: 'Distribusi K/L' },
          { label: 'Wilayah' },
        ],
        counters: [],
        panels: [
          {
            tab: 'Kondisi Pinjaman',
            title: 'Kondisi pinjaman',
            hint: 'LA',
            kind: 'status',
            span: 'wide',
            data: loanAgreementStatusData,
          },
          {
            tab: 'Lender & Funding',
            title: 'Komposisi lender',
            hint: `${countFormatter.format(totalDistributionItems(loanAgreementLenderTypes.value))} proyek LA`,
            kind: 'donutbar',
            span: 'medium',
            data: loanAgreementLenderTypes.value,
          },
          {
            tab: 'Lender & Funding',
            title: 'Top lender Loan Agreement',
            hint: `Top ${loanAgreementTopLenders.value.length}`,
            kind: 'bars',
            span: 'medium',
            data: loanAgreementTopLenders.value,
          },
          {
            tab: 'Distribusi K/L',
            title: 'Distribusi K/L berdasarkan EA tertinggi',
            hint: `Top ${loanAgreementTopLevelAgencies.value.length}`,
            kind: 'bars',
            span: 'medium',
            data: loanAgreementTopLevelAgencies.value,
          },
          {
            tab: 'Distribusi K/L',
            title: 'Program prioritas',
            hint: `${loanAgreementPrograms.value.length} program`,
            kind: 'bars',
            span: 'medium',
            data: loanAgreementPrograms.value,
          },
          {
            tab: 'Wilayah',
            title: 'Sebaran wilayah kandidat LA',
            hint: 'Jumlah proyek',
            kind: 'regions',
            span: 'wide',
            data: loanAgreementRegionItems,
          },
        ],
      },
    },
  ]
})
</script>

<template>
  <section class="dashboard-page space-y-6">
    <PageHeader
      title="Dashboard Pinjaman Luar Negeri"
      subtitle="Ikhtisar portofolio pinjaman luar negeri dari perencanaan hingga perjanjian pinjaman."
    >
      <template #actions>
        <div
          class="dashboard-page-actions flex w-full flex-wrap items-center justify-end gap-2 sm:w-auto"
        >
          <MultiSelect
            v-if="periodFilterReady"
            :key="periodFilterKey"
            v-model="selectedPeriodIds"
            :options="periodOptions"
            option-label="name"
            option-value="id"
            placeholder="Semua periode"
            filter
            filter-placeholder="Cari periode"
            empty-message="Belum ada periode"
            empty-filter-message="Periode tidak ditemukan"
            :max-selected-labels="1"
            scroll-height="16rem"
            append-to="body"
            :overlay-style="{ width: '18rem', minWidth: '18rem' }"
            class="min-w-56 flex-1 sm:w-72 sm:flex-none"
          >
            <template #value>
              <span class="truncate">{{ selectedPeriodLabel }}</span>
            </template>
            <template #option="{ option }">
              <span class="block min-w-0">
                <span class="block truncate font-medium">{{ option.name }}</span>
                <span class="mt-0.5 block text-xs text-surface-500">
                  {{ option.year_start }}-{{ option.year_end }}
                </span>
              </span>
            </template>
          </MultiSelect>
          <Button
            v-else
            :label="periodFilterFallbackLabel"
            :icon="dashboardPeriodsLoading ? 'pi pi-spin pi-spinner' : 'pi pi-calendar'"
            severity="secondary"
            outlined
            disabled
            class="min-w-56 flex-1 justify-center sm:w-72 sm:flex-none"
          />
        </div>
      </template>
    </PageHeader>

    <section class="overflow-hidden rounded-lg border border-surface-200 bg-white shadow-sm">
      <header
        class="flex flex-col gap-3 border-b border-surface-100 px-4 py-3 xl:flex-row xl:items-center xl:justify-between"
      >
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase text-primary-700">Sebaran wilayah</p>
          <h2 class="mt-1 text-base font-semibold text-surface-950">
            Peta Choropleth Distribusi Wilayah
          </h2>
          <p class="mt-1 max-w-3xl text-sm leading-5 text-surface-500">
            Distribusi proyek berdasarkan wilayah dan periode dashboard.
          </p>
        </div>
        <Button
          v-if="showDashboardMapReset"
          :label="dashboardMapResetLabel"
          :icon="dashboardMapResetIcon"
          severity="secondary"
          outlined
          size="small"
          class="shrink-0"
          @click="resetDashboardMapFocus"
        />
      </header>

      <div
        class="grid gap-3 bg-surface-50/60 p-3 xl:grid-cols-[minmax(0,1fr)_22rem] xl:items-stretch"
      >
        <SpatialChoroplethMap
          :level="dashboardMapLevel"
          :metric="dashboardMapMetric"
          :province-code="dashboardMapProvinceCode"
          :province-name="dashboardMapProvinceName"
          :regions="dashboardMap.regions"
          :selected-region-code="selectedDashboardMapRegion?.region_code"
          :loading="dashboardMapLoading"
          compact
          @select="selectDashboardMapRegion"
        />

        <aside
          class="flex flex-col rounded-lg border border-surface-200 bg-white p-3 shadow-sm xl:h-full xl:min-h-[24rem]"
        >
          <div class="min-w-0">
            <p class="text-xs font-semibold uppercase tracking-[0.16em] text-prism-teal-dark">
              Fokus peta
            </p>
            <h3 class="mt-0.5 truncate text-xl font-semibold text-surface-950">
              {{ dashboardMapFocusRegion }}
            </h3>
            <p class="mt-0.5 text-xs text-surface-500">
              {{ dashboardMapScopeLabel }} - {{ selectedPeriodLabel }}
            </p>
          </div>

          <p
            v-if="dashboardMapError"
            class="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700"
          >
            {{ dashboardMapError }}
          </p>

          <div class="mt-3 space-y-1.5">
            <div
              v-for="stage in dashboardMapStageBreakdown"
              :key="stage.key"
              class="rounded-lg border border-surface-100 bg-surface-50 px-3 py-2"
            >
              <div class="flex items-center justify-between gap-2">
                <span
                  class="inline-flex min-w-0 items-center gap-2 text-xs font-semibold text-surface-800"
                >
                  <span
                    class="h-2 w-2 shrink-0 rounded-sm"
                    :style="{ backgroundColor: stage.color }"
                    aria-hidden="true"
                  />
                  <span class="truncate">{{ stage.label }}</span>
                </span>
                <span class="inline-flex shrink-0 items-baseline gap-1.5">
                  <span class="text-base font-semibold text-surface-950">
                    {{ countFormatter.format(stage.projectCount) }}
                  </span>
                  <span class="text-[11px] font-medium text-surface-500">
                    {{ formatDashboardMapUsd(stage.totalLoanUsd) }}
                  </span>
                </span>
              </div>
            </div>
          </div>

          <div class="mt-3 grid gap-2 sm:grid-cols-2">
            <div class="rounded-lg border border-surface-100 bg-white px-3 py-2">
              <p class="text-xs font-semibold uppercase text-surface-500">Wilayah aktif</p>
              <p class="mt-0.5 text-sm font-semibold text-surface-950">
                {{ countFormatter.format(dashboardMap.summary.active_regions) }}
                <span class="text-xs font-medium text-surface-500">
                  / {{ countFormatter.format(dashboardMap.summary.total_regions) }}
                </span>
              </p>
            </div>

            <div class="rounded-lg border border-surface-100 bg-white px-2 py-2">
              <p class="px-1 text-xs font-semibold uppercase text-surface-500">Metrik peta</p>
              <div
                class="mt-1 grid grid-cols-2 rounded-lg border border-surface-200 bg-surface-50 p-1"
              >
                <button
                  type="button"
                  class="min-h-7 rounded-md text-xs font-semibold transition"
                  :class="
                    dashboardMapMetric === 'count'
                      ? 'bg-white text-surface-950 shadow-sm'
                      : 'text-surface-500 hover:text-surface-800'
                  "
                  @click="dashboardMapMetric = 'count'"
                >
                  Proyek
                </button>
                <button
                  type="button"
                  class="min-h-7 rounded-md text-xs font-semibold transition"
                  :class="
                    dashboardMapMetric === 'value'
                      ? 'bg-white text-surface-950 shadow-sm'
                      : 'text-surface-500 hover:text-surface-800'
                  "
                  @click="dashboardMapMetric = 'value'"
                >
                  Nilai
                </button>
              </div>
            </div>
          </div>

          <div class="mt-auto space-y-1.5 pt-2">
            <Button
              v-if="canDrillDownDashboardMap"
              label="Lihat Kab/Kota"
              icon="pi pi-map-marker"
              size="small"
              class="w-full"
              @click="drillDownDashboardMapProvince"
            />

            <RouterLink
              :to="dashboardMapDetailRoute"
              class="inline-flex min-h-8 w-full items-center justify-center gap-2 rounded-lg border border-surface-200 px-3 text-xs font-semibold text-surface-700 no-underline transition hover:border-primary-200 hover:bg-primary-50 hover:text-primary-700 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary"
            >
              Buka sebaran wilayah
              <i class="pi pi-arrow-up-right text-xs" aria-hidden="true" />
            </RouterLink>
          </div>
        </aside>
      </div>
    </section>

    <PlanningFunnelFlow :stages="stages" />
  </section>
</template>
