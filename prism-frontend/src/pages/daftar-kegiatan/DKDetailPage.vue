<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Accordion from 'primevue/accordion'
import AccordionContent from 'primevue/accordioncontent'
import AccordionHeader from 'primevue/accordionheader'
import AccordionPanel from 'primevue/accordionpanel'
import Button from 'primevue/button'
import MultiSelect from 'primevue/multiselect'
import Tag from 'primevue/tag'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ListPaginationFooter from '@/components/common/ListPaginationFooter.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import TableReloadShell from '@/components/common/TableReloadShell.vue'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useDaftarKegiatanStore } from '@/stores/daftar-kegiatan.store'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useMasterStore } from '@/stores/master.store'
import type { DKProject, DKProjectListParams, GBProjectSummary } from '@/types/daftar-kegiatan.types'
import type { Institution, Region } from '@/types/master.types'
import { formatDate, joinNames } from './daftar-kegiatan-page-utils'

const route = useRoute()
const router = useRouter()
const dkStore = useDaftarKegiatanStore()
const greenBookStore = useGreenBookStore()
const masterStore = useMasterStore()
const { can } = usePermission()
interface DKProjectFilterState {
  gb_project_ids: string[]
  executing_agency_ids: string[]
  location_ids: string[]
  lender_ids: string[]
}

const dkId = computed(() => String(route.params.id ?? ''))
const projectControls = useListControls<DKProjectFilterState>({
  initialFilters: {
    gb_project_ids: [],
    executing_agency_ids: [],
    location_ids: [],
    lender_ids: [],
  },
  filterLabels: {
    gb_project_ids: 'Proyek Green Book',
    executing_agency_ids: 'Executing Agency',
    location_ids: 'Location',
    lender_ids: 'Lender',
  },
  formatFilterValue: (key, value) => {
    if (key === 'gb_project_ids' && Array.isArray(value)) return selectedGBProjectSummary(value)
    if (key === 'executing_agency_ids' && Array.isArray(value)) return selectedInstitutionSummary(value)
    if (key === 'location_ids' && Array.isArray(value)) return selectedRegionSummary(value)
    if (key === 'lender_ids' && Array.isArray(value)) return selectedLenderSummary(value)
    return Array.isArray(value) ? selectedLabelSummary(value) : String(value)
  },
})
const selectedCountryCodes = computed(() => {
  const selected = new Set(projectControls.draftFilters.location_ids)

  return masterStore.regions
    .filter((region) => region.type === 'COUNTRY' && selected.has(region.id))
    .map((region) => region.code)
})
const gbProjectOptions = computed(() =>
  greenBookStore.projectOptions.map((project) => ({
    ...project,
    label: `${project.gb_code} - ${project.project_name}`,
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
const lenderOptions = computed(() =>
  masterStore.lenders.map((lender) => ({
    ...lender,
    label: lender.short_name ? `${lender.name} (${lender.short_name})` : lender.name,
    value: lender.id,
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
const initialProjectsLoading = computed(() => dkStore.loading && dkStore.projects.length === 0)
const refreshingProjects = computed(() => dkStore.loading && dkStore.projects.length > 0)

function buildProjectParams(): DKProjectListParams {
  const params = projectControls.buildParams() as DKProjectListParams
  const locationIDs = expandLocationFilterIds(projectControls.appliedFilters.location_ids)
  if (locationIDs.length > 0) {
    params.location_ids = locationIDs
  }
  return params
}

async function loadData() {
  await Promise.all([
    dkStore.fetchDK(dkId.value),
    greenBookStore.fetchProjectOptions(),
    masterStore.fetchLenders(true, { limit: 1000, sort: 'name', order: 'asc' }),
    masterStore.fetchInstitutions(true, { limit: 1000, sort: 'name', order: 'asc' }),
    masterStore.fetchAllRegionLevels(true),
    loadProjects(),
  ])
}

async function loadProjects() {
  await dkStore.fetchProjects(dkId.value, buildProjectParams())
}

function gbProjectRoute(project: GBProjectSummary) {
  return project.green_book_id
    ? { name: 'gb-project-detail', params: { gbId: project.green_book_id, id: project.id } }
    : { name: 'green-books' }
}

function loanAgreementRoute(project: DKProject) {
  if (project.loan_agreement) {
    return { name: 'loan-agreement-detail', params: { id: project.loan_agreement.id } }
  }

  return {
    name: 'loan-agreement-create',
    query: {
      dk_id: dkId.value,
      dk_project_id: project.id,
    },
  }
}

function canUseLoanAgreementAction(project: DKProject) {
  return project.loan_agreement
    ? can('loan_agreement', 'read')
    : can('loan_agreement', 'create')
}

function hasFinancingLender(project: DKProject) {
  return project.financing_details.some((detail) => Boolean(detail.lender?.id))
}

function isLoanAgreementActionDisabled(project: DKProject) {
  return !project.loan_agreement && !hasFinancingLender(project)
}

function loanAgreementActionLabel(project: DKProject) {
  return project.loan_agreement ? 'Buka Loan Agreement' : 'Buat Loan Agreement'
}

function loanAgreementActionIcon(project: DKProject) {
  return project.loan_agreement ? 'pi pi-external-link' : 'pi pi-file-plus'
}

function loanAgreementActionTitle(project: DKProject) {
  if (project.loan_agreement) {
    return `Loan Agreement ${project.loan_agreement.loan_code}`
  }

  if (!hasFinancingLender(project)) {
    return 'Tambahkan lender pada rincian pembiayaan sebelum membuat Loan Agreement'
  }

  return 'Buat Loan Agreement dari proyek ini'
}

function goToLoanAgreement(project: DKProject) {
  if (isLoanAgreementActionDisabled(project)) return
  void router.push(loanAgreementRoute(project))
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

function selectedGBProjectSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    greenBookStore.projectOptions
      .filter((project) => selected.has(project.id))
      .map((project) => project.gb_code),
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

function selectedLenderSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    masterStore.lenders
      .filter((lender) => selected.has(lender.id))
      .map((lender) => lender.short_name || lender.name),
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
      :title="dkStore.currentDK?.subject ?? 'Detail Daftar Kegiatan'"
      :subtitle="dkStore.currentDK ? formatDate(dkStore.currentDK.date) : undefined"
    >
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'daftar-kegiatan' })"
        />
        <Button
          v-if="can('daftar_kegiatan', 'create')"
          as="router-link"
          :to="{ name: 'dk-project-create', params: { dkId } }"
          label="Tambah Proyek ke Surat"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <div
      v-if="dkStore.currentDK"
      class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-3"
    >
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Perihal</p>
        <p class="font-semibold text-surface-950">{{ dkStore.currentDK.subject }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Tanggal</p>
        <p class="font-semibold text-surface-950">{{ formatDate(dkStore.currentDK.date) }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Nomor Surat</p>
        <p class="font-semibold text-surface-950">{{ dkStore.currentDK.letter_number || '-' }}</p>
      </div>
    </div>

    <SearchFilterBar
      v-model:search="projectControls.search.value"
      search-placeholder="Cari nama proyek, proyek Green Book, lender, lokasi, atau komponen"
      :active-filters="projectControls.activeFilterPills.value"
      :filter-count="projectControls.activeFilterCount.value"
      @apply="projectControls.applyFilters"
      @reset="projectControls.resetFilters"
      @remove="projectControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-2">
          <span class="text-sm font-medium text-surface-700">Proyek Green Book</span>
          <MultiSelect
            v-model="projectControls.draftFilters.gb_project_ids"
            :options="gbProjectOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua proyek Green Book"
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
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <MultiSelect
            v-model="projectControls.draftFilters.lender_ids"
            :options="lenderOptions"
            option-label="label"
            option-value="value"
            placeholder="Semua lender"
            filter
            display="chip"
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

    <div v-if="initialProjectsLoading" class="rounded-lg border border-surface-200 bg-white p-8">
      <EmptyState title="Memuat proyek" description="Data proyek Daftar Kegiatan sedang dimuat." />
    </div>

    <TableReloadShell v-else-if="dkStore.projects.length" :refreshing="refreshingProjects">
      <Accordion value="0" class="space-y-3">
        <AccordionPanel
          v-for="(project, index) in dkStore.projects"
          :key="project.id"
          :value="String(index)"
          class="overflow-hidden rounded-lg border border-surface-200 bg-white"
        >
        <AccordionHeader>
          <div class="flex w-full flex-wrap items-center justify-between gap-3 pr-4">
            <span class="min-w-0 flex-1">{{ project.project_name }}</span>
            <div class="flex flex-wrap items-center justify-end gap-3">
              <Button
                v-if="canUseLoanAgreementAction(project)"
                :label="loanAgreementActionLabel(project)"
                :icon="loanAgreementActionIcon(project)"
                :disabled="isLoanAgreementActionDisabled(project)"
                :title="loanAgreementActionTitle(project)"
                size="small"
                outlined
                @click.stop="goToLoanAgreement(project)"
              />
              <span class="text-sm font-normal text-surface-500">{{
                project.duration ? `${project.duration} bulan` : '-'
              }}</span>
            </div>
          </div>
        </AccordionHeader>
        <AccordionContent>
          <div class="space-y-5 p-2">
            <div class="grid gap-4 md:grid-cols-2">
              <div class="md:col-span-2">
                <p class="text-xs uppercase tracking-wide text-surface-500">
                  Nama Proyek Daftar Kegiatan
                </p>
                <p class="font-medium text-surface-900">{{ project.project_name }}</p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-surface-500">Executing Agency</p>
                <p class="font-medium text-surface-900">
                  {{ project.institution?.name ?? project.institution_id ?? '-' }}
                </p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-surface-500">Lokasi</p>
                <p class="font-medium text-surface-900">{{ joinNames(project.locations) }}</p>
              </div>
              <div class="md:col-span-2">
                <p class="text-xs uppercase tracking-wide text-surface-500">Mitra Kerja Bappenas</p>
                <p class="font-medium text-surface-900">
                  {{ joinNames(project.bappenas_partners) }}
                </p>
              </div>
              <div class="md:col-span-2">
                <p class="text-xs uppercase tracking-wide text-surface-500">Objectives</p>
                <p class="text-surface-800">{{ project.objectives || '-' }}</p>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">
                Snapshot Proyek Green Book yang Dipakai
              </h3>
              <div class="flex flex-wrap gap-2">
                <RouterLink
                  v-for="gbProject in project.gb_projects"
                  :key="gbProject.id"
                  :to="gbProjectRoute(gbProject)"
                  class="inline-flex items-center gap-2 rounded-full border border-surface-200 px-3 py-1.5 text-sm font-medium text-primary"
                >
                  <span>{{ gbProject.gb_code }} - {{ gbProject.project_name }}</span>
                  <Tag
                    v-if="gbProject.has_newer_revision"
                    value="Ada revisi lebih baru"
                    severity="warn"
                    rounded
                  />
                  <Tag v-else-if="gbProject.is_latest" value="Terbaru" severity="success" rounded />
                </RouterLink>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">Rincian Pembiayaan</h3>
              <div class="overflow-auto rounded-lg border border-surface-200">
                <table class="w-full min-w-[56rem] text-left text-sm">
                  <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
                    <tr>
                      <th class="px-4 py-3">Lender</th>
                      <th class="px-4 py-3">Mata Uang</th>
                      <th class="px-4 py-3">Nilai Asli</th>
                      <th class="px-4 py-3">USD</th>
                      <th class="px-4 py-3">Catatan</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-surface-100">
                    <tr v-for="row in project.financing_details" :key="row.id">
                      <td class="px-4 py-3">{{ row.lender?.name ?? '-' }}</td>
                      <td class="px-4 py-3">{{ row.currency }}</td>
                      <td class="px-4 py-3">
                        <CurrencyDisplay :amount="row.amount_original" :currency="row.currency" />
                      </td>
                      <td class="px-4 py-3"><CurrencyDisplay :amount="row.amount_usd" /></td>
                      <td class="px-4 py-3">{{ row.remarks || '-' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">Alokasi Pinjaman</h3>
              <div class="overflow-auto rounded-lg border border-surface-200">
                <table class="w-full min-w-[56rem] text-left text-sm">
                  <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
                    <tr>
                      <th class="px-4 py-3">Instansi</th>
                      <th class="px-4 py-3">Mata Uang</th>
                      <th class="px-4 py-3">Nilai Asli</th>
                      <th class="px-4 py-3">USD</th>
                      <th class="px-4 py-3">Catatan</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-surface-100">
                    <tr v-for="row in project.loan_allocations" :key="row.id">
                      <td class="px-4 py-3">{{ row.institution?.name ?? '-' }}</td>
                      <td class="px-4 py-3">{{ row.currency }}</td>
                      <td class="px-4 py-3">
                        <CurrencyDisplay :amount="row.amount_original" :currency="row.currency" />
                      </td>
                      <td class="px-4 py-3"><CurrencyDisplay :amount="row.amount_usd" /></td>
                      <td class="px-4 py-3">{{ row.remarks || '-' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">Rincian Kegiatan</h3>
              <ol class="space-y-2">
                <li
                  v-for="activity in project.activity_details"
                  :key="activity.id"
                  class="rounded-lg border border-surface-200 p-3 text-sm"
                >
                  <span class="font-semibold">{{ activity.activity_number }}.</span>
                  {{ activity.activity_name }}
                </li>
              </ol>
            </div>
          </div>
        </AccordionContent>
        </AccordionPanel>
      </Accordion>
    </TableReloadShell>

    <div
      v-else
      class="rounded-lg border border-dashed border-surface-300 bg-white p-8 text-center text-surface-500"
    >
      Belum ada proyek dalam surat ini.
    </div>

    <ListPaginationFooter
      v-model:page="projectControls.page.value"
      v-model:limit="projectControls.limit.value"
      :total="dkStore.projectTotal"
    />
  </section>
</template>
