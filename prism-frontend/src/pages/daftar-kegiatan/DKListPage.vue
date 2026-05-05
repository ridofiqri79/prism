<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import { useConfirm } from '@/composables/useConfirm'
import { useListControls } from '@/composables/useListControls'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { daftarKegiatanSchema } from '@/schemas/daftar-kegiatan.schema'
import { useDaftarKegiatanStore } from '@/stores/daftar-kegiatan.store'
import type {
  DaftarKegiatan,
  DaftarKegiatanListParams,
  DaftarKegiatanPayload,
} from '@/types/daftar-kegiatan.types'
import { formatApiError } from '@/utils/api-error'
import { formatDate, toFormErrors, type FormErrors } from './daftar-kegiatan-page-utils'

type DKField = keyof DaftarKegiatanPayload
interface DKFilterState {
  date_from: string
  date_to: string
}

const dkStore = useDaftarKegiatanStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const listControls = useListControls<DKFilterState>({
  initialFilters: {
    date_from: '',
    date_to: '',
  },
  filterLabels: {
    date_from: 'Tanggal dari',
    date_to: 'Tanggal sampai',
  },
})
const dialogVisible = ref(false)
const form = reactive<DaftarKegiatanPayload>({
  subject: '',
  date: new Date().toISOString().slice(0, 10),
  letter_number: '',
})
const errors = ref<FormErrors<DKField>>({})
const columns: ColumnDef[] = [
  { field: 'subject', header: 'Perihal' },
  { field: 'date', header: 'Tanggal' },
  { field: 'letter_number', header: 'Nomor Surat' },
  { field: 'project_count', header: 'Jumlah Proyek' },
  { field: 'actions', header: 'Aksi' },
]

function buildListParams(): DaftarKegiatanListParams {
  return listControls.buildParams() as DaftarKegiatanListParams
}

async function loadData() {
  await dkStore.fetchDaftarKegiatan(buildListParams())
}

function openCreate() {
  Object.assign(form, {
    subject: '',
    date: new Date().toISOString().slice(0, 10),
    letter_number: '',
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = daftarKegiatanSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['subject', 'date', 'letter_number'])
    return
  }

  await dkStore.createDK({
    ...parsed.data,
    letter_number: parsed.data.letter_number || null,
  })
  toast.success('Berhasil', 'Daftar Kegiatan berhasil dibuat')
  dialogVisible.value = false
  await loadData()
}

function deleteDaftarKegiatan(row: DaftarKegiatan) {
  if ((row.project_count ?? 0) > 0) {
    toast.warn('Tidak Bisa Menghapus', 'Daftar Kegiatan masih memiliki Project di Daftar Kegiatan.')
    return
  }

  confirm.confirmDelete(`Daftar Kegiatan ${row.letter_number || row.subject}`, async () => {
    try {
      await dkStore.deleteDK(row.id)
      toast.success('Berhasil', 'Daftar Kegiatan berhasil dihapus permanen')
      if (dkStore.daftarKegiatan.length === 1 && listControls.page.value > 1) {
        listControls.page.value -= 1
      } else {
        await loadData()
      }
    } catch (error) {
      toast.warn(
        'Tidak Bisa Menghapus',
        formatApiError(error, 'Daftar Kegiatan masih memiliki Project di Daftar Kegiatan'),
        12000,
      )
    }
  })
}

watch(
  [
    listControls.page,
    listControls.limit,
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
    <PageHeader title="Daftar Kegiatan" subtitle="Header surat dan proyek dalam Daftar Kegiatan">
      <template #actions>
        <Button
          v-if="can('daftar_kegiatan', 'create')"
          label="Buat Daftar Kegiatan"
          icon="pi pi-plus"
          @click="openCreate"
        />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="listControls.search.value"
      search-placeholder="Cari perihal, nomor surat, atau tanggal"
      :active-filters="listControls.activeFilterPills.value"
      :filter-count="listControls.activeFilterCount.value"
      @apply="listControls.applyFilters"
      @reset="listControls.resetFilters"
      @remove="listControls.removeFilter"
    >
      <template #filters>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Tanggal dari</span>
          <InputText v-model="listControls.draftFilters.date_from" type="date" class="w-full" />
        </label>
        <label class="block space-y-2 xl:col-span-3">
          <span class="text-sm font-medium text-surface-700">Tanggal sampai</span>
          <InputText v-model="listControls.draftFilters.date_to" type="date" class="w-full" />
        </label>
      </template>
    </SearchFilterBar>

    <DataTable
      v-model:page="listControls.page.value"
      v-model:limit="listControls.limit.value"
      :data="dkStore.daftarKegiatan as unknown as Record<string, unknown>[]"
      :columns="columns"
      :loading="dkStore.loading"
      :total="dkStore.total"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'date'">{{ formatDate(String(row.date ?? '')) }}</span>
        <span v-else-if="column.field === 'letter_number'">{{ row.letter_number || '-' }}</span>
        <span v-else-if="column.field === 'project_count'">{{ (row as unknown as DaftarKegiatan).project_count ?? '-' }}</span>
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'daftar-kegiatan-detail', params: { id: row.id } }"
            icon="pi pi-eye"
            label="Detail"
            size="small"
            outlined
          />
          <Button
            v-if="can('daftar_kegiatan', 'delete') && ((row as unknown as DaftarKegiatan).project_count ?? 0) === 0"
            v-tooltip.top="'Hapus Daftar Kegiatan'"
            icon="pi pi-trash"
            size="small"
            severity="danger"
            outlined
            rounded
            aria-label="Hapus Daftar Kegiatan"
            @click="deleteDaftarKegiatan(row as unknown as DaftarKegiatan)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Buat Daftar Kegiatan" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Perihal</span>
          <InputText v-model="form.subject" class="w-full" />
          <small v-if="errors.subject" class="text-red-600">{{ errors.subject }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tanggal</span>
          <InputText v-model="form.date" type="date" class="w-full" />
          <small v-if="errors.date" class="text-red-600">{{ errors.date }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nomor Surat</span>
          <InputText v-model="form.letter_number" class="w-full" />
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
