<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import Select from 'primevue/select'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import { usePagination } from '@/composables/usePagination'
import { usePermission } from '@/composables/usePermission'
import { useLoanAgreementStore } from '@/stores/loan-agreement.store'
import { useMasterStore } from '@/stores/master.store'
import type { LoanAgreement } from '@/types/loan-agreement.types'
import { formatDate, toDateString } from './loan-agreement-page-utils'

const loanAgreementStore = useLoanAgreementStore()
const masterStore = useMasterStore()
const { can } = usePermission()
const pagination = usePagination()

const lenderId = ref<string | null>(null)
const isExtended = ref<boolean | null>(null)
const closingDateBefore = ref<Date | null>(null)
const isExtendedOptions = [
  { label: 'Semua', value: null },
  { label: 'Diperpanjang', value: true },
  { label: 'Tidak diperpanjang', value: false },
]
const columns: ColumnDef[] = [
  { field: 'loan_code', header: 'Kode Loan' },
  { field: 'lender', header: 'Lender' },
  { field: 'effective_date', header: 'Tanggal Efektif' },
  { field: 'closing_date', header: 'Tanggal Closing' },
  { field: 'currency', header: 'Mata Uang' },
  { field: 'amount_usd', header: 'Nilai USD' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Aksi' },
]

const closingDateBeforeString = computed(() => toDateString(closingDateBefore.value))
const filteredRows = computed(() => {
  return loanAgreementStore.loanAgreements.filter((item) => {
    if (lenderId.value && item.lender.id !== lenderId.value) return false
    if (isExtended.value !== null && item.is_extended !== isExtended.value) return false
    if (closingDateBeforeString.value && item.closing_date > closingDateBeforeString.value) return false
    return true
  })
})

async function loadData() {
  await loanAgreementStore.fetchLoanAgreements({
    page: pagination.page.value,
    limit: 1000,
    lender_id: lenderId.value || undefined,
    is_extended: isExtended.value ?? undefined,
    closing_date_before: closingDateBeforeString.value || undefined,
  })
}

function resetFilters() {
  lenderId.value = null
  isExtended.value = null
  closingDateBefore.value = null
}

watch(
  [lenderId, isExtended, closingDateBefore],
  () => {
    pagination.resetPage()
    void loadData()
  },
  { deep: true },
)

onMounted(() => {
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
          label="Buat LA"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <section class="rounded-lg border border-surface-200 bg-white p-4">
      <div class="grid gap-4 md:grid-cols-4">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <LenderSelect v-model="lenderId" placeholder="Semua lender" />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Status Perpanjangan</span>
          <Select
            v-model="isExtended"
            :options="isExtendedOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Closing Sebelum Tanggal</span>
          <DatePicker v-model="closingDateBefore" date-format="yy-mm-dd" show-icon class="w-full" />
        </label>
        <div class="flex items-end gap-2">
          <Button
            :label="isExtended === true ? 'Hanya diperpanjang aktif' : 'Hanya diperpanjang'"
            :severity="isExtended === true ? 'warn' : 'secondary'"
            outlined
            @click="isExtended = isExtended === true ? null : true"
          />
          <Button label="Reset" severity="secondary" outlined @click="resetFilters" />
        </div>
      </div>
    </section>

    <DataTable
      v-model:page="pagination.page.value"
      v-model:limit="pagination.limit.value"
      :data="filteredRows as unknown as Record<string, unknown>[]"
      :columns="columns"
      :loading="loanAgreementStore.loading"
      :total="filteredRows.length"
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
