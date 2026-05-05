<script setup lang="ts" generic="T extends MultiCurrencyRow">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import CurrencySelect from '@/components/forms/CurrencySelect.vue'

/**
 * Minimal shape required by the multi-currency table.
 * Both DKFinancingDetailPayload and DKLoanAllocationPayload satisfy this.
 */
export interface MultiCurrencyRow {
  currency: string
  amount_original: number
  grant_original: number
  counterpart_original: number
  amount_usd: number
  grant_usd: number
  counterpart_usd: number
  remarks?: string | null
}

const props = defineProps<{
  rows: T[]
  /** Placeholder text shown when the table is empty */
  emptyText?: string
}>()

const emit = defineEmits<{
  'update:rows': [value: T[]]
  add: []
  remove: [index: number]
}>()

function updateRow(index: number, patch: Partial<MultiCurrencyRow>) {
  const next = props.rows.map((row, rowIndex) => {
    const value: T = rowIndex === index ? { ...row, ...patch } : { ...row }
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

function isUSD(row: T) {
  return (row.currency || 'USD').trim().toUpperCase() === 'USD'
}
</script>

<template>
  <div class="overflow-x-auto rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[90rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <!--
            First-column header provided by the consumer.
            Usage: <template #first-col-header>Lender</template>
          -->
          <th class="px-4 py-3">
            <slot name="first-col-header" />
          </th>
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
          <!--
            First-column cell — the selector differs per table.
            Usage: <template #first-col="{ row, index }">...</template>
          -->
          <td class="px-4 py-3">
            <slot name="first-col" :row="row" :index="index" />
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
            {{ emptyText ?? 'Belum ada data.' }}
          </td>
        </tr>
      </tbody>
    </table>
    <div class="border-t border-surface-200 p-3">
      <Button label="Tambah Baris" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
