<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useLoanAgreementStore } from '@/stores/loan-agreement.store'
import { useMasterStore } from '@/stores/master.store'
import type {
  LoanAgreement,
  LoanAgreementListParams,
  LoanAgreementRiskCode,
} from '@/types/loan-agreement.types'
import { formatDate } from './loan-agreement-page-utils'

const loanAgreementStore = useLoanAgreementStore()
const masterStore = useMasterStore()
const { can } = usePermission()
const route = useRoute()
const hydratingRouteQuery = ref(false)
interface LAFilterState {
  lender_id: string | null
  is_extended: boolean | null
  closing_date_before: string
  risk_codes: LoanAgreementRiskCode[]
}

const listControls = useListControls<LAFilterState>({
  initialFilters: {
    lender_id: null,
    is_extended: null,
    closing_date_before: '',
    risk_codes: [],
  },
  filterLabels: {
    lender_id: 'Lender',
    is_extended: 'Status Perpanjangan',
    closing_date_before: 'Penutupan sebelum',
    risk_codes: 'Risiko',
  },
  formatFilterValue: (key, value) => {
    if (key === 'lender_id' && typeof value === 'string') {
      return masterStore.lenders.find((lender) => lender.id === value)?.name ?? value
    }
    if (key === 'is_extended') {
      return value ? 'Diperpanjang' : 'Tidak diperpanjang'
    }
    if (Array.isArray(value) && key === 'risk_codes') {
      return selectedLabelSummary(value.map((item) => riskCodeLabel(item)))
    }
    return String(value)
  },
})

const isExtendedOptions = [
  { label: 'Semua', value: null },
  { label: 'Diperpanjang', value: true },
  { label: 'Tidak diperpanjang', value: false },
]
const riskCodeOptions: Array<{ label: string; value: LoanAgreementRiskCode }> = [
  { label: 'Mendekati closing date', value: 'CLOSING_RISK' },
  { label: 'Loan Agreement diperpanjang', value: 'EXTENDED_LOAN' },
]
const columns: ColumnDef[] = [
  { field: 'loan_code', header: 'Kode Pinjaman' },
  { field: 'lender', header: 'Lender' },
  { field: 'effective_date', header: 'Tanggal Efektif' },
  { field: 'closing_date', header: 'Tanggal Penutupan' },
  { field: 'currency', header: 'Mata Uang' },
  { field: 'amount_usd', header: 'Nilai USD' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Aksi' },
]

function buildListParams(): LoanAgreementListParams {
  return listControls.buildParams() as LoanAgreementListParams
}

async function loadData() {
  await loanAgreementStore.fetchLoanAgreements(buildListParams())
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

function routeRiskCodes(key: string): LoanAgreementRiskCode[] {
  return routeQueryValues(key).filter(
    (value): value is LoanAgreementRiskCode =>
      value === 'CLOSING_RISK' || value === 'EXTENDED_LOAN',
  )
}

function hydrateFiltersFromRouteQuery() {
  hydratingRouteQuery.value = true
  listControls.draftFilters.lender_id = routeQueryString('lender_id') || null
  listControls.draftFilters.is_extended = routeQueryBoolean('is_extended')
  listControls.draftFilters.closing_date_before = routeQueryString('closing_date_before')
  listControls.draftFilters.risk_codes = routeRiskCodes('risk_codes')
  listControls.appliedFilters.lender_id = listControls.draftFilters.lender_id
  listControls.appliedFilters.is_extended = listControls.draftFilters.is_extended
  listControls.appliedFilters.closing_date_before = listControls.draftFilters.closing_date_before
  listControls.appliedFilters.risk_codes = [...listControls.draftFilters.risk_codes]
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

function riskCodeLabel(code: LoanAgreementRiskCode) {
  return riskCodeOptions.find((option) => option.value === code)?.label ?? code
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
  void Promise.all([masterStore.fetchLenders(true, { limit: 1000 }), loadData()])
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Loan Agreement" subtitle="Daftar Loan Agreement dan status perpanjangan">
      <template #actions>
        <Button
          v-if="can('loan_agreement', 'create')"
          as="router-link"
          :to="{ name: 'loan-agreement-create' }"
          label="Buat Loan Agreement"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari kode pinjaman atau lender"
      :active-filters="listControls.activeFilterPills.value"
      :filter-count="listControls.activeFilterCount.value"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <LenderSelect v-model="listControls.draftFilters.lender_id" placeholder="Semua lender" />
        </label>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Status Perpanjangan</span>
          <Select
            v-model="listControls.draftFilters.is_extended"
            :options="isExtendedOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Penutupan Sebelum Tanggal</span>
          <InputText v-model="listControls.draftFilters.closing_date_before" type="date" class="w-full" />
        </label>
        <label class="block space-y-2 xl:col-span-2">
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
      </template>
    </SearchFilterBar>

    <DataTable
      v-model:page="listControls.page.value"
      v-model:limit="listControls.limit.value"
      :data="loanAgreementStore.loanAgreements as unknown as Record<string, unknown>[]"
      :columns="columns"
      :loading="loanAgreementStore.loading"
      :total="loanAgreementStore.total"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'lender'">{{ (row as LoanAgreement).lender?.name || '-' }}</span>
        <span v-else-if="column.field === 'effective_date'">{{ formatDate(String(row.effective_date)) }}</span>
        <span v-else-if="column.field === 'closing_date'">{{ formatDate(String(row.closing_date)) }}</span>
        <CurrencyDisplay
          v-else-if="column.field === 'amount_usd'"
          :amount="Number(row.amount_usd)"
          currency="USD"
        />
        <StatusBadge
          v-else-if="column.field === 'status' && (row as LoanAgreement).is_extended"
          status="extended"
        />
        <span v-else-if="column.field === 'status'">Normal</span>
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'loan-agreement-detail', params: { id: row.id } }"
            icon="pi pi-eye"
            label="Detail"
            size="small"
            outlined
          />
          <Button
            v-if="can('loan_agreement', 'update')"
            as="router-link"
            :to="{ name: 'loan-agreement-edit', params: { id: row.id } }"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            severity="secondary"
            outlined
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>
  </section>
</template>
