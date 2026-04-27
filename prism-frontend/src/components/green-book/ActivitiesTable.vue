<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
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
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[56rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">No</th>
          <th class="px-4 py-3">Nama Kegiatan</th>
          <th class="px-4 py-3">Lokasi Pelaksanaan</th>
          <th class="px-4 py-3">PIU</th>
          <th v-if="editable" class="w-36 px-4 py-3"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="(row, index) in rows" :key="index">
          <td class="px-4 py-3 font-medium text-surface-500">{{ index + 1 }}</td>
          <td class="px-4 py-3">
            <InputText
              v-if="editable"
              :model-value="row.activity_name"
              class="w-full"
              @update:model-value="updateRow(index, { activity_name: String($event ?? '') })"
            />
            <span v-else>{{ row.activity_name }}</span>
          </td>
          <td class="px-4 py-3">
            <InputText
              v-if="editable"
              :model-value="row.implementation_location ?? ''"
              class="w-full"
              @update:model-value="updateRow(index, { implementation_location: String($event ?? '') })"
            />
            <span v-else>{{ row.implementation_location || '-' }}</span>
          </td>
          <td class="px-4 py-3">
            <InputText
              v-if="editable"
              :model-value="row.piu ?? ''"
              class="w-full"
              @update:model-value="updateRow(index, { piu: String($event ?? '') })"
            />
            <span v-else>{{ row.piu || '-' }}</span>
          </td>
          <td v-if="editable" class="px-4 py-3">
            <div class="flex justify-end gap-1">
              <Button
                icon="pi pi-arrow-up"
                text
                rounded
                aria-label="Naikkan kegiatan"
                :disabled="index === 0"
                @click="emit('reorder', index, index - 1)"
              />
              <Button
                icon="pi pi-arrow-down"
                text
                rounded
                aria-label="Turunkan kegiatan"
                :disabled="index === rows.length - 1"
                @click="emit('reorder', index, index + 1)"
              />
              <Button
                icon="pi pi-trash"
                severity="danger"
                text
                rounded
                aria-label="Hapus kegiatan"
                @click="emit('remove', index)"
              />
            </div>
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td :colspan="editable ? 5 : 4" class="px-4 py-6 text-center text-surface-500">
            Belum ada kegiatan.
          </td>
        </tr>
      </tbody>
    </table>
    <div v-if="editable" class="border-t border-surface-200 p-3">
      <Button label="Tambah Kegiatan" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
