<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AutoComplete, { type AutoCompleteCompleteEvent } from 'primevue/autocomplete'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import CurrencySelect from '@/components/forms/CurrencySelect.vue'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useLAForm } from '@/composables/forms/useLAForm'
import { useToast } from '@/composables/useToast'
import { useLoanAgreementStore } from '@/stores/loan-agreement.store'
import type { DKProjectLoanOption } from '@/types/loan-agreement.types'
import { formatDKProjectLabel, parseDateModel, toDateString } from './loan-agreement-page-utils'

const route = useRoute()
const router = useRouter()
const loanAgreementStore = useLoanAgreementStore()
const toast = useToast()

const loanAgreementId = computed(() => String(route.params.id ?? ''))
const isEditMode = computed(() => route.name === 'loan-agreement-edit')
const pageTitle = computed(() => (isEditMode.value ? 'Edit Loan Agreement' : 'Buat Loan Agreement'))
const sourceDKId = computed(() => queryString(route.query.dk_id))
const sourceDKProjectId = computed(() => queryString(route.query.dk_project_id))
const backRoute = computed(() =>
  sourceDKId.value && !isEditMode.value
    ? { name: 'daftar-kegiatan-detail', params: { id: sourceDKId.value } }
    : { name: 'loan-agreements' },
)
const form = useLAForm(null, {
  dkProjects: () => loanAgreementStore.dkProjectOptions,
})
const dkProjectModel = computed<DKProjectLoanOption | null>({
  get: () => loanAgreementStore.dkProjectOptionMap.get(form.values.dk_project_id) ?? null,
  set: (value) => {
    form.values.dk_project_id = value?.id ?? ''
  },
})
const agreementDateModel = computed({
  get: () => parseDateModel(form.values.agreement_date),
  set: (value: Date | null) => {
    form.values.agreement_date = toDateString(value)
  },
})
const effectiveDateModel = computed({
  get: () => parseDateModel(form.values.effective_date),
  set: (value: Date | null) => {
    form.values.effective_date = toDateString(value)
  },
})
const originalClosingDateModel = computed({
  get: () => parseDateModel(form.values.original_closing_date),
  set: (value: Date | null) => {
    form.values.original_closing_date = toDateString(value)
  },
})
const closingDateModel = computed({
  get: () => parseDateModel(form.values.closing_date),
  set: (value: Date | null) => {
    form.values.closing_date = toDateString(value)
  },
})

async function searchDKProjects(event: AutoCompleteCompleteEvent) {
  await loanAgreementStore.fetchDKProjectOptions(event.query)
}

async function loadData() {
  if (isEditMode.value) {
    await loanAgreementStore.fetchDKProjectOptions()
    const loanAgreement = await loanAgreementStore.fetchLoanAgreement(loanAgreementId.value)
    form.applyLoanAgreement(loanAgreement)
    return
  }

  if (sourceDKId.value && sourceDKProjectId.value) {
    const selectedProject = await loanAgreementStore.fetchDKProjectOption(
      sourceDKId.value,
      sourceDKProjectId.value,
    )
    form.values.dk_project_id = selectedProject.id
    return
  }

  await loanAgreementStore.fetchDKProjectOptions()
}

function queryString(value: unknown) {
  if (Array.isArray(value)) return typeof value[0] === 'string' ? value[0] : ''
  return typeof value === 'string' ? value : ''
}

const onSubmit = form.submit(async (values) => {
  if (isEditMode.value) {
    await loanAgreementStore.updateLoanAgreement(loanAgreementId.value, values)
    toast.success('Berhasil', 'Loan Agreement berhasil diperbarui')
    await router.push({ name: 'loan-agreement-detail', params: { id: loanAgreementId.value } })
    return
  }

  const created = await loanAgreementStore.createLoanAgreement(values)
  toast.success('Berhasil', 'Loan Agreement berhasil dibuat')
  await router.push({ name: 'loan-agreement-detail', params: { id: created.id } })
})

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader :title="pageTitle" subtitle="Lengkapi data Loan Agreement">
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push(backRoute)"
        />
      </template>
    </PageHeader>

    <form class="space-y-6" @submit.prevent="onSubmit">
      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Referensi Daftar Kegiatan dan Lender</h2>
          <p class="text-sm text-surface-500">
            Lender hanya dapat dipilih dari rincian pembiayaan Proyek Daftar Kegiatan terkait.
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Proyek Daftar Kegiatan</span>
            <AutoComplete
              v-model="dkProjectModel"
              :suggestions="loanAgreementStore.dkProjectOptions"
              option-label="label"
              placeholder="Cari tujuan atau kode Green Book"
              dropdown
              force-selection
              class="w-full"
              @complete="searchDKProjects"
            >
              <template #option="{ option }">
                <div class="space-y-1">
                  <p class="font-medium text-surface-900">{{ formatDKProjectLabel(option) }}</p>
                  <p class="text-xs text-surface-500">{{ option.daftar_kegiatan_subject }}</p>
                </div>
              </template>
            </AutoComplete>
            <small v-if="form.errors.dk_project_id" class="text-red-600">{{ form.errors.dk_project_id }}</small>
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Lender</span>
            <LenderSelect
              v-model="form.values.lender_id"
              :allowed-ids="form.allowedLenderIds.value"
              :disabled="!form.values.dk_project_id"
              placeholder="Pilih Proyek Daftar Kegiatan dulu"
            />
            <small v-if="form.errors.lender_id" class="text-red-600">{{ form.errors.lender_id }}</small>
          </label>
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Informasi Pinjaman</h2>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Kode Loan</span>
            <InputText v-model="form.values.loan_code" class="w-full" placeholder="IP-603" />
            <small v-if="form.errors.loan_code" class="text-red-600">{{ form.errors.loan_code }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Mata Uang</span>
            <CurrencySelect
              v-model="form.values.currency"
              :invalid="Boolean(form.errors.currency)"
              placeholder="Pilih mata uang pinjaman"
            />
            <small v-if="form.errors.currency" class="text-red-600">{{ form.errors.currency }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Tanggal Agreement</span>
            <DatePicker v-model="agreementDateModel" date-format="yy-mm-dd" show-icon class="w-full" />
            <small v-if="form.errors.agreement_date" class="text-red-600">{{ form.errors.agreement_date }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Tanggal Efektif</span>
            <DatePicker v-model="effectiveDateModel" date-format="yy-mm-dd" show-icon class="w-full" />
            <small v-if="form.errors.effective_date" class="text-red-600">{{ form.errors.effective_date }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Tanggal Closing Awal</span>
            <DatePicker v-model="originalClosingDateModel" date-format="yy-mm-dd" show-icon class="w-full" />
            <small v-if="form.errors.original_closing_date" class="text-red-600">
              {{ form.errors.original_closing_date }}
            </small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Tanggal Closing</span>
            <DatePicker v-model="closingDateModel" date-format="yy-mm-dd" show-icon class="w-full" />
            <small v-if="form.errors.closing_date" class="text-red-600">{{ form.errors.closing_date }}</small>
          </label>
        </div>

        <Message v-if="form.isExtended.value" severity="warn" :closable="false">
          Perpanjangan terdeteksi: +{{ form.extensionDays.value }} hari
        </Message>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <h2 class="text-lg font-semibold text-surface-950">Nilai Pinjaman</h2>
          <p class="text-sm text-surface-500">
            Konversi ke USD diisi manual oleh staf untuk mata uang selain USD.
          </p>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">
              {{ form.values.currency || 'Mata uang pinjaman' }} (mata uang lender)
            </span>
            <CurrencyInput v-model="form.values.amount_original" :currency="form.values.currency || 'USD'" />
            <small v-if="form.errors.amount_original" class="text-red-600">{{ form.errors.amount_original }}</small>
          </label>
          <label v-if="!form.isUSD.value" class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">USD</span>
            <CurrencyInput v-model="form.values.amount_usd" currency="USD" />
            <small v-if="form.errors.amount_usd" class="text-red-600">{{ form.errors.amount_usd }}</small>
          </label>
        </div>
      </section>

      <div class="sticky bottom-0 flex justify-end gap-2 border-t border-surface-200 bg-surface-50/95 py-4 backdrop-blur">
        <Button label="Batal" severity="secondary" outlined @click="router.push(backRoute)" />
        <Button type="submit" label="Simpan" icon="pi pi-save" :loading="loanAgreementStore.loading" />
      </div>
    </form>
  </section>
</template>
