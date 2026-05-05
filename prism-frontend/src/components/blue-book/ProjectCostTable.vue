<script setup lang="ts">
import Button from 'primevue/button'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import Select from 'primevue/select'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import { categoriesForFundingType } from '@/composables/forms/useBBProjectForm'
import type { FundingType, ProjectCostPayload } from '@/types/blue-book.types'

const props = withDefaults(
  defineProps<{
    rows: ProjectCostPayload[]
    editable?: boolean
  }>(),
  {
    editable: true,
  },
)

const emit = defineEmits<{
  'update:rows': [value: ProjectCostPayload[]]
  add: []
  remove: [index: number]
}>()

const fundingTypes: FundingType[] = ['Foreign', 'Counterpart']

function updateRow(index: number, patch: Partial<ProjectCostPayload>) {
  const next = props.rows.map((row, rowIndex) => {
    if (rowIndex !== index) return row
    const updated = { ...row, ...patch }
    if (patch.funding_type) {
      updated.funding_category = categoriesForFundingType(patch.funding_type)[0] ?? ''
    }
    return updated
  })

  emit('update:rows', next)
}

function updateFundingType(index: number, value: unknown) {
  updateRow(index, { funding_type: value as FundingType })
}

function updateFundingCategory(index: number, value: unknown) {
  updateRow(index, { funding_category: String(value ?? '') })
}
</script>

<template>
  <div class="overflow-visible rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[48rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">Tipe Pendanaan</th>
          <th class="px-4 py-3">Kategori Pendanaan</th>
          <th class="px-4 py-3">Nilai USD</th>
          <th v-if="editable" class="w-28 px-4 py-3"></th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="(row, index) in rows" :key="index">
          <td class="px-4 py-3">
            <Select
              v-if="editable"
              :model-value="row.funding_type"
              :options="fundingTypes"
              :show-clear="false"
              class="w-full"
              @update:model-value="updateFundingType(index, $event)"
            />
            <span v-else>{{ row.funding_type }}</span>
          </td>
          <td class="px-4 py-3">
            <Select
              v-if="editable"
              :model-value="row.funding_category"
              :options="categoriesForFundingType(row.funding_type)"
              :show-clear="false"
              class="w-full"
              @update:model-value="updateFundingCategory(index, $event)"
            />
            <span v-else>{{ row.funding_category }}</span>
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="row.amount_usd"
              @update:model-value="updateRow(index, { amount_usd: $event })"
            />
            <CurrencyDisplay v-else :amount="row.amount_usd" />
          </td>
          <td v-if="editable" class="px-4 py-3 text-right">
            <Button
              icon="pi pi-trash"
              severity="danger"
              text
              rounded
              aria-label="Hapus baris biaya proyek"
              @click="emit('remove', index)"
            />
          </td>
        </tr>
        <tr v-if="rows.length === 0">
          <td :colspan="editable ? 4 : 3" class="px-4 py-6 text-center text-surface-500">
            Belum ada biaya proyek.
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="editable" class="border-t border-surface-200 p-3">
      <Button label="Tambah Baris" icon="pi pi-plus" outlined size="small" @click="emit('add')" />
    </div>
  </div>
</template>
