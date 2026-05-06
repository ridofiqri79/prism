<script setup lang="ts">
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import type { DKActivityDetailPayload } from '@/types/daftar-kegiatan.types'

const props = defineProps<{
  rows: DKActivityDetailPayload[]
}>()

const emit = defineEmits<{
  'update:rows': [value: DKActivityDetailPayload[]]
  add: []
  remove: [index: number]
}>()

function updateName(index: number, activityName: string) {
  emit(
    'update:rows',
    props.rows.map((row, rowIndex) =>
      rowIndex === index ? { ...row, activity_name: activityName } : row,
    ),
  )
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[42rem] text-sm">
      <thead class="bg-surface-50 text-xs font-semibold uppercase tracking-wide text-surface-500">
        <tr>
          <th class="w-24 px-4 py-3 text-left">No.</th>
          <th class="px-4 py-3 text-left">Nama Aktivitas</th>
          <th class="w-20 px-4 py-3"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="(row, index) in rows" :key="index">
          <td class="px-4 py-2.5 text-sm font-semibold text-surface-700">{{ row.activity_number }}</td>
          <td class="px-4 py-2.5 text-sm text-surface-800">
            <InputText
              :model-value="row.activity_name"
              class="w-full"
              placeholder="Nama aktivitas"
              @update:model-value="updateName(index, String($event ?? ''))"
            />
          </td>
          <td class="px-4 py-2.5 text-right">
            <Button
              icon="pi pi-trash"
              severity="danger"
              text
              rounded
              :aria-label="`Hapus kegiatan ${row.activity_number}`"
              @click="emit('remove', index)"
            />
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td colspan="3" class="px-4 py-6 text-center text-sm text-surface-500">
            Belum ada rincian kegiatan.
          </td>
        </tr>
      </tbody>
    </table>
    <div class="border-t border-surface-200 p-3">
      <Button label="Tambah Kegiatan" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
