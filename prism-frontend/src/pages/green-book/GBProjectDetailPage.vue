<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
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
import StatusBadge from '@/components/common/StatusBadge.vue'
import { usePermission } from '@/composables/usePermission'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useMasterStore } from '@/stores/master.store'
import { joinNames } from './green-book-page-utils'

const route = useRoute()
const router = useRouter()
const greenBookStore = useGreenBookStore()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const { can } = usePermission()

const greenBookId = computed(() => String(route.params.gbId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const project = computed(() => greenBookStore.currentProject)
const programTitleName = computed(
  () =>
    project.value?.program_title?.title ??
    masterStore.programTitles.find((item) => item.id === project.value?.program_title_id)?.title ??
    '-',
)

function bbProjectBlueBookId(id: string) {
  return blueBookStore.projectOptions.find((item) => item.id === id)?.blue_book_id
}

async function loadData() {
  await Promise.all([
    greenBookStore.fetchProject(greenBookId.value, projectId.value),
    blueBookStore.fetchProjectOptions(),
    masterStore.fetchProgramTitles(true, { limit: 1000 }),
  ])
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader :title="project?.gb_code ?? 'GB Project Detail'" :subtitle="project?.project_name">
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
      <div class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-3">
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Status</p>
          <StatusBadge :status="project.status" />
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Program Title</p>
          <p class="font-semibold text-surface-950">{{ programTitleName }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Duration</p>
          <p class="font-semibold text-surface-950">{{ project.duration || '-' }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Executing Agency</p>
          <p class="font-semibold text-surface-950">{{ joinNames(project.executing_agencies) }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Implementing Agency</p>
          <p class="font-semibold text-surface-950">{{ joinNames(project.implementing_agencies) }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Lokasi</p>
          <p class="font-semibold text-surface-950">{{ joinNames(project.locations) }}</p>
        </div>
      </div>

      <div class="rounded-lg border border-surface-200 bg-white p-5">
        <p class="text-xs uppercase tracking-wide text-surface-500">Referensi BB Projects</p>
        <div class="mt-3 flex flex-wrap gap-2">
          <RouterLink
            v-for="bbProject in project.bb_projects"
            :key="bbProject.id"
            :to="
              bbProjectBlueBookId(bbProject.id)
                ? { name: 'bb-project-detail', params: { bbId: bbProjectBlueBookId(bbProject.id), id: bbProject.id } }
                : { name: 'blue-books' }
            "
            class="rounded-full border border-surface-200 px-3 py-1.5 text-sm font-medium text-primary"
          >
            {{ bbProject.bb_code }} - {{ bbProject.project_name }}
          </RouterLink>
        </div>
      </div>

      <Tabs value="0" class="rounded-lg border border-surface-200 bg-white p-2">
        <TabList>
          <Tab value="0">Activities</Tab>
          <Tab value="1">Funding Source</Tab>
          <Tab value="2">Disbursement Plan</Tab>
          <Tab value="3">Funding Allocation</Tab>
        </TabList>
        <TabPanels>
          <TabPanel value="0">
            <div class="p-3">
              <ActivitiesTable :rows="project.activities" :editable="false" />
            </div>
          </TabPanel>
          <TabPanel value="1">
            <div class="p-3">
              <FundingSourceTable :rows="project.funding_sources" :editable="false" />
            </div>
          </TabPanel>
          <TabPanel value="2">
            <div class="p-3">
              <DisbursementPlanTable :rows="project.disbursement_plan" :editable="false" />
            </div>
          </TabPanel>
          <TabPanel value="3">
            <div class="p-3">
              <FundingAllocationTable
                :activities="project.activities"
                :rows="project.funding_allocations"
                :editable="false"
              />
            </div>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </div>
  </section>
</template>

