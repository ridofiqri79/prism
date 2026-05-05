<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import MultiSelect from '@/components/common/MultiSelectDropdown.vue'
import LenderIndicationTable from '@/components/blue-book/LenderIndicationTable.vue'
import ProjectCostTable from '@/components/blue-book/ProjectCostTable.vue'
import FormActionBar from '@/components/common/FormActionBar.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import InstitutionSelect from '@/components/forms/InstitutionSelect.vue'
import LocationMultiSelect from '@/components/forms/LocationMultiSelect.vue'
import NationalPriorityMultiSelect from '@/components/forms/NationalPriorityMultiSelect.vue'
import ProgramTitleSelect from '@/components/forms/ProgramTitleSelect.vue'
import RichTextEditor from '@/components/forms/RichTextEditor.vue'
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
const currentProject = computed(() => blueBookStore.currentProject)
const bappenasPartnerLookupOptions = ref<BappenasPartner[]>([])
const bappenasPartnerParentOptions = ref<BappenasPartner[]>([])

const bappenasPartnerParentMap = computed(() => {
  const byId = new Map<string, BappenasPartner>()

  for (const partner of uniquePartners([
    ...bappenasPartnerParentOptions.value,
    ...masterStore.bappenasPartners.filter((item) => item.level === 'Eselon I'),
    ...bappenasPartnerLookupOptions.value.flatMap((item) => (item.parent ? [item.parent] : [])),
    ...(currentProject.value?.bappenas_partners ?? []).flatMap((item) =>
      item.parent ? [item.parent] : [],
    ),
  ])) {
    byId.set(partner.id, partner)
  }

  return byId
})

const bappenasPartnerOptions = computed(() => {
  const byId = new Map<string, BappenasPartner>()

  for (const partner of uniquePartners([
    ...bappenasPartnerLookupOptions.value,
    ...masterStore.bappenasPartners,
    ...(currentProject.value?.bappenas_partners ?? []),
  ])) {
    if (partner.level === 'Eselon II') {
      byId.set(partner.id, partner)
    }
  }

  return [...byId.values()]
})
const selectedPartners = computed(() =>
  form.values.bappenas_partner_ids
    .map((id) => bappenasPartnerOptions.value.find((partner) => partner.id === id))
    .filter((partner): partner is BappenasPartner => Boolean(partner)),
)
const selectedPartnerParents = computed(() => {
  const parents = selectedPartners.value
    .map((partner) => findPartnerParentName(partner))
    .filter((name) => name !== '-')
  return [...new Set(parents)].join(', ') || '-'
})
const selectedPartnerParentGroups = computed(() => {
  const groups = new Map<string, string[]>()

  for (const partner of selectedPartners.value) {
    const parentName = findPartnerParentName(partner)

    if (parentName === '-') {
      continue
    }

    groups.set(parentName, [...(groups.get(parentName) ?? []), partner.name])
  }

  return [...groups.entries()].map(([parentName, partnerNames]) => ({
    parentName,
    partnerNames,
  }))
})

function uniquePartners(partners: BappenasPartner[]) {
  const byId = new Map<string, BappenasPartner>()

  for (const partner of partners) {
    byId.set(partner.id, partner)
  }

  return [...byId.values()]
}

function findPartnerParentName(partner?: BappenasPartner) {
  if (!partner?.parent_id) return partner?.parent?.name ?? '-'
  return (
    bappenasPartnerParentMap.value.get(partner.parent_id)?.name ??
    partner.parent?.name ??
    '-'
  )
}

function optionName(option: Record<string, unknown>) {
  return String(option.name ?? '')
}

function optionParentName(option: Record<string, unknown>) {
  return findPartnerParentName(option as unknown as BappenasPartner)
}

function bappenasPartnerParams(search?: string) {
  return {
    limit: 50,
    level: ['Eselon II'],
    search: search?.trim() || undefined,
    sort: 'name',
    order: 'asc' as const,
  }
}

function bappenasPartnerParentParams() {
  return {
    limit: 200,
    level: ['Eselon I'],
    sort: 'name',
    order: 'asc' as const,
  }
}

async function loadBappenasPartnerOptions(search?: string, force = true) {
  const response = await masterStore.fetchBappenasPartners(force, bappenasPartnerParams(search))

  if (response?.data) {
    bappenasPartnerLookupOptions.value = uniquePartners([
      ...bappenasPartnerLookupOptions.value,
      ...response.data,
    ])
  }
}

async function loadBappenasPartnerParents() {
  const response = await masterStore.fetchBappenasPartnerTree(bappenasPartnerParentParams())
  bappenasPartnerParentOptions.value = response.data
}

async function loadData() {
  await Promise.all([
    blueBookStore.fetchBlueBook(blueBookId.value),
    loadBappenasPartnerOptions(undefined, true),
    loadBappenasPartnerParents(),
  ])

  if (isEditMode.value) {
    const project = await blueBookStore.fetchProject(blueBookId.value, projectId.value)
    form.applyProject(project)
  }
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
      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">Informasi Umum</h2>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Judul Program</span>
            <ProgramTitleSelect
              v-model="form.values.program_title_id"
              :extra-options="currentProject?.program_title ? [currentProject.program_title] : []"
            />
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
              append-to="body"
              :overlay-style="{ minWidth: '100%' }"
              class="w-full"
              @filter="loadBappenasPartnerOptions($event.value)"
            >
              <template #option="{ option }">
                <span class="block truncate font-medium">{{ optionName(option) }}</span>
                <span class="mt-0.5 block truncate text-xs text-surface-500">
                  Induk Eselon I: {{ optionParentName(option) }}
                </span>
              </template>
            </MultiSelect>
            <small v-if="form.errors.bappenas_partner_ids" class="text-red-600">
              {{ form.errors.bappenas_partner_ids }}
            </small>
          </label>
        </div>
        <div class="rounded-lg border border-surface-200 bg-surface-50 p-3 text-sm">
          <div class="flex flex-col gap-2 md:flex-row md:items-start">
            <span class="shrink-0 font-medium text-surface-700">Induk Eselon I</span>
            <div class="min-w-0 flex-1">
              <div v-if="selectedPartnerParentGroups.length" class="flex flex-wrap gap-2">
                <span
                  v-for="group in selectedPartnerParentGroups"
                  :key="group.parentName"
                  class="inline-flex max-w-full items-center rounded-md border border-primary-100 bg-primary-50 px-2 py-1 text-xs font-medium text-primary-700"
                  :title="group.partnerNames.join(', ')"
                >
                  <span class="truncate">{{ group.parentName }}</span>
                </span>
              </div>
              <span v-else class="text-surface-500">{{ selectedPartnerParents }}</span>
            </div>
          </div>
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
          <div class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Tujuan</span>
            <RichTextEditor v-model="form.values.objective" placeholder="Tulis tujuan proyek" />
          </div>
          <div class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Lingkup Pekerjaan</span>
            <RichTextEditor
              v-model="form.values.scope_of_work"
              placeholder="Tulis lingkup pekerjaan"
            />
          </div>
          <div class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Outputs</span>
            <RichTextEditor v-model="form.values.outputs" placeholder="Tulis outputs proyek" />
          </div>
          <div class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Outcomes</span>
            <RichTextEditor v-model="form.values.outcomes" placeholder="Tulis outcomes proyek" />
          </div>
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="mt-1 text-lg font-semibold text-surface-950">Pihak Terlibat</h2>
        </div>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Executing Agency</span>
            <InstitutionSelect
              v-model="form.values.executing_agency_ids"
              :extra-options="currentProject?.executing_agencies ?? []"
              multiple
            />
            <small v-if="form.errors.executing_agency_ids" class="text-red-600">
              {{ form.errors.executing_agency_ids }}
            </small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Implementing Agency</span>
            <InstitutionSelect
              v-model="form.values.implementing_agency_ids"
              :extra-options="currentProject?.implementing_agencies ?? []"
              multiple
            />
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
          <div class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Lokasi</span>
            <LocationMultiSelect v-model="form.values.location_ids" />
            <small v-if="form.errors.location_ids" class="text-red-600">{{
              form.errors.location_ids
            }}</small>
          </div>
          <div class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Prioritas Nasional</span>
            <NationalPriorityMultiSelect
              v-model="form.values.national_priority_ids"
              :extra-options="currentProject?.national_priorities ?? []"
            />
          </div>
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
          :extra-lender-options="currentProject?.lender_indications.map((item) => item.lender) ?? []"
          @add="form.addIndication"
          @remove="form.removeIndication"
        />
      </section>

      <FormActionBar
        :loading="blueBookStore.loading"
        @cancel="router.push({ name: 'blue-book-detail', params: { id: blueBookId } })"
      />
    </form>
  </section>
</template>
