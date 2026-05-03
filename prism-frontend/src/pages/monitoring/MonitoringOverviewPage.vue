<script setup lang="ts">
import { onMounted, watch } from 'vue'
import Button from 'primevue/button'
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
  MonitoringLoanAgreementListParams,
  MonitoringLoanAgreementReference,
} from '@/types/monitoring.types'
import { formatDate } from '@/pages/loan-agreement/loan-agreement-page-utils'

const monitoringStore = useMonitoringStore()
const { can } = usePermission()

interface MonitoringOverviewFilterState {
  is_effective: boolean | null
}

const listControls = useListControls<MonitoringOverviewFilterState>({
  initialFilters: {
    is_effective: null,
  },
  filterLabels: {
    is_effective: 'Status Efektif',
  },
  formatFilterValue: (key, value) => {
    if (key === 'is_effective') {
      return value ? 'Efektif' : 'Belum efektif'
    }
    return String(value)
  },
})

const effectiveStatusOptions = [
  { label: 'Semua', value: null },
  { label: 'Efektif', value: true },
  { label: 'Belum efektif', value: false },
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
  return listControls.buildParams() as MonitoringLoanAgreementListParams
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

watch(
  [
    listControls.page,
    listControls.limit,
    listControls.debouncedSearch,
    () => JSON.stringify(listControls.appliedFilters),
  ],
  () => {
    void loadData()
  },
)

onMounted(() => {
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
