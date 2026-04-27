<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import Message from 'primevue/message'
import Select from 'primevue/select'
import ToggleSwitch from 'primevue/toggleswitch'
import AbsorptionBar from '@/components/monitoring/AbsorptionBar.vue'
import KomponenTable from '@/components/monitoring/KomponenTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useMonitoringForm } from '@/composables/forms/useMonitoringForm'
import { useToast } from '@/composables/useToast'
import { useMonitoringStore } from '@/stores/monitoring.store'
import type { Quarter } from '@/types/monitoring.types'
import { formatDate } from '@/pages/loan-agreement/loan-agreement-page-utils'

const route = useRoute()
const router = useRouter()
const monitoringStore = useMonitoringStore()
const toast = useToast()
const form = useMonitoringForm()

const loanAgreementId = computed(() => String(route.params.laId ?? ''))
const monitoringId = computed(() => (route.params.id ? String(route.params.id) : ''))
const isEdit = computed(() => Boolean(monitoringId.value))
const currentLA = computed(() => monitoringStore.currentLA)
const todayString = computed(() => new Date().toISOString().slice(0, 10))
const isNotEffective = computed(() => {
  if (!currentLA.value?.effective_date) return false
  return currentLA.value.effective_date.slice(0, 10) > todayString.value
})
const quarterOptions: { label: string; value: Quarter }[] = [
  { label: 'TW1 (Apr-Jun)', value: 'TW1' },
  { label: 'TW2 (Jul-Sep)', value: 'TW2' },
  { label: 'TW3 (Okt-Des)', value: 'TW3' },
  { label: 'TW4 (Jan-Mar)', value: 'TW4' },
]

const onSubmit = form.submit(async (values) => {
  if (isNotEffective.value) return

  const saved = isEdit.value
    ? await monitoringStore.updateMonitoring(loanAgreementId.value, monitoringId.value, values)
    : await monitoringStore.createMonitoring(loanAgreementId.value, values)

  toast.success('Berhasil', `Monitoring berhasil ${isEdit.value ? 'diperbarui' : 'disimpan'}`)
  await router.push({ name: 'monitoring-list', params: { laId: loanAgreementId.value }, hash: `#${saved.id}` })
})

onMounted(async () => {
  await monitoringStore.fetchLoanAgreement(loanAgreementId.value)

  if (isEdit.value) {
    const monitoring = await monitoringStore.fetchMonitoring(loanAgreementId.value, monitoringId.value)
    if (monitoring) {
      form.applyMonitoring(monitoring)
    }
  }
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="isEdit ? 'Edit Monitoring' : 'Tambah Monitoring'"
      subtitle="Input rencana dan realisasi per triwulan"
    >
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'monitoring-list', params: { laId: loanAgreementId } })"
        />
      </template>
    </PageHeader>

    <Message v-if="isNotEffective" severity="warn" :closable="false">
      LA belum efektif - monitoring belum bisa diinput. Effective Date:
      {{ formatDate(currentLA?.effective_date ?? '') }}
    </Message>

    <form class="space-y-6" @submit.prevent="onSubmit">
      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <p class="text-xs font-semibold uppercase tracking-wide text-primary">Section 1</p>
          <h2 class="text-lg font-semibold text-surface-950">Periode dan Kurs</h2>
          <p class="text-sm text-surface-500">Triwulan mengikuti tahun anggaran PRISM; kurs diinput manual.</p>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Budget Year</span>
            <InputNumber v-model="form.values.budget_year" :use-grouping="false" class="w-full" />
            <small v-if="form.errors.budget_year" class="text-red-600">{{ form.errors.budget_year }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Quarter</span>
            <Select
              v-model="form.values.quarter"
              :options="quarterOptions"
              option-label="label"
              option-value="value"
              class="w-full"
            />
            <small v-if="form.errors.quarter" class="text-red-600">{{ form.errors.quarter }}</small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Exchange Rate USD/IDR</span>
            <InputNumber
              v-model="form.values.exchange_rate_usd_idr"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
            />
            <small v-if="form.errors.exchange_rate_usd_idr" class="text-red-600">
              {{ form.errors.exchange_rate_usd_idr }}
            </small>
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Exchange Rate LA/IDR</span>
            <InputNumber
              v-model="form.values.exchange_rate_la_idr"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
            />
            <small v-if="form.errors.exchange_rate_la_idr" class="text-red-600">
              {{ form.errors.exchange_rate_la_idr }}
            </small>
          </label>
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div>
          <p class="text-xs font-semibold uppercase tracking-wide text-primary">Section 2</p>
          <h2 class="text-lg font-semibold text-surface-950">Rencana vs Realisasi</h2>
          <p class="text-sm text-surface-500">Simpan nilai LA, USD, dan IDR secara manual tanpa auto-convert.</p>
        </div>

        <div class="overflow-auto rounded-lg border border-surface-200">
          <table class="w-full min-w-[42rem] text-left text-sm">
            <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
              <tr>
                <th class="px-4 py-3">Mata Uang</th>
                <th class="px-4 py-3">Rencana</th>
                <th class="px-4 py-3">Realisasi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-surface-100">
              <tr>
                <td class="px-4 py-3 font-medium text-surface-900">Mata Uang LA</td>
                <td class="px-4 py-3"><InputNumber v-model="form.values.planned_la" :min="0" :min-fraction-digits="2" class="w-full" /></td>
                <td class="px-4 py-3"><InputNumber v-model="form.values.realized_la" :min="0" :min-fraction-digits="2" class="w-full" /></td>
              </tr>
              <tr>
                <td class="px-4 py-3 font-medium text-surface-900">USD</td>
                <td class="px-4 py-3"><InputNumber v-model="form.values.planned_usd" :min="0" :min-fraction-digits="2" class="w-full" /></td>
                <td class="px-4 py-3"><InputNumber v-model="form.values.realized_usd" :min="0" :min-fraction-digits="2" class="w-full" /></td>
              </tr>
              <tr>
                <td class="px-4 py-3 font-medium text-surface-900">IDR</td>
                <td class="px-4 py-3"><InputNumber v-model="form.values.planned_idr" :min="0" :min-fraction-digits="2" class="w-full" /></td>
                <td class="px-4 py-3"><InputNumber v-model="form.values.realized_idr" :min="0" :min-fraction-digits="2" class="w-full" /></td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="rounded-lg bg-surface-50 p-4">
          <p class="mb-2 text-sm font-medium text-surface-700">Absorption real-time</p>
          <AbsorptionBar :pct="form.absorptionPct.value" />
        </div>
      </section>

      <section class="space-y-4 rounded-lg border border-surface-200 bg-white p-5">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p class="text-xs font-semibold uppercase tracking-wide text-primary">Section 3</p>
            <h2 class="text-lg font-semibold text-surface-950">Breakdown Komponen</h2>
            <p class="text-sm text-surface-500">Opsional; total komponen tidak wajib sama dengan level LA.</p>
          </div>
          <label class="flex items-center gap-3 text-sm font-medium text-surface-700">
            <ToggleSwitch v-model="form.showKomponen.value" />
            Tambah Breakdown per Komponen
          </label>
        </div>

        <div v-if="form.showKomponen.value" class="space-y-3">
          <KomponenTable
            v-model:komponen="form.komponen.value"
            editable
            @remove="form.removeKomponen"
          />
          <Button label="Tambah Komponen" icon="pi pi-plus" severity="secondary" outlined @click="form.addKomponen" />
        </div>
      </section>

      <div class="sticky bottom-0 flex justify-end gap-2 border-t border-surface-200 bg-surface-50/95 py-4 backdrop-blur">
        <Button
          label="Batal"
          severity="secondary"
          outlined
          @click="router.push({ name: 'monitoring-list', params: { laId: loanAgreementId } })"
        />
        <Button
          type="submit"
          :label="isEdit ? 'Perbarui' : 'Simpan'"
          icon="pi pi-save"
          :loading="monitoringStore.loading"
          :disabled="isNotEffective"
        />
      </div>
    </form>
  </section>
</template>
