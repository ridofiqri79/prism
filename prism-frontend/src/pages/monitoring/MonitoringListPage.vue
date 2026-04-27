<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Message from 'primevue/message'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import AbsorptionBar from '@/components/monitoring/AbsorptionBar.vue'
import MonitoringCard from '@/components/monitoring/MonitoringCard.vue'
import MonitoringChart from '@/components/monitoring/MonitoringChart.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePagination } from '@/composables/usePagination'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { useMonitoringStore } from '@/stores/monitoring.store'
import type { MonitoringDisbursement } from '@/types/monitoring.types'
import { formatDate } from '@/pages/loan-agreement/loan-agreement-page-utils'

const route = useRoute()
const router = useRouter()
const monitoringStore = useMonitoringStore()
const pagination = usePagination()
const { can } = usePermission()
const confirm = useConfirm()
const toast = useToast()

const loanAgreementId = computed(() => String(route.params.laId ?? ''))
const currentLA = computed(() => monitoringStore.currentLA)
const todayString = computed(() => new Date().toISOString().slice(0, 10))
const isNotEffective = computed(() => {
  if (!currentLA.value?.effective_date) return false
  return currentLA.value.effective_date.slice(0, 10) > todayString.value
})
const columns: ColumnDef[] = [
  { field: 'budget_year', header: 'Budget Year' },
  { field: 'quarter', header: 'Quarter' },
  { field: 'exchange_rate_usd_idr', header: 'USD/IDR' },
  { field: 'exchange_rate_la_idr', header: 'LA/IDR' },
  { field: 'planned_usd', header: 'Planned USD' },
  { field: 'realized_usd', header: 'Realized USD' },
  { field: 'absorption_pct', header: 'Absorption' },
  { field: 'actions', header: 'Actions' },
]

async function loadData() {
  await Promise.all([
    monitoringStore.fetchLoanAgreement(loanAgreementId.value),
    monitoringStore.fetchMonitorings(loanAgreementId.value, {
      page: pagination.page.value,
      limit: pagination.limit.value,
    }),
  ])
}

function deleteMonitoring(row: MonitoringDisbursement) {
  confirm.confirmDelete(`monitoring ${row.quarter} ${row.budget_year}`, async () => {
    await monitoringStore.deleteMonitoring(loanAgreementId.value, row.id)
    toast.success('Berhasil', 'Monitoring berhasil dihapus')
  })
}

watch(
  () => [pagination.page.value, pagination.limit.value],
  () => {
    void monitoringStore.fetchMonitorings(loanAgreementId.value, {
      page: pagination.page.value,
      limit: pagination.limit.value,
    })
  },
)

onMounted(() => {
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
          label="Kembali ke LA"
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
      LA belum efektif - monitoring belum bisa diinput. Effective Date:
      {{ formatDate(currentLA?.effective_date ?? '') }}
    </Message>

    <section v-if="currentLA" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-4">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Loan Code</p>
        <p class="font-semibold text-surface-950">{{ currentLA.loan_code }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Lender</p>
        <p class="font-semibold text-surface-950">{{ currentLA.lender.name }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Currency</p>
        <p class="font-semibold text-surface-950">{{ currentLA.currency }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Effective Date</p>
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

    <DataTable
      v-model:page="pagination.page.value"
      v-model:limit="pagination.limit.value"
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
