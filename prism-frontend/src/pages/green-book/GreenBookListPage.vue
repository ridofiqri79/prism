<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { greenBookSchema } from '@/schemas/green-book.schema'
import { useGreenBookStore } from '@/stores/green-book.store'
import type {
  GreenBook,
  GreenBookListParams,
  GreenBookPayload,
  GreenBookStatus,
} from '@/types/green-book.types'
import { formatApiError } from '@/utils/api-error'
import {
  formatGBRevision,
  formatGreenBookStatus,
  toFormErrors,
  type FormErrors,
} from './green-book-page-utils'

type GreenBookField = keyof GreenBookPayload
interface GreenBookFilterState {
  publish_year: number[]
  status: GreenBookStatus[]
}

const greenBookStore = useGreenBookStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const listControls = useListControls<GreenBookFilterState>({
  initialFilters: {
    publish_year: [],
    status: [],
  },
  filterLabels: {
    publish_year: 'Tahun Terbit',
    status: 'Status',
  },
  formatFilterValue: (key, value) => {
    if (key === 'status' && Array.isArray(value)) {
      return value.map((item) => formatGreenBookStatus(String(item))).join(', ')
    }
    return Array.isArray(value) ? value.join(', ') : String(value)
  },
})
const dialogVisible = ref(false)
const form = reactive<GreenBookPayload>({
  publish_year: new Date().getFullYear(),
  revision_number: 0,
  status: 'active',
})
const errors = ref<FormErrors<GreenBookField>>({})
const columns: ColumnDef[] = [
  { field: 'publish_year', header: 'Tahun Terbit', sortable: true },
  { field: 'revision', header: 'Revisi', sortable: true },
  { field: 'status', header: 'Status', sortable: true },
  { field: 'actions', header: 'Aksi' },
]
const statusOptions: Array<{ label: string; value: GreenBookStatus }> = [
  { label: 'Berlaku', value: 'active' },
  { label: 'Tidak Berlaku', value: 'superseded' },
]
const publishYearOptions = computed(() => {
  const currentYear = new Date().getFullYear()
  const years = new Set<number>(greenBookStore.greenBooks.map((greenBook) => greenBook.publish_year))

  for (let year = currentYear - 5; year <= currentYear + 5; year += 1) {
    years.add(year)
  }

  return [...years].sort((a, b) => b - a)
})

function buildListParams(): GreenBookListParams {
  return listControls.buildParams() as GreenBookListParams
}

async function loadData() {
  await greenBookStore.fetchGreenBooks(buildListParams())
}

function openCreate() {
  Object.assign(form, { publish_year: new Date().getFullYear(), revision_number: 0, status: 'active' })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = greenBookSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['publish_year', 'revision_number', 'status'])
    return
  }

  await greenBookStore.createGreenBook(parsed.data)
  toast.success('Berhasil', 'Green Book berhasil dibuat')
  dialogVisible.value = false
  await loadData()
}

function deleteGreenBook(greenBook: GreenBook) {
  if (greenBook.project_count > 0) {
    toast.warn('Tidak Bisa Menghapus', 'Green Book masih memiliki Project Green Book.')
    return
  }

  confirm.confirmDelete(`Green Book ${greenBook.publish_year}`, async () => {
    try {
      await greenBookStore.deleteGreenBook(greenBook.id)
      toast.success('Berhasil', 'Green Book berhasil dihapus permanen')
      if (greenBookStore.greenBooks.length === 1 && listControls.page.value > 1) {
        listControls.page.value -= 1
      } else {
        await loadData()
      }
    } catch (error) {
      toast.warn(
        'Tidak Bisa Menghapus',
        formatApiError(error, 'Green Book masih memiliki Project Green Book'),
        12000,
      )
    }
  })
}

watch(
  [
    listControls.page,
    listControls.limit,
    listControls.sort,
    listControls.order,
    listControls.debouncedSearch,
    () => JSON.stringify(listControls.appliedFilters),
  ],
  () => {
    void loadData()
  },
)

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Green Book" subtitle="Header Green Book per tahun terbit">
      <template #actions>
        <Button
          v-if="can('green_book', 'create')"
          label="Buat Green Book"
          icon="pi pi-plus"
          @click="openCreate"
        />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari tahun atau status Green Book"
      :active-filters="listControls.activeFilterPills.value"
      :filter-count="listControls.activeFilterCount.value"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Tahun Terbit</span>
          <MultiSelect
            v-model="listControls.draftFilters.publish_year"
            :options="publishYearOptions"
            placeholder="Semua tahun"
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
            option-label="label"
            option-value="value"
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
      :data="greenBookStore.greenBooks"
      :columns="columns"
      :loading="greenBookStore.loading"
      :total="greenBookStore.total"
      :sort-field="listControls.sort.value"
      :sort-order="listControls.order.value"
      @sort="listControls.setSort"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'revision'">
          {{ formatGBRevision((row as GreenBook).revision_number) }}
        </span>
        <StatusBadge
          v-else-if="column.field === 'status'"
          :status="String(row.status)"
          :label="formatGreenBookStatus(String(row.status))"
        />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'green-book-detail', params: { id: row.id } }"
            icon="pi pi-eye"
            label="Detail"
            size="small"
            outlined
          />
          <Button
            v-if="can('green_book', 'delete') && (row as GreenBook).project_count === 0"
            icon="pi pi-trash"
            size="small"
            severity="danger"
            outlined
            aria-label="Hapus Green Book"
            @click="deleteGreenBook(row as GreenBook)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Buat Green Book" class="w-[32rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tahun Terbit</span>
          <InputNumber v-model="form.publish_year" :use-grouping="false" class="w-full" />
          <small v-if="errors.publish_year" class="text-red-600">{{ errors.publish_year }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nomor Revisi</span>
          <InputNumber v-model="form.revision_number" :min="0" class="w-full" />
          <small v-if="errors.revision_number" class="text-red-600">{{ errors.revision_number }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Status</span>
          <Select
            v-model="form.status"
            :options="statusOptions"
            option-label="label"
            option-value="value"
            placeholder="Pilih status"
            class="w-full"
            :invalid="Boolean(errors.status)"
          />
          <small v-if="errors.status" class="text-red-600">{{ errors.status }}</small>
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
