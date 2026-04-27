<script setup lang="ts">
import { computed } from 'vue'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import CurrencyInput from '@/components/forms/CurrencyInput.vue'
import type {
  GBActivity,
  GBActivityPayload,
  GBAllocationValues,
  GBFundingAllocation,
} from '@/types/green-book.types'

const props = withDefaults(
  defineProps<{
    activities: GBActivityPayload[] | GBActivity[]
    rows: GBAllocationValues[] | GBFundingAllocation[]
    editable?: boolean
  }>(),
  {
    editable: true,
  },
)

const emit = defineEmits<{
  'update:rows': [value: GBAllocationValues[]]
}>()

const totals = computed(() =>
  props.rows.reduce(
    (sum, row) => ({
      services: sum.services + row.services,
      constructions: sum.constructions + row.constructions,
      goods: sum.goods + row.goods,
      trainings: sum.trainings + row.trainings,
      other: sum.other + row.other,
    }),
    { services: 0, constructions: 0, goods: 0, trainings: 0, other: 0 },
  ),
)

function toValues(row: GBAllocationValues | GBFundingAllocation): GBAllocationValues {
  return {
    services: row.services,
    constructions: row.constructions,
    goods: row.goods,
    trainings: row.trainings,
    other: row.other,
  }
}

function updateRow(index: number, patch: Partial<GBAllocationValues>) {
  const next = props.rows.map((row, rowIndex) =>
    rowIndex === index ? { ...toValues(row), ...patch } : toValues(row),
  )

  emit('update:rows', next)
}
</script>

<template>
  <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <table class="w-full min-w-[64rem] text-left text-sm">
      <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
        <tr>
          <th class="px-4 py-3">Activity</th>
          <th class="px-4 py-3">Services</th>
          <th class="px-4 py-3">Constructions</th>
          <th class="px-4 py-3">Goods</th>
          <th class="px-4 py-3">Trainings</th>
          <th class="px-4 py-3">Other</th>
        </tr>
      </thead>
      <tbody class="divide-y divide-surface-100">
        <tr v-for="(activity, index) in activities" :key="index">
          <td class="px-4 py-3 font-medium text-surface-950">
            {{ activity.activity_name || `Activity ${index + 1}` }}
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="rows[index]?.services ?? 0"
              @update:model-value="updateRow(index, { services: $event })"
            />
            <CurrencyDisplay v-else :amount="rows[index]?.services ?? 0" />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="rows[index]?.constructions ?? 0"
              @update:model-value="updateRow(index, { constructions: $event })"
            />
            <CurrencyDisplay v-else :amount="rows[index]?.constructions ?? 0" />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="rows[index]?.goods ?? 0"
              @update:model-value="updateRow(index, { goods: $event })"
            />
            <CurrencyDisplay v-else :amount="rows[index]?.goods ?? 0" />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="rows[index]?.trainings ?? 0"
              @update:model-value="updateRow(index, { trainings: $event })"
            />
            <CurrencyDisplay v-else :amount="rows[index]?.trainings ?? 0" />
          </td>
          <td class="px-4 py-3">
            <CurrencyInput
              v-if="editable"
              :model-value="rows[index]?.other ?? 0"
              @update:model-value="updateRow(index, { other: $event })"
            />
            <CurrencyDisplay v-else :amount="rows[index]?.other ?? 0" />
          </td>
        </tr>
        <tr v-if="activities.length === 0">
          <td colspan="6" class="px-4 py-6 text-center text-surface-500">
            Tambahkan activity terlebih dahulu.
          </td>
        </tr>
      </tbody>
      <tfoot class="border-t border-surface-200 bg-surface-50 font-semibold">
        <tr>
          <td class="px-4 py-3">Total</td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.services" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.constructions" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.goods" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.trainings" /></td>
          <td class="px-4 py-3"><CurrencyDisplay :amount="totals.other" /></td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

