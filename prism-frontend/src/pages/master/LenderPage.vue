<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import MultiSelect from '@/components/common/MultiSelectDropdown.vue'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { lenderSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { Lender, LenderPayload, LenderType } from '@/types/master.types'
import { toFormErrors, useMasterListControls, type FormErrors } from './master-page-utils'

type LenderField = keyof LenderPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
const controls = useMasterListControls('name', 'asc')

const dialogVisible = ref(false)
const editing = ref<Lender | null>(null)
const selectedTypes = ref<LenderType[]>([])
const form = reactive<LenderPayload>({
  name: '',
  short_name: '',
  type: 'Bilateral',
  country_id: undefined,
})
const errors = ref<FormErrors<LenderField>>({})
const typeOptions: LenderType[] = ['Bilateral', 'Multilateral', 'KSA']
const showCountry = computed(() => form.type !== 'Multilateral')

const columns: ColumnDef[] = [
  { field: 'name', header: 'Nama', sortable: true },
  { field: 'short_name', header: 'Nama Singkat', sortable: true },
  { field: 'type', header: 'Tipe', sortable: true },
  { field: 'country', header: 'Negara', sortable: true },
  { field: 'actions', header: 'Aksi' },
]

async function loadData() {
  controls.loading.value = true
  try {
    const [, lendersResponse] = await Promise.all([
      masterStore.fetchCountries(true, { limit: 100, sort: 'name', order: 'asc' }),
      masterStore.fetchLenders(true, controls.params({ type: selectedTypes.value })),
    ])
    if (lendersResponse) controls.syncMeta(lendersResponse.meta)
  } finally {
    controls.loading.value = false
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', short_name: '', type: 'Bilateral', country_id: undefined })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(lender: Lender) {
  editing.value = lender
  Object.assign(form, {
    name: lender.name,
    short_name: lender.short_name ?? '',
    type: lender.type,
    country_id: lender.country_id ?? lender.country?.id,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = lenderSchema.safeParse({
    ...form,
    country_id: form.type === 'Multilateral' ? undefined : form.country_id,
  })

  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['name', 'short_name', 'type', 'country_id'])
    return
  }

  const payload: LenderPayload = {
    ...parsed.data,
    short_name: parsed.data.short_name || undefined,
    country_id: parsed.data.type === 'Multilateral' ? null : parsed.data.country_id,
  }

  if (editing.value) {
    await masterStore.updateLender(editing.value.id, payload)
    toast.success('Berhasil', 'Lender berhasil diperbarui')
  } else {
    await masterStore.createLender(payload)
    toast.success('Berhasil', 'Lender berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(lender: Lender) {
  confirm.confirmDelete(`lender ${lender.name}`, async () => {
    await masterStore.deleteLender(lender.id)
    await loadData()
    toast.success('Berhasil', 'Lender berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})

watch(controls.search, () => {
  controls.resetAndLoadDebounced(loadData)
})

watch(selectedTypes, () => {
  controls.resetAndLoad(loadData)
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Lender" subtitle="Master lender dengan aturan country untuk Bilateral dan KSA">
      <template #actions>
        <Button v-if="can('lender', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <div class="grid gap-4 rounded-lg border border-surface-200 bg-white p-4 md:grid-cols-[minmax(0,1fr)_16rem]">
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Cari Lender</span>
        <span class="relative block">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400" />
          <InputText v-model="controls.search.value" class="w-full pl-10" placeholder="Nama atau nama singkat" />
        </span>
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Filter Tipe</span>
        <MultiSelect
          v-model="selectedTypes"
          :options="typeOptions"
          placeholder="Semua tipe"
          display="chip"
          :max-selected-labels="2"
          class="w-full"
        />
      </label>
    </div>

    <DataTable
      :data="masterStore.lenders"
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
        <Tag v-if="column.field === 'type'" :value="row.type" severity="info" rounded />
        <span v-else-if="column.field === 'country'">{{ (row as Lender).country?.name ?? '-' }}</span>
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            v-if="can('lender', 'update')"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
            @click="openEdit(row as Lender)"
          />
          <Button
            v-if="can('lender', 'delete')"
            icon="pi pi-trash"
            label="Hapus"
            size="small"
            severity="danger"
            outlined
            @click="deleteItem(row as Lender)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Lender' : 'Tambah Lender'" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama</span>
          <InputText v-model="form.name" class="w-full" :invalid="Boolean(errors.name)" />
          <small v-if="errors.name" class="text-red-600">{{ errors.name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama Singkat</span>
          <InputText v-model="form.short_name" class="w-full" :invalid="Boolean(errors.short_name)" />
          <small v-if="errors.short_name" class="text-red-600">{{ errors.short_name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tipe</span>
          <Select v-model="form.type" :options="typeOptions" class="w-full" />
          <small v-if="errors.type" class="text-red-600">{{ errors.type }}</small>
        </label>

        <label v-if="showCountry" class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Negara</span>
          <Select
            v-model="form.country_id"
            :options="masterStore.countries"
            option-label="name"
            option-value="id"
            placeholder="Pilih negara"
            filter
            show-clear
            class="w-full"
            :invalid="Boolean(errors.country_id)"
          />
          <small v-if="errors.country_id" class="text-red-600">{{ errors.country_id }}</small>
        </label>

        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
