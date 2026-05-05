<script setup lang="ts">
import InstitutionSelect from '@/components/forms/InstitutionSelect.vue'
import type { DKLoanAllocationPayload } from '@/types/daftar-kegiatan.types'
import MultiCurrencyTable from './MultiCurrencyTable.vue'

const props = defineProps<{
  rows: DKLoanAllocationPayload[]
}>()

const emit = defineEmits<{
  'update:rows': [value: DKLoanAllocationPayload[]]
  add: []
  remove: [index: number]
}>()
</script>

<template>
  <MultiCurrencyTable
    :rows="props.rows"
    empty-text="Belum ada alokasi pinjaman."
    @update:rows="emit('update:rows', $event as DKLoanAllocationPayload[])"
    @add="emit('add')"
    @remove="emit('remove', $event)"
  >
    <template #first-col-header>Instansi</template>
    <template #first-col="{ row, index }">
      <InstitutionSelect
        :model-value="(row as DKLoanAllocationPayload).institution_id"
        @update:model-value="
          emit('update:rows', props.rows.map((r, i) => i === index ? { ...r, institution_id: String($event ?? '') } : { ...r }) as DKLoanAllocationPayload[])
        "
      />
    </template>
  </MultiCurrencyTable>
</template>
