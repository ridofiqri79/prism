<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { useLoanAgreementStore } from '@/stores/loan-agreement.store'
import { formatDate, formatDKProjectLabel } from './loan-agreement-page-utils'

const route = useRoute()
const router = useRouter()
const loanAgreementStore = useLoanAgreementStore()
const toast = useToast()
const { can } = usePermission()

const loanAgreementId = computed(() => String(route.params.id ?? ''))
const loanAgreement = computed(() => loanAgreementStore.currentLoanAgreement)
const dkProject = computed(() => {
  const id = loanAgreement.value?.dk_project?.id ?? loanAgreement.value?.dk_project_id
  return id ? loanAgreementStore.dkProjectOptionMap.get(id) : null
})

async function deleteLoanAgreement() {
  await loanAgreementStore.deleteLoanAgreement(loanAgreementId.value)
  toast.success('Berhasil', 'Loan Agreement berhasil dihapus')
  await router.push({ name: 'loan-agreements' })
}

onMounted(() => {
  void Promise.all([
    loanAgreementStore.fetchDKProjectOptions(),
    loanAgreementStore.fetchLoanAgreement(loanAgreementId.value),
  ])
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="loanAgreement?.loan_code || 'Detail Loan Agreement'"
      subtitle="Detail Loan Agreement dan status perpanjangan"
    >
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'loan-agreements' })"
        />
        <Button
          v-if="can('loan_agreement', 'update') && loanAgreement"
          as="router-link"
          :to="{ name: 'loan-agreement-edit', params: { id: loanAgreement.id } }"
          label="Edit"
          icon="pi pi-pencil"
          severity="secondary"
          outlined
        />
        <Button
          v-if="can('loan_agreement', 'delete')"
          label="Hapus"
          icon="pi pi-trash"
          severity="danger"
          outlined
          @click="deleteLoanAgreement"
        />
      </template>
    </PageHeader>

    <section v-if="loanAgreement" class="grid gap-6 lg:grid-cols-[1fr_22rem]">
      <div class="space-y-6">
        <section class="rounded-lg border border-surface-200 bg-white p-5">
          <div class="mb-4 flex flex-wrap items-center gap-3">
            <h2 class="text-lg font-semibold text-surface-950">Informasi Pinjaman</h2>
            <StatusBadge v-if="loanAgreement.is_extended" status="Extended" />
            <span v-if="loanAgreement.is_extended" class="text-sm font-medium text-prism-gold-deep">
              {{ loanAgreement.extension_days }} hari
            </span>
          </div>
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">Kode Pinjaman</dt>
              <dd class="mt-1 font-medium text-surface-900">{{ loanAgreement.loan_code }}</dd>
            </div>
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">Lender</dt>
              <dd class="mt-1 font-medium text-surface-900">{{ loanAgreement.lender.name }}</dd>
            </div>
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">Tanggal Perjanjian</dt>
              <dd class="mt-1 font-medium text-surface-900">
                {{ formatDate(loanAgreement.agreement_date) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">Tanggal Efektif</dt>
              <dd class="mt-1 font-medium text-surface-900">
                {{ formatDate(loanAgreement.effective_date) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">Tanggal Penutupan Awal</dt>
              <dd class="mt-1 font-medium text-surface-900">
                {{ formatDate(loanAgreement.original_closing_date) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">Tanggal Penutupan</dt>
              <dd class="mt-1 font-medium text-surface-900">
                {{ formatDate(loanAgreement.closing_date) }}
              </dd>
            </div>
          </dl>
        </section>

        <section class="rounded-lg border border-surface-200 bg-white p-5">
          <h2 class="mb-4 text-lg font-semibold text-surface-950">Nilai Pinjaman</h2>
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">
                {{ loanAgreement.currency }}
              </dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                <CurrencyDisplay
                  :amount="loanAgreement.amount_original"
                  :currency="loanAgreement.currency"
                />
              </dd>
            </div>
            <div>
              <dt class="text-xs uppercase tracking-wide text-surface-500">USD</dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                <CurrencyDisplay :amount="loanAgreement.amount_usd" currency="USD" />
              </dd>
            </div>
          </dl>
        </section>
      </div>

      <aside class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <h2 class="text-lg font-semibold text-surface-950">Relasi Alur Kerja</h2>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Proyek Daftar Kegiatan</p>
          <p class="mt-1 font-medium text-surface-900">{{ formatDKProjectLabel(dkProject) }}</p>
        </div>
        <Button
          v-if="dkProject?.dk_id"
          as="router-link"
          :to="{ name: 'daftar-kegiatan-detail', params: { id: dkProject.dk_id } }"
          label="Lihat Proyek Daftar Kegiatan"
          icon="pi pi-list"
          outlined
          class="w-full"
        />
        <Button
          as="router-link"
          :to="{ name: 'monitoring-list', params: { laId: loanAgreement.id } }"
          label="Lihat Monitoring"
          icon="pi pi-chart-line"
          class="w-full"
        />
      </aside>
    </section>
  </section>
</template>
