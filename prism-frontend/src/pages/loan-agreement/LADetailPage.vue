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
import { formatDate, formatPercent, formatRatio } from './loan-agreement-page-utils'

const route = useRoute()
const router = useRouter()
const loanAgreementStore = useLoanAgreementStore()
const toast = useToast()
const { can } = usePermission()

const loanAgreementId = computed(() => String(route.params.id ?? ''))
const loanAgreement = computed(() => loanAgreementStore.currentLoanAgreement)
const relatedDKProjects = computed(() => loanAgreement.value?.dk_projects ?? [])
const relatedDKHeaders = computed(() => {
  const groups = new Map<
    string,
    {
      id: string
      subject: string
      date: string
      letterNumber?: string | null
      projects: typeof relatedDKProjects.value
    }
  >()
  relatedDKProjects.value.forEach((project) => {
    const header = project.daftar_kegiatan
    const existing = groups.get(header.id)
    if (existing) {
      existing.projects.push(project)
      return
    }
    groups.set(header.id, {
      id: header.id,
      subject: header.subject,
      date: header.date,
      letterNumber: header.letter_number,
      projects: [project],
    })
  })
  return [...groups.values()]
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
            <StatusBadge
              v-if="loanAgreement.performance_status"
              :status="loanAgreement.performance_status"
              domain="loan_agreement"
            />
            <StatusBadge v-if="loanAgreement.is_extended" status="Extended" />
            <span v-if="loanAgreement.is_extended" class="text-sm font-medium text-prism-gold-deep">
              {{ loanAgreement.extension_days }} hari
            </span>
          </div>
          <dl class="grid gap-4 md:grid-cols-2">
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Kode Pinjaman
              </dt>
              <dd class="mt-1 font-medium text-surface-900">{{ loanAgreement.loan_code }}</dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">Lender</dt>
              <dd class="mt-1 font-medium text-surface-900">{{ loanAgreement.lender.name }}</dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Tanggal Perjanjian
              </dt>
              <dd class="mt-1 font-medium text-surface-900">
                {{ formatDate(loanAgreement.agreement_date) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Tanggal Efektif
              </dt>
              <dd class="mt-1 font-medium text-surface-900">
                {{ formatDate(loanAgreement.effective_date) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Tanggal Penutupan Awal
              </dt>
              <dd class="mt-1 font-medium text-surface-900">
                {{ formatDate(loanAgreement.original_closing_date) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Tanggal Penutupan
              </dt>
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
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
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
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">USD</dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                <CurrencyDisplay :amount="loanAgreement.amount_usd" currency="USD" />
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Cumulative Disbursement ({{ loanAgreement.currency }})
              </dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                <CurrencyDisplay
                  :amount="loanAgreement.cumulative_disbursement"
                  :currency="loanAgreement.currency"
                />
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Cumulative Disbursement USD
              </dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                <CurrencyDisplay
                  v-if="loanAgreement.cumulative_disbursement_usd !== null"
                  :amount="loanAgreement.cumulative_disbursement_usd"
                  currency="USD"
                />
                <span v-else>-</span>
              </dd>
            </div>
          </dl>
        </section>

        <section class="rounded-lg border border-surface-200 bg-white p-5">
          <h2 class="mb-4 text-lg font-semibold text-surface-950">Perhitungan Kinerja</h2>
          <dl class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Disbursement Ratio
              </dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                {{ formatPercent(loanAgreement.disbursement_ratio) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                Estimated Time Ratio
              </dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                {{ formatPercent(loanAgreement.estimated_time_ratio) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                PV (DR/ETR)
              </dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                {{ formatRatio(loanAgreement.performance_value) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs font-semibold uppercase tracking-wide text-surface-500">Status</dt>
              <dd class="mt-1 text-xl font-semibold text-surface-950">
                {{ loanAgreement.performance_status ?? '-' }}
              </dd>
            </div>
          </dl>
          <p v-if="loanAgreement.kurs_cut_off_date" class="mt-4 text-sm text-surface-500">
            Cut off Kurs Tengah BI: {{ formatDate(loanAgreement.kurs_cut_off_date) }}
          </p>
        </section>
      </div>

      <aside class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <h2 class="text-lg font-semibold text-surface-950">Relasi Alur Kerja</h2>
        <div v-if="relatedDKHeaders.length === 0" class="text-sm text-surface-500">
          Belum ada relasi Proyek Daftar Kegiatan.
        </div>
        <article
          v-for="header in relatedDKHeaders"
          :key="header.id"
          class="space-y-3 rounded-md border border-surface-200 p-3"
        >
          <div>
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Daftar Kegiatan
            </p>
            <p class="mt-1 font-medium text-surface-900">
              {{ header.letterNumber || header.subject }}
            </p>
            <p class="text-xs text-surface-500">{{ formatDate(header.date) }}</p>
          </div>
          <div class="space-y-2">
            <div v-for="project in header.projects" :key="project.id" class="space-y-1">
              <p class="text-sm font-medium text-surface-900">{{ project.project_name }}</p>
              <p class="text-xs text-surface-500">{{ project.gb_codes }}</p>
              <p class="text-sm text-surface-700">
                Alokasi:
                <CurrencyDisplay
                  :amount="project.allocation_original"
                  :currency="loanAgreement.currency"
                />
                <span class="text-surface-400"> / </span>
                <CurrencyDisplay :amount="project.allocation_usd" currency="USD" />
              </p>
            </div>
          </div>
          <Button
            as="router-link"
            :to="{ name: 'daftar-kegiatan-detail', params: { id: header.id } }"
            label="Lihat Daftar Kegiatan"
            icon="pi pi-list"
            outlined
            class="w-full"
          />
        </article>
      </aside>
    </section>
  </section>
</template>
