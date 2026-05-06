<script setup lang="ts">
import Button from 'primevue/button'
import type { LoI } from '@/types/blue-book.types'

defineProps<{
  rows: LoI[]
  canAdd?: boolean
}>()

const emit = defineEmits<{
  add: []
}>()
</script>

<template>
  <div class="overflow-x-auto rounded-lg border border-surface-200 bg-white">
    <div class="flex items-center justify-between border-b border-surface-200 px-4 py-3">
      <h2 class="font-semibold text-surface-950">Letter of Intent</h2>
      <Button
        v-if="canAdd"
        label="Tambah LoI"
        icon="pi pi-plus"
        size="small"
        @click="emit('add')"
      />
    </div>

    <table class="w-full min-w-[44rem] text-sm">
      <thead class="bg-surface-50 text-xs font-semibold uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3 text-left">Lender</th>
          <th class="px-4 py-3 text-left">Subject</th>
          <th class="px-4 py-3 text-left">Date</th>
          <th class="px-4 py-3 text-left">Letter Number</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="row in rows" :key="row.id">
          <td class="px-4 py-2.5 text-sm text-surface-800">{{ row.lender.name }}</td>
          <td class="px-4 py-2.5 text-sm text-surface-800">{{ row.subject }}</td>
          <td class="px-4 py-2.5 text-sm text-surface-800">{{ row.date }}</td>
          <td class="px-4 py-2.5 text-sm text-surface-800">{{ row.letter_number || '-' }}</td>
        </tr>
        <tr v-if="rows.length === 0">
          <td colspan="4" class="px-4 py-6 text-center text-sm text-surface-500">
            Belum ada LoI.
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

