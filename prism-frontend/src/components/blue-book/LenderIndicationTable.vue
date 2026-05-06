<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import type { LenderIndication, LenderIndicationPayload } from '@/types/blue-book.types'
import type { Lender } from '@/types/master.types'

const props = withDefaults(
  defineProps<{
    rows: LenderIndicationPayload[] | LenderIndication[]
    editable?: boolean
    extraLenderOptions?: Lender[]
  }>(),
  {
    editable: true,
    extraLenderOptions: () => [],
  },
)

const emit = defineEmits<{
  'update:rows': [value: LenderIndicationPayload[]]
  add: []
  remove: [index: number]
}>()

function rowLenderId(row: LenderIndicationPayload | LenderIndication) {
  return 'lender' in row ? row.lender.id : row.lender_id
}

function rowLenderName(row: LenderIndicationPayload | LenderIndication) {
  return 'lender' in row ? row.lender.name : '-'
}

function rowRemarks(row: LenderIndicationPayload | LenderIndication) {
  return row.remarks ?? ''
}

function updateRow(index: number, patch: Partial<LenderIndicationPayload>) {
  const next = props.rows.map((row, rowIndex) => {
    const current = { lender_id: rowLenderId(row), remarks: rowRemarks(row) }
    return rowIndex === index ? { ...current, ...patch } : current
  })

  emit('update:rows', next)
}
</script>

<template>
  <div class="overflow-x-auto rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[42rem] text-left text-sm">
      <thead class="bg-surface-50 text-left text-xs font-semibold uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">Lender</th>
          <th class="px-4 py-3">Remarks</th>
          <th v-if="editable" class="w-28 px-4 py-3"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="(row, index) in rows" :key="index">
          <td class="px-4 py-3">
            <LenderSelect
              v-if="editable"
              :model-value="rowLenderId(row)"
              :extra-options="props.extraLenderOptions"
              @update:model-value="updateRow(index, { lender_id: String($event ?? '') })"
            />
            <span v-else>{{ rowLenderName(row) }}</span>
          </td>
          <td class="px-4 py-3">
            <InputText
              v-if="editable"
              :model-value="rowRemarks(row)"
              class="w-full"
              @update:model-value="updateRow(index, { remarks: String($event ?? '') })"
            />
            <span v-else>{{ rowRemarks(row) || '-' }}</span>
          </td>
          <td v-if="editable" class="px-4 py-3 text-right">
            <Button
              icon="pi pi-trash"
              severity="danger"
              text
              rounded
              aria-label="Hapus indikasi lender"
              @click="emit('remove', index)"
            />
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td :colspan="editable ? 3 : 2" class="px-4 py-6 text-center text-surface-500">
            Belum ada lender indication.
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="editable" class="border-t border-surface-200 p-3">
      <Button label="Tambah Indikasi" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
