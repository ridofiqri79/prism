<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { blueBookSchema } from '@/schemas/blue-book.schema'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useMasterStore } from '@/stores/master.store'
import type { BlueBook, BlueBookListParams, BlueBookPayload, BlueBookStatus } from '@/types/blue-book.types'
import { formatRevision, toFormErrors, type FormErrors } from './blue-book-page-utils'

type BlueBookField = keyof BlueBookPayload
interface BlueBookFilterState {
  period_id: string[]
  status: BlueBookStatus[]
}

const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()
const { can } = usePermission()

const listControls = useListControls<BlueBookFilterState>({
  initialFilters: {
    period_id: [],
    status: [],
  },
  filterLabels: {
    period_id: 'Periode',
    status: 'Status',
  },
  formatFilterValue: (key, value) => {
    if (key === 'period_id' && Array.isArray(value)) {
      return selectedPeriodSummary(value)
    }
    if (key === 'status' && Array.isArray(value)) {
      return value.join(', ')
    }
    return Array.isArray(value) ? value.join(', ') : String(value)
  },
})
const dialogVisible = ref(false)
const form = reactive<BlueBookPayload>({
  period_id: '',
  publish_date: '',
  revision_number: 0,
  revision_year: null,
})
const errors = ref<FormErrors<BlueBookField>>({})

const columns: ColumnDef[] = [
  { field: 'period', header: 'Periode' },
  { field: 'publish_date', header: 'Tanggal Terbit' },
  { field: 'revision', header: 'Revision' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Aksi' },
]
const statusOptions = ['active', 'superseded']

function buildListParams(): BlueBookListParams {
  return listControls.buildParams() as BlueBookListParams
}

async function loadBlueBooks() {
  await blueBookStore.fetchBlueBooks(buildListParams())
}

async function loadData() {
  await Promise.all([
    masterStore.fetchPeriods(true, { limit: 1000, sort: 'year_start', order: 'desc' }),
    loadBlueBooks(),
  ])
}

function openCreate() {
  Object.assign(form, {
    period_id: listControls.draftFilters.period_id[0] ?? '',
    publish_date: '',
    revision_number: 0,
    revision_year: null,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = blueBookSchema.safeParse({
    ...form,
    revision_year: form.revision_year ?? undefined,
  })

  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, [
      'period_id',
      'publish_date',
      'revision_number',
      'revision_year',
    ])
    return
  }

  await blueBookStore.createBlueBook({
    ...parsed.data,
    revision_year: parsed.data.revision_year ?? null,
  })
  toast.success('Berhasil', 'Blue Book berhasil dibuat')
  dialogVisible.value = false
  await loadBlueBooks()
}

function selectedLabelSummary(labels: string[]) {
  if (labels.length <= 2) {
    return labels.join(', ')
  }

  return `${labels.slice(0, 2).join(', ')} +${labels.length - 2}`
}

function selectedPeriodSummary(ids: string[]) {
  const selected = new Set(ids)
  return selectedLabelSummary(
    masterStore.periods.filter((period) => selected.has(period.id)).map((period) => period.name),
  )
}

watch(
  [
    listControls.page,
    listControls.limit,
    listControls.debouncedSearch,
    () => JSON.stringify(listControls.appliedFilters),
  ],
  () => {
    void loadBlueBooks()
  },
)

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Blue Book" subtitle="Header Blue Book per periode dan status revisi">
      <template #actions>
        <Button
          v-if="can('blue_book', 'create')"
          label="Buat Blue Book"
          icon="pi pi-plus"
          @click="openCreate"
        />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari periode atau status Blue Book"
      :active-filters="listControls.activeFilterPills.value"
      :filter-count="listControls.activeFilterCount.value"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Periode</span>
          <MultiSelect
            v-model="listControls.draftFilters.period_id"
            :options="masterStore.periods"
            option-label="name"
            option-value="id"
            placeholder="Semua periode"
            filter
            display="chip"
            class="w-full"
          />
        </label>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Status</span>
          <MultiSelect
            v-model="listControls.draftFilters.status"
            :options="statusOptions"
            placeholder="Semua status"
            display="chip"
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

    <DataTable
      v-model:page="listControls.page.value"
      v-model:limit="listControls.limit.value"
      :data="blueBookStore.blueBooks"
      :columns="columns"
      :loading="blueBookStore.loading"
      :total="blueBookStore.total"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'period'">{{ (row as BlueBook).period.name }}</span>
        <span v-else-if="column.field === 'revision'">
          {{ formatRevision((row as BlueBook).revision_number, (row as BlueBook).revision_year) }}
        </span>
        <StatusBadge v-else-if="column.field === 'status'" :status="String(row.status)" />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'blue-book-detail', params: { id: row.id } }"
            icon="pi pi-eye"
            label="Detail"
            size="small"
            outlined
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Buat Blue Book" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Periode</span>
          <Select
            v-model="form.period_id"
            :options="masterStore.periods"
            option-label="name"
            option-value="id"
            placeholder="Pilih period"
            class="w-full"
            :invalid="Boolean(errors.period_id)"
          />
          <small v-if="errors.period_id" class="text-red-600">{{ errors.period_id }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tanggal Terbit</span>
          <InputText v-model="form.publish_date" type="date" class="w-full" :invalid="Boolean(errors.publish_date)" />
          <small v-if="errors.publish_date" class="text-red-600">{{ errors.publish_date }}</small>
        </label>
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Revision Number</span>
            <InputNumber v-model="form.revision_number" :min="0" class="w-full" />
          </label>
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Revision Year</span>
            <InputNumber v-model="form.revision_year" :use-grouping="false" class="w-full" />
          </label>
        </div>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
