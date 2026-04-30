<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import MultiSelect from 'primevue/multiselect'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { greenBookSchema } from '@/schemas/green-book.schema'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useMasterStore } from '@/stores/master.store'
import type { GBProject, GBProjectListParams, GBProjectStatus, GreenBookPayload } from '@/types/green-book.types'
import type { Institution, Region } from '@/types/master.types'
import { formatGBRevision, joinNames, toFormErrors, type FormErrors } from './green-book-page-utils'

type GreenBookField = keyof GreenBookPayload

const route = useRoute()
const router = useRouter()
const greenBookStore = useGreenBookStore()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
interface GBProjectFilterState {
  bb_project_ids: string[]
  executing_agency_ids: string[]
  location_ids: string[]
  status: GBProjectStatus[]
}

const greenBookId = computed(() => String(route.params.id ?? ''))
const projectControls = useListControls<GBProjectFilterState>({
  initialFilters: {
    bb_project_ids: [],
    executing_agency_ids: [],
    location_ids: [],
    status: [],
  },
  filterLabels: {
    bb_project_ids: 'Proyek Blue Book',
    executing_agency_ids: 'Executing Agency',
    location_ids: 'Location',
    status: 'Status',
  },
  formatFilterValue: (key, value) => {
    if (key === 'bb_project_ids' && Array.isArray(value)) return selectedBBProjectSummary(value)
    if (key === 'executing_agency_ids' && Array.isArray(value)) return selectedInstitutionSummary(value)
    if (key === 'location_ids' && Array.isArray(value)) return selectedRegionSummary(value)
    return Array.isArray(value) ? selectedLabelSummary(value) : String(value)
  },
})
const dialogVisible = ref(false)
const form = reactive<GreenBookPayload>({
  publish_year: new Date().getFullYear(),
  revision_number: 0,
})
const errors = ref<FormErrors<GreenBookField>>({})
const columns: ColumnDef[] = [
  { field: 'gb_code', header: 'Kode Green Book' },
  { field: 'project_name', header: 'Nama Proyek' },
  { field: 'bb_projects', header: 'Proyek Blue Book' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Aksi' },
]
const projectStatusOptions: GBProjectStatus[] = ['active', 'deleted']
const selectedCountryCodes = computed(() => {
  const selected = new Set(projectControls.draftFilters.location_ids)

  return masterStore.regions
    .filter((region) => region.type === 'COUNTRY' && selected.has(region.id))
    .map((region) => region.code)
})
const bbProjectOptions = computed(() =>
  blueBookStore.projectOptions.map((project) => ({
    ...project,
    label: `${project.bb_code} - ${project.project_name}`,
    value: project.id,
  })),
)
const institutionFilterOptions = computed(() =>
  masterStore.institutions.map((institution) => ({
    ...institution,
    label: formatInstitution(institution),
    value: institution.id,
  })),
)
const locationFilterOptions = computed(() =>
  masterStore.regions.map((region) => ({
    ...region,
    label: formatRegion(region),
    value: region.id,
    disabled:
      selectedCountryCodes.value.length > 0 &&
      region.type !== 'COUNTRY' &&
      isCoveredBySelectedCountry(region),
  })),
)

function buildProjectParams(): GBProjectListParams {
  const params = projectControls.buildParams() as GBProjectListParams
  const locationIDs = expandLocationFilterIds(projectControls.appliedFilters.location_ids)
  if (locationIDs.length > 0) {
    params.location_ids = locationIDs
  }
  return params
}

async function loadData() {
  await Promise.all([
    greenBookStore.fetchGreenBook(greenBookId.value),
    blueBookStore.fetchProjectOptions(),
    masterStore.fetchInstitutions(true, { limit: 1000, sort: 'name', order: 'asc' }),
    masterStore.fetchAllRegionLevels(true),
    loadProjects(),
  ])
}

async function loadProjects() {
  await greenBookStore.fetchProjects(greenBookId.value, buildProjectParams())
}

function openEdit() {
  const current = greenBookStore.currentGreenBook
  if (!current) return
  Object.assign(form, {
    publish_year: current.publish_year,
    revision_number: current.revision_number,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = greenBookSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['publish_year', 'revision_number'])
    return
  }

  await greenBookStore.updateGreenBook(greenBookId.value, parsed.data)
  toast.success('Berhasil', 'Green Book berhasil diperbarui')
  dialogVisible.value = false
  await loadData()
}

function deleteProject(project: GBProject) {
  confirm.confirmDelete(`Proyek Green Book ${project.gb_code}`, async () => {
    await greenBookStore.deleteProject(greenBookId.value, project.id)
    toast.success('Berhasil', 'Proyek Green Book berhasil dihapus')
    if (greenBookStore.projects.length === 1 && projectControls.page.value > 1) {
      projectControls.page.value -= 1
    } else {
      await loadProjects()
    }
  })
}

function formatInstitution(institution: Institution) {
  return institution.short_name ? `${institution.name} (${institution.short_name})` : institution.name
}

function formatRegion(region: Region) {
  if (region.type === 'COUNTRY') return `${region.name} (Nasional)`
  if (region.type === 'CITY') return `-- ${region.name}`
  return `- ${region.name}`
}

function isCoveredBySelectedCountry(region: Region) {
  if (!region.parent_code) return false
  if (selectedCountryCodes.value.includes(region.parent_code)) return true
  const parent = masterStore.regions.find((item) => item.code === region.parent_code)
  return parent?.parent_code ? selectedCountryCodes.value.includes(parent.parent_code) : false
}

function expandLocationFilterIds(locationIDs: string[]) {
  if (locationIDs.length === 0) return []

  const expanded = new Set(locationIDs)
  const selectedRegions = masterStore.regions.filter((region) => expanded.has(region.id))

  selectedRegions.forEach((selectedRegion) => {
    if (selectedRegion.type === 'COUNTRY') {
      masterStore.regions.forEach((region) => {
        const parent = region.parent_code
          ? masterStore.regions.find((item) => item.code === region.parent_code)
          : undefined

        if (region.parent_code === selectedRegion.code || parent?.parent_code === selectedRegion.code) {
          expanded.add(region.id)
        }
      })
    }

    if (selectedRegion.type === 'PROVINCE') {
      masterStore.regions
        .filter((region) => region.parent_code === selectedRegion.code)
        .forEach((region) => expanded.add(region.id))
    }
  })

  return [...expanded]
}

function selectedLabelSummary(labels: string[]) {
  if (labels.length <= 2) return labels.join(', ')
  return `${labels.slice(0, 2).join(', ')} +${labels.length - 2}`
}

function selectedBBProjectSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    blueBookStore.projectOptions
      .filter((project) => selected.has(project.id))
      .map((project) => project.bb_code),
  )
}

function selectedInstitutionSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    masterStore.institutions
      .filter((institution) => selected.has(institution.id))
      .map((institution) => institution.short_name || institution.name),
  )
}

function selectedRegionSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    masterStore.regions.filter((region) => selected.has(region.id)).map((region) => region.name),
  )
}

onMounted(() => {
  void loadData()
})

watch(
  [
    projectControls.page,
    projectControls.limit,
    projectControls.debouncedSearch,
    () => JSON.stringify(projectControls.appliedFilters),
  ],
  () => {
    void loadProjects()
  },
)
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="greenBookStore.currentGreenBook ? `Green Book ${greenBookStore.currentGreenBook.publish_year}` : 'Detail Green Book'"
      :subtitle="greenBookStore.currentGreenBook ? formatGBRevision(greenBookStore.currentGreenBook.revision_number) : undefined"
    >
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'green-books' })" />
        <Button v-if="can('green_book', 'update')" label="Edit Green Book" icon="pi pi-pencil" outlined @click="openEdit" />
        <Button
          v-if="can('gb_project', 'create')"
          as="router-link"
          :to="{ name: 'gb-project-create', params: { gbId: greenBookId } }"
          label="Tambah Proyek"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <div v-if="greenBookStore.currentGreenBook" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-3">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Tahun Terbit</p>
        <p class="font-semibold text-surface-950">{{ greenBookStore.currentGreenBook.publish_year }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Revisi</p>
        <p class="font-semibold text-surface-950">{{ formatGBRevision(greenBookStore.currentGreenBook.revision_number) }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Status</p>
        <StatusBadge :status="greenBookStore.currentGreenBook.status" />
      </div>
    </div>

    <SearchFilterBar
      v-model:search="projectControls.search.value"
      search-placeholder="Cari kode, nama proyek, atau relasi Blue Book"
      :active-filters="projectControls.activeFilterPills.value"
      :filter-count="projectControls.activeFilterCount.value"
      @apply="projectControls.applyFilters"
      @reset="projectControls.resetFilters"
      @remove="projectControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Proyek Blue Book</span>
          <MultiSelect
            v-model="projectControls.draftFilters.bb_project_ids"
            :options="bbProjectOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua proyek Blue Book"
            filter
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Executing Agency</span>
          <MultiSelect
            v-model="projectControls.draftFilters.executing_agency_ids"
            :options="institutionFilterOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua executing agency"
            filter
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-1">
          <span class="text-sm font-medium text-surface-700">Location</span>
          <MultiSelect
            v-model="projectControls.draftFilters.location_ids"
            :options="locationFilterOptions"
            option-label="label"
            option-value="value"
            option-disabled="disabled"
            placeholder="Semua lokasi"
            filter
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-1">
          <span class="text-sm font-medium text-surface-700">Status</span>
          <MultiSelect
            v-model="projectControls.draftFilters.status"
            :options="projectStatusOptions"
            placeholder="Semua status"
            display="chip"
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

    <DataTable
      v-model:page="projectControls.page.value"
      v-model:limit="projectControls.limit.value"
      :data="greenBookStore.projects"
      :columns="columns"
      :loading="greenBookStore.loading"
      :total="greenBookStore.projectTotal"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'bb_projects'">{{ joinNames((row as GBProject).bb_projects) }}</span>
        <StatusBadge v-else-if="column.field === 'status'" :status="String(row.status)" />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'gb-project-detail', params: { gbId: greenBookId, id: row.id } }"
            icon="pi pi-eye"
            label="Lihat"
            size="small"
            outlined
          />
          <Button
            v-if="can('gb_project', 'update')"
            as="router-link"
            :to="{ name: 'gb-project-edit', params: { gbId: greenBookId, id: row.id } }"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
          />
          <Button
            v-if="can('gb_project', 'delete')"
            icon="pi pi-trash"
            label="Hapus"
            size="small"
            severity="danger"
            outlined
            @click="deleteProject(row as GBProject)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Edit Green Book" class="w-[32rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tahun Terbit</span>
          <InputNumber v-model="form.publish_year" :use-grouping="false" class="w-full" />
          <small v-if="errors.publish_year" class="text-red-600">{{ errors.publish_year }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nomor Revisi</span>
          <InputNumber v-model="form.revision_number" :min="0" class="w-full" />
          <small v-if="errors.revision_number" class="text-red-600">{{ errors.revision_number }}</small>
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
