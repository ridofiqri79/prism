<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import MultiSelect from 'primevue/multiselect'
import PlanningFunnelFlow from '@/components/dashboard/PlanningFunnelFlow.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { DashboardService } from '@/services/dashboard.service'
import { LoanAgreementService } from '@/services/loan-agreement.service'
import { ProjectService } from '@/services/project.service'
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

const dashboardStore = useDashboardStore()
const masterStore = useMasterStore()
const {
  blueBookDistribution,
  greenBookDistribution,
  daftarKegiatanDistribution,
  loanAgreementDistribution,
} = storeToRefs(dashboardStore)
const { periods } = storeToRefs(masterStore)
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

const dashboardStageKeys: DashboardStageKey[] = ['BB', 'GB', 'DK', 'LA']

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

const periodOptions = computed(() => periods.value)
const availablePeriodIds = computed(() => periodOptions.value.map((period) => period.id))

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

let stageSummaryRequestId = 0
let loanAgreementStatusRequestId = 0

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

function withSelectedPeriodQuery(query: DashboardInsightTarget['query']) {
  const periodIds = selectedPeriodFilterIds()
  if (!periodIds.length) return query

  return {
    ...query,
    period_ids: periodIds,
  }
}

function projectTarget(
  query: DashboardInsightTarget['query'],
  label: string,
): DashboardInsightTarget {
  return {
    name: 'project-master',
    query: withSelectedPeriodQuery(query),
    label,
    exact: true,
  }
}

function spatialTarget(
  query: DashboardInsightTarget['query'],
  label: string,
): DashboardInsightTarget {
  return {
    name: 'spatial-distribution',
    query: withSelectedPeriodQuery(query),
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

const blueBookAgencyGroups = computed(() => {
  const items = blueBookDistribution.value.agency_groups
  if (!items.length) return fallbackBlueBookAgencyGroups
  return items.map((item, index) => distributionDatum(item, index))
})

const blueBookTopLevelAgencies = computed(() => {
  const items = blueBookDistribution.value.top_agencies
  if (!items.length) return fallbackBlueBookTopLevelAgencies
  return items.map((item, index) => distributionDatum(item, index, true))
})

const blueBookPrograms = computed(() => {
  const items = blueBookDistribution.value.programs
  if (!items.length) return fallbackBlueBookPrograms
  return items.map((item, index) => programDatum(item, index))
})

const greenBookLenderTypes = computed(() => {
  const items = greenBookDistribution.value.lender_types
  if (!items.length) return fallbackGreenBookLenderTypes
  return items.map((item, index) => greenBookLenderTypeDatum(item, index))
})

const greenBookTopLenders = computed(() => {
  const items = greenBookDistribution.value.top_lenders
  if (!items.length) return fallbackGreenBookTopLenders
  return items.map((item, index) => greenBookLenderDatum(item, index))
})

const greenBookTopLevelAgencies = computed(() => {
  const items = greenBookDistribution.value.top_agencies
  if (!items.length) return fallbackGreenBookTopLevelAgencies
  return items.map((item, index) => greenBookAgencyDatum(item, index))
})

const daftarKegiatanLenderTypes = computed(() => {
  const items = daftarKegiatanDistribution.value.lender_types
  if (!items.length) return fallbackDaftarKegiatanLenderTypes
  return items.map((item, index) => stageLenderTypeDatum('DK', 'Daftar Kegiatan', item, index))
})

const daftarKegiatanTopLenders = computed(() => {
  const items = daftarKegiatanDistribution.value.top_lenders
  if (!items.length) return fallbackDaftarKegiatanTopLenders
  return items.map((item, index) => stageLenderDatum('DK', 'Daftar Kegiatan', item, index))
})

const daftarKegiatanTopLevelAgencies = computed(() => {
  const items = daftarKegiatanDistribution.value.top_agencies
  if (!items.length) return fallbackDaftarKegiatanTopLevelAgencies
  return items.map((item, index) => stageAgencyDatum('DK', 'Daftar Kegiatan', item, index))
})

const daftarKegiatanPrograms = computed(() => {
  const items = daftarKegiatanDistribution.value.programs ?? []
  if (!items.length) return fallbackDaftarKegiatanPrograms
  return items.map((item, index) => stageProgramDatum('DK', 'Daftar Kegiatan', item, index))
})

const loanAgreementLenderTypes = computed(() => {
  const items = loanAgreementDistribution.value.lender_types
  if (!items.length) return fallbackLoanAgreementLenderTypes
  return items.map((item, index) => stageLenderTypeDatum('LA', 'Loan Agreement', item, index))
})

const loanAgreementTopLenders = computed(() => {
  const items = loanAgreementDistribution.value.top_lenders
  if (!items.length) return fallbackLoanAgreementTopLenders
  return items.map((item, index) => stageLenderDatum('LA', 'Loan Agreement', item, index))
})

const loanAgreementTopLevelAgencies = computed(() => {
  const items = loanAgreementDistribution.value.top_agencies
  if (!items.length) return fallbackLoanAgreementTopLevelAgencies
  return items.map((item, index) => stageAgencyDatum('LA', 'Loan Agreement', item, index))
})

const loanAgreementPrograms = computed(() => {
  const items = loanAgreementDistribution.value.programs ?? []
  if (!items.length) return fallbackLoanAgreementPrograms
  return items.map((item, index) => stageProgramDatum('LA', 'Loan Agreement', item, index))
})

async function fetchStageSummaries() {
  const requestId = ++stageSummaryRequestId
  const periodIds = selectedPeriodFilterIds()
  const [overviewResponse, blueBookResponse] = await Promise.allSettled([
    DashboardService.getStageOverview(periodIds.length ? { period_ids: periodIds } : undefined),
    ProjectService.getProjectMaster({
      page: 1,
      limit: 1000,
      reached_stages: ['BB'],
      ...selectedPeriodParams(),
    }),
  ])

  if (requestId !== stageSummaryRequestId) return

  const nextSummaries: Record<DashboardStageKey, DashboardStageSummary> = {
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
    const remainingBeforeGreenBook = Math.max(
      nextSummaries.BB.count - nextSummaries.GB.count,
      0,
    )
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
    fetchStageSummaries(),
    fetchLoanAgreementStatuses(),
  ])
}

onMounted(() => {
  void masterStore.fetchPeriods(false, { limit: 1000, sort: 'year_start', order: 'desc' })
  void refreshDashboard()
})

watch(
  periodOptions,
  (options) => {
    const ids = options.map((period) => period.id)
    if (!ids.length) return

    const allowedIds = new Set(ids)
    const nextSelectedIds = selectedPeriodIds.value.filter((periodId) => allowedIds.has(periodId))

    if (nextSelectedIds.length === 0 || nextSelectedIds.length === ids.length) {
      if (selectedPeriodIds.value.length !== ids.length || nextSelectedIds.length !== ids.length) {
        selectedPeriodIds.value = [...ids]
      }
      return
    }

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
    buildBlueBookPipelineStatusData({
      advancedToGreenBook: greenBook.count,
      loiOnly: 18,
      indicationOnly: 24,
      withoutIndication: 18,
    })
  const blueBookLoiData =
    blueBookLoiCards.value ??
    buildBlueBookLenderMixData('loi', {
      bilateral: 16,
      multilateral: 13,
      ksa: 5,
      other: 2,
    })
  const blueBookIndicationData =
    blueBookIndicationCards.value ??
    buildBlueBookLenderMixData('indication', {
      bilateral: 22,
      multilateral: 14,
      ksa: 6,
    })
  const blueBookRegionItems = stageRegionData.value.BB ?? fallbackStageLocationData('BB')
  const greenBookRegionItems = stageRegionData.value.GB ?? fallbackStageLocationData('GB')
  const daftarKegiatanRegionItems = stageRegionData.value.DK ?? fallbackStageLocationData('DK')
  const loanAgreementRegionItems = stageRegionData.value.LA ?? fallbackStageLocationData('LA')
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
            data: greenBookLenderTypes.value.length
              ? greenBookLenderTypes.value
              : [
                  {
                    label: 'Bilateral',
                    value: 16,
                    valueLabel: '16 · Rp 8,7 T',
                    color: '#0b6f73',
                    target: projectTarget(
                      { reached_stages: 'GB', loan_types: 'Bilateral' },
                      'Buka Project Master untuk proyek Green Book Bilateral',
                    ),
                  },
                  {
                    label: 'Multilateral',
                    value: 14,
                    valueLabel: '14 · Rp 7,6 T',
                    color: '#1fb5b2',
                    target: projectTarget(
                      { reached_stages: 'GB', loan_types: 'Multilateral' },
                      'Buka Project Master untuk proyek Green Book Multilateral',
                    ),
                  },
                  {
                    label: 'KSA',
                    value: 6,
                    valueLabel: '6 · Rp 1,1 T',
                    color: '#fdb813',
                    target: projectTarget(
                      { reached_stages: 'GB', loan_types: 'KSA' },
                      'Buka Project Master untuk proyek Green Book KSA',
                    ),
                  },
                ],
          },
          {
            tab: 'Lender & Funding',
            title: 'Top lender',
            hint: `${greenBookTopLenders.value.length || 5} lender`,
            kind: 'bars',
            span: 'medium',
            data: greenBookTopLenders.value.length
              ? greenBookTopLenders.value
              : [
                  { label: 'JICA', value: 10, valueLabel: '10 · Rp 4,1 T', color: '#0b6f73' },
                  { label: 'World Bank', value: 7, valueLabel: '7 · Rp 3,0 T', color: '#1fb5b2' },
                  { label: 'ADB', value: 6, valueLabel: '6 · Rp 2,4 T', color: '#1fa06f' },
                  { label: 'KfW', value: 4, valueLabel: '4 · Rp 1,5 T', color: '#fdb813' },
                  { label: 'AIIB', value: 3, valueLabel: '3 · Rp 1,1 T', color: '#64748b' },
                ],
          },

          {
            tab: 'Distribusi K/L',
            title: 'Distribusi K/L berdasarkan EA tertinggi',
            hint: `${countFormatter.format(totalDistributionItems(greenBookTopLevelAgencies.value))} relasi EA`,
            kind: 'bars',
            span: 'wide',
            data: greenBookTopLevelAgencies.value.length
              ? greenBookTopLevelAgencies.value
              : [
                  { label: 'PUPR', value: 11, valueLabel: '11 · Rp 5,1 T', color: '#0b6f73' },
                  { label: 'Kemenhub', value: 8, valueLabel: '8 · Rp 3,4 T', color: '#1fb5b2' },
                  { label: 'ESDM', value: 5, valueLabel: '5 · Rp 2,0 T', color: '#1fa06f' },
                  {
                    label: 'SDA dan pangan',
                    value: 6,
                    valueLabel: '6 · Rp 2,8 T',
                    color: '#fdb813',
                  },
                  { label: 'Digital', value: 4, valueLabel: '4 · Rp 1,2 T', color: '#64748b' },
                ],
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
            v-model="selectedPeriodIds"
            :options="periodOptions"
            option-label="name"
            option-value="id"
            placeholder="Semua periode"
            filter
            filter-placeholder="Cari periode"
            :max-selected-labels="1"
            append-to="body"
            class="min-w-56 flex-1 sm:w-72 sm:flex-none"
          >
            <template #value>
              <span class="truncate">{{ selectedPeriodLabel }}</span>
            </template>
          </MultiSelect>
        </div>
      </template>
    </PageHeader>

    <PlanningFunnelFlow :stages="stages" />
  </section>
</template>
