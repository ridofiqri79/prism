<script setup lang="ts">
import { computed } from 'vue'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import type {
  DKProjectJourney,
  GBProjectJourney,
  JourneyFundingGroup,
  JourneyMatrixStage,
  JourneyMatrixRow,
  JourneyResponse,
  JourneySnapshotStep,
  JourneyStageState,
  JourneySummaryMetric,
  LAJourney,
} from '@/types/journey.types'

const props = defineProps<{
  journey: JourneyResponse
}>()

const numberFormatter = new Intl.NumberFormat('id-ID')

const gbProjects = computed(() => props.journey.gb_projects)
const dkProjects = computed(() => gbProjects.value.flatMap((project) => project.dk_projects))
const loanAgreements = computed(() =>
  dkProjects.value.flatMap((project) => loanAgreementsForProject(project)),
)
const monitoringRows = computed(() =>
  loanAgreements.value.flatMap((loanAgreement) => loanAgreement.monitoring),
)

const fundingGroups = computed<JourneyFundingGroup[]>(() => {
  const groups = new Map<string, JourneyFundingGroup>()

  for (const source of gbProjects.value.flatMap((project) => project.funding_sources)) {
    const label = source.lender?.short_name || source.lender?.name || 'Lender belum terisi'
    const currency = source.currency || 'USD'
    const key = `${source.lender?.id || label}-${currency}`
    const current =
      groups.get(key) ??
      ({
        key,
        label,
        currency,
        loan_usd: 0,
        grant_usd: 0,
        local_usd: 0,
        total_usd: 0,
      } satisfies JourneyFundingGroup)

    current.loan_usd += source.loan_usd ?? 0
    current.grant_usd += source.grant_usd ?? 0
    current.local_usd += source.local_usd ?? 0
    current.total_usd = current.loan_usd + current.grant_usd + current.local_usd
    groups.set(key, current)
  }

  return Array.from(groups.values()).sort((a, b) => b.total_usd - a.total_usd)
})

const totalFundingUsd = computed(() =>
  fundingGroups.value.reduce((sum, group) => sum + group.total_usd, 0),
)
const maxFundingUsd = computed(() =>
  fundingGroups.value.reduce((max, group) => Math.max(max, group.total_usd), 0),
)

const latestGreenBookWarnings = computed(
  () => gbProjects.value.filter((project) => project.has_newer_revision).length,
)

const missingItems = computed(() => {
  const items: string[] = []

  if (gbProjects.value.length === 0) {
    items.push('Green Book belum tersedia untuk proyek Blue Book ini.')
  }

  if (gbProjects.value.length > 0 && dkProjects.value.length === 0) {
    items.push('Daftar Kegiatan belum tersedia pada seluruh proyek Green Book.')
  }

  const dkWithoutLoanAgreement = dkProjects.value.filter(
    (project) => loanAgreementsForProject(project).length === 0,
  ).length
  if (dkWithoutLoanAgreement > 0) {
    items.push(
      `${numberFormatter.format(dkWithoutLoanAgreement)} Daftar Kegiatan belum memiliki Loan Agreement.`,
    )
  }

  const loanAgreementWithoutMonitoring = loanAgreements.value.filter(
    (loanAgreement) => loanAgreement.monitoring.length === 0,
  ).length
  if (loanAgreementWithoutMonitoring > 0) {
    items.push(
      `${numberFormatter.format(
        loanAgreementWithoutMonitoring,
      )} Loan Agreement belum memiliki Monitoring Disbursement.`,
    )
  }

  if (props.journey.bb_project.has_newer_revision) {
    items.push('Blue Book yang sedang dilihat bukan revisi terbaru.')
  }

  if (latestGreenBookWarnings.value > 0) {
    items.push(
      `${numberFormatter.format(
        latestGreenBookWarnings.value,
      )} proyek Green Book memiliki revisi lebih baru.`,
    )
  }

  return items
})

const summaryMetrics = computed<JourneySummaryMetric[]>(() => [
  {
    label: 'Blue Book',
    value: '1',
    hint: props.journey.bb_project.blue_book_revision_label || props.journey.bb_project.bb_code,
    icon: 'pi pi-book',
    state: props.journey.bb_project.has_newer_revision ? 'warning' : 'completed',
  },
  {
    label: 'Green Book',
    value: numberFormatter.format(gbProjects.value.length),
    hint:
      gbProjects.value.length > 0
        ? `${numberFormatter.format(gbProjects.value.length)} proyek terkait`
        : 'Belum ada proyek terkait',
    icon: 'pi pi-folder',
    state: gbProjects.value.length > 0 ? 'completed' : 'pending',
  },
  {
    label: 'Daftar Kegiatan',
    value: numberFormatter.format(dkProjects.value.length),
    hint:
      dkProjects.value.length > 0
        ? 'Snapshot downstream sudah terbentuk'
        : 'Belum ada snapshot downstream',
    icon: 'pi pi-list',
    state: dkProjects.value.length > 0 ? 'completed' : 'pending',
  },
  {
    label: 'Loan Agreement',
    value: numberFormatter.format(loanAgreements.value.length),
    hint:
      dkProjects.value.length === 0
        ? 'Menunggu Daftar Kegiatan'
        : `${numberFormatter.format(
            dkProjects.value.filter((project) => loanAgreementsForProject(project).length === 0)
              .length,
          )} Daftar Kegiatan belum legal binding`,
    icon: 'pi pi-file-edit',
    state:
      dkProjects.value.length > 0 &&
      dkProjects.value.every((project) => loanAgreementsForProject(project).length > 0)
        ? 'completed'
        : 'pending',
  },
  {
    label: 'Monitoring Disbursement',
    value: numberFormatter.format(monitoringRows.value.length),
    hint:
      loanAgreements.value.length === 0
        ? 'Menunggu Loan Agreement'
        : `${numberFormatter.format(loanAgreements.value.length)} Loan Agreement tercakup`,
    icon: 'pi pi-chart-line',
    state: monitoringRows.value.length > 0 ? 'completed' : 'pending',
  },
  {
    label: 'Total Pendanaan',
    value: formatCompactUsd(totalFundingUsd.value),
    hint: 'Akumulasi sumber pendanaan Green Book',
    icon: 'pi pi-wallet',
    state: totalFundingUsd.value > 0 ? 'completed' : 'pending',
  },
])

const snapshotSteps = computed<JourneySnapshotStep[]>(() => {
  const greenBookRevisionLabels = Array.from(
    new Set(
      gbProjects.value
        .map((project) => project.green_book_revision_label || project.gb_code)
        .filter(Boolean),
    ),
  )
  const extendedLoanAgreementCount = loanAgreements.value.filter(
    (loanAgreement) => loanAgreement.is_extended,
  ).length

  return [
    {
      key: 'blue-book',
      label: 'Blue Book',
      value: props.journey.bb_project.blue_book_revision_label || props.journey.bb_project.bb_code,
      state: props.journey.bb_project.has_newer_revision ? 'warning' : 'completed',
      hint: props.journey.bb_project.has_newer_revision
        ? `Versi terbaru: ${props.journey.bb_project.latest_blue_book_revision_label || '-'}`
        : 'Snapshot yang sedang dibuka',
    },
    {
      key: 'green-book',
      label: 'Green Book',
      value:
        gbProjects.value.length === 0
          ? 'Belum ada'
          : greenBookRevisionLabels.length === 1
            ? (greenBookRevisionLabels[0] ?? 'Green Book')
            : `${numberFormatter.format(greenBookRevisionLabels.length)} revisi Green Book`,
      state:
        gbProjects.value.length === 0
          ? 'pending'
          : latestGreenBookWarnings.value > 0
            ? 'warning'
            : 'completed',
      hint:
        latestGreenBookWarnings.value > 0
          ? `${numberFormatter.format(latestGreenBookWarnings.value)} punya revisi lebih baru`
          : `${numberFormatter.format(gbProjects.value.length)} proyek Green Book`,
    },
    {
      key: 'daftar-kegiatan',
      label: 'Daftar Kegiatan',
      value:
        dkProjects.value.length === 0
          ? 'Belum ada'
          : `${numberFormatter.format(dkProjects.value.length)} snapshot`,
      state: dkProjects.value.length > 0 ? 'completed' : 'pending',
      hint: 'Tetap menunjuk snapshot Green Book saat dibuat',
    },
    {
      key: 'loan-agreement',
      label: 'Loan Agreement',
      value:
        loanAgreements.value.length === 0
          ? 'Belum ada'
          : `${numberFormatter.format(loanAgreements.value.length)} dokumen`,
      state:
        extendedLoanAgreementCount > 0
          ? 'extended'
          : loanAgreements.value.length > 0
            ? 'completed'
            : 'pending',
      hint:
        extendedLoanAgreementCount > 0
          ? `${numberFormatter.format(extendedLoanAgreementCount)} diperpanjang`
          : 'Legal binding per Daftar Kegiatan',
    },
    {
      key: 'monitoring',
      label: 'Monitoring Disbursement',
      value:
        monitoringRows.value.length === 0
          ? 'Belum ada'
          : `${numberFormatter.format(monitoringRows.value.length)} entri`,
      state: monitoringRows.value.length > 0 ? 'completed' : 'pending',
      hint: 'Rencana dan realisasi per triwulan',
    },
  ]
})

const matrixRows = computed<JourneyMatrixRow[]>(() => {
  if (gbProjects.value.length === 0) {
    return [
      {
        key: props.journey.bb_project.id,
        project_label: props.journey.bb_project.bb_code,
        project_name: props.journey.bb_project.project_name,
        funding_usd: 0,
        stages: [
          stage('blue-book', 'Blue Book', props.journey.bb_project.bb_code, 'completed'),
          stage('green-book', 'Green Book', 'Belum ada', 'pending'),
          stage('daftar-kegiatan', 'Daftar Kegiatan', 'Belum ada', 'pending'),
          stage('loan-agreement', 'Loan Agreement', 'Belum ada', 'pending'),
          stage('monitoring', 'Monitoring', 'Belum ada', 'pending'),
        ],
      },
    ]
  }

  return gbProjects.value.flatMap((greenBookProject) => {
    if (greenBookProject.dk_projects.length === 0) {
      return [
        {
          key: greenBookProject.id,
          project_label: greenBookProject.gb_code,
          project_name: greenBookProject.project_name,
          funding_usd: fundingTotalForGreenBook(greenBookProject),
          stages: [
            stage('blue-book', 'Blue Book', props.journey.bb_project.bb_code, 'completed'),
            stage('green-book', 'Green Book', greenBookProject.gb_code, 'completed'),
            stage('daftar-kegiatan', 'Daftar Kegiatan', 'Belum ada', 'pending'),
            stage('loan-agreement', 'Loan Agreement', 'Belum ada', 'pending'),
            stage('monitoring', 'Monitoring', 'Belum ada', 'pending'),
          ],
        },
      ]
    }

    return greenBookProject.dk_projects.map((dkProject) => {
      const projectLoanAgreements = loanAgreementsForProject(dkProject)
      const monitoringCount = projectLoanAgreements.reduce(
        (sum, loanAgreement) => sum + loanAgreement.monitoring.length,
        0,
      )

      return {
        key: dkProject.id,
        project_label: `${greenBookProject.gb_code} / ${dkLabel(dkProject)}`,
        project_name: dkProject.project_name || greenBookProject.project_name,
        funding_usd: fundingTotalForGreenBook(greenBookProject),
        stages: [
          stage('blue-book', 'Blue Book', props.journey.bb_project.bb_code, 'completed'),
          stage('green-book', 'Green Book', greenBookProject.gb_code, 'completed'),
          stage('daftar-kegiatan', 'Daftar Kegiatan', dkLabel(dkProject), 'completed'),
          stage(
            'loan-agreement',
            'Loan Agreement',
            loanAgreementStageValue(projectLoanAgreements),
            loanAgreementStageState(projectLoanAgreements),
          ),
          stage(
            'monitoring',
            'Monitoring',
            monitoringCount > 0 ? `${numberFormatter.format(monitoringCount)} entri` : 'Belum ada',
            monitoringCount > 0 ? 'completed' : 'pending',
          ),
        ],
      }
    })
  })
})

function stage(
  key: string,
  label: string,
  value: string,
  state: JourneyStageState,
): JourneyMatrixStage {
  return { key, label, value, state }
}

function loanAgreementsForProject(project: DKProjectJourney) {
  return project.loan_agreements ?? []
}

function loanAgreementStageValue(loanAgreements: LAJourney[]) {
  if (loanAgreements.length === 0) return 'Belum ada'
  if (loanAgreements.length === 1) return loanAgreements[0]?.loan_code ?? '1 dokumen'
  return `${numberFormatter.format(loanAgreements.length)} dokumen`
}

function loanAgreementStageState(loanAgreements: LAJourney[]): JourneyStageState {
  if (loanAgreements.length === 0) return 'pending'
  return loanAgreements.some((loanAgreement) => loanAgreement.is_extended)
    ? 'extended'
    : 'completed'
}

function dkLabel(project: DKProjectJourney) {
  return (
    project.daftar_kegiatan?.letter_number ||
    project.daftar_kegiatan?.subject ||
    project.project_name ||
    project.id
  )
}

function fundingTotalForGreenBook(project: GBProjectJourney) {
  return project.funding_sources.reduce(
    (sum, source) =>
      sum + (source.loan_usd ?? 0) + (source.grant_usd ?? 0) + (source.local_usd ?? 0),
    0,
  )
}

function formatCompactUsd(value: number) {
  return new Intl.NumberFormat('en-US', {
    notation: 'compact',
    maximumFractionDigits: 1,
  }).format(value)
}

function stateClass(state: JourneyStageState) {
  if (state === 'extended') return 'border-amber-200 bg-amber-50 text-amber-700'
  if (state === 'warning') return 'border-orange-200 bg-orange-50 text-orange-700'
  if (state === 'completed') return 'border-emerald-200 bg-emerald-50 text-emerald-700'
  return 'border-surface-200 bg-surface-50 text-surface-500'
}

function iconClass(state: JourneyStageState) {
  if (state === 'extended') return 'bg-amber-100 text-amber-700'
  if (state === 'warning') return 'bg-orange-100 text-orange-700'
  if (state === 'completed') return 'bg-emerald-100 text-emerald-700'
  return 'bg-surface-100 text-surface-500'
}

function dotClass(state: JourneyStageState) {
  if (state === 'extended') return 'bg-amber-500'
  if (state === 'warning') return 'bg-orange-500'
  if (state === 'completed') return 'bg-emerald-500'
  return 'bg-surface-300'
}

function fundingWidth(value: number) {
  if (maxFundingUsd.value <= 0 || value <= 0) return '0%'
  return `${Math.max(6, Math.round((value / maxFundingUsd.value) * 100))}%`
}
</script>

<template>
  <section class="space-y-4">
    <div class="rounded-lg border border-surface-200 bg-white p-4">
      <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-base font-semibold text-surface-950">Snapshot Perjalanan</h2>
          <p class="text-sm text-surface-500">
            Jalur konkret yang sedang dipakai downstream, termasuk indikator revisi.
          </p>
        </div>
        <span
          class="rounded-full border border-surface-200 bg-surface-50 px-3 py-1 text-xs font-medium text-surface-600"
        >
          {{ journey.bb_project.bb_code }}
        </span>
      </div>

      <div class="grid gap-3 lg:grid-cols-5">
        <div
          v-for="(step, index) in snapshotSteps"
          :key="step.key"
          class="relative rounded-lg border p-3"
          :class="stateClass(step.state)"
        >
          <span
            v-if="index < snapshotSteps.length - 1"
            class="absolute right-[-1.15rem] top-1/2 hidden h-px w-5 bg-surface-200 lg:block"
          />
          <p class="text-xs font-semibold uppercase tracking-wide">{{ step.label }}</p>
          <p class="mt-1 line-clamp-2 text-sm font-semibold">{{ step.value }}</p>
          <p v-if="step.hint" class="mt-1 text-xs opacity-80">{{ step.hint }}</p>
        </div>
      </div>
    </div>

    <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
      <div
        v-for="metric in summaryMetrics"
        :key="metric.label"
        class="rounded-lg border border-surface-200 bg-white p-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <p class="text-sm font-medium text-surface-500">{{ metric.label }}</p>
            <p class="mt-1 text-2xl font-semibold text-surface-950">{{ metric.value }}</p>
          </div>
          <span
            class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full"
            :class="iconClass(metric.state)"
          >
            <i :class="metric.icon" />
          </span>
        </div>
        <p class="mt-3 line-clamp-2 text-sm text-surface-500">{{ metric.hint }}</p>
      </div>
    </div>

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1.25fr)_minmax(22rem,0.75fr)]">
      <div class="rounded-lg border border-surface-200 bg-white p-4">
        <div class="mb-4">
          <h2 class="text-base font-semibold text-surface-950">Stage Matrix</h2>
          <p class="text-sm text-surface-500">
            Setiap baris menunjukkan posisi proyek dari Blue Book sampai Monitoring.
          </p>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-[56rem] w-full border-separate border-spacing-0 text-left text-sm">
            <thead class="bg-surface-50 text-left text-xs font-semibold uppercase tracking-wide text-surface-500">
              <tr>
                <th class="border-b border-surface-100 px-3 py-2">Proyek</th>
                <th class="border-b border-surface-100 px-3 py-2">Pendanaan</th>
                <th
                  v-for="stageHeader in matrixRows[0]?.stages ?? []"
                  :key="stageHeader.key"
                  class="border-b border-surface-100 px-3 py-2"
                >
                  {{ stageHeader.label }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in matrixRows" :key="row.key" class="align-top">
                <td class="border-b border-surface-100 px-3 py-3">
                  <p class="font-semibold text-surface-900">{{ row.project_label }}</p>
                  <p class="mt-1 line-clamp-2 text-xs text-surface-500">{{ row.project_name }}</p>
                </td>
                <td class="border-b border-surface-100 px-3 py-3 font-medium text-surface-700">
                  <CurrencyDisplay :amount="row.funding_usd" currency="USD" compact />
                </td>
                <td
                  v-for="stageItem in row.stages"
                  :key="`${row.key}-${stageItem.key}`"
                  class="border-b border-surface-100 px-3 py-3"
                >
                  <div class="inline-flex max-w-[11rem] items-center gap-2">
                    <span
                      class="h-2 w-2 shrink-0 rounded-full"
                      :class="dotClass(stageItem.state)"
                    />
                    <span
                      class="truncate text-xs font-medium"
                      :class="
                        stageItem.state === 'pending'
                          ? 'italic text-surface-400'
                          : 'text-surface-700'
                      "
                    >
                      {{ stageItem.value }}
                    </span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="space-y-4">
        <div class="rounded-lg border border-surface-200 bg-white p-4">
          <div class="mb-4">
            <h2 class="text-base font-semibold text-surface-950">Gap Utama</h2>
            <p class="text-sm text-surface-500">Tahap yang perlu ditindaklanjuti.</p>
          </div>
          <div
            v-if="missingItems.length === 0"
            class="flex items-center gap-2 text-sm text-emerald-700"
          >
            <i class="pi pi-check-circle" />
            Semua tahap utama sudah memiliki data.
          </div>
          <ul v-else class="space-y-2">
            <li
              v-for="item in missingItems"
              :key="item"
              class="flex gap-2 text-sm text-surface-700"
            >
              <span class="mt-1 h-2 w-2 shrink-0 rounded-full bg-amber-500" />
              <span>{{ item }}</span>
            </li>
          </ul>
        </div>

        <div class="rounded-lg border border-surface-200 bg-white p-4">
          <div class="mb-4">
            <h2 class="text-base font-semibold text-surface-950">Funding Breakdown</h2>
            <p class="text-sm text-surface-500">
              Akumulasi lender dari sumber pendanaan Green Book.
            </p>
          </div>
          <p v-if="fundingGroups.length === 0" class="text-sm italic text-surface-400">
            Belum ada funding source.
          </p>
          <div v-else class="space-y-3">
            <div v-for="group in fundingGroups" :key="group.key" class="space-y-1.5">
              <div class="flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-sm font-semibold text-surface-900">{{ group.label }}</p>
                  <p class="text-xs text-surface-500">{{ group.currency }}</p>
                </div>
                <span class="text-sm font-semibold text-surface-800">
                  <CurrencyDisplay :amount="group.total_usd" currency="USD" compact />
                </span>
              </div>
              <div class="h-2 overflow-hidden rounded-full bg-surface-100">
                <div
                  class="h-full rounded-full bg-prism-teal"
                  :style="{ width: fundingWidth(group.total_usd) }"
                />
              </div>
              <div class="flex flex-wrap gap-x-3 gap-y-1 text-xs text-surface-500">
                <span
                  >Pinjaman: <CurrencyDisplay :amount="group.loan_usd" currency="USD" compact
                /></span>
                <span
                  >Hibah: <CurrencyDisplay :amount="group.grant_usd" currency="USD" compact
                /></span>
                <span
                  >Dana pendamping:
                  <CurrencyDisplay :amount="group.local_usd" currency="USD" compact
                /></span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
