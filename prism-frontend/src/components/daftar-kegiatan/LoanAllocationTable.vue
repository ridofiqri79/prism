<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import CurrencySelect from '@/components/forms/CurrencySelect.vue'
import InstitutionSelect from '@/components/forms/InstitutionSelect.vue'
import type { DKLoanAllocationPayload } from '@/types/daftar-kegiatan.types'

const props = defineProps<{
  rows: DKLoanAllocationPayload[]
}>()

const emit = defineEmits<{
  'update:rows': [value: DKLoanAllocationPayload[]]
  add: []
  remove: [index: number]
}>()

function updateRow(index: number, patch: Partial<DKLoanAllocationPayload>) {
  const next = props.rows.map((row, rowIndex) => {
    const value = rowIndex === index ? { ...row, ...patch } : { ...row }
    value.currency = (value.currency || 'USD').trim().toUpperCase()
    if (value.currency === 'USD') {
      if (value.amount_original === 0 && value.amount_usd !== 0) value.amount_original = value.amount_usd
      if (value.grant_original === 0 && value.grant_usd !== 0) value.grant_original = value.grant_usd
      if (value.counterpart_original === 0 && value.counterpart_usd !== 0) {
        value.counterpart_original = value.counterpart_usd
      }
      value.amount_usd = value.amount_original
      value.grant_usd = value.grant_original
      value.counterpart_usd = value.counterpart_original
    }
    return value
  })

  emit('update:rows', next)
}

function isUSD(row: DKLoanAllocationPayload) {
  return (row.currency || 'USD').trim().toUpperCase() === 'USD'
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[90rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">Instansi</th>
          <th class="px-4 py-3">Mata Uang</th>
          <th class="px-4 py-3">Pinjaman Original</th>
          <th class="px-4 py-3">Hibah Original</th>
          <th class="px-4 py-3">Counterpart Original</th>
          <th class="px-4 py-3">Pinjaman USD</th>
          <th class="px-4 py-3">Hibah USD</th>
          <th class="px-4 py-3">Counterpart USD</th>
          <th class="px-4 py-3">Catatan</th>
          <th class="w-20 px-4 py-3"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="(row, index) in rows" :key="index">
          <td class="px-4 py-3">
            <InstitutionSelect
              :model-value="row.institution_id"
              @update:model-value="updateRow(index, { institution_id: String($event ?? '') })"
            />
          </td>
          <td class="px-4 py-3">
            <CurrencySelect
              :model-value="row.currency"
              placeholder="Pilih mata uang"
              @update:model-value="updateRow(index, { currency: String($event ?? '').toUpperCase() })"
            />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              :model-value="row.amount_original"
              :currency="row.currency"
              @update:model-value="updateRow(index, { amount_original: $event })"
            />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              :model-value="row.grant_original"
              :currency="row.currency"
              @update:model-value="updateRow(index, { grant_original: $event })"
            />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              :model-value="row.counterpart_original"
              :currency="row.currency"
              @update:model-value="updateRow(index, { counterpart_original: $event })"
            />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="!isUSD(row)"
              :model-value="row.amount_usd"
              @update:model-value="updateRow(index, { amount_usd: $event })"
            />
            <CurrencyInput v-else :model-value="row.amount_usd" disabled />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="!isUSD(row)"
              :model-value="row.grant_usd"
              @update:model-value="updateRow(index, { grant_usd: $event })"
            />
            <CurrencyInput v-else :model-value="row.grant_usd" disabled />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="!isUSD(row)"
              :model-value="row.counterpart_usd"
              @update:model-value="updateRow(index, { counterpart_usd: $event })"
            />
            <CurrencyInput v-else :model-value="row.counterpart_usd" disabled />
          </td>
          <td class="px-4 py-3">
            <InputText
              :model-value="row.remarks ?? ''"
              class="w-full"
              @update:model-value="updateRow(index, { remarks: String($event ?? '') })"
            />
          </td>
          <td class="px-4 py-3 text-right">
            <Button icon="pi pi-trash" severity="danger" text rounded @click="emit('remove', index)" />
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td colspan="10" class="px-4 py-6 text-center text-surface-500">
            Belum ada alokasi pinjaman.
          </td>
        </tr>
      </tbody>
    </table>
    <div class="border-t border-surface-200 p-3">
      <Button label="Tambah Baris" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
