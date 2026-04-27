<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import type { MonitoringKomponen } from '@/types/monitoring.types'

const props = withDefaults(
  defineProps<{
    komponen: MonitoringKomponen[]
    editable?: boolean
  }>(),
  {
    editable: false,
  },
)

const emit = defineEmits<{
  'update:komponen': [value: MonitoringKomponen[]]
  remove: [index: number]
}>()

const totals = computed(() =>
  props.komponen.reduce(
    (acc, row) => ({
      planned_la: acc.planned_la + row.planned_la,
      planned_usd: acc.planned_usd + row.planned_usd,
      planned_idr: acc.planned_idr + row.planned_idr,
      realized_la: acc.realized_la + row.realized_la,
      realized_usd: acc.realized_usd + row.realized_usd,
      realized_idr: acc.realized_idr + row.realized_idr,
    }),
    {
      planned_la: 0,
      planned_usd: 0,
      planned_idr: 0,
      realized_la: 0,
      realized_usd: 0,
      realized_idr: 0,
    },
  ),
)

function updateRow(index: number, patch: Partial<MonitoringKomponen>) {
  emit(
    'update:komponen',
    props.komponen.map((row, rowIndex) => (rowIndex === index ? { ...row, ...patch } : row)),
  )
}
</script>

<template>
  <div class="overflow-auto rounded-lg border border-surface-200">
    <table class="w-full min-w-[72rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">Komponen</th>
          <th class="px-4 py-3">Rencana LA</th>
          <th class="px-4 py-3">Rencana USD</th>
          <th class="px-4 py-3">Rencana IDR</th>
          <th class="px-4 py-3">Realisasi LA</th>
          <th class="px-4 py-3">Realisasi USD</th>
          <th class="px-4 py-3">Realisasi IDR</th>
          <th v-if="editable" class="px-4 py-3">Aksi</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-if="komponen.length === 0">
          <td :colspan="editable ? 8 : 7" class="px-4 py-6 text-center text-surface-500">
            Belum ada komponen.
          </td>
        </tr>
        <tr v-for="(row, index) in komponen" :key="row.id ?? index">
          <td class="px-4 py-3">
            <InputText
              v-if="editable"
              :model-value="row.component_name"
              class="w-full"
              placeholder="Konstruksi"
              @update:model-value="updateRow(index, { component_name: String($event ?? '') })"
            />
            <span v-else>{{ row.component_name }}</span>
          </td>
          <td class="px-4 py-3">
            <InputNumber
              v-if="editable"
              :model-value="row.planned_la"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
              @update:model-value="updateRow(index, { planned_la: Number($event ?? 0) })"
            />
            <CurrencyDisplay v-else :amount="row.planned_la" />
          </td>
          <td class="px-4 py-3">
            <InputNumber
              v-if="editable"
              :model-value="row.planned_usd"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
              @update:model-value="updateRow(index, { planned_usd: Number($event ?? 0) })"
            />
            <CurrencyDisplay v-else :amount="row.planned_usd" currency="USD" />
          </td>
          <td class="px-4 py-3">
            <InputNumber
              v-if="editable"
              :model-value="row.planned_idr"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
              @update:model-value="updateRow(index, { planned_idr: Number($event ?? 0) })"
            />
            <CurrencyDisplay v-else :amount="row.planned_idr" currency="IDR" />
          </td>
          <td class="px-4 py-3">
            <InputNumber
              v-if="editable"
              :model-value="row.realized_la"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
              @update:model-value="updateRow(index, { realized_la: Number($event ?? 0) })"
            />
            <CurrencyDisplay v-else :amount="row.realized_la" />
          </td>
          <td class="px-4 py-3">
            <InputNumber
              v-if="editable"
              :model-value="row.realized_usd"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
              @update:model-value="updateRow(index, { realized_usd: Number($event ?? 0) })"
            />
            <CurrencyDisplay v-else :amount="row.realized_usd" currency="USD" />
          </td>
          <td class="px-4 py-3">
            <InputNumber
              v-if="editable"
              :model-value="row.realized_idr"
              :min="0"
              :min-fraction-digits="2"
              class="w-full"
              @update:model-value="updateRow(index, { realized_idr: Number($event ?? 0) })"
            />
            <CurrencyDisplay v-else :amount="row.realized_idr" currency="IDR" />
          </td>
          <td v-if="editable" class="px-4 py-3">
            <Button icon="pi pi-trash" label="Hapus" severity="danger" size="small" outlined @click="emit('remove', index)" />
          </td>
        </tr>
      </tbody>
      <tfoot class="bg-surface-50 font-semibold text-surface-900">
        <tr>
          <td class="px-4 py-3">Total</td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.planned_la" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.planned_usd" currency="USD" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.planned_idr" currency="IDR" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.realized_la" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.realized_usd" currency="USD" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.realized_idr" currency="IDR" /></td>
          <td v-if="editable" />
        </tr>
      </tfoot>
    </table>
  </div>
</template>
