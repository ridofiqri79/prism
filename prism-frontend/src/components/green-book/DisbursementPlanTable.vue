<script setup lang="ts">
import { computed, ref } from 'vue'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import type { GBDisbursementPlan, GBDisbursementPlanPayload } from '@/types/green-book.types'

const props = withDefaults(
  defineProps<{
    rows: GBDisbursementPlanPayload[] | GBDisbursementPlan[]
    selectedCurrency?: string
    error?: string
    editable?: boolean
  }>(),
  {
    selectedCurrency: 'USD',
    error: '',
    editable: true,
  },
)

const emit = defineEmits<{
  'update:rows': [value: GBDisbursementPlanPayload[]]
  addYear: [year: number]
  updateYear: [index: number, year: number]
  remove: [index: number]
}>()

const newYear = ref(new Date().getFullYear())
const displayCurrency = computed(() => (props.selectedCurrency || 'USD').trim().toUpperCase())
const total = computed(() => props.rows.reduce((sum, row) => sum + row.amount_usd, 0))

function toPayload(row: GBDisbursementPlanPayload | GBDisbursementPlan): GBDisbursementPlanPayload {
  return {
    year: row.year,
    amount_usd: row.amount_usd,
  }
}

function updateAmount(index: number, amount: number) {
  const next = props.rows.map((row, rowIndex) =>
    rowIndex === index ? { ...toPayload(row), amount_usd: amount } : toPayload(row),
  )

  emit('update:rows', next)
}
</script>

<template>
  <div class="space-y-3">
    <div v-if="editable" class="flex flex-wrap items-end gap-3 rounded-lg border border-surface-200 bg-white p-3">
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Tahun</span>
        <InputNumber v-model="newYear" :use-grouping="false" input-class="w-32" />
      </label>
      <Button label="Tambah Tahun" icon="pi pi-plus" outlined @click="emit('addYear', newYear)" />
      <small v-if="error" class="text-red-600">{{ error }}</small>
    </div>

    <div class="overflow-x-auto rounded-lg border border-surface-200 bg-white">
      <table class="w-full min-w-[32rem] text-sm">
        <thead class="bg-surface-50 text-xs font-semibold uppercase tracking-wide text-surface-500">
          <tr>
            <th class="px-4 py-3 text-left">Tahun</th>
            <th class="px-4 py-3 text-right">Nilai ({{ displayCurrency }})</th>
            <th v-if="editable" class="w-24 px-4 py-3"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-surface-100">
          <tr v-for="(row, index) in rows" :key="index">
            <td class="px-4 py-2.5 text-sm text-surface-800">
              <InputNumber
                v-if="editable"
                :model-value="row.year"
                :use-grouping="false"
                input-class="w-32"
                @update:model-value="emit('updateYear', index, Number($event ?? 0))"
              />
              <span v-else>{{ row.year }}</span>
            </td>
            <td class="px-4 py-2.5 text-right text-sm text-surface-800">
              <CurrencyInput
                v-if="editable"
                :model-value="row.amount_usd"
                :currency="displayCurrency"
                @update:model-value="updateAmount(index, $event)"
              />
              <CurrencyDisplay v-else :amount="row.amount_usd" :currency="displayCurrency" />
            </td>
            <td v-if="editable" class="px-4 py-2.5 text-right">
              <Button
                icon="pi pi-trash"
                severity="danger"
                text
                rounded
                aria-label="Hapus tahun disbursement"
                @click="emit('remove', index)"
              />
            </td>
          </tr>
          <tr v-if="rows.length === 0">
            <td :colspan="editable ? 3 : 2" class="px-4 py-6 text-center text-sm text-surface-500">
              Belum ada rencana disbursement.
            </td>
          </tr>
        </tbody>
        <tfoot class="border-t border-surface-200 bg-surface-50 text-sm font-semibold">
          <tr>
            <td class="px-4 py-2.5">Grand Total</td>
            <td class="px-4 py-2.5 text-right"><CurrencyDisplay :amount="total" :currency="displayCurrency" /></td>
            <td v-if="editable"></td>
          </tr>
        </tfoot>
      </table>
    </div>
  </div>
</template>
