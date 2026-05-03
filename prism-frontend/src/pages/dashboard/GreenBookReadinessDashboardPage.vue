<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import Message from 'primevue/message'
import Select from 'primevue/select'
import DashboardFilterBar from '@/components/dashboard/DashboardFilterBar.vue'
import DisbursementPlanChart from '@/components/dashboard/DisbursementPlanChart.vue'
import FundingAllocationChart from '@/components/dashboard/FundingAllocationChart.vue'
import ReadinessScoreCard from '@/components/dashboard/ReadinessScoreCard.vue'
import ReadinessWorklistTable from '@/components/dashboard/ReadinessWorklistTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type { GreenBookReadinessParams, GreenBookReadinessStatus } from '@/types/dashboard.types'

const dashboard = useDashboardStore()

const filters = reactive<{
  publish_year: number | null
  green_book_id: string | null
  institution_id: string | null
  lender_id: string | null
  readiness_status: GreenBookReadinessStatus | null
}>({
  publish_year: null,
  green_book_id: null,
  institution_id: null,
  lender_id: null,
  readiness_status: null,
})

const publishYearOptions = computed(() =>
  (dashboard.filterOptions.publish_year ?? [])
    .map((item) => ({
      label: item.label,
      value: Number(item.key),
    }))
    .filter((item) => Number.isFinite(item.value)),
)

const greenBookOptions = computed(() =>
  (dashboard.filterOptions.green_book ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const institutionOptions = computed(() =>
  (dashboard.filterOptions.institution ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const lenderOptions = computed(() =>
  (dashboard.filterOptions.lender ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const readinessStatusOptions = computed(() => [
  { label: 'Semua status', value: null },
  { label: 'Ready', value: 'READY' },
  { label: 'Partial', value: 'PARTIAL' },
  { label: 'Incomplete', value: 'INCOMPLETE' },
  { label: 'Cofinancing', value: 'COFINANCING' },
])

const readiness = computed(() => dashboard.greenBookReadiness)
const summary = computed(() => readiness.value?.summary)
const allocation = computed(
  () =>
    readiness.value?.funding_allocation ?? {
      services: 0,
      constructions: 0,
      goods: 0,
      trainings: 0,
      other: 0,
    },
)

function buildParams(): GreenBookReadinessParams {
  return {
    publish_year: filters.publish_year ?? undefined,
    green_book_id: filters.green_book_id ?? undefined,
    institution_id: filters.institution_id ?? undefined,
    lender_id: filters.lender_id ?? undefined,
    readiness_status: filters.readiness_status ?? undefined,
  }
}

async function loadReadiness() {
  await dashboard.fetchGreenBookReadiness(buildParams())
}

async function loadInitialData() {
  await Promise.allSettled([dashboard.fetchFilterOptions(), loadReadiness()])
}

function clearFilters() {
  filters.publish_year = null
  filters.green_book_id = null
  filters.institution_id = null
  filters.lender_id = null
  filters.readiness_status = null
  void loadReadiness()
}

onMounted(() => {
  void loadInitialData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Green Book Readiness"
      subtitle="Kesiapan proyek Green Book berdasarkan referensi Blue Book, K/L, lokasi, pendanaan, activities, disbursement plan, dan funding allocation."
    />

    <DashboardFilterBar
      :loading="dashboard.greenBookReadinessLoading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[10rem_1fr_1fr_1fr_11rem]"
      @apply="loadReadiness"
      @reset="clearFilters"
    >
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">GB Year</span>
          <Select
            v-model="filters.publish_year"
            :options="publishYearOptions"
            option-label="label"
            option-value="value"
            show-clear
            class="w-full"
            placeholder="Semua"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Green Book</span>
          <Select
            v-model="filters.green_book_id"
            :options="greenBookOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua Green Book"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Institution</span>
          <Select
            v-model="filters.institution_id"
            :options="institutionOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua K/L"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <Select
            v-model="filters.lender_id"
            :options="lenderOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua lender"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Status</span>
          <Select
            v-model="filters.readiness_status"
            :options="readinessStatusOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>
    </DashboardFilterBar>

    <Message v-if="dashboard.greenBookReadinessError" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.greenBookReadinessError }}
    </Message>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-6">
      <ReadinessScoreCard
        label="Total Green Book Projects"
        :value="summary?.total_projects ?? 0"
        unit="project"
        tone="neutral"
      />
      <ReadinessScoreCard label="Loan" :value="summary?.total_loan_usd ?? 0" unit="USD" tone="ready" />
      <ReadinessScoreCard label="Grant" :value="summary?.total_grant_usd ?? 0" unit="USD" tone="partial" />
      <ReadinessScoreCard label="Local" :value="summary?.total_local_usd ?? 0" unit="USD" tone="neutral" />
      <ReadinessScoreCard
        label="Cofinancing"
        :value="summary?.projects_with_cofinancing ?? 0"
        unit="project"
        tone="cofinancing"
      />
      <ReadinessScoreCard
        label="Incomplete"
        :value="summary?.projects_incomplete ?? 0"
        unit="project"
        tone="incomplete"
      />
    </section>

    <section class="grid gap-4 xl:grid-cols-2">
      <DisbursementPlanChart
        :data="readiness?.disbursement_plan_by_year ?? []"
        :loading="dashboard.greenBookReadinessLoading"
      />
      <FundingAllocationChart
        :allocation="allocation"
        :loading="dashboard.greenBookReadinessLoading"
      />
    </section>

    <ReadinessWorklistTable
      :items="readiness?.readiness_items ?? []"
      :loading="dashboard.greenBookReadinessLoading"
    />
  </section>
</template>
