<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
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
  emit(
    'update:rows',
    props.rows.map((row, rowIndex) => (rowIndex === index ? { ...row, ...patch } : row)),
  )
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[90rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">Institution</th>
          <th class="px-4 py-3">Currency</th>
          <th class="px-4 py-3">Loan Original</th>
          <th class="px-4 py-3">Grant Original</th>
          <th class="px-4 py-3">Counterpart Original</th>
          <th class="px-4 py-3">Loan USD</th>
          <th class="px-4 py-3">Grant USD</th>
          <th class="px-4 py-3">Counterpart USD</th>
          <th class="px-4 py-3">Remarks</th>
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
            <InputText
              :model-value="row.currency"
              class="w-full"
              maxlength="3"
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
              :model-value="row.amount_usd"
              @update:model-value="updateRow(index, { amount_usd: $event })"
            />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              :model-value="row.grant_usd"
              @update:model-value="updateRow(index, { grant_usd: $event })"
            />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              :model-value="row.counterpart_usd"
              @update:model-value="updateRow(index, { counterpart_usd: $event })"
            />
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
            Belum ada loan allocation.
          </td>
        </tr>
      </tbody>
    </table>
    <div class="border-t border-surface-200 p-3">
      <Button label="Tambah Baris" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
