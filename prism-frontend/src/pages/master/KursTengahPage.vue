<script setup lang="ts">
import { computed, onMounted, ref, watch, type CSSProperties } from 'vue'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import PrimeColumn from 'primevue/column'
import PrimeDataTable from 'primevue/datatable'
import DatePicker from 'primevue/datepicker'
import InputNumber from 'primevue/inputnumber'
import Select from 'primevue/select'
import Skeleton from 'primevue/skeleton'
import Tag from 'primevue/tag'
import EmptyState from '@/components/common/EmptyState.vue'
import ListPaginationFooter from '@/components/common/ListPaginationFooter.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import TableReloadShell from '@/components/common/TableReloadShell.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { kursTengahCreateSchema, kursTengahUpdateSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { KursTengah, KursTengahPayload, KursTengahUpdatePayload } from '@/types/master.types'
import { primeTablePt } from '@/utils/table-styles'
import { toFormErrors, useMasterListControls, type FormErrors } from './master-page-utils'

type KursTengahField = keyof KursTengahPayload

interface SortEvent {
  sortField?: unknown
  sortOrder?: unknown
}

interface KursTengahDraft extends KursTengahPayload {
  id?: string
  row_key: string
  is_new: boolean
  dirty: boolean
  selected: boolean
  errors: FormErrors<KursTengahField | 'id'>
}

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
const controls = useMasterListControls('cut_off_date', 'desc')

const rows = ref<KursTengahDraft[]>([])
const saving = ref(false)
const deleting = ref(false)
const nextDraftID = ref(1)

const tableStyle: CSSProperties = {
  minWidth: '64rem',
  tableLayout: 'auto',
  width: '100%',
}
const selectionColumnStyle: CSSProperties = {
  minWidth: '3.25rem',
  textAlign: 'center',
  width: '3.25rem',
}
const tableSortOrder = computed(() => (controls.pagination.order.value === 'asc' ? 1 : -1))
const skeletonRows = computed(() =>
  Array.from({ length: controls.pagination.limit.value }, (_, index) => index),
)
const refreshingRows = computed(() => controls.loading.value && rows.value.length > 0)
const initialLoading = computed(() => controls.loading.value && rows.value.length === 0)
const selectedRows = computed(() => rows.value.filter((row) => row.selected))
const pendingCreateRows = computed(() => rows.value.filter((row) => row.is_new))
const pendingUpdateRows = computed(() => rows.value.filter((row) => !row.is_new && row.dirty))
const hasPendingChanges = computed(
  () => pendingCreateRows.value.length > 0 || pendingUpdateRows.value.length > 0,
)
const currencyOptions = computed(() =>
  masterStore.currencies.map((currency) => ({
    label: `${currency.code} - ${currency.name}`,
    value: currency.id,
  })),
)

async function loadData() {
  controls.loading.value = true
  try {
    const response = await masterStore.fetchKursTengah(true, controls.params())
    if (response) {
      rows.value = response.data.map(toDraft)
      controls.syncMeta(response.meta)
    }
  } finally {
    controls.loading.value = false
  }
}

function toDraft(item: KursTengah): KursTengahDraft {
  return {
    id: item.id,
    row_key: item.id,
    currency_id: item.currency_id,
    kurs: item.kurs,
    kurs_tengah_bi: item.kurs_tengah_bi,
    cut_off_date: item.cut_off_date,
    is_new: false,
    dirty: false,
    selected: false,
    errors: {},
  }
}

function newDraft(): KursTengahDraft {
  return {
    row_key: `new-${nextDraftID.value++}`,
    currency_id: '',
    kurs: 0,
    kurs_tengah_bi: 0,
    cut_off_date: '',
    is_new: true,
    dirty: true,
    selected: false,
    errors: {},
  }
}

function addRow() {
  rows.value = [newDraft(), ...rows.value]
}

function updateRow(row: KursTengahDraft, patch: Partial<KursTengahPayload>) {
  Object.assign(row, patch)
  row.errors = {}
  if (!row.is_new) {
    row.dirty = true
  }
}

function datePickerValue(value: string) {
  if (!value) return null
  const [year, month, day] = value.split('-').map(Number)
  if (!year || !month || !day) return null

  return new Date(year, month - 1, day)
}

function updateCutOffDate(
  row: KursTengahDraft,
  value: Date | Date[] | (Date | null)[] | null | undefined,
) {
  if (!(value instanceof Date)) {
    updateRow(row, { cut_off_date: '' })
    return
  }
  updateRow(row, { cut_off_date: toISODate(value) })
}

function toISODate(value: Date) {
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')

  return `${year}-${month}-${day}`
}

function validateCreateRow(row: KursTengahDraft): KursTengahPayload | null {
  const parsed = kursTengahCreateSchema.safeParse(rowPayload(row))
  if (!parsed.success) {
    row.errors = toFormErrors(parsed.error, [
      'currency_id',
      'kurs',
      'kurs_tengah_bi',
      'cut_off_date',
    ])
    return null
  }
  row.errors = {}
  return parsed.data
}

function validateUpdateRow(row: KursTengahDraft): KursTengahUpdatePayload | null {
  const parsed = kursTengahUpdateSchema.safeParse({ id: row.id, ...rowPayload(row) })
  if (!parsed.success) {
    row.errors = toFormErrors(parsed.error, [
      'id',
      'currency_id',
      'kurs',
      'kurs_tengah_bi',
      'cut_off_date',
    ])
    return null
  }
  row.errors = {}
  return parsed.data
}

function rowPayload(row: KursTengahDraft): KursTengahPayload {
  return {
    currency_id: row.currency_id,
    kurs: row.kurs,
    kurs_tengah_bi: row.kurs_tengah_bi,
    cut_off_date: row.cut_off_date,
  }
}

async function saveBulk() {
  const createPayloads: KursTengahPayload[] = []
  const updatePayloads: KursTengahUpdatePayload[] = []
  let valid = true

  for (const row of rows.value) {
    if (row.is_new) {
      const payload = validateCreateRow(row)
      if (payload) createPayloads.push(payload)
      else valid = false
    } else if (row.dirty) {
      const payload = validateUpdateRow(row)
      if (payload) updatePayloads.push(payload)
      else valid = false
    }
  }

  if (!valid) {
    toast.warn('Validasi belum lengkap', 'Periksa baris yang masih memiliki error')
    return
  }
  if (createPayloads.length === 0 && updatePayloads.length === 0) {
    toast.info('Tidak ada perubahan', 'Belum ada baris kurs tengah yang perlu disimpan')
    return
  }
  if (createPayloads.length > 0 && !can('currency', 'create')) {
    toast.error('Akses ditolak', 'Anda tidak memiliki akses tambah Kurs Tengah BI')
    return
  }
  if (updatePayloads.length > 0 && !can('currency', 'update')) {
    toast.error('Akses ditolak', 'Anda tidak memiliki akses ubah Kurs Tengah BI')
    return
  }

  saving.value = true
  try {
    if (createPayloads.length > 0) {
      await masterStore.createKursTengahBulk(createPayloads)
    }
    if (updatePayloads.length > 0) {
      await masterStore.updateKursTengahBulk(updatePayloads)
    }
    toast.success(
      'Berhasil',
      `${createPayloads.length + updatePayloads.length} baris kurs tengah disimpan`,
    )
    await loadData()
  } finally {
    saving.value = false
  }
}

function removeLocalRow(row: KursTengahDraft) {
  rows.value = rows.value.filter((item) => item.row_key !== row.row_key)
}

function deleteSelected() {
  const selected = selectedRows.value
  if (selected.length === 0) {
    toast.info('Belum ada pilihan', 'Pilih minimal satu baris kurs tengah')
    return
  }

  const existingIDs = selected.filter((row) => !row.is_new && row.id).map((row) => row.id as string)
  const localKeys = selected.filter((row) => row.is_new).map((row) => row.row_key)

  if (existingIDs.length === 0) {
    rows.value = rows.value.filter((row) => !localKeys.includes(row.row_key))
    return
  }
  if (!can('currency', 'delete')) {
    toast.error('Akses ditolak', 'Anda tidak memiliki akses hapus Kurs Tengah BI')
    return
  }

  confirm.confirmDelete(`${selected.length} baris kurs tengah`, async () => {
    deleting.value = true
    try {
      await masterStore.deleteKursTengahBulk(existingIDs)
      rows.value = rows.value.filter((row) => !localKeys.includes(row.row_key))
      toast.success('Berhasil', `${selected.length} baris kurs tengah dihapus`)
      await loadData()
    } finally {
      deleting.value = false
    }
  })
}

function handleSort(event: SortEvent) {
  if (typeof event.sortField !== 'string' || event.sortOrder === 0) {
    return
  }

  controls.handleSort(
    {
      sort: event.sortField,
      order: event.sortOrder === 1 ? 'asc' : 'desc',
    },
    loadData,
  )
}

onMounted(async () => {
  await Promise.all([
    masterStore.fetchCurrencies(true, { limit: 10000, sort: 'code', order: 'asc' }),
    loadData(),
  ])
})

watch(controls.search, () => {
  controls.resetAndLoadDebounced(loadData)
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Kurs Tengah BI"
      subtitle="Kelola nilai kurs per mata uang dan tanggal cut off"
    >
      <template #actions>
        <Button
          v-if="can('currency', 'create')"
          label="Tambah"
          icon="pi pi-plus"
          outlined
          @click="addRow"
        />
        <Button
          label="Simpan Perubahan"
          icon="pi pi-save"
          :disabled="!hasPendingChanges || saving || deleting"
          :loading="saving"
          @click="saveBulk"
        />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="controls.search.value"
      search-placeholder="Cari kode atau nama mata uang"
    />

    <div
      class="flex flex-col gap-3 rounded-lg border border-surface-200 bg-white px-4 py-3 shadow-sm shadow-surface-200/60 sm:flex-row sm:items-center sm:justify-between"
    >
      <div class="flex min-w-0 flex-wrap gap-2">
        <Tag :value="`${pendingCreateRows.length} Baru`" severity="info" rounded />
        <Tag :value="`${pendingUpdateRows.length} Diubah`" severity="warning" rounded />
        <Tag :value="`${selectedRows.length} Dipilih`" severity="secondary" rounded />
      </div>
      <Button
        label="Hapus Terpilih"
        icon="pi pi-trash"
        severity="danger"
        outlined
        :disabled="selectedRows.length === 0 || saving || deleting"
        :loading="deleting"
        @click="deleteSelected"
      />
    </div>

    <div
      v-if="initialLoading"
      class="overflow-hidden rounded-lg border border-surface-200 bg-white"
    >
      <div
        v-for="row in skeletonRows"
        :key="row"
        class="grid grid-cols-[2rem_minmax(14rem,1.6fr)_minmax(8rem,1fr)_minmax(8rem,1fr)_minmax(10rem,1fr)_minmax(7rem,0.7fr)_minmax(4rem,0.4fr)] gap-4 border-b border-surface-100 p-4 last:border-b-0"
      >
        <Skeleton v-for="column in 7" :key="column" height="1.5rem" />
      </div>
    </div>

    <EmptyState v-else-if="rows.length === 0" compact />

    <TableReloadShell v-else :refreshing="refreshingRows">
      <div class="overflow-x-auto">
        <PrimeDataTable
          :value="rows"
          lazy
          striped-rows
          removable-sort
          resizable-columns
          column-resize-mode="fit"
          data-key="row_key"
          :sort-field="controls.pagination.sort.value"
          :sort-order="tableSortOrder"
          :table-style="tableStyle"
          :pt="primeTablePt"
          class="prism-data-table w-full rounded-lg border border-surface-200"
          @sort="handleSort"
        >
          <PrimeColumn
            header=""
            header-class="prism-table-align-center"
            body-class="prism-table-align-center"
            :header-style="selectionColumnStyle"
            :body-style="selectionColumnStyle"
          >
            <template #body="{ data }">
              <div class="flex items-center justify-center">
                <Checkbox v-model="(data as KursTengahDraft).selected" binary />
              </div>
            </template>
          </PrimeColumn>

          <PrimeColumn field="currency" header="Mata Uang" sortable>
            <template #body="{ data }">
              <div class="min-w-64 space-y-1">
                <Select
                  :model-value="(data as KursTengahDraft).currency_id"
                  :options="currencyOptions"
                  option-label="label"
                  option-value="value"
                  filter
                  placeholder="Pilih mata uang"
                  class="w-full"
                  :invalid="Boolean((data as KursTengahDraft).errors.currency_id)"
                  @update:model-value="
                    updateRow(data as KursTengahDraft, { currency_id: String($event ?? '') })
                  "
                />
                <small v-if="(data as KursTengahDraft).errors.currency_id" class="text-red-600">
                  {{ (data as KursTengahDraft).errors.currency_id }}
                </small>
              </div>
            </template>
          </PrimeColumn>

          <PrimeColumn field="kurs" header="Kurs" sortable>
            <template #body="{ data }">
              <div class="min-w-44 space-y-1">
                <InputNumber
                  :model-value="(data as KursTengahDraft).kurs"
                  :min="0"
                  :min-fraction-digits="0"
                  :max-fraction-digits="6"
                  :use-grouping="false"
                  class="w-full"
                  :invalid="Boolean((data as KursTengahDraft).errors.kurs)"
                  @update:model-value="
                    updateRow(data as KursTengahDraft, { kurs: Number($event ?? 0) })
                  "
                />
                <small v-if="(data as KursTengahDraft).errors.kurs" class="text-red-600">
                  {{ (data as KursTengahDraft).errors.kurs }}
                </small>
              </div>
            </template>
          </PrimeColumn>

          <PrimeColumn field="kurs_tengah_bi" header="Kurs Tengah BI" sortable>
            <template #body="{ data }">
              <div class="min-w-44 space-y-1">
                <InputNumber
                  :model-value="(data as KursTengahDraft).kurs_tengah_bi"
                  :min="0"
                  :min-fraction-digits="0"
                  :max-fraction-digits="6"
                  :use-grouping="false"
                  class="w-full"
                  :invalid="Boolean((data as KursTengahDraft).errors.kurs_tengah_bi)"
                  @update:model-value="
                    updateRow(data as KursTengahDraft, { kurs_tengah_bi: Number($event ?? 0) })
                  "
                />
                <small v-if="(data as KursTengahDraft).errors.kurs_tengah_bi" class="text-red-600">
                  {{ (data as KursTengahDraft).errors.kurs_tengah_bi }}
                </small>
              </div>
            </template>
          </PrimeColumn>

          <PrimeColumn field="cut_off_date" header="Tanggal Cut Off" sortable>
            <template #body="{ data }">
              <div class="min-w-48 space-y-1">
                <DatePicker
                  :model-value="datePickerValue((data as KursTengahDraft).cut_off_date)"
                  date-format="yy-mm-dd"
                  show-icon
                  class="w-full"
                  :invalid="Boolean((data as KursTengahDraft).errors.cut_off_date)"
                  @update:model-value="updateCutOffDate(data as KursTengahDraft, $event)"
                />
                <small v-if="(data as KursTengahDraft).errors.cut_off_date" class="text-red-600">
                  {{ (data as KursTengahDraft).errors.cut_off_date }}
                </small>
              </div>
            </template>
          </PrimeColumn>

          <PrimeColumn header="Status">
            <template #body="{ data }">
              <Tag v-if="(data as KursTengahDraft).is_new" value="Baru" severity="info" rounded />
              <Tag
                v-else-if="(data as KursTengahDraft).dirty"
                value="Diubah"
                severity="warning"
                rounded
              />
              <Tag v-else value="Tersimpan" severity="secondary" rounded />
            </template>
          </PrimeColumn>

          <PrimeColumn header="Aksi">
            <template #body="{ data }">
              <Button
                icon="pi pi-times"
                rounded
                outlined
                severity="secondary"
                aria-label="Lepas baris"
                @click="removeLocalRow(data as KursTengahDraft)"
              />
            </template>
          </PrimeColumn>
        </PrimeDataTable>
      </div>
    </TableReloadShell>

    <ListPaginationFooter
      :page="controls.pagination.page.value"
      :limit="controls.pagination.limit.value"
      :total="controls.total.value"
      @update:page="(value) => controls.handlePage(value, loadData)"
      @update:limit="(value) => controls.handleLimit(value, loadData)"
    />
  </section>
</template>
