<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Message from 'primevue/message'
import Select from 'primevue/select'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import AbsorptionBar from '@/components/monitoring/AbsorptionBar.vue'
import MonitoringCard from '@/components/monitoring/MonitoringCard.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { useMonitoringStore } from '@/stores/monitoring.store'
import type {
  MonitoringDataQualityCode,
  MonitoringDisbursement,
  MonitoringListParams,
  MonitoringRiskCode,
  Quarter,
} from '@/types/monitoring.types'
import { formatDate } from '@/pages/loan-agreement/loan-agreement-page-utils'

const MonitoringChart = defineAsyncComponent(() => import('@/components/monitoring/MonitoringChart.vue'))
const route = useRoute()
const router = useRouter()
const monitoringStore = useMonitoringStore()
const { can } = usePermission()
const confirm = useConfirm()
const toast = useToast()
const hydratingRouteQuery = ref(false)
interface MonitoringFilterState {
  budget_year: string
  quarter: Quarter | ''
  risk_codes: MonitoringRiskCode[]
  data_quality_codes: MonitoringDataQualityCode[]
}

const listControls = useListControls<MonitoringFilterState>({
  initialFilters: {
    budget_year: '',
    quarter: '',
    risk_codes: [],
    data_quality_codes: [],
  },
  filterLabels: {
    budget_year: 'Tahun Anggaran',
    quarter: 'Triwulan',
    risk_codes: 'Risiko',
    data_quality_codes: 'Kelengkapan Data',
  },
  formatFilterValue: (key, value) => {
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

const loanAgreementId = computed(() => String(route.params.laId ?? ''))
const currentLA = computed(() => monitoringStore.currentLA)
const todayString = computed(() => new Date().toISOString().slice(0, 10))
const isNotEffective = computed(() => {
  if (!currentLA.value?.effective_date) return false
  return currentLA.value.effective_date.slice(0, 10) > todayString.value
})
const columns: ColumnDef[] = [
  { field: 'budget_year', header: 'Tahun Anggaran' },
  { field: 'quarter', header: 'Triwulan' },
  { field: 'exchange_rate_usd_idr', header: 'USD/IDR' },
  { field: 'exchange_rate_la_idr', header: 'Loan Agreement/IDR' },
  { field: 'planned_usd', header: 'Rencana USD' },
  { field: 'realized_usd', header: 'Realisasi USD' },
  { field: 'absorption_pct', header: 'Penyerapan' },
  { field: 'actions', header: 'Aksi' },
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

function buildListParams(): MonitoringListParams {
  const raw = listControls.buildParams() as Record<string, unknown>
  const params: MonitoringListParams = {
    page: Number(raw.page),
    limit: Number(raw.limit),
  }
  if (typeof raw.search === 'string') params.search = raw.search
  if (typeof raw.budget_year === 'string' && raw.budget_year.trim()) {
    const year = Number(raw.budget_year)
    if (Number.isFinite(year)) params.budget_year = year
  }
  if (raw.quarter === 'TW1' || raw.quarter === 'TW2' || raw.quarter === 'TW3' || raw.quarter === 'TW4') {
    params.quarter = raw.quarter
  }
  if (Array.isArray(raw.risk_codes)) params.risk_codes = raw.risk_codes as MonitoringRiskCode[]
  if (Array.isArray(raw.data_quality_codes)) {
    params.data_quality_codes = raw.data_quality_codes as MonitoringDataQualityCode[]
  }
  return params
}

async function loadData() {
  await Promise.all([
    monitoringStore.fetchLoanAgreement(loanAgreementId.value),
    monitoringStore.fetchMonitorings(loanAgreementId.value, buildListParams()),
  ])
}

function deleteMonitoring(row: MonitoringDisbursement) {
  confirm.confirmDelete(`monitoring ${row.quarter} ${row.budget_year}`, async () => {
    await monitoringStore.deleteMonitoring(loanAgreementId.value, row.id)
    toast.success('Berhasil', 'Monitoring berhasil dihapus')
  })
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
  listControls.draftFilters.budget_year = routeQueryString('budget_year')
  listControls.draftFilters.quarter = routeQuarter('quarter')
  listControls.draftFilters.risk_codes = routeRiskCodes('risk_codes')
  listControls.draftFilters.data_quality_codes = routeDataQualityCodes('data_quality_codes')
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
    void monitoringStore.fetchMonitorings(loanAgreementId.value, buildListParams())
  },
)

watch(
  () => route.query,
  () => {
    hydrateFiltersFromRouteQuery()
    void monitoringStore.fetchMonitorings(loanAgreementId.value, buildListParams())
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
      :title="currentLA ? `Monitoring ${currentLA.loan_code}` : 'Monitoring Disbursement'"
      subtitle="Rencana dan realisasi per triwulan"
    >
      <template #actions>
        <Button
          label="Kembali ke Loan Agreement"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'loan-agreement-detail', params: { id: loanAgreementId } })"
        />
        <Button
          v-if="can('monitoring_disbursement', 'create') && isNotEffective"
          label="Tambah Monitoring"
          icon="pi pi-plus"
          disabled
        />
        <Button
          v-else-if="can('monitoring_disbursement', 'create')"
          as="router-link"
          :to="{ name: 'monitoring-create', params: { laId: loanAgreementId } }"
          label="Tambah Monitoring"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <Message v-if="isNotEffective" severity="warn" :closable="false">
      Loan Agreement belum efektif - monitoring belum bisa diinput. Tanggal efektif:
      {{ formatDate(currentLA?.effective_date ?? '') }}
    </Message>

    <section v-if="currentLA" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-4">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Kode Loan</p>
        <p class="font-semibold text-surface-950">{{ currentLA.loan_code }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Lender</p>
        <p class="font-semibold text-surface-950">{{ currentLA.lender.name }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Mata Uang</p>
        <p class="font-semibold text-surface-950">{{ currentLA.currency }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Tanggal Efektif</p>
        <p class="font-semibold text-surface-950">{{ formatDate(currentLA.effective_date) }}</p>
      </div>
    </section>

    <MonitoringChart :data="monitoringStore.monitorings" />

    <div v-if="monitoringStore.monitorings.length" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <MonitoringCard
        v-for="monitoring in monitoringStore.monitorings.slice(0, 4)"
        :key="monitoring.id"
        :monitoring="monitoring"
      />
    </div>

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari tahun, triwulan, atau komponen"
      :active-filters="listControls.activeFilterPills.value"
      :filter-count="listControls.activeFilterCount.value"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
      <template #filters>
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
      :data="monitoringStore.monitorings as unknown as Record<string, unknown>[]"
      :columns="columns"
      :loading="monitoringStore.loading"
      :total="monitoringStore.total"
    >
      <template #body-row="{ row, column }">
        <StatusBadge v-if="column.field === 'quarter'" :status="String(row.quarter)" />
        <CurrencyDisplay
          v-else-if="column.field === 'planned_usd'"
          :amount="Number(row.planned_usd)"
          currency="USD"
        />
        <CurrencyDisplay
          v-else-if="column.field === 'realized_usd'"
          :amount="Number(row.realized_usd)"
          currency="USD"
        />
        <AbsorptionBar
          v-else-if="column.field === 'absorption_pct'"
          :pct="Number(row.absorption_pct)"
        />
        <span v-else-if="column.field === 'exchange_rate_usd_idr'">
          {{ Number(row.exchange_rate_usd_idr).toLocaleString('id-ID') }}
        </span>
        <span v-else-if="column.field === 'exchange_rate_la_idr'">
          {{ Number(row.exchange_rate_la_idr).toLocaleString('id-ID') }}
        </span>
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap items-center gap-2">
          <span class="mr-1 text-xs text-surface-500">
            {{ ((row as unknown as MonitoringDisbursement).komponen ?? []).length }} komponen
          </span>
          <Button
            v-if="can('monitoring_disbursement', 'update')"
            as="router-link"
            :to="{ name: 'monitoring-edit', params: { laId: loanAgreementId, id: row.id } }"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            severity="secondary"
            outlined
          />
          <Button
            v-if="can('monitoring_disbursement', 'delete')"
            icon="pi pi-trash"
            label="Hapus"
            size="small"
            severity="danger"
            outlined
            @click="deleteMonitoring(row as unknown as MonitoringDisbursement)"
          />
        </div>
        <span v-else>{{ row[column.field] ?? '-' }}</span>
      </template>
    </DataTable>
  </section>
</template>
