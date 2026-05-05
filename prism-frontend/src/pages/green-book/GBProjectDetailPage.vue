<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import Tab from 'primevue/tab'
import TabList from 'primevue/tablist'
import TabPanel from 'primevue/tabpanel'
import TabPanels from 'primevue/tabpanels'
import Tabs from 'primevue/tabs'
import ActivitiesTable from '@/components/green-book/ActivitiesTable.vue'
import DisbursementPlanTable from '@/components/green-book/DisbursementPlanTable.vue'
import FundingAllocationTable from '@/components/green-book/FundingAllocationTable.vue'
import FundingSourceTable from '@/components/green-book/FundingSourceTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import ProjectAuditRail from '@/components/common/ProjectAuditRail.vue'
import ProjectInstitutionGrid from '@/components/common/ProjectInstitutionGrid.vue'
import ProjectRevisionHistory from '@/components/common/ProjectRevisionHistory.vue'
import type { RevisionHistoryItem } from '@/components/common/ProjectRevisionHistory.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { usePermission } from '@/composables/usePermission'
import { useGreenBookStore } from '@/stores/green-book.store'
import type { BBProjectSummary, GBProjectHistoryItem } from '@/types/green-book.types'
import { toNameList } from '@/utils/formatters'
import { formatGreenBookStatus } from './green-book-page-utils'

const route = useRoute()
const router = useRouter()
const greenBookStore = useGreenBookStore()
const { can } = usePermission()

const greenBookId = computed(() => String(route.params.gbId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const project = computed(() => greenBookStore.currentProject)
const executingAgencyNames = computed(() => toNameList(project.value?.executing_agencies))
const implementingAgencyNames = computed(() => toNameList(project.value?.implementing_agencies))
const locationNames = computed(() => toNameList(project.value?.locations))
const bappenasPartnerNames = computed(() => toNameList(project.value?.bappenas_partners))
const selectedCurrency = computed(() => project.value?.funding_sources[0]?.currency ?? 'USD')
const programTitleName = computed(() => project.value?.program_title?.title ?? '-')

const revisionHistoryItems = computed<RevisionHistoryItem[]>(() =>
  greenBookStore.projectHistory.map((item) => ({
    id: item.id,
    label: item.book_label,
    code: item.gb_code,
    book_status: item.book_status,
    status_label: formatGreenBookStatus(item.book_status),
    is_latest: item.is_latest,
    route: { name: 'gb-project-detail', params: { gbId: item.green_book_id, id: item.id } },
  })),
)

const auditRailItems = computed(() =>
  greenBookStore.projectHistory.flatMap((item) =>
    (item.audit_entries ?? []).map((entry) => ({
      ...entry,
      snapshot_label: item.book_label,
    })),
  ),
)

async function loadData() {
  await Promise.all([
    greenBookStore.fetchProject(greenBookId.value, projectId.value),
    greenBookStore.fetchProjectHistory(projectId.value),
  ])
}

function bbProjectRoute(bbProject: BBProjectSummary) {
  const bbId = bbProject.blue_book_id
  return bbId
    ? { name: 'bb-project-detail', params: { bbId, id: bbProject.id } }
    : { name: 'blue-books' }
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="project?.gb_code ?? 'Detail Proyek Green Book'"
      :subtitle="project?.project_name"
    >
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'green-book-detail', params: { id: greenBookId } })"
        />
        <Button
          v-if="can('gb_project', 'update')"
          as="router-link"
          :to="{ name: 'gb-project-edit', params: { gbId: greenBookId, id: projectId } }"
          label="Edit"
          icon="pi pi-pencil"
          outlined
        />
      </template>
    </PageHeader>

    <div v-if="project" class="space-y-6">
      <!-- Top card: Judul Program + Durasi + Status -->
      <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
        <div class="p-5">
          <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
            <div class="min-w-0 space-y-3">
              <div class="space-y-1.5">
                <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Judul Program</p>
                <p class="truncate text-base font-semibold text-surface-950">{{ programTitleName }}</p>
              </div>
              <div class="space-y-1.5">
                <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Durasi</p>
                <p class="text-sm font-semibold text-surface-950">
                  {{ project.duration ? `${project.duration} bulan` : '-' }}
                </p>
              </div>
            </div>
            <div class="flex flex-col gap-2 lg:items-end">
              <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Status Snapshot</p>
              <div class="flex flex-wrap items-center gap-2 lg:justify-end">
                <StatusBadge :status="project.status" :label="formatGreenBookStatus(project.status)" />
                <Tag v-if="project.is_latest" value="Versi terbaru" severity="success" rounded />
                <Tag
                  v-else-if="project.has_newer_revision"
                  value="Ada revisi lebih baru"
                  severity="warn"
                  rounded
                />
              </div>
            </div>
          </div>
        </div>
        <!-- Institusi & Lokasi (reusable grid) -->
        <ProjectInstitutionGrid
          :executing-agencies="executingAgencyNames"
          :implementing-agencies="implementingAgencyNames"
          :bappenas-partners="bappenasPartnerNames"
          :locations="locationNames"
          class="border-t border-surface-100"
        />
      </div>

      <!-- Referensi Proyek Blue Book -->
      <div class="rounded-lg border border-surface-200 bg-white p-5">
        <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Referensi Proyek Blue Book</p>
        <div class="mt-3 flex flex-wrap gap-2">
          <RouterLink
            v-for="bbProject in project.bb_projects"
            :key="bbProject.id"
            :to="bbProjectRoute(bbProject)"
            class="inline-flex items-center gap-2 rounded-full border border-surface-200 px-3 py-1.5 text-sm font-medium text-primary"
          >
            <span>{{ bbProject.bb_code }} - {{ bbProject.project_name }}</span>
            <Tag
              v-if="bbProject.has_newer_revision"
              value="Ada revisi lebih baru"
              severity="warn"
              rounded
            />
            <Tag v-else-if="bbProject.is_latest" value="Terbaru" severity="success" rounded />
          </RouterLink>
        </div>
      </div>

      <!-- Tabs: Kegiatan / Funding / Disbursement / Alokasi -->
      <Tabs value="0" class="rounded-lg border border-surface-200 bg-white p-2">
        <TabList>
          <Tab value="0">Kegiatan</Tab>
          <Tab value="1">Funding Source</Tab>
          <Tab value="2">Rencana Disbursement</Tab>
          <Tab value="3">Alokasi Funding</Tab>
        </TabList>
        <TabPanels>
          <TabPanel value="0">
            <div class="p-3">
              <ActivitiesTable :rows="project.activities" :editable="false" />
            </div>
          </TabPanel>
          <TabPanel value="1">
            <div class="p-3">
              <FundingSourceTable
                :rows="project.funding_sources"
                :selected-currency="selectedCurrency"
                :editable="false"
              />
            </div>
          </TabPanel>
          <TabPanel value="2">
            <div class="p-3">
              <DisbursementPlanTable
                :rows="project.disbursement_plan"
                :selected-currency="selectedCurrency"
                :editable="false"
              />
            </div>
          </TabPanel>
          <TabPanel value="3">
            <div class="p-3">
              <FundingAllocationTable
                :activities="project.activities"
                :rows="project.funding_allocations"
                :selected-currency="selectedCurrency"
                :editable="false"
              />
            </div>
          </TabPanel>
        </TabPanels>
      </Tabs>

      <ProjectRevisionHistory :items="revisionHistoryItems" />

      <ProjectAuditRail :items="auditRailItems" />
    </div>
  </section>
</template>
