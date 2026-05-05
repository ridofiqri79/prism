<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import InputNumber from 'primevue/inputnumber'
import Message from 'primevue/message'
import Select from 'primevue/select'
import AuditSummaryTable from '@/components/dashboard/AuditSummaryTable.vue'
import DashboardFilterBar from '@/components/dashboard/DashboardFilterBar.vue'
import DataQualityIssueTable from '@/components/dashboard/DataQualityIssueTable.vue'
import MetricCard from '@/components/dashboard/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useAuthStore } from '@/stores/auth.store'
import { useDashboardStore } from '@/stores/dashboard.store'
import type {
  DataQualityGovernanceParams,
  DataQualityIssueItem,
  DataQualitySeverity,
  MetricCard as MetricCardType,
} from '@/types/dashboard.types'

const dashboard = useDashboardStore()
const auth = useAuthStore()
const router = useRouter()

const props = withDefaults(
  defineProps<{
    embedded?: boolean
  }>(),
  {
    embedded: false,
  },
)

const filters = reactive<{
  severity: DataQualitySeverity | null
  module: string | null
  issue_type: string | null
  only_unresolved: boolean
  audit_days: number
}>({
  severity: null,
  module: null,
  issue_type: null,
  only_unresolved: true,
  audit_days: 30,
})

const activePanel = ref<'issues' | 'audit'>('issues')

const severityOptions = [
  { label: 'Semua severity', value: null },
  { label: 'Error', value: 'error' },
  { label: 'Warning', value: 'warning' },
  { label: 'Info', value: 'info' },
]

const moduleOptions = [
  { label: 'Semua module', value: null },
  { label: 'Blue Book Project', value: 'bb_project' },
  { label: 'Green Book Project', value: 'gb_project' },
  { label: 'Daftar Kegiatan Project', value: 'dk_project' },
  { label: 'Loan Agreement', value: 'loan_agreement' },
]

const issueTypeOptions = [
  { label: 'Semua issue type', value: null },
  { label: 'BB without Bappenas partner', value: 'BB_WITHOUT_BAPPENAS_PARTNER' },
  { label: 'BB indication without Letter of Intent', value: 'BB_INDICATION_WITHOUT_LOI' },
  { label: 'Letter of Intent without Green Book', value: 'LOI_WITHOUT_GB' },
  { label: 'Green Book without Blue Book reference', value: 'GB_WITHOUT_BB_REFERENCE' },
  { label: 'Green Book without funding source', value: 'GB_WITHOUT_FUNDING_SOURCE' },
  { label: 'Green Book without disbursement plan', value: 'GB_WITHOUT_DISBURSEMENT_PLAN' },
  { label: 'Green Book without activity', value: 'GB_WITHOUT_ACTIVITY' },
  { label: 'Daftar Kegiatan without financing detail', value: 'DK_WITHOUT_FINANCING_DETAIL' },
  { label: 'Daftar Kegiatan without activity detail', value: 'DK_WITHOUT_ACTIVITY_DETAIL' },
  { label: 'Daftar Kegiatan without Loan Agreement', value: 'DK_WITHOUT_LA' },
  { label: 'Loan Agreement not effective', value: 'LA_NOT_EFFECTIVE' },
  { label: 'Currency USD mismatch', value: 'CURRENCY_USD_MISMATCH' },
]

const isAdmin = computed(() => auth.user?.role === 'ADMIN')
const governance = computed(() => dashboard.dataQualityGovernance)
const summary = computed(() => governance.value?.summary)
const auditSummary = computed(() => governance.value?.audit_summary)
const visibleIssues = computed(() => governance.value?.issues ?? [])
const visibleIssueSummary = computed(() => ({
  total_issues: visibleIssues.value.length,
  error_count: visibleIssues.value.filter((item) => item.severity === 'error').length,
  warning_count: visibleIssues.value.filter((item) => item.severity === 'warning').length,
  info_count: visibleIssues.value.filter((item) => item.severity === 'info').length,
  audit_events: summary.value?.audit_events ?? 0,
}))

const cards = computed<MetricCardType[]>(() => {
  const baseCards: MetricCardType[] = [
    {
      key: 'total_issues',
      label: 'Total Issue',
      value: visibleIssueSummary.value.total_issues,
      unit: 'issue',
      category: 'quality',
    },
    {
      key: 'error_count',
      label: 'Error',
      value: visibleIssueSummary.value.error_count,
      unit: 'issue',
      category: 'critical',
    },
    {
      key: 'warning_count',
      label: 'Warning',
      value: visibleIssueSummary.value.warning_count,
      unit: 'issue',
      category: 'watch',
    },
    {
      key: 'info_count',
      label: 'Info',
      value: visibleIssueSummary.value.info_count,
      unit: 'issue',
      category: 'info',
    },
  ]

  if (isAdmin.value) {
    baseCards.push({
      key: 'audit_events',
      label: 'Audit Events',
      value: visibleIssueSummary.value.audit_events,
      unit: 'event',
      category: 'admin',
    })
  }

  return baseCards
})

function buildParams(): DataQualityGovernanceParams {
  return {
    severity: filters.severity ?? undefined,
    module: filters.module ?? undefined,
    issue_type: filters.issue_type ?? undefined,
    only_unresolved: filters.only_unresolved,
    audit_days: filters.audit_days,
  }
}

async function loadDashboard() {
  await dashboard.fetchDataQualityGovernance(buildParams())
  if (!isAdmin.value) {
    activePanel.value = 'issues'
  }
}

function clearFilters() {
  filters.severity = null
  filters.module = null
  filters.issue_type = null
  filters.only_unresolved = true
  filters.audit_days = 30
  void loadDashboard()
}

function openIssue(item: DataQualityIssueItem) {
  if (item.module === 'loan_agreement' && router.hasRoute('loan-agreement-detail')) {
    void router.push({ name: 'loan-agreement-detail', params: { id: item.record_id } })
    return
  }
  if (
    item.issue_type === 'BB_WITHOUT_BAPPENAS_PARTNER' &&
    router.hasRoute('project-journey')
  ) {
    void router.push({ name: 'project-journey', params: { bbProjectId: item.record_id } })
  }
}

onMounted(() => {
  void loadDashboard()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      v-if="!props.embedded"
      title="Data Quality & Governance"
      subtitle="Kontrol kelengkapan data, konsistensi business rule, integritas relasi, dan ringkasan audit ADMIN."
    />

    <DashboardFilterBar
      :loading="dashboard.dataQualityGovernanceLoading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[12rem_14rem_1fr_11rem_10rem]"
      @apply="loadDashboard"
      @reset="clearFilters"
    >
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Severity</span>
          <Select
            v-model="filters.severity"
            :options="severityOptions"
            option-label="label"
            option-value="value"
            class="w-full"
            placeholder="Semua"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Module</span>
          <Select
            v-model="filters.module"
            :options="moduleOptions"
            option-label="label"
            option-value="value"
            class="w-full"
            placeholder="Semua module"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Issue Type</span>
          <Select
            v-model="filters.issue_type"
            :options="issueTypeOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua issue"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Audit Days</span>
          <InputNumber v-model="filters.audit_days" :min="1" :max="365" class="w-full" />
        </label>

        <label class="flex items-end gap-2 pb-2">
          <Checkbox v-model="filters.only_unresolved" binary input-id="only-unresolved" />
          <span class="text-sm font-medium text-surface-700">Open only</span>
        </label>
    </DashboardFilterBar>

    <Message v-if="dashboard.dataQualityGovernanceError" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.dataQualityGovernanceError }}
    </Message>

    <section class="grid gap-4 md:grid-cols-2" :class="isAdmin ? 'xl:grid-cols-5' : 'xl:grid-cols-4'">
      <MetricCard v-for="card in cards" :key="card.key" :card="card" />
    </section>

    <section class="flex flex-wrap gap-2">
      <Button
        label="Issues"
        icon="pi pi-list-check"
        :outlined="activePanel !== 'issues'"
        @click="activePanel = 'issues'"
      />
      <Button
        v-if="isAdmin"
        label="Audit Summary"
        icon="pi pi-shield"
        :outlined="activePanel !== 'audit'"
        @click="activePanel = 'audit'"
      />
    </section>

    <DataQualityIssueTable
      v-if="activePanel === 'issues'"
      :items="visibleIssues"
      :loading="dashboard.dataQualityGovernanceLoading"
      @open="openIssue"
    />

    <AuditSummaryTable
      v-if="isAdmin && activePanel === 'audit'"
      :by-user="auditSummary?.by_user ?? []"
      :by-table="auditSummary?.by_table ?? []"
      :recent-activity="auditSummary?.recent_activity ?? []"
      :loading="dashboard.dataQualityGovernanceLoading"
    />
  </section>
</template>
