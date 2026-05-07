<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import PlanningFunnelFlow from '@/components/dashboard/PlanningFunnelFlow.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type { DashboardDatum, DashboardInsightTarget, DashboardStage } from '@/types/dashboard-flow.types'
import type { DashboardDistributionItem } from '@/types/dashboard.types'

const dashboardStore = useDashboardStore()
const {
  blueBookDistribution,
  greenBookDistribution,
  daftarKegiatanDistribution,
  loanAgreementDistribution,
} = storeToRefs(dashboardStore)

const countFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 0,
})

const compactUsdFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 1,
  notation: 'compact',
})

const agencyColors = ['#0b6f73', '#1fb5b2', '#1fa06f', '#fdb813', '#64748b', '#7c3aed']

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

onMounted(() => {
  void dashboardStore.fetchBlueBookDistribution().catch(() => undefined)
  void dashboardStore.fetchGreenBookDistribution().catch(() => undefined)
  void dashboardStore.fetchDaftarKegiatanDistribution().catch(() => undefined)
  void dashboardStore.fetchLoanAgreementDistribution().catch(() => undefined)
})

function projectTarget(query: DashboardInsightTarget['query'], label: string): DashboardInsightTarget {
  return {
    name: 'project-master',
    query,
    label,
    exact: true,
  }
}

function spatialTarget(query: DashboardInsightTarget['query'], label: string): DashboardInsightTarget {
  return {
    name: 'spatial-distribution',
    query,
    label,
    exact: true,
  }
}

const stages = computed<DashboardStage[]>(() => [
  {
    key: 'BB',
    stepLabel: 'Tahap 01',
    title: 'Blue Book',
    subtitle: 'Daftar usulan sebagai dasar penyusunan rencana pinjaman luar negeri.',
    count: 96,
    amountLabel: 'Rp 29,28 T',
    pipelineShare: 100,
    color: '#0b6f73',
    colorSoft: '#0f8f8c',
    target: projectTarget({ reached_stages: 'BB' }, 'Buka Project Master untuk proyek yang sudah mencapai Blue Book'),
    nextLabel: 'Green Book',
    conversionLabel: '37,5%',
    blockedLabel: '60 proyek menunggu tindak lanjut',
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
          hint: '96 proyek BB',
          kind: 'card',
          span: 'wide',
          description:
            'Dari 96 proyek Blue Book, 36 sudah masuk Green Book. Sisa pipeline masih bergantung pada tindak lanjut LoI, penguatan indikasi lender, dan pencarian lender lead.',
          data: [
            {
              label: 'Sudah lanjut ke Green Book',
              value: 36,
              valueLabel: '36 proyek',
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
              value: 18,
              valueLabel: '18 proyek',
              description: 'Ada minat tertulis dari lender, tetapi belum dipaketkan menjadi proyek Green Book.',
              color: '#fdb813',
              tone: 'warning',
              target: projectTarget(
                { reached_stages: 'BB', missing_stages: 'GB', has_loi: true },
                'Buka Project Master untuk proyek yang sudah LoI tetapi belum Green Book',
              ),
            },
            {
              label: 'Indikasi saja, tanpa LoI',
              value: 24,
              valueLabel: '24 proyek',
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
              value: 18,
              valueLabel: '18 proyek',
              description: 'Belum punya lender lead dan menjadi risiko utama pada konversi Blue Book.',
              color: '#dc2626',
              tone: 'danger',
              target: projectTarget(
                { reached_stages: 'BB', missing_stages: 'GB', has_lender_indication: false },
                'Buka Project Master untuk proyek yang belum memiliki indikasi lender',
              ),
            },
          ],
        },
        {
          tab: 'Status Pipeline',
          title: 'Tipe lender yang sudah mengirim LoI',
          hint: '36 LoI masuk',
          kind: 'donutbar',
          span: 'medium',
          data: [
            {
              label: 'Bilateral',
              value: 16,
              valueLabel: '16 LoI',
              color: '#0b6f73',
              target: projectTarget(
                { reached_stages: 'BB', has_loi: true, loan_types: 'Bilateral' },
                'Buka Project Master untuk proyek LoI Bilateral',
              ),
            },
            {
              label: 'Multilateral',
              value: 13,
              valueLabel: '13 LoI',
              color: '#1fb5b2',
              target: projectTarget(
                { reached_stages: 'BB', has_loi: true, loan_types: 'Multilateral' },
                'Buka Project Master untuk proyek LoI Multilateral',
              ),
            },
            {
              label: 'KSA',
              value: 5,
              valueLabel: '5 LoI',
              color: '#fdb813',
              target: projectTarget(
                { reached_stages: 'BB', has_loi: true, loan_types: 'KSA' },
                'Buka Project Master untuk proyek LoI KSA',
              ),
            },
            { label: 'Lainnya', value: 2, valueLabel: '2 LoI', color: '#94a3b8' },
          ],
        },
        {
          tab: 'Status Pipeline',
          title: 'Tipe lender terindikasi',
          hint: '42 indikasi',
          kind: 'donutbar',
          span: 'medium',
          data: [
            {
              label: 'Bilateral',
              value: 22,
              valueLabel: '22 indikasi',
              color: '#0b6f73',
              target: projectTarget(
                { reached_stages: 'BB', has_lender_indication: true, loan_types: 'Bilateral' },
                'Buka Project Master untuk indikasi lender Bilateral',
              ),
            },
            {
              label: 'Multilateral',
              value: 14,
              valueLabel: '14 indikasi',
              color: '#1fb5b2',
              target: projectTarget(
                { reached_stages: 'BB', has_lender_indication: true, loan_types: 'Multilateral' },
                'Buka Project Master untuk indikasi lender Multilateral',
              ),
            },
            {
              label: 'KSA',
              value: 6,
              valueLabel: '6 indikasi',
              color: '#fdb813',
              target: projectTarget(
                { reached_stages: 'BB', has_lender_indication: true, loan_types: 'KSA' },
                'Buka Project Master untuk indikasi lender KSA',
              ),
            },
          ],
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
          data: [
            { label: 'Jawa', value: 28 },
            { label: 'Sumatera', value: 17 },
            { label: 'Sulawesi', value: 12 },
            { label: 'Kalimantan', value: 9 },
            { label: 'Papua', value: 6 },
            {
              label: 'Nasional',
              value: 24,
              target: spatialTarget(
                { level: 'province', region_code: 'ID', reached_stages: 'BB', metric: 'count' },
                'Buka Sebaran Wilayah untuk proyek nasional Blue Book',
              ),
            },
          ],
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
    count: 36,
    amountLabel: 'Rp 20,00 T',
    pipelineShare: 37.5,
    color: '#1fb5b2',
    colorSoft: '#26b7a5',
    target: projectTarget({ reached_stages: 'GB' }, 'Buka Project Master untuk proyek yang sudah mencapai Green Book'),
    nextLabel: 'Daftar Kegiatan',
    conversionLabel: '97,2%',
    blockedLabel: '1 proyek menunggu penetapan',
    details: {
      tabs: [
        { label: 'Lender & Funding' },
        { label: 'Distribusi K/L' },
        { label: 'Wilayah' },
      ],
      counters: [],


      panels: [

        {
          tab: 'Lender & Funding',
          title: 'Komposisi lender',
          hint: '36 proyek',
          kind: 'donutbar',
          span: 'medium',
          data: greenBookLenderTypes.value.length ? greenBookLenderTypes.value : [
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
          data: greenBookTopLenders.value.length ? greenBookTopLenders.value : [
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
          data: greenBookTopLevelAgencies.value.length ? greenBookTopLevelAgencies.value : [
            { label: 'PUPR', value: 11, valueLabel: '11 · Rp 5,1 T', color: '#0b6f73' },
            { label: 'Kemenhub', value: 8, valueLabel: '8 · Rp 3,4 T', color: '#1fb5b2' },
            { label: 'ESDM', value: 5, valueLabel: '5 · Rp 2,0 T', color: '#1fa06f' },
            { label: 'SDA dan pangan', value: 6, valueLabel: '6 · Rp 2,8 T', color: '#fdb813' },
            { label: 'Digital', value: 4, valueLabel: '4 · Rp 1,2 T', color: '#64748b' },
          ],
        },
        {
          tab: 'Wilayah',
          title: 'Wilayah Green Book',
          hint: 'Jumlah proyek',
          kind: 'regions',
          span: 'wide',
          data: [
            { label: 'Jawa', value: 10 },
            { label: 'Sumatera', value: 7 },
            { label: 'Sulawesi', value: 5 },
            { label: 'Kalimantan', value: 4 },
            { label: 'Papua', value: 3 },
            { label: 'Nasional', value: 7 },
          ],
        },
      ],
    },
  },
  {
    key: 'DK',
    stepLabel: 'Tahap 03',
    title: 'Daftar Kegiatan',
    subtitle: 'Kegiatan yang telah ditetapkan dalam surat dan rincian pembiayaan.',
    count: 35,
    amountLabel: 'Rp 13,90 T',
    pipelineShare: 36.4,
    color: '#1fa06f',
    colorSoft: '#3ccb6b',
    target: projectTarget({ reached_stages: 'DK' }, 'Buka Project Master untuk proyek yang sudah mencapai Daftar Kegiatan'),
    nextLabel: 'Loan Agreement',
    conversionLabel: '0,0%',
    blockedLabel: '35 proyek menunggu perjanjian',
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
          data: [
            { label: 'Jawa', value: 11 },
            { label: 'Sumatera', value: 8 },
            { label: 'Sulawesi', value: 6 },
            { label: 'Kalimantan', value: 5 },
            { label: 'Papua', value: 3 },
            { label: 'Bali, NT, Maluku', value: 2 },
          ],
        },
      ],
    },
  },
  {
    key: 'LA',
    stepLabel: 'Tahap 04',
    title: 'Loan Agreement',
    subtitle: 'Perjanjian pinjaman yang telah memiliki dasar hukum pelaksanaan.',
    count: 0,
    amountLabel: 'Rp 0',
    pipelineShare: 0,
    color: '#f6a800',
    colorSoft: '#fdb813',
    target: projectTarget({ reached_stages: 'LA' }, 'Buka Project Master untuk proyek yang sudah mencapai Loan Agreement'),
    nextLabel: 'Monitoring',
    conversionLabel: '0,0%',
    blockedLabel: 'Belum ada perjanjian tercatat',
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
          data: [
            { label: 'On Schedule', value: 0, valueLabel: '0', amountLabel: 'Rp 0', color: '#10a36b' },
            { label: 'Behind', value: 0, valueLabel: '0', amountLabel: 'Rp 0', color: '#d97706' },
            { label: 'At Risk', value: 0, valueLabel: '0', amountLabel: 'Rp 0', color: '#dc2626' },
          ],
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
          data: [
            { label: 'Jawa', value: 11 },
            { label: 'Sumatera', value: 8 },
            { label: 'Sulawesi', value: 6 },
            { label: 'Kalimantan', value: 5 },
            { label: 'Papua', value: 3 },
            { label: 'Bali, NT, Maluku', value: 2 },
          ],
        },
      ],
    },
  },
])
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Dashboard Pinjaman Luar Negeri"
      subtitle="Ikhtisar portofolio pinjaman luar negeri dari perencanaan hingga perjanjian pinjaman."
    >
      <template #actions>
        <Tag severity="success" rounded>
          <span class="inline-flex items-center gap-2">
            <span class="h-1.5 w-1.5 rounded-full bg-prism-green-dark" />
            Data contoh
          </span>
        </Tag>
        <Button
          label="Filter"
          icon="pi pi-filter"
          severity="secondary"
          outlined
          size="small"
          disabled
        />
        <Button
          label="Ekspor"
          icon="pi pi-download"
          severity="secondary"
          outlined
          size="small"
          disabled
        />
      </template>
    </PageHeader>

    <PlanningFunnelFlow :stages="stages" />
  </section>
</template>
