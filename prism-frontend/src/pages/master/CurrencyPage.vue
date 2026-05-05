<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import ToggleSwitch from 'primevue/toggleswitch'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { currencySchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { Currency, CurrencyPayload } from '@/types/master.types'
import { toFormErrors, useMasterListControls, type FormErrors } from './master-page-utils'

type CurrencyField = keyof CurrencyPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
const controls = useMasterListControls('sort_order', 'asc')

const dialogVisible = ref(false)
const editing = ref<Currency | null>(null)
const form = reactive<CurrencyPayload>({
  code: '',
  name: '',
  symbol: '',
  is_active: true,
  sort_order: 0,
})
const errors = ref<FormErrors<CurrencyField>>({})

const columns: ColumnDef[] = [
  { field: 'code', header: 'Kode', sortable: true },
  { field: 'name', header: 'Nama', sortable: true },
  { field: 'symbol', header: 'Simbol' },
  { field: 'is_active', header: 'Status', sortable: true },
  { field: 'sort_order', header: 'Urutan', sortable: true },
  { field: 'actions', header: 'Aksi' },
]

async function loadData() {
  controls.loading.value = true
  try {
    const response = await masterStore.fetchCurrencies(true, controls.params())
    if (response) controls.syncMeta(response.meta)
  } finally {
    controls.loading.value = false
  }
}

function resetForm() {
  Object.assign(form, {
    code: '',
    name: '',
    symbol: '',
    is_active: true,
    sort_order: masterStore.currencies.length * 10 + 10,
  })
  errors.value = {}
}

function openCreate() {
  editing.value = null
  resetForm()
  dialogVisible.value = true
}

function openEdit(currency: Currency) {
  editing.value = currency
  Object.assign(form, {
    code: currency.code,
    name: currency.name,
    symbol: currency.symbol ?? '',
    is_active: currency.is_active,
    sort_order: currency.sort_order,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = currencySchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, [
      'code',
      'name',
      'symbol',
      'is_active',
      'sort_order',
    ])
    return
  }

  if (editing.value) {
    await masterStore.updateCurrency(editing.value.id, parsed.data)
    toast.success('Berhasil', 'Currency berhasil diperbarui')
  } else {
    await masterStore.createCurrency(parsed.data)
    toast.success('Berhasil', 'Currency berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(currency: Currency) {
  confirm.confirmDelete(`currency ${currency.code}`, async () => {
    await masterStore.deleteCurrency(currency.id)
    await loadData()
    toast.success('Berhasil', 'Currency berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})

watch(controls.search, () => {
  controls.resetAndLoadDebounced(loadData)
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Currency" subtitle="Master mata uang yang tersedia di form transaksi">
      <template #actions>
        <Button v-if="can('currency', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="controls.search.value"
      search-placeholder="Cari kode atau nama currency"
    />


    <DataTable
      :data="masterStore.currencies"
      :columns="columns"
      :loading="controls.loading.value"
      :total="controls.total.value"
      :page="controls.pagination.page.value"
      :limit="controls.pagination.limit.value"
      :sort-field="controls.pagination.sort.value"
      :sort-order="controls.pagination.order.value"
      @update:page="(value) => controls.handlePage(value, loadData)"
      @update:limit="(value) => controls.handleLimit(value, loadData)"
      @sort="(value) => controls.handleSort(value, loadData)"
    >
      <template #body-row="{ row, column }">
        <div v-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            v-if="can('currency', 'update')"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
            @click="openEdit(row as Currency)"
          />
          <Button
            v-if="can('currency', 'delete')"
            icon="pi pi-trash"
            label="Hapus"
            size="small"
            severity="danger"
            outlined
            @click="deleteItem(row as Currency)"
          />
        </div>
        <Tag
          v-else-if="column.field === 'is_active'"
          :value="(row as Currency).is_active ? 'Aktif' : 'Nonaktif'"
          :severity="(row as Currency).is_active ? 'success' : 'secondary'"
          rounded
        />
        <span v-else>{{ row[column.field] }}</span>
      </template>
    </DataTable>

    <Dialog
      v-model:visible="dialogVisible"
      modal
      :header="editing ? 'Edit Currency' : 'Tambah Currency'"
      class="w-[36rem] max-w-[95vw]"
    >
      <form class="space-y-4" @submit.prevent="save">
        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Kode ISO 4217</span>
            <InputText
              v-model="form.code"
              class="w-full uppercase"
              maxlength="3"
              :invalid="Boolean(errors.code)"
            />
            <small v-if="errors.code" class="text-red-600">{{ errors.code }}</small>
          </label>

          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Simbol</span>
            <InputText v-model="form.symbol" class="w-full" :invalid="Boolean(errors.symbol)" />
            <small v-if="errors.symbol" class="text-red-600">{{ errors.symbol }}</small>
          </label>
        </div>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama</span>
          <InputText v-model="form.name" class="w-full" :invalid="Boolean(errors.name)" />
          <small v-if="errors.name" class="text-red-600">{{ errors.name }}</small>
        </label>

        <div class="grid gap-4 md:grid-cols-2">
          <label class="block space-y-2">
            <span class="text-sm font-medium text-surface-700">Urutan</span>
            <InputNumber
              v-model="form.sort_order"
              class="w-full"
              :min="0"
              :use-grouping="false"
              :invalid="Boolean(errors.sort_order)"
            />
            <small v-if="errors.sort_order" class="text-red-600">{{ errors.sort_order }}</small>
          </label>

          <label class="flex items-center justify-between gap-4 rounded-lg border border-surface-200 px-3 py-2">
            <span class="text-sm font-medium text-surface-700">Aktif</span>
            <ToggleSwitch v-model="form.is_active" />
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
