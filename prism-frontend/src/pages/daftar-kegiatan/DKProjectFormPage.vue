<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Textarea from 'primevue/textarea'
import ActivityDetailsTable from '@/components/daftar-kegiatan/ActivityDetailsTable.vue'
import FinancingDetailTable from '@/components/daftar-kegiatan/FinancingDetailTable.vue'
import LoanAllocationTable from '@/components/daftar-kegiatan/LoanAllocationTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import InstitutionSelect from '@/components/forms/InstitutionSelect.vue'
import LocationMultiSelect from '@/components/forms/LocationMultiSelect.vue'
import ProgramTitleSelect from '@/components/forms/ProgramTitleSelect.vue'
import { useDKProjectForm } from '@/composables/forms/useDKProjectForm'
import { useToast } from '@/composables/useToast'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useDaftarKegiatanStore } from '@/stores/daftar-kegiatan.store'
import { useGreenBookStore } from '@/stores/green-book.store'
import { useMasterStore } from '@/stores/master.store'

const route = useRoute()
const router = useRouter()
const dkStore = useDaftarKegiatanStore()
const greenBookStore = useGreenBookStore()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()

const dkId = computed(() => String(route.params.dkId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const isEditMode = computed(() => route.name === 'dk-project-edit')
const pageTitle = computed(() =>
  isEditMode.value ? 'Edit Proyek Daftar Kegiatan' : 'Tambah Proyek Daftar Kegiatan',
)
const form = useDKProjectForm(null, {
  gbProjects: () => greenBookStore.projectOptions,
  bbProjects: () => blueBookStore.projectOptions,
})

const gbProjectOptions = computed(() =>
  greenBookStore.projectOptions
    .filter((project) => project.is_latest !== false)
    .map((project) => ({
      ...project,
      label: `${project.gb_code} - ${project.project_name}`,
    })),
)
const bappenasPartnerOptions = computed(() =>
  masterStore.bappenasPartners.filter((partner) => partner.level === 'Eselon II'),
)

async function loadData() {
  await Promise.all([
    dkStore.fetchDK(dkId.value),
    greenBookStore.fetchProjectOptions(),
    blueBookStore.fetchProjectOptions(),
    masterStore.fetchProgramTitles(true, { limit: 1000 }),
    masterStore.fetchBappenasPartners(true, { limit: 1000 }),
    masterStore.fetchInstitutions(true, { limit: 1000 }),
    masterStore.fetchAllRegionLevels(true),
    masterStore.fetchLenders(true, { limit: 1000 }),
  ])

  if (isEditMode.value) {
    const project = await dkStore.fetchProject(dkId.value, projectId.value)
    form.applyProject(project)
  }
}

const onSubmit = form.submit(async (values) => {
  if (isEditMode.value) {
    await dkStore.updateProject(dkId.value, projectId.value, values)
    toast.success('Berhasil', 'Proyek Daftar Kegiatan berhasil diperbarui')
    await router.push({ name: 'daftar-kegiatan-detail', params: { id: dkId.value } })
    return
  }

  await dkStore.createProject(dkId.value, values)
  toast.success('Berhasil', 'Proyek Daftar Kegiatan berhasil dibuat')
  await router.push({ name: 'daftar-kegiatan-detail', params: { id: dkId.value } })
})

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader :title="pageTitle" subtitle="Lengkapi data proyek dalam Daftar Kegiatan">
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'daftar-kegiatan-detail', params: { id: dkId } })"
        />
      </template>
    </PageHeader>

    <form class="space-y-6" @submit.prevent="onSubmit">
      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Header Proyek</h2>
        </div>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Proyek Green Book</span>
          <MultiSelect
            v-model="form.values.gb_project_ids"
            :options="gbProjectOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih Proyek Green Book"
            filter
            display="chip"
            class="w-full"
            @change="form.applySelectedGBProjects"
          />
          <small v-if="form.errors.gb_project_ids" class="text-red-600">{{
            form.errors.gb_project_ids
          }}</small>
        </label>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2 md:col-span-2">
            <span class="text-sm font-medium text-surface-700">Nama Proyek Daftar Kegiatan</span>
            <InputText v-model="form.values.project_name" class="w-full" />
            <small v-if="form.errors.project_name" class="text-red-600">{{
              form.errors.project_name
            }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Judul Program</span>
            <ProgramTitleSelect v-model="form.values.program_title_id" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Executing Agency</span>
            <InstitutionSelect v-model="form.values.institution_id" />
            <small v-if="form.errors.institution_id" class="text-red-600">{{
              form.errors.institution_id
            }}</small>
          </label>
          <label class="block space-y-2 md:col-span-2">
            <span class="text-sm font-medium text-surface-700">Mitra Kerja Bappenas</span>
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
          <label class="block space-y-2">
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
          <div class="block space-y-2 md:col-span-2">
            <span class="text-sm font-medium text-surface-700">Lokasi</span>
            <LocationMultiSelect v-model="form.values.location_ids" />
            <small v-if="form.errors.location_ids" class="text-red-600">{{
              form.errors.location_ids
            }}</small>
          </div>
          <label class="block space-y-2 md:col-span-2">
            <span class="text-sm font-medium text-surface-700">Tujuan</span>
            <Textarea v-model="form.values.objectives" auto-resize rows="3" class="w-full" />
          </label>
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Rincian Pembiayaan</h2>
          <p class="text-sm text-surface-500">
            Pilihan lender hanya berisi lender dari funding source Green Book dan indikasi lender
            Blue Book proyek terpilih.
          </p>
          <small v-if="form.errors.financing_details" class="text-red-600">{{
            form.errors.financing_details
          }}</small>
        </div>
        <FinancingDetailTable
          v-model:rows="form.financingDetails.value"
          :allowed-lender-ids="form.allowedLenderIds.value"
          @add="form.addFinancing"
          @remove="form.removeFinancing"
        />
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Alokasi Pinjaman</h2>
          <small v-if="form.errors.loan_allocations" class="text-red-600">{{
            form.errors.loan_allocations
          }}</small>
        </div>
        <LoanAllocationTable
          v-model:rows="form.loanAllocations.value"
          @add="form.addAllocation"
          @remove="form.removeAllocation"
        />
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Rincian Kegiatan</h2>
          <p class="text-sm text-surface-500">
            Nomor urut otomatis dihitung ulang saat baris dihapus.
          </p>
          <small v-if="form.errors.activity_details" class="text-red-600">{{
            form.errors.activity_details
          }}</small>
        </div>
        <ActivityDetailsTable
          v-model:rows="form.activityDetails.value"
          @add="form.addActivity"
          @remove="form.removeActivity"
        />
      </section>

      <div
        class="sticky bottom-0 flex justify-end gap-2 border-t border-surface-200 bg-surface-50/95 py-4 backdrop-blur"
      >
        <Button
          label="Batal"
          severity="secondary"
          outlined
          @click="router.push({ name: 'daftar-kegiatan-detail', params: { id: dkId } })"
        />
        <Button type="submit" label="Simpan" icon="pi pi-save" :loading="dkStore.loading" />
      </div>
    </form>
  </section>
</template>
