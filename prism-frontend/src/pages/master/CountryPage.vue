<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { countrySchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { Country, CountryPayload } from '@/types/master.types'
import { toFormErrors, type FormErrors } from './master-page-utils'

type CountryField = keyof CountryPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const dialogVisible = ref(false)
const editing = ref<Country | null>(null)
const form = reactive<CountryPayload>({ name: '', code: '' })
const errors = ref<FormErrors<CountryField>>({})

const columns: ColumnDef[] = [
  { field: 'code', header: 'Kode', sortable: true },
  { field: 'name', header: 'Nama', sortable: true },
  { field: 'actions', header: 'Aksi' },
]

async function loadData() {
  await masterStore.fetchCountries(true, { limit: 1000, sort: 'name', order: 'asc' })
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', code: '' })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(country: Country) {
  editing.value = country
  Object.assign(form, { name: country.name, code: country.code })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = countrySchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['name', 'code'])
    return
  }

  if (editing.value) {
    await masterStore.updateCountry(editing.value.id, parsed.data)
    toast.success('Berhasil', 'Negara berhasil diperbarui')
  } else {
    await masterStore.createCountry(parsed.data)
    toast.success('Berhasil', 'Negara berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(country: Country) {
  confirm.confirmDelete(`country ${country.name}`, async () => {
    await masterStore.deleteCountry(country.id)
    await loadData()
    toast.success('Berhasil', 'Negara berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Negara" subtitle="Master negara pemberi pinjaman dan referensi lender">
      <template #actions>
        <Button v-if="can('country', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <DataTable
      :data="masterStore.countries"
      :columns="columns"
      :loading="false"
      :total="masterStore.countries.length"
      :page="1"
      :limit="1000"
    >
      <template #body-row="{ row, column }">
        <div v-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            v-if="can('country', 'update')"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
            @click="openEdit(row as Country)"
          />
          <Button
            v-if="can('country', 'delete')"
            icon="pi pi-trash"
            label="Hapus"
            size="small"
            severity="danger"
            outlined
            @click="deleteItem(row as Country)"
          />
        </div>
        <span v-else>{{ row[column.field] }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Negara' : 'Tambah Negara'" class="w-[32rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama</span>
          <InputText v-model="form.name" class="w-full" :invalid="Boolean(errors.name)" />
          <small v-if="errors.name" class="text-red-600">{{ errors.name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Kode ISO 3</span>
          <InputText v-model="form.code" class="w-full uppercase" maxlength="3" :invalid="Boolean(errors.code)" />
          <small v-if="errors.code" class="text-red-600">{{ errors.code }}</small>
        </label>

        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
