<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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
import StatusBadge from '@/components/common/StatusBadge.vue'
import ValueChipList from '@/components/common/ValueChipList.vue'
import { usePermission } from '@/composables/usePermission'
import { useGreenBookStore } from '@/stores/green-book.store'
import type { BBProjectSummary, GBProjectHistoryItem } from '@/types/green-book.types'
import { formatDateTime } from '@/utils/formatters'
import { formatGreenBookStatus } from './green-book-page-utils'

const route = useRoute()
const router = useRouter()
const greenBookStore = useGreenBookStore()
const { can } = usePermission()

const greenBookId = computed(() => String(route.params.gbId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const isRevisionHistoryOpen = ref(false)
const project = computed(() => greenBookStore.currentProject)
const executingAgencyNames = computed(() => toNameList(project.value?.executing_agencies))
const implementingAgencyNames = computed(() => toNameList(project.value?.implementing_agencies))
const locationNames = computed(() => toNameList(project.value?.locations))
const bappenasPartnerNames = computed(() => toNameList(project.value?.bappenas_partners))
const selectedCurrency = computed(() => project.value?.funding_sources[0]?.currency ?? 'USD')
const programTitleName = computed(
  () => project.value?.program_title?.title ?? '-',
)
const auditRailItems = computed(() =>
  greenBookStore.projectHistory.flatMap((item) =>
    (item.audit_entries ?? []).map((entry) => ({
      ...entry,
      snapshot_label: item.book_label,
    })),
  ),
)
const hasAuditRail = computed(() => auditRailItems.value.length > 0)

function toNameList(items?: { name?: string; title?: string }[]) {
  return items?.map((item) => item.name ?? item.title).filter((item): item is string => Boolean(item)) ?? []
}

async function loadData() {
  await Promise.all([
    greenBookStore.fetchProject(greenBookId.value, projectId.value),
    greenBookStore.fetchProjectHistory(projectId.value),
  ])
}

function historyRoute(item: GBProjectHistoryItem) {
  return { name: 'gb-project-detail', params: { gbId: item.green_book_id, id: item.id } }
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
      <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
        <div class="grid gap-5 p-5 md:grid-cols-3">
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Status</p>
            <div class="mt-1 flex flex-wrap items-center gap-2">
              <StatusBadge :status="project.status" :label="formatGreenBookStatus(project.status)" />
              <Tag v-if="project.is_latest" value="Terbaru" severity="success" rounded />
              <Tag
                v-else-if="project.has_newer_revision"
                value="Ada revisi lebih baru"
                severity="warn"
                rounded
              />
            </div>
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Judul Program
            </p>
            <p class="text-sm font-semibold text-surface-950">{{ programTitleName }}</p>
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Durasi</p>
            <p class="text-sm font-semibold text-surface-950">
              {{ project.duration ? `${project.duration} bulan` : '-' }}
            </p>
          </div>
        </div>
        <div class="border-t border-surface-100 px-5 py-4">
          <h2 class="text-lg font-semibold text-surface-950">Profil Kelembagaan</h2>
        </div>
        <div class="grid gap-x-8 gap-y-5 px-5 pb-5 md:grid-cols-2">
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Executing Agency
            </p>
            <ValueChipList :items="executingAgencyNames" />
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Implementing Agency
            </p>
            <ValueChipList :items="implementingAgencyNames" />
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Lokasi</p>
            <ValueChipList :items="locationNames" />
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Mitra Kerja Bappenas
            </p>
            <ValueChipList :items="bappenasPartnerNames" />
          </div>
        </div>
      </section>

      <div class="rounded-lg border border-surface-200 bg-white p-5">
        <p class="text-xs uppercase tracking-wide text-surface-500">Referensi Proyek Blue Book</p>
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

      <section class="space-y-3 rounded-lg border border-surface-200 bg-white p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="text-lg font-semibold text-surface-950">Histori Revisi</h2>
            <Tag
              :value="`${greenBookStore.projectHistory.length} snapshot`"
              severity="secondary"
              rounded
            />
          </div>
          <Button
            :label="isRevisionHistoryOpen ? 'Tutup' : 'Detail'"
            :icon="isRevisionHistoryOpen ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
            severity="secondary"
            size="small"
            outlined
            @click="isRevisionHistoryOpen = !isRevisionHistoryOpen"
          />
        </div>
        <div
          v-if="isRevisionHistoryOpen"
          class="overflow-auto rounded-lg border border-surface-200"
        >
          <table class="w-full min-w-[60rem] text-left text-sm">
            <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
              <tr>
                <th class="px-4 py-3">Green Book</th>
                <th class="px-4 py-3">Kode</th>
                <th class="px-4 py-3">Status Dokumen</th>
                <th class="px-4 py-3">Snapshot</th>
                <th class="px-4 py-3">Referensi Blue Book</th>
                <th class="px-4 py-3">Downstream</th>
                <th v-if="hasAuditRail" class="px-4 py-3">Perubahan Terakhir</th>
                <th class="px-4 py-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-surface-100">
              <tr v-for="item in greenBookStore.projectHistory" :key="item.id">
                <td class="px-4 py-3 font-medium text-surface-900">{{ item.book_label }}</td>
                <td class="px-4 py-3 text-surface-700">{{ item.gb_code }}</td>
                <td class="px-4 py-3">
                  <StatusBadge
                    :status="item.book_status"
                    :label="formatGreenBookStatus(item.book_status)"
                  />
                </td>
                <td class="px-4 py-3">
                  <Tag
                    :value="item.is_latest ? 'Terbaru' : 'Historis'"
                    :severity="item.is_latest ? 'success' : 'secondary'"
                    rounded
                  />
                </td>
                <td class="px-4 py-3 text-surface-700">
                  {{ item.bb_projects?.map((bbProject) => bbProject.bb_code).join(', ') || '-' }}
                </td>
                <td class="px-4 py-3">
                  <Tag
                    :value="item.used_by_downstream ? 'Dipakai tahap lanjutan' : 'Belum dipakai'"
                    :severity="item.used_by_downstream ? 'info' : 'secondary'"
                    rounded
                  />
                </td>
                <td v-if="hasAuditRail" class="px-4 py-3 text-surface-700">
                  <div v-if="item.last_change_summary">
                    <p class="font-medium text-surface-900">{{ item.last_change_summary }}</p>
                    <p class="text-xs text-surface-500">
                      {{ item.last_changed_by }} - {{ formatDateTime(item.last_changed_at) }}
                    </p>
                  </div>
                  <span v-else>-</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <Button
                    as="router-link"
                    :to="historyRoute(item)"
                    icon="pi pi-eye"
                    severity="secondary"
                    size="small"
                    outlined
                    rounded
                    aria-label="Lihat snapshot"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

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

      <ProjectAuditRail :items="auditRailItems" />
    </div>
  </section>
</template>
