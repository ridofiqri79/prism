<script setup lang="ts">
import LenderSelect from '@/components/forms/LenderSelect.vue'
import type { DKFinancingDetailPayload } from '@/types/daftar-kegiatan.types'
import MultiCurrencyTable from './MultiCurrencyTable.vue'

const props = defineProps<{
  rows: DKFinancingDetailPayload[]
  allowedLenderIds: string[]
}>()

const emit = defineEmits<{
  'update:rows': [value: DKFinancingDetailPayload[]]
  add: []
  remove: [index: number]
}>()
</script>

<template>
  <div class="space-y-3">
    <div class="rounded-lg border border-primary/20 bg-primary/5 p-3 text-sm text-surface-700">
      Konversi ke USD dilakukan manual. Isi nilai original dan nilai USD sesuai hasil konversi staf.
    </div>
    <MultiCurrencyTable
      :rows="props.rows"
      empty-text="Pilih Proyek Green Book di header proyek, lalu tambah rincian pembiayaan."
      @update:rows="emit('update:rows', $event as DKFinancingDetailPayload[])"
      @add="emit('add')"
      @remove="emit('remove', $event)"
    >
      <template #first-col-header>Lender</template>
      <template #first-col="{ row, index }">
        <LenderSelect
          :model-value="(row as DKFinancingDetailPayload).lender_id"
          :allowed-ids="allowedLenderIds"
          :disabled="allowedLenderIds.length === 0"
          placeholder="Pilih lender"
          @update:model-value="
            emit('update:rows', props.rows.map((r, i) => i === index ? { ...r, lender_id: String($event ?? '') } : { ...r }) as DKFinancingDetailPayload[])
          "
        />
      </template>
    </MultiCurrencyTable>
  </div>
</template>
