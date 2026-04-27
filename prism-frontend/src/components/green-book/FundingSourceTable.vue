<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import InstitutionSelect from '@/components/forms/InstitutionSelect.vue'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import type { GBFundingSource, GBFundingSourcePayload } from '@/types/green-book.types'

const props = withDefaults(
  defineProps<{
    rows: GBFundingSourcePayload[] | GBFundingSource[]
    editable?: boolean
  }>(),
  {
    editable: true,
  },
)

const emit = defineEmits<{
  'update:rows': [value: GBFundingSourcePayload[]]
  add: []
  remove: [index: number]
}>()

const totals = computed(() =>
  props.rows.reduce(
    (sum, row) => ({
      loan_usd: sum.loan_usd + row.loan_usd,
      grant_usd: sum.grant_usd + row.grant_usd,
      local_usd: sum.local_usd + row.local_usd,
    }),
    { loan_usd: 0, grant_usd: 0, local_usd: 0 },
  ),
)

function toPayload(row: GBFundingSourcePayload | GBFundingSource): GBFundingSourcePayload {
  if ('lender_id' in row) {
    return {
      lender_id: row.lender_id,
      institution_id: row.institution_id ?? null,
      loan_usd: row.loan_usd,
      grant_usd: row.grant_usd,
      local_usd: row.local_usd,
    }
  }

  return {
    lender_id: row.lender.id,
    institution_id: row.institution?.id ?? null,
    loan_usd: row.loan_usd,
    grant_usd: row.grant_usd,
    local_usd: row.local_usd,
  }
}

function rowLenderName(row: GBFundingSourcePayload | GBFundingSource) {
  return 'lender' in row ? row.lender.name : '-'
}

function rowInstitutionName(row: GBFundingSourcePayload | GBFundingSource) {
  return 'institution' in row ? (row.institution?.name ?? '-') : '-'
}

function updateRow(index: number, patch: Partial<GBFundingSourcePayload>) {
  const next = props.rows.map((row, rowIndex) =>
    rowIndex === index ? { ...toPayload(row), ...patch } : toPayload(row),
  )

  emit('update:rows', next)
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[58rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">Lender</th>
          <th class="px-4 py-3">Instansi</th>
          <th class="px-4 py-3">Pinjaman USD</th>
          <th class="px-4 py-3">Hibah USD</th>
          <th class="px-4 py-3">Local USD</th>
          <th v-if="editable" class="w-24 px-4 py-3"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="(row, index) in rows" :key="index">
          <td class="px-4 py-3">
            <LenderSelect
              v-if="editable"
              :model-value="toPayload(row).lender_id"
              @update:model-value="updateRow(index, { lender_id: String($event ?? '') })"
            />
            <span v-else>{{ rowLenderName(row) }}</span>
          </td>
          <td class="px-4 py-3">
            <InstitutionSelect
              v-if="editable"
              :model-value="toPayload(row).institution_id ?? null"
              @update:model-value="updateRow(index, { institution_id: String($event ?? '') || null })"
            />
            <span v-else>{{ rowInstitutionName(row) }}</span>
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="row.loan_usd"
              @update:model-value="updateRow(index, { loan_usd: $event })"
            />
            <CurrencyDisplay v-else :amount="row.loan_usd" />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="row.grant_usd"
              @update:model-value="updateRow(index, { grant_usd: $event })"
            />
            <CurrencyDisplay v-else :amount="row.grant_usd" />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="row.local_usd"
              @update:model-value="updateRow(index, { local_usd: $event })"
            />
            <CurrencyDisplay v-else :amount="row.local_usd" />
          </td>
          <td v-if="editable" class="px-4 py-3 text-right">
            <Button
              icon="pi pi-trash"
              severity="danger"
              text
              rounded
              aria-label="Hapus funding source"
              @click="emit('remove', index)"
            />
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td :colspan="editable ? 6 : 5" class="px-4 py-6 text-center text-surface-500">
            Belum ada funding source.
          </td>
        </tr>
      </tbody>
      <tfoot class="border-t border-surface-200 bg-surface-50 font-semibold">
        <tr>
          <td class="px-4 py-3" colspan="2">Total</td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.loan_usd" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.grant_usd" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.local_usd" /></td>
          <td v-if="editable"></td>
        </tr>
      </tfoot>
    </table>
    <div v-if="editable" class="border-t border-surface-200 p-3">
      <Button label="Tambah Funding Source" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
