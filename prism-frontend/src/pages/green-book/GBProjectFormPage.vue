<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Tab from 'primevue/tab'
import TabList from 'primevue/tablist'
import TabPanel from 'primevue/tabpanel'
import TabPanels from 'primevue/tabpanels'
import Tabs from 'primevue/tabs'
import Textarea from 'primevue/textarea'
import ActivitiesTable from '@/components/green-book/ActivitiesTable.vue'
import DisbursementPlanTable from '@/components/green-book/DisbursementPlanTable.vue'
import FundingAllocationTable from '@/components/green-book/FundingAllocationTable.vue'
import FundingSourceTable from '@/components/green-book/FundingSourceTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import InstitutionSelect from '@/components/forms/InstitutionSelect.vue'
import LocationMultiSelect from '@/components/forms/LocationMultiSelect.vue'
import ProgramTitleSelect from '@/components/forms/ProgramTitleSelect.vue'
import { useGBProjectForm } from '@/composables/forms/useGBProjectForm'
import { useToast } from '@/composables/useToast'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useMasterStore } from '@/stores/master.store'

const route = useRoute()
const router = useRouter()
const greenBookStore = useGreenBookStore()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()

const greenBookId = computed(() => String(route.params.gbId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const isEditMode = computed(() => route.name === 'gb-project-edit')
const pageTitle = computed(() => (isEditMode.value ? 'Edit GB Project' : 'Tambah GB Project'))
const form = useGBProjectForm()

const bbProjectOptions = computed(() =>
  blueBookStore.projectOptions
    .filter((project) => project.is_latest !== false)
    .map((project) => ({
      ...project,
      label: `${project.bb_code} - ${project.project_name}`,
    })),
)

async function loadData() {
  await Promise.all([
    greenBookStore.fetchGreenBook(greenBookId.value),
    blueBookStore.fetchProjectOptions(),
    masterStore.fetchProgramTitles(true, { limit: 1000 }),
    masterStore.fetchInstitutions(true, { limit: 1000 }),
    masterStore.fetchRegions(true, { limit: 1000 }),
    masterStore.fetchLenders(true, { limit: 1000 }),
  ])

  if (isEditMode.value) {
    const project = await greenBookStore.fetchProject(greenBookId.value, projectId.value)
    form.applyProject(project)
  }
}

const onSubmit = form.submit(async (values) => {
  if (isEditMode.value) {
    await greenBookStore.updateProject(greenBookId.value, projectId.value, values)
    toast.success('Berhasil', 'GB Project berhasil diperbarui')
    await router.push({ name: 'gb-project-detail', params: { gbId: greenBookId.value, id: projectId.value } })
    return
  }

  const created = await greenBookStore.createProject(greenBookId.value, values)
  toast.success('Berhasil', 'GB Project berhasil dibuat')
  await router.push({ name: 'gb-project-detail', params: { gbId: greenBookId.value, id: created.id } })
})

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader :title="pageTitle" subtitle="Lengkapi data proyek Green Book">
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'green-book-detail', params: { id: greenBookId } })"
        />
      </template>
    </PageHeader>

    <form class="space-y-6" @submit.prevent="onSubmit">
      <Tabs value="0" class="rounded-lg border border-surface-200 bg-white p-2">
        <TabList>
          <Tab value="0">Informasi Umum</Tab>
          <Tab value="1">Kegiatan</Tab>
          <Tab value="2">Funding Source</Tab>
          <Tab value="3">Rencana Disbursement</Tab>
          <Tab value="4">Alokasi Funding</Tab>
        </TabList>

        <TabPanels>
          <TabPanel value="0">
            <div class="space-y-4 p-3">
              <div class="grid gap-4 md:grid-cols-2">
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Judul Program</span>
                  <ProgramTitleSelect v-model="form.values.program_title_id" />
                  <small v-if="form.errors.program_title_id" class="text-red-600">
                    {{ form.errors.program_title_id }}
                  </small>
                </label>
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">BB Project</span>
                  <MultiSelect
                    v-model="form.values.bb_project_ids"
                    :options="bbProjectOptions"
                    option-label="label"
                    option-value="id"
                    placeholder="Pilih BB Project"
                    filter
                    display="chip"
                    class="w-full"
                  />
                  <small v-if="form.errors.bb_project_ids" class="text-red-600">
                    {{ form.errors.bb_project_ids }}
                  </small>
                </label>
              </div>

              <div class="grid gap-4 md:grid-cols-2">
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Kode GB</span>
                  <InputText v-model="form.values.gb_code" class="w-full" :disabled="isEditMode" />
                  <small v-if="form.errors.gb_code" class="text-red-600">{{ form.errors.gb_code }}</small>
                </label>
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Nama Proyek</span>
                  <InputText v-model="form.values.project_name" class="w-full" />
                  <small v-if="form.errors.project_name" class="text-red-600">{{ form.errors.project_name }}</small>
                </label>
                <label class="block space-y-2 md:col-span-2">
                  <span class="text-sm font-medium text-surface-700">Durasi</span>
                  <InputText v-model="form.values.duration" class="w-full" placeholder="2025-2030" />
                </label>
              </div>

              <div class="grid gap-4 md:grid-cols-2">
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Tujuan</span>
                  <Textarea v-model="form.values.objective" auto-resize rows="3" class="w-full" />
                </label>
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Lingkup Proyek</span>
                  <Textarea v-model="form.values.scope_of_project" auto-resize rows="3" class="w-full" />
                </label>
              </div>

              <div class="grid gap-4 md:grid-cols-3">
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Executing Agency</span>
                  <InstitutionSelect v-model="form.values.executing_agency_ids" multiple />
                  <small v-if="form.errors.executing_agency_ids" class="text-red-600">
                    {{ form.errors.executing_agency_ids }}
                  </small>
                </label>
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Implementing Agency</span>
                  <InstitutionSelect v-model="form.values.implementing_agency_ids" multiple />
                  <small v-if="form.errors.implementing_agency_ids" class="text-red-600">
                    {{ form.errors.implementing_agency_ids }}
                  </small>
                </label>
                <label class="block space-y-2">
                  <span class="text-sm font-medium text-surface-700">Lokasi</span>
                  <LocationMultiSelect v-model="form.values.location_ids" />
                  <small v-if="form.errors.location_ids" class="text-red-600">{{ form.errors.location_ids }}</small>
                </label>
              </div>
            </div>
          </TabPanel>

          <TabPanel value="1">
            <div class="space-y-3 p-3">
              <ActivitiesTable
                v-model:rows="form.activities.value"
                @add="form.addActivity"
                @remove="form.removeActivity"
                @reorder="form.reorderActivities"
              />
            </div>
          </TabPanel>

          <TabPanel value="2">
            <div class="space-y-3 p-3">
              <FundingSourceTable
                v-model:rows="form.fundingSources.value"
                @add="form.addFundingSource"
                @remove="form.removeFundingSource"
              />
            </div>
          </TabPanel>

          <TabPanel value="3">
            <div class="space-y-3 p-3">
              <DisbursementPlanTable
                v-model:rows="form.disbursementPlan.value"
                :error="form.disbursementError.value"
                @add-year="form.addDisbursementYear"
                @update-year="form.updateDisbursementYear"
                @remove="form.removeDisbursementYear"
              />
            </div>
          </TabPanel>

          <TabPanel value="4">
            <div class="space-y-3 p-3">
              <div class="rounded-lg border border-primary/20 bg-primary/5 p-3 text-sm text-surface-700">
                Alokasi funding selalu mengikuti jumlah dan urutan kegiatan. Saat kegiatan dihapus,
                baris alokasi ikut berkurang otomatis.
              </div>
              <FundingAllocationTable
                :activities="form.activities.value"
                v-model:rows="form.allocationValues.value"
              />
            </div>
          </TabPanel>
        </TabPanels>
      </Tabs>

      <div class="sticky bottom-0 flex justify-end gap-2 border-t border-surface-200 bg-surface-50/95 py-4 backdrop-blur">
        <Button
          label="Batal"
          severity="secondary"
          outlined
          @click="router.push({ name: 'green-book-detail', params: { id: greenBookId } })"
        />
        <Button type="submit" label="Simpan" icon="pi pi-save" :loading="greenBookStore.loading" />
      </div>
    </form>
  </section>
</template>
