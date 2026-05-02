<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useMonitoringStore } from '@/stores/monitoring.store'
import type {
  MonitoringDataQualityCode,
  MonitoringLoanAgreementListParams,
  MonitoringLoanAgreementReference,
  MonitoringRiskCode,
  Quarter,
} from '@/types/monitoring.types'
import { formatDate } from '@/pages/loan-agreement/loan-agreement-page-utils'

const monitoringStore = useMonitoringStore()
const { can } = usePermission()
const route = useRoute()
const hydratingRouteQuery = ref(false)

interface MonitoringOverviewFilterState {
  is_effective: boolean | null
  budget_year: string
  quarter: Quarter | ''
  risk_codes: MonitoringRiskCode[]
  data_quality_codes: MonitoringDataQualityCode[]
}

const listControls = useListControls<MonitoringOverviewFilterState>({
  initialFilters: {
    is_effective: null,
    budget_year: '',
    quarter: '',
    risk_codes: [],
    data_quality_codes: [],
  },
  filterLabels: {
    is_effective: 'Status Efektif',
    budget_year: 'Tahun Anggaran',
    quarter: 'Triwulan',
    risk_codes: 'Risiko',
    data_quality_codes: 'Kelengkapan Data',
  },
  formatFilterValue: (key, value) => {
    if (key === 'is_effective') {
      return value ? 'Efektif' : 'Belum efektif'
    }
    if (Array.isArray(value) && key === 'risk_codes') {
      return selectedLabelSummary(value.map((item) => riskCodeLabel(item as MonitoringRiskCode)))
    }
    if (Array.isArray(value) && key === 'data_quality_codes') {
      return selectedLabelSummary(
        value.map((item) => dataQualityCodeLabel(item as MonitoringDataQualityCode)),
      )
    }
    return String(value)
  },
})

const effectiveStatusOptions = [
  { label: 'Semua', value: null },
  { label: 'Efektif', value: true },
  { label: 'Belum efektif', value: false },
]
const quarterOptions: Array<{ label: string; value: Quarter }> = [
  { label: 'TW1', value: 'TW1' },
  { label: 'TW2', value: 'TW2' },
  { label: 'TW3', value: 'TW3' },
  { label: 'TW4', value: 'TW4' },
]
const riskCodeOptions: Array<{ label: string; value: MonitoringRiskCode }> = [
  { label: 'Loan Agreement efektif tanpa monitoring', value: 'EFFECTIVE_WITHOUT_MONITORING' },
  { label: 'Penyerapan rendah', value: 'LOW_ABSORPTION' },
]
const dataQualityCodeOptions: Array<{ label: string; value: MonitoringDataQualityCode }> = [
  { label: 'Loan Agreement efektif tanpa monitoring', value: 'EFFECTIVE_NO_MONITORING' },
  { label: 'Rencana nol, realisasi positif', value: 'PLANNED_ZERO_REALIZED_POSITIVE' },
]

const columns: ColumnDef[] = [
  { field: 'loan_code', header: 'Kode Pinjaman' },
  { field: 'lender', header: 'Lender' },
  { field: 'effective_date', header: 'Tanggal Efektif' },
  { field: 'amount_usd', header: 'Nilai USD' },
  { field: 'monitoring_count', header: 'Jumlah Monitoring' },
  { field: 'latest_monitoring_at', header: 'Update Terakhir' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Aksi' },
]

function buildListParams(): MonitoringLoanAgreementListParams {
  const raw = listControls.buildParams() as Record<string, unknown>
  const params: MonitoringLoanAgreementListParams = {
    page: Number(raw.page),
    limit: Number(raw.limit),
  }
  if (typeof raw.search === 'string') params.search = raw.search
  if (typeof raw.is_effective === 'boolean') params.is_effective = raw.is_effective
  if (typeof raw.budget_year === 'string' && raw.budget_year.trim()) {
    const year = Number(raw.budget_year)
    if (Number.isFinite(year)) params.budget_year = year
  }
  if (raw.quarter === 'TW1' || raw.quarter === 'TW2' || raw.quarter === 'TW3' || raw.quarter === 'TW4') {
    params.quarter = raw.quarter
  }
  if (Array.isArray(raw.risk_codes)) {
    params.risk_codes = raw.risk_codes as MonitoringRiskCode[]
  }
  if (Array.isArray(raw.data_quality_codes)) {
    params.data_quality_codes = raw.data_quality_codes as MonitoringDataQualityCode[]
  }
  return params
}

async function loadData() {
  await monitoringStore.fetchLoanAgreementReferences(buildListParams())
}

function formatDateTime(value?: string) {
  if (!value) return '-'
  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
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

function routeQueryBoolean(key: string): boolean | null {
  const value = routeQueryString(key).toLowerCase()
  if (value === 'true' || value === '1') return true
  if (value === 'false' || value === '0') return false
  return null
}

function routeQuarter(key: string): Quarter | '' {
  const value = routeQueryString(key).toUpperCase()
  return value === 'TW1' || value === 'TW2' || value === 'TW3' || value === 'TW4' ? value : ''
}

function routeRiskCodes(key: string): MonitoringRiskCode[] {
  return routeQueryValues(key).filter(
    (value): value is MonitoringRiskCode =>
      value === 'EFFECTIVE_WITHOUT_MONITORING' || value === 'LOW_ABSORPTION',
  )
}

function routeDataQualityCodes(key: string): MonitoringDataQualityCode[] {
  return routeQueryValues(key).filter(
    (value): value is MonitoringDataQualityCode =>
      value === 'EFFECTIVE_NO_MONITORING' || value === 'PLANNED_ZERO_REALIZED_POSITIVE',
  )
}

function hydrateFiltersFromRouteQuery() {
  hydratingRouteQuery.value = true
  listControls.draftFilters.is_effective = routeQueryBoolean('is_effective')
  listControls.draftFilters.budget_year = routeQueryString('budget_year')
  listControls.draftFilters.quarter = routeQuarter('quarter')
  listControls.draftFilters.risk_codes = routeRiskCodes('risk_codes')
  listControls.draftFilters.data_quality_codes = routeDataQualityCodes('data_quality_codes')

  listControls.appliedFilters.is_effective = listControls.draftFilters.is_effective
  listControls.appliedFilters.budget_year = listControls.draftFilters.budget_year
  listControls.appliedFilters.quarter = listControls.draftFilters.quarter
  listControls.appliedFilters.risk_codes = [...listControls.draftFilters.risk_codes]
  listControls.appliedFilters.data_quality_codes = [...listControls.draftFilters.data_quality_codes]
  listControls.search.value = routeQueryString('search')
  listControls.debouncedSearch.value = routeQueryString('search')

  window.queueMicrotask(() => {
    hydratingRouteQuery.value = false
  })
}

function selectedLabelSummary(labels: string[]) {
  if (labels.length <= 2) return labels.join(', ')
  return `${labels.slice(0, 2).join(', ')} +${labels.length - 2}`
}

function riskCodeLabel(code: MonitoringRiskCode) {
  return riskCodeOptions.find((option) => option.value === code)?.label ?? code
}

function dataQualityCodeLabel(code: MonitoringDataQualityCode) {
  return dataQualityCodeOptions.find((option) => option.value === code)?.label ?? code
}

watch(
  [
    listControls.page,
    listControls.limit,
    listControls.debouncedSearch,
    () => JSON.stringify(listControls.appliedFilters),
  ],
  () => {
    if (hydratingRouteQuery.value) return
    void loadData()
  },
)

watch(
  () => route.query,
  () => {
    hydrateFiltersFromRouteQuery()
    void loadData()
  },
)

onMounted(() => {
  hydrateFiltersFromRouteQuery()
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Monitoring Disbursement"
      subtitle="Pilih Loan Agreement untuk melihat dan menginput monitoring triwulan"
    />

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari kode pinjaman, lender, atau proyek"
      :active-filters="listControls.activeFilterPills.value"
      :filter-count="listControls.activeFilterCount.value"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Status Efektif</span>
          <Select
            v-model="listControls.draftFilters.is_effective"
            :options="effectiveStatusOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Tahun Anggaran</span>
          <InputText
            v-model="listControls.draftFilters.budget_year"
            inputmode="numeric"
            placeholder="Contoh: 2026"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Triwulan</span>
          <Select
            v-model="listControls.draftFilters.quarter"
            :options="quarterOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua triwulan"
            show-clear
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Risiko</span>
          <MultiSelect
            v-model="listControls.draftFilters.risk_codes"
            :options="riskCodeOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua risiko"
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Kelengkapan Data</span>
          <MultiSelect
            v-model="listControls.draftFilters.data_quality_codes"
            :options="dataQualityCodeOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua isu data"
            display="chip"
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

    <DataTable
      v-model:page="listControls.page.value"
      v-model:limit="listControls.limit.value"
      :data="monitoringStore.loanAgreementReferences as unknown as Record<string, unknown>[]"
      :columns="columns"
      :loading="monitoringStore.loading"
      :total="monitoringStore.referenceTotal"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'lender'">
          {{ (row as unknown as MonitoringLoanAgreementReference).lender?.name ?? '-' }}
        </span>
        <span v-else-if="column.field === 'effective_date'">
          {{ formatDate(String(row.effective_date)) }}
        </span>
        <CurrencyDisplay
          v-else-if="column.field === 'amount_usd'"
          :amount="Number(row.amount_usd)"
          currency="USD"
        />
        <span v-else-if="column.field === 'latest_monitoring_at'">
          {{ formatDateTime(String(row.latest_monitoring_at || '')) }}
        </span>
        <Tag
          v-else-if="column.field === 'status'"
          :value="(row as unknown as MonitoringLoanAgreementReference).is_effective ? 'Efektif' : 'Belum efektif'"
          :severity="(row as unknown as MonitoringLoanAgreementReference).is_effective ? 'success' : 'warn'"
          rounded
        />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'monitoring-list', params: { laId: row.id } }"
            icon="pi pi-chart-line"
            label="Buka Monitoring"
            size="small"
            outlined
          />
          <Button
            v-if="
              can('monitoring_disbursement', 'create') &&
              (row as unknown as MonitoringLoanAgreementReference).is_effective
            "
            as="router-link"
            :to="{ name: 'monitoring-create', params: { laId: row.id } }"
            icon="pi pi-plus"
            label="Tambah"
            size="small"
            severity="secondary"
            outlined
          />
        </div>
        <span v-else>{{ row[column.field] ?? '-' }}</span>
      </template>
    </DataTable>
  </section>
</template>
