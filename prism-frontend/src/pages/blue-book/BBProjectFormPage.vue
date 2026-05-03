<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import Textarea from 'primevue/textarea'
import LenderIndicationTable from '@/components/blue-book/LenderIndicationTable.vue'
import ProjectCostTable from '@/components/blue-book/ProjectCostTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import InstitutionSelect from '@/components/forms/InstitutionSelect.vue'
import LocationMultiSelect from '@/components/forms/LocationMultiSelect.vue'
import NationalPriorityMultiSelect from '@/components/forms/NationalPriorityMultiSelect.vue'
import ProgramTitleSelect from '@/components/forms/ProgramTitleSelect.vue'
import { useBBProjectForm } from '@/composables/forms/useBBProjectForm'
import { useToast } from '@/composables/useToast'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useMasterStore } from '@/stores/master.store'
import type { BappenasPartner } from '@/types/master.types'

const route = useRoute()
const router = useRouter()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()

const blueBookId = computed(() => String(route.params.bbId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const isEditMode = computed(() => route.name === 'bb-project-edit')
const pageTitle = computed(() =>
  isEditMode.value ? 'Edit Proyek Blue Book' : 'Tambah Proyek Blue Book',
)

const form = useBBProjectForm()
const selectedRevisionSourceProjectId = ref('')

const bappenasPartnerOptions = computed(() =>
  masterStore.bappenasPartners.filter((partner) => partner.level === 'Eselon II'),
)
const selectedPartners = computed(() =>
  form.values.bappenas_partner_ids
    .map((id) => masterStore.bappenasPartners.find((partner) => partner.id === id))
    .filter((partner): partner is BappenasPartner => Boolean(partner)),
)
const selectedPartnerParents = computed(() => {
  const parents = selectedPartners.value
    .map((partner) => findPartnerParent(partner))
    .filter((name) => name !== '-')
  return [...new Set(parents)].join(', ') || '-'
})
const currentBlueBook = computed(() => blueBookStore.currentBlueBook)
const isRevisionBlueBook = computed(() =>
  Boolean(currentBlueBook.value?.replaces_blue_book_id) ||
  Number(currentBlueBook.value?.revision_number ?? 0) > 0,
)
const showRevisionSourcePicker = computed(() => !isEditMode.value && isRevisionBlueBook.value)
const selectedRevisionSourceProject = computed(() =>
  blueBookStore.revisionSourceProjectOptions.find(
    (project) => project.id === selectedRevisionSourceProjectId.value,
  ),
)
const revisionSourceProjectOptions = computed(() =>
  blueBookStore.revisionSourceProjectOptions.map((project) => ({
    ...project,
    label: revisionSourceOptionLabel(project),
  })),
)
function findPartnerParent(partner?: BappenasPartner) {
  if (!partner?.parent_id) return partner?.parent?.name ?? '-'
  return (
    masterStore.bappenasPartners.find((item) => item.id === partner.parent_id)?.name ??
    partner.parent?.name ??
    '-'
  )
}

async function loadData() {
  selectedRevisionSourceProjectId.value = ''
  delete form.errors.project_identity_id

  const [blueBook] = await Promise.all([
    blueBookStore.fetchBlueBook(blueBookId.value),
    masterStore.fetchProgramTitles(true, { limit: 1000 }),
    masterStore.fetchBappenasPartners(true, { limit: 1000 }),
    masterStore.fetchInstitutions(true, { limit: 1000 }),
    masterStore.fetchAllRegionLevels(true),
    masterStore.fetchNationalPriorities(true, { limit: 1000 }),
    masterStore.fetchLenders(true, { limit: 1000 }),
  ])

  if (isEditMode.value) {
    const project = await blueBookStore.fetchProject(blueBookId.value, projectId.value)
    form.applyProject(project)
    blueBookStore.clearRevisionSourceProjectOptions()
    return
  }

  if (isRevisionBlueBook.value) {
    await blueBookStore.fetchRevisionSourceProjectOptions(blueBook)
  } else {
    blueBookStore.clearRevisionSourceProjectOptions()
  }
}

function revisionSourceOptionLabel(project: { bb_code: string; project_name: string }) {
  return `${project.bb_code} - ${project.project_name}`
}

function normalizeProjectCode(code: string) {
  return code.trim().toLowerCase()
}

function findRevisionSourceProjectByCode(code: string) {
  const normalizedCode = normalizeProjectCode(code)
  return blueBookStore.revisionSourceProjectOptions.find(
    (project) => normalizeProjectCode(project.bb_code) === normalizedCode,
  )
}

const onSubmit = form.submit(async (values) => {
  if (isEditMode.value) {
    await blueBookStore.updateProject(blueBookId.value, projectId.value, values)
    toast.success('Berhasil', 'Proyek Blue Book berhasil diperbarui')
    await router.push({
      name: 'bb-project-detail',
      params: { bbId: blueBookId.value, id: projectId.value },
    })
    return
  }

  if (showRevisionSourcePicker.value && !values.project_identity_id) {
    const sourceProject = findRevisionSourceProjectByCode(values.bb_code)

    if (sourceProject) {
      selectedRevisionSourceProjectId.value = sourceProject.id
      form.applyProject(sourceProject)
      form.errors.project_identity_id =
        'Kode Blue Book ditemukan di revisi sebelumnya. Pilih proyek eksisting untuk melanjutkan histori.'
      toast.warn(
        'Pilih Proyek Eksisting',
        'Data proyek dari revisi sebelumnya sudah diambil. Periksa data form lalu simpan kembali.',
      )
      return
    }
  }

  const created = await blueBookStore.createProject(blueBookId.value, values)
  toast.success('Berhasil', 'Proyek Blue Book berhasil dibuat')
  await router.push({
    name: 'bb-project-detail',
    params: { bbId: blueBookId.value, id: created.id },
  })
})

onMounted(() => {
  void loadData()
})

watch(selectedRevisionSourceProjectId, () => {
  if (!selectedRevisionSourceProjectId.value) {
    form.reset()
    delete form.errors.project_identity_id
    return
  }

  if (selectedRevisionSourceProject.value) {
    delete form.errors.project_identity_id
    form.applyProject(selectedRevisionSourceProject.value)
  }
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader :title="pageTitle" subtitle="Lengkapi data proyek Blue Book">
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'blue-book-detail', params: { id: blueBookId } })"
        />
      </template>
    </PageHeader>

    <form class="space-y-6" @submit.prevent="onSubmit">
      <section
        v-if="showRevisionSourcePicker"
        class="space-y-4 rounded-lg border border-surface-200 bg-white p-5"
      >
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">
            Ambil dari Proyek Blue Book Eksisting
          </h2>
        </div>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Proyek dari Revisi Sebelumnya</span>
          <Select
            v-model="selectedRevisionSourceProjectId"
            :options="revisionSourceProjectOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih proyek Blue Book eksisting"
            filter
            show-clear
            class="w-full"
            :loading="blueBookStore.revisionSourceProjectLoading"
          >
            <template #option="{ option }">
              <div class="space-y-1">
                <p class="font-medium text-surface-950">
                  {{ option.bb_code }} - {{ option.project_name }}
                </p>
                <p class="text-xs text-surface-500">{{ option.source_blue_book_label }}</p>
              </div>
            </template>
          </Select>
          <small v-if="form.errors.project_identity_id" class="text-red-600">
            {{ form.errors.project_identity_id }}
          </small>
        </label>
        <div
          v-if="selectedRevisionSourceProject"
          class="rounded-lg border border-primary-100 bg-primary-50 p-3 text-sm text-primary-900"
        >
          Proyek ini akan disimpan sebagai revisi dari
          <strong>
            {{ selectedRevisionSourceProject.bb_code }} -
            {{ selectedRevisionSourceProject.project_name }}
          </strong>
          dan data form tetap bisa diedit sebelum disimpan.
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">Informasi Umum</h2>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Judul Program</span>
            <ProgramTitleSelect v-model="form.values.program_title_id" />
            <small v-if="form.errors.program_title_id" class="text-red-600">
              {{ form.errors.program_title_id }}
            </small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700"
              >Mitra Kerja Bappenas (Eselon II)</span
            >
            <MultiSelect
              v-model="form.values.bappenas_partner_ids"
              :options="bappenasPartnerOptions"
              option-label="name"
              option-value="id"
              placeholder="Pilih mitra kerja Bappenas"
              filter
              display="chip"
              class="w-full"
            />
            <small v-if="form.errors.bappenas_partner_ids" class="text-red-600">
              {{ form.errors.bappenas_partner_ids }}
            </small>
          </label>
        </div>
        <div
          class="rounded-lg border border-surface-200 bg-surface-50 p-3 text-sm text-surface-700"
        >
          Induk Eselon I: <strong>{{ selectedPartnerParents }}</strong>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Kode Blue Book</span>
            <InputText v-model="form.values.bb_code" class="w-full" :disabled="isEditMode" />
            <small v-if="form.errors.bb_code" class="text-red-600">{{ form.errors.bb_code }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Nama Proyek</span>
            <InputText v-model="form.values.project_name" class="w-full" />
            <small v-if="form.errors.project_name" class="text-red-600">{{
              form.errors.project_name
            }}</small>
          </label>
          <label class="block space-y-2 md:col-span-2">
            <span class="text-sm font-medium text-surface-700">Durasi (bulan)</span>
            <InputNumber
              v-model="form.values.duration"
              class="w-full"
              :min="1"
              :use-grouping="false"
            />
            <small v-if="form.errors.duration" class="text-red-600">{{
              form.errors.duration
            }}</small>
          </label>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Tujuan</span>
            <Textarea v-model="form.values.objective" auto-resize rows="3" class="w-full" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Lingkup Pekerjaan</span>
            <Textarea v-model="form.values.scope_of_work" auto-resize rows="3" class="w-full" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Outputs</span>
            <Textarea v-model="form.values.outputs" auto-resize rows="3" class="w-full" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Outcomes</span>
            <Textarea v-model="form.values.outcomes" auto-resize rows="3" class="w-full" />
          </label>
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">Pihak Terlibat</h2>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
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
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">Lokasi & Prioritas</h2>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Lokasi</span>
            <LocationMultiSelect v-model="form.values.location_ids" />
            <small v-if="form.errors.location_ids" class="text-red-600">{{
              form.errors.location_ids
            }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Prioritas Nasional</span>
            <NationalPriorityMultiSelect v-model="form.values.national_priority_ids" />
          </label>
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">Biaya Proyek</h2>
        </div>
        <ProjectCostTable
          v-model:rows="form.projectCosts.value"
          @add="form.addCost"
          @remove="form.removeCost"
        />
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">Indikasi Lender</h2>
        </div>
        <LenderIndicationTable
          v-model:rows="form.lenderIndications.value"
          @add="form.addIndication"
          @remove="form.removeIndication"
        />
      </section>

      <div
        class="sticky bottom-0 flex justify-end gap-2 border-t border-surface-200 bg-surface-50/95 py-4 backdrop-blur"
      >
        <Button
          label="Batal"
          severity="secondary"
          outlined
          @click="router.push({ name: 'blue-book-detail', params: { id: blueBookId } })"
        />
        <Button type="submit" label="Simpan" icon="pi pi-save" :loading="blueBookStore.loading" />
      </div>
    </form>
  </section>
</template>
