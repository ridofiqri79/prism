<script setup lang="ts">
import { computed, markRaw, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Tag from 'primevue/tag'
import AmountDisplay from '@/components/dashboard/AmountDisplay.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import DataQualityGovernanceDashboardPage from '@/pages/dashboard/DataQualityGovernanceDashboardPage.vue'
import ExecutivePortfolioDashboardPage from '@/pages/dashboard/ExecutivePortfolioDashboardPage.vue'
import GreenBookReadinessDashboardPage from '@/pages/dashboard/GreenBookReadinessDashboardPage.vue'
import KLPortfolioPerformanceDashboardPage from '@/pages/dashboard/KLPortfolioPerformanceDashboardPage.vue'
import LenderFinancingMixDashboardPage from '@/pages/dashboard/LenderFinancingMixDashboardPage.vue'
import PipelineBottleneckDashboardPage from '@/pages/dashboard/PipelineBottleneckDashboardPage.vue'

const route = useRoute()
const router = useRouter()
const dashboard = useDashboardStore()

const dashboardTabs = [
  {
    key: 'executive',
    label: 'Ringkasan Eksekutif',
    shortLabel: 'Eksekutif',
    icon: 'pi pi-briefcase',
    component: markRaw(ExecutivePortfolioDashboardPage),
    caption: 'Kesehatan portofolio dan risiko utama',
    question: 'Apakah portofolio pinjaman luar negeri sedang sehat dari Blue Book sampai Loan Agreement?',
    scope: 'Portofolio nasional, funnel tahap, K/L teratas, lender utama, dan risiko lintas alur kerja.',
    outcomes: ['Baca kesehatan portofolio', 'Identifikasi risiko utama', 'Lihat eksposur terbesar'],
  },
  {
    key: 'pipeline',
    label: 'Alur Perencanaan',
    shortLabel: 'Alur',
    icon: 'pi pi-sort-amount-down',
    component: markRaw(PipelineBottleneckDashboardPage),
    caption: 'Hambatan sebelum komitmen legal',
    question: 'Di tahap mana proyek tertahan sebelum menjadi komitmen legal?',
    scope: 'Blue Book, Letter of Intent, Green Book, Daftar Kegiatan, dan Loan Agreement.',
    outcomes: ['Cari hambatan per tahap', 'Prioritaskan usia tertahan', 'Buka perjalanan proyek'],
  },
  {
    key: 'readiness',
    label: 'Kesiapan Green Book',
    shortLabel: 'Green Book',
    icon: 'pi pi-check-square',
    component: markRaw(GreenBookReadinessDashboardPage),
    caption: 'Kesiapan proyek Green Book',
    question: 'Proyek Green Book mana yang siap masuk Daftar Kegiatan?',
    scope: 'Sumber pendanaan, aktivitas, rencana penarikan, alokasi pendanaan, dan pendanaan bersama.',
    outcomes: ['Ukur skor kesiapan', 'Temukan isian kosong', 'Validasi struktur pendanaan'],
  },
  {
    key: 'financing',
    label: 'Pembiayaan & Lender',
    shortLabel: 'Lender',
    icon: 'pi pi-building-columns',
    component: markRaw(LenderFinancingMixDashboardPage),
    caption: 'Konversi lender dan mata uang',
    question: 'Seberapa kuat kepastian lender dari indikasi sampai Loan Agreement?',
    scope: 'Indikasi lender, Letter of Intent, sumber pendanaan Green Book, Daftar Kegiatan, Loan Agreement, dan mata uang.',
    outcomes: ['Pantau konversi lender', 'Baca eksposur mata uang', 'Cek pendanaan bersama'],
  },
  {
    key: 'institution',
    label: 'Kinerja K/L',
    shortLabel: 'K/L',
    icon: 'pi pi-sitemap',
    component: markRaw(KLPortfolioPerformanceDashboardPage),
    caption: 'Eksposur dan risiko K/L',
    question: 'Kementerian/Lembaga mana yang memegang eksposur dan risiko terbesar?',
    scope: 'Agregasi K/L, peran Executing Agency/Implementing Agency, nilai alur perencanaan, komitmen legal, dan jumlah risiko.',
    outcomes: ['Bandingkan kinerja K/L', 'Urutkan eksposur', 'Lihat risiko portofolio'],
  },
  {
    key: 'quality',
    label: 'Kualitas Data',
    shortLabel: 'Data',
    icon: 'pi pi-shield',
    component: markRaw(DataQualityGovernanceDashboardPage),
    caption: 'Isu data dan audit',
    question: 'Data mana yang menghambat alur kerja, analitik, atau kepatuhan aturan bisnis?',
    scope: 'Relasi kosong, konsistensi aturan bisnis, integritas data, dan ringkasan audit untuk ADMIN.',
    outcomes: ['Tindaklanjuti isu', 'Validasi relasi data', 'Tinjau aktivitas audit'],
  },
] as const

type DashboardTabKey = (typeof dashboardTabs)[number]['key']

const defaultTabKey: DashboardTabKey = 'executive'
const dashboardTabKeys = new Set<DashboardTabKey>(dashboardTabs.map((tab) => tab.key))

function normalizeTab(value: unknown): DashboardTabKey {
  const candidate = Array.isArray(value) ? value[0] : value

  if (typeof candidate === 'string' && dashboardTabKeys.has(candidate as DashboardTabKey)) {
    return candidate as DashboardTabKey
  }

  return defaultTabKey
}

const activeTabKey = ref<DashboardTabKey>(normalizeTab(route.query.tab))

const activeTab = computed(
  () => dashboardTabs.find((tab) => tab.key === activeTabKey.value) ?? dashboardTabs[0],
)

const activeComponentProps = computed(() => ({
  embedded: true,
  ...(activeTabKey.value === 'executive' ? { autoload: false } : {}),
}))

const projectStageMetrics = computed(() => dashboard.executivePortfolio?.funnel ?? [])

function projectStageCount(stage: string, fallback = 0) {
  return projectStageMetrics.value.find((item) => item.stage === stage)?.project_count ?? fallback
}

const overviewMetrics = computed(() => [
  {
    key: 'bb',
    label: 'Proyek Blue Book',
    value: projectStageCount('BB', dashboard.summary?.total_bb_projects ?? 0),
    unit: 'project',
    hint: 'Total proyek di tahap Blue Book',
  },
  {
    key: 'gb',
    label: 'Proyek Green Book',
    value: projectStageCount('GB', dashboard.summary?.total_gb_projects ?? 0),
    unit: 'project',
    hint: 'Total proyek di tahap Green Book',
  },
  {
    key: 'dk',
    label: 'Proyek Daftar Kegiatan',
    value: projectStageCount('DK'),
    unit: 'project',
    hint: 'Total proyek yang sudah masuk Daftar Kegiatan',
  },
  {
    key: 'la',
    label: 'Proyek Loan Agreement',
    value: projectStageCount('LA', dashboard.summary?.total_loan_agreements ?? 0),
    unit: 'project',
    hint: 'Total proyek yang sudah memiliki Loan Agreement',
  },
])

watch(
  () => route.query.tab,
  (value) => {
    activeTabKey.value = normalizeTab(value)
  },
)

function selectTab(tabKey: DashboardTabKey) {
  activeTabKey.value = tabKey

  const query = { ...route.query }
  if (tabKey === defaultTabKey) {
    delete query.tab
  } else {
    query.tab = tabKey
  }

  void router.replace({ name: 'dashboard', query })
}

function tabButtonClass(tabKey: DashboardTabKey) {
  const isActive = tabKey === activeTabKey.value

  return [
    'group flex min-h-[4.75rem] min-w-[13rem] flex-1 items-start gap-3 rounded-lg border px-3 py-3 text-left transition-colors',
    isActive
      ? 'border-primary-500 bg-primary-50 text-primary-950'
      : 'border-surface-200 bg-white text-surface-700 hover:border-primary-300 hover:bg-surface-50',
  ]
}

onMounted(() => {
  void Promise.all([
    dashboard.fetchSummary(),
    dashboard.fetchFilterOptions(),
    dashboard.fetchExecutivePortfolio(),
  ])
})
</script>

<template>
  <section class="space-y-5">
    <PageHeader
      title="Dasbor"
      subtitle="Satu ruang kerja untuk membaca portofolio, hambatan, kesiapan, pembiayaan, kinerja K/L, dan kualitas data."
    >
      <template #actions>
        <Tag value="Analitik baca-saja" severity="secondary" />
      </template>
    </PageHeader>

    <section class="rounded-lg border border-surface-200 bg-white p-3">
      <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <div
          v-for="metric in overviewMetrics"
          :key="metric.key"
          class="min-w-0 rounded-md border border-surface-200 bg-surface-0 p-3"
        >
          <p class="text-xs font-medium text-surface-500">{{ metric.label }}</p>
          <p class="mt-2 break-words text-xl font-semibold text-surface-950">
            <AmountDisplay
              :value="metric.value"
              :unit="metric.unit"
              :maximum-fraction-digits="0"
            />
          </p>
          <p class="mt-1 text-xs text-surface-500">{{ metric.hint }}</p>
        </div>
      </div>

      <div class="mt-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase text-primary">Alur bisnis dashboard</p>
          <h2 class="mt-1 text-xl font-semibold text-surface-950">{{ activeTab.label }}</h2>
          <p class="mt-1 max-w-3xl text-sm leading-6 text-surface-600">{{ activeTab.question }}</p>
        </div>
        <div class="flex shrink-0 flex-wrap gap-2">
          <Tag value="Blue Book -> Green Book -> Daftar Kegiatan -> Loan Agreement" severity="info" />
          <Tag value="Snapshot terbaru" severity="secondary" />
        </div>
      </div>

      <div class="mt-4 flex gap-2 overflow-x-auto pb-1">
        <button
          v-for="tab in dashboardTabs"
          :key="tab.key"
          type="button"
          :class="tabButtonClass(tab.key)"
          @click="selectTab(tab.key)"
        >
          <span
            class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-md"
            :class="tab.key === activeTabKey ? 'bg-primary text-primary-contrast' : 'bg-surface-100 text-surface-500 group-hover:text-primary'"
          >
            <i :class="[tab.icon, 'text-sm']" />
          </span>
          <span class="min-w-0">
            <span class="block text-sm font-semibold leading-5">{{ tab.shortLabel }}</span>
            <span class="mt-1 block text-xs leading-5 text-surface-500">
              {{ tab.caption }}
            </span>
          </span>
        </button>
      </div>
    </section>

    <section class="rounded-lg border border-surface-200 bg-surface-0 p-4">
      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_22rem] lg:items-start">
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase text-surface-500">Fokus analisis</p>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">{{ activeTab.label }}</h2>
          <p class="mt-2 text-sm leading-6 text-surface-600">{{ activeTab.scope }}</p>
        </div>

        <div class="grid gap-2 sm:grid-cols-3 lg:grid-cols-1">
          <div
            v-for="outcome in activeTab.outcomes"
            :key="outcome"
            class="flex min-h-10 items-center gap-2 rounded-md border border-surface-200 bg-white px-3 py-2 text-sm font-medium text-surface-700"
          >
            <i class="pi pi-check-circle shrink-0 text-primary" />
            <span class="min-w-0">{{ outcome }}</span>
          </div>
        </div>
      </div>
    </section>

    <KeepAlive>
      <component :is="activeTab.component" :key="activeTab.key" v-bind="activeComponentProps" />
    </KeepAlive>
  </section>
</template>
