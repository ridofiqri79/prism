<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import DataTable, { type DataTableRowReorderEvent } from 'primevue/datatable'
import Column from 'primevue/column'
import type { GBActivity, GBActivityPayload } from '@/types/green-book.types'

const props = withDefaults(
  defineProps<{
    rows: GBActivityPayload[] | GBActivity[]
    editable?: boolean
  }>(),
  {
    editable: true,
  },
)

const emit = defineEmits<{
  'update:rows': [value: GBActivityPayload[]]
  add: []
  remove: [index: number]
  reorder: [from: number, to: number]
}>()

function toPayload(row: GBActivityPayload | GBActivity): GBActivityPayload {
  return {
    activity_name: row.activity_name,
    implementation_location: row.implementation_location ?? '',
    piu: row.piu ?? '',
    sort_order: row.sort_order,
  }
}

function updateRow(index: number, patch: Partial<GBActivityPayload>) {
  const next = props.rows.map((row, rowIndex) =>
    rowIndex === index ? { ...toPayload(row), ...patch } : toPayload(row),
  )

  emit('update:rows', next)
}

function onRowReorder(event: DataTableRowReorderEvent) {
  emit('reorder', event.dragIndex, event.dropIndex)
}
</script>

<template>
  <div class="overflow-x-auto rounded-lg border border-surface-200 bg-white">
    <DataTable
      :value="rows"
      dataKey="sort_order"
      @rowReorder="onRowReorder"
      class="min-w-[56rem]"
      :pt="{
        thead: { class: 'bg-surface-50 text-left text-xs font-semibold uppercase tracking-wide text-surface-500' },
        headerCell: { class: 'px-4 py-3 text-xs font-semibold uppercase tracking-wide text-surface-500' },
        columnHeaderContent: { class: 'gap-2' },
        bodyCell: { class: 'px-4 py-2.5 text-sm text-surface-800' },
      }"
    >
      <Column v-if="editable" rowReorder style="width: 3rem" :reorderableColumn="false" />
      <Column header="No" style="width: 4rem">
        <template #body="{ index }">
          <span class="font-medium text-surface-500">{{ index + 1 }}</span>
        </template>
      </Column>
      <Column header="Nama Kegiatan">
        <template #body="{ data, index }">
          <InputText
            v-if="editable"
            :model-value="data.activity_name"
            class="w-full"
            @update:model-value="updateRow(index, { activity_name: String($event ?? '') })"
          />
          <span v-else>{{ data.activity_name }}</span>
        </template>
      </Column>
      <Column header="Lokasi Pelaksanaan">
        <template #body="{ data, index }">
          <InputText
            v-if="editable"
            :model-value="data.implementation_location ?? ''"
            class="w-full"
            @update:model-value="updateRow(index, { implementation_location: String($event ?? '') })"
          />
          <span v-else>{{ data.implementation_location || '-' }}</span>
        </template>
      </Column>
      <Column header="PIU">
        <template #body="{ data, index }">
          <InputText
            v-if="editable"
            :model-value="data.piu ?? ''"
            class="w-full"
            @update:model-value="updateRow(index, { piu: String($event ?? '') })"
          />
          <span v-else>{{ data.piu || '-' }}</span>
        </template>
      </Column>
      <Column v-if="editable" style="width: 5rem">
        <template #body="{ index }">
          <div class="flex justify-end">
            <Button
              icon="pi pi-trash"
              severity="danger"
              text
              rounded
              aria-label="Hapus kegiatan"
              @click="emit('remove', index)"
            />
          </div>
        </template>
      </Column>
      <template #empty>
        <div class="p-6 text-center text-surface-500">Belum ada kegiatan.</div>
      </template>
    </DataTable>

    <div v-if="editable" class="border-t border-surface-200 p-3">
      <Button label="Tambah Kegiatan" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
