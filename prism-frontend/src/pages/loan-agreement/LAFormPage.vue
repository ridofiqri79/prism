<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AutoComplete, { type AutoCompleteCompleteEvent } from 'primevue/autocomplete'
import Button from 'primevue/button'
import DatePicker from 'primevue/datepicker'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Tag from 'primevue/tag'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import CurrencySelect from '@/components/forms/CurrencySelect.vue'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useLAForm } from '@/composables/forms/useLAForm'
import { useToast } from '@/composables/useToast'
import { useLoanAgreementStore } from '@/stores/loan-agreement.store'
import type { DKProjectLoanOption, LoanAgreementPayload } from '@/types/loan-agreement.types'
import { formatApiError } from '@/utils/api-error'
import { formatDate, formatDKProjectLabel, parseDateModel, toDateString } from './loan-agreement-page-utils'

const route = useRoute()
const router = useRouter()
const loanAgreementStore = useLoanAgreementStore()
const toast = useToast()

interface DraftLoanAgreement extends LoanAgreementPayload {
  local_id: string
}

const loanAgreementId = computed(() => String(route.params.id ?? ''))
const isEditMode = computed(() => route.name === 'loan-agreement-edit')
const pageTitle = computed(() => (isEditMode.value ? 'Edit Loan Agreement' : 'Buat Loan Agreement'))
const sourceDKId = computed(() => queryString(route.query.dk_id))
const sourceDKProjectId = computed(() => queryString(route.query.dk_project_id))
const draftLoanAgreements = ref<DraftLoanAgreement[]>([])
const editingDraftIndex = ref<number | null>(null)
const savingDrafts = ref(false)
const isLoanInputExpanded = ref(true)
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
const selectedProject = computed(() => form.selectedDKProject.value)
const persistedLoanAgreements = computed(() => selectedProject.value?.loan_agreements ?? [])
const canChangeProject = computed(
  () => !isEditMode.value && draftLoanAgreements.value.length === 0 && !savingDrafts.value,
)

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

function nextDraftId() {
  return `draft-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

function createDraft(values: LoanAgreementPayload): DraftLoanAgreement {
  return { ...values, local_id: nextDraftId() }
}

function toPayload(draft: DraftLoanAgreement): LoanAgreementPayload {
  return {
    dk_project_id: draft.dk_project_id,
    lender_id: draft.lender_id,
    loan_code: draft.loan_code,
    agreement_date: draft.agreement_date,
    effective_date: draft.effective_date,
    original_closing_date: draft.original_closing_date,
    closing_date: draft.closing_date,
    currency: draft.currency,
    amount_original: draft.amount_original,
    amount_usd: draft.amount_usd,
    cumulative_disbursement: draft.cumulative_disbursement,
  }
}

function lenderNameForDraft(draft: DraftLoanAgreement) {
  return (
    selectedProject.value?.financing_details.find((detail) => detail.lender?.id === draft.lender_id)?.lender
      ?.name ?? '-'
  )
}

function clearDraftEditor() {
  form.reset({
    dk_project_id: form.values.dk_project_id,
  })
  editingDraftIndex.value = null
}

function editDraft(index: number) {
  const draft = draftLoanAgreements.value[index]
  if (!draft) return

  editingDraftIndex.value = index
  Object.assign(form.values, {
    dk_project_id: draft.dk_project_id,
    lender_id: draft.lender_id,
    loan_code: draft.loan_code,
    agreement_date: draft.agreement_date,
    effective_date: draft.effective_date,
    original_closing_date: draft.original_closing_date,
    closing_date: draft.closing_date,
    currency: draft.currency,
    amount_original: draft.amount_original,
    amount_usd: draft.amount_usd,
    cumulative_disbursement: draft.cumulative_disbursement,
  })
}

function removeDraft(index: number) {
  draftLoanAgreements.value.splice(index, 1)
  if (editingDraftIndex.value === index) {
    editingDraftIndex.value = null
  } else if (editingDraftIndex.value !== null && editingDraftIndex.value > index) {
    editingDraftIndex.value -= 1
  }
}

function clearDrafts() {
  draftLoanAgreements.value = []
  editingDraftIndex.value = null
}

async function refreshCurrentProjectOption(projectId: string) {
  const dkId = selectedProject.value?.dk_id ?? sourceDKId.value
  if (dkId) {
    await loanAgreementStore.fetchDKProjectOption(dkId, projectId)
    return
  }

  await loanAgreementStore.fetchDKProjectOptions()
}

const addDraftToStack = form.submit(async (values) => {
  const normalizedLoanCode = values.loan_code.trim().toLowerCase()
  const duplicateIndex = draftLoanAgreements.value.findIndex(
    (draft, index) =>
      index !== editingDraftIndex.value && draft.loan_code.trim().toLowerCase() === normalizedLoanCode,
  )
  const duplicatePersisted = persistedLoanAgreements.value.some(
    (loanAgreement) => loanAgreement.loan_code.trim().toLowerCase() === normalizedLoanCode,
  )

  if (duplicateIndex >= 0) {
    toast.warn('Loan Agreement Duplikat', 'Kode pinjaman sudah ada di daftar draft.')
    return
  }
  if (duplicatePersisted) {
    toast.warn('Loan Agreement Duplikat', 'Kode pinjaman sudah tersimpan untuk proyek ini.')
    return
  }

  const draft = createDraft(values)
  if (editingDraftIndex.value !== null) {
    draftLoanAgreements.value.splice(editingDraftIndex.value, 1, draft)
    editingDraftIndex.value = null
  } else {
    draftLoanAgreements.value.push(draft)
  }

  toast.success('Berhasil', 'Loan Agreement masuk ke daftar draft')
  clearDraftEditor()
})

const updateExistingLoanAgreement = form.submit(async (values) => {
  if (isEditMode.value) {
    await loanAgreementStore.updateLoanAgreement(loanAgreementId.value, values)
    toast.success('Berhasil', 'Loan Agreement berhasil diperbarui')
    await router.push({ name: 'loan-agreement-detail', params: { id: loanAgreementId.value } })
    return
  }
})

async function saveDrafts() {
  if (draftLoanAgreements.value.length === 0) {
    toast.warn('Belum Ada Draft', 'Tambahkan Loan Agreement ke daftar terlebih dahulu.')
    return
  }

  savingDrafts.value = true
  try {
    while (draftLoanAgreements.value.length > 0) {
      const draft = draftLoanAgreements.value[0]
      if (!draft) break

      await loanAgreementStore.createLoanAgreement(toPayload(draft))
      draftLoanAgreements.value.shift()
      await refreshCurrentProjectOption(draft.dk_project_id)
    }

    toast.success(
      'Berhasil',
      `${selectedProject.value?.project_name ?? 'Loan Agreement'} berhasil disimpan ke daftar.`,
    )
  } catch (error) {
    toast.warn('Gagal Menyimpan', formatApiError(error, 'Gagal menyimpan Loan Agreement'), 12000)
  } finally {
    editingDraftIndex.value = null
    savingDrafts.value = false
  }
}

async function onSubmit() {
  if (isEditMode.value) {
    await updateExistingLoanAgreement()
    return
  }

  await addDraftToStack()
}

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
          <h2 class="text-lg font-semibold text-surface-950">Referensi Daftar Kegiatan</h2>
          <p class="text-sm text-surface-500">
            Pilih Proyek Daftar Kegiatan sebagai referensi untuk semua Loan Agreement yang akan ditambahkan.
          </p>
        </div>

        <div class="grid gap-4">
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
              :disabled="!canChangeProject"
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
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-surface-950">Informasi Pinjaman</h2>
            <p class="text-sm text-surface-500">
              Lengkapi lender, tanggal, mata uang, dan nilai pinjaman untuk satu record Loan Agreement.
            </p>
          </div>
          <Button
            type="button"
            :label="isLoanInputExpanded ? 'Tutup' : 'Buka'"
            :icon="isLoanInputExpanded ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
            severity="secondary"
            outlined
            @click="isLoanInputExpanded = !isLoanInputExpanded"
          />
        </div>

        <div v-if="isLoanInputExpanded" class="space-y-5">
          <div class="grid gap-4 md:grid-cols-2">
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
              <small class="text-surface-500">Opsional. Isi hanya jika pinjaman mengalami perpanjangan.</small>
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

          <div class="space-y-4 border-t border-surface-100 pt-5">
            <div>
              <h3 class="font-semibold text-surface-950">Nilai Pinjaman</h3>
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
              <label class="block space-y-2">
                <span class="text-sm font-medium text-surface-700">
                  Cumulative Disbursement ({{ form.values.currency || 'Mata uang pinjaman' }})
                </span>
                <CurrencyInput
                  v-model="form.values.cumulative_disbursement"
                  :currency="form.values.currency || 'USD'"
                />
                <small class="text-surface-500">
                  Nilai kumulatif berdasarkan mata uang pinjaman yang dipilih.
                </small>
                <small v-if="form.errors.cumulative_disbursement" class="text-red-600">
                  {{ form.errors.cumulative_disbursement }}
                </small>
              </label>
              <label v-if="!form.isUSD.value" class="block space-y-2">
                <span class="text-sm font-medium text-surface-700">USD</span>
                <CurrencyInput v-model="form.values.amount_usd" currency="USD" />
                <small v-if="form.errors.amount_usd" class="text-red-600">{{ form.errors.amount_usd }}</small>
              </label>
            </div>
          </div>
        </div>
      </section>

      <section
        v-if="!isEditMode && selectedProject"
        class="space-y-4 rounded-lg border border-surface-200 bg-white p-5"
      >
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-lg font-semibold text-surface-950">Daftar Loan Agreement</h2>
            <p class="text-sm text-surface-500">
              Loan Agreement yang ditambahkan akan muncul di daftar ini sebelum disimpan.
            </p>
          </div>
          <Button
            v-if="draftLoanAgreements.length > 0"
            type="button"
            label="Kosongkan Daftar"
            icon="pi pi-trash"
            severity="secondary"
            outlined
            @click="clearDrafts"
          />
        </div>

        <div class="grid gap-4 lg:grid-cols-2">
          <div class="space-y-3">
            <div class="flex items-center justify-between gap-2">
              <h3 class="font-semibold text-surface-950">Draft</h3>
              <Tag :value="`${draftLoanAgreements.length} Draft`" severity="info" rounded />
            </div>
            <div
              v-if="draftLoanAgreements.length === 0"
              class="rounded-lg border border-dashed border-surface-200 p-4 text-sm text-surface-500"
            >
              Belum ada Loan Agreement di draft.
            </div>
            <article
              v-for="(draft, index) in draftLoanAgreements"
              :key="draft.local_id"
              class="space-y-3 rounded-lg border border-surface-200 p-4"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0 space-y-1">
                  <p class="font-semibold text-surface-950">{{ draft.loan_code }}</p>
                  <p class="text-sm text-surface-600">{{ lenderNameForDraft(draft) }}</p>
                  <p class="text-xs text-surface-500">
                    {{ formatDate(draft.agreement_date) }} - {{ formatDate(draft.closing_date) }}
                  </p>
                </div>
                <div class="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    label="Ubah"
                    icon="pi pi-pencil"
                    size="small"
                    outlined
                    @click="editDraft(index)"
                  />
                  <Button
                    type="button"
                    label="Hapus"
                    icon="pi pi-trash"
                    size="small"
                    severity="danger"
                    outlined
                    @click="removeDraft(index)"
                  />
                </div>
              </div>
              <div class="grid gap-3 text-sm md:grid-cols-2">
                <div>
                  <p class="text-xs uppercase tracking-wide text-surface-500">Mata Uang</p>
                  <p class="font-medium text-surface-900">{{ draft.currency }}</p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-surface-500">Nilai Asli</p>
                  <p class="font-medium text-surface-900">
                    <CurrencyDisplay :amount="draft.amount_original" :currency="draft.currency" />
                  </p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-surface-500">Cumulative Disbursement</p>
                  <p class="font-medium text-surface-900">
                    <CurrencyDisplay :amount="draft.cumulative_disbursement" :currency="draft.currency" />
                  </p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-surface-500">USD</p>
                  <p class="font-medium text-surface-900">
                    <CurrencyDisplay :amount="draft.amount_usd" currency="USD" />
                  </p>
                </div>
                <div>
                  <p class="text-xs uppercase tracking-wide text-surface-500">Status</p>
                  <p class="font-medium text-surface-900">
                    {{
                      draft.original_closing_date && draft.original_closing_date !== draft.closing_date
                        ? 'Perpanjangan'
                        : 'Normal'
                    }}
                  </p>
                </div>
              </div>
            </article>
          </div>

          <div class="space-y-3">
            <div class="flex items-center justify-between gap-2">
              <h3 class="font-semibold text-surface-950">Yang Sudah Tersimpan</h3>
              <Tag :value="`${persistedLoanAgreements.length} Tersimpan`" severity="success" rounded />
            </div>
            <div
              v-if="persistedLoanAgreements.length === 0"
              class="rounded-lg border border-dashed border-surface-200 p-4 text-sm text-surface-500"
            >
              Belum ada Loan Agreement tersimpan untuk proyek ini.
            </div>
            <article
              v-for="loanAgreement in persistedLoanAgreements"
              :key="loanAgreement.id"
              class="rounded-lg border border-surface-200 p-4"
            >
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div class="min-w-0 space-y-1">
                  <p class="font-semibold text-surface-950">{{ loanAgreement.loan_code }}</p>
                  <p class="text-xs uppercase tracking-wide text-surface-500">Cumulative Disbursement</p>
                  <p class="text-sm font-medium text-surface-700">
                    <CurrencyDisplay
                      :amount="loanAgreement.cumulative_disbursement"
                      :currency="loanAgreement.currency"
                    />
                  </p>
                  <Tag value="Tersimpan" severity="success" rounded />
                </div>
                <Button
                  as="router-link"
                  :to="{ name: 'loan-agreement-detail', params: { id: loanAgreement.id } }"
                  label="Detail"
                  icon="pi pi-external-link"
                  size="small"
                  outlined
                />
              </div>
            </article>
          </div>
        </div>
      </section>

      <div class="sticky bottom-0 flex flex-wrap justify-end gap-2 border-t border-surface-200 bg-surface-50/95 py-4 backdrop-blur">
        <Button type="button" label="Batal" severity="secondary" outlined @click="router.push(backRoute)" />
        <Button
          v-if="isEditMode"
          type="submit"
          label="Simpan Perubahan"
          icon="pi pi-save"
          :loading="loanAgreementStore.loading"
        />
        <Button
          v-else
          type="submit"
          :label="editingDraftIndex === null ? 'Tambahkan ke Daftar' : 'Perbarui Draft'"
          icon="pi pi-plus"
          severity="secondary"
          outlined
        />
        <Button
          v-if="!isEditMode"
          type="button"
          label="Simpan Daftar Loan Agreement"
          icon="pi pi-save"
          :disabled="draftLoanAgreements.length === 0"
          :loading="savingDrafts"
          @click="saveDrafts"
        />
      </div>
    </form>
  </section>
</template>
