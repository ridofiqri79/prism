<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { periodSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { Period, PeriodPayload } from '@/types/master.types'
import { toFormErrors, type FormErrors } from './master-page-utils'

type PeriodField = keyof PeriodPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const dialogVisible = ref(false)
const editing = ref<Period | null>(null)
const form = reactive<PeriodPayload>({ name: '', year_start: 2025, year_end: 2029 })
const errors = ref<FormErrors<PeriodField>>({})
const columns: ColumnDef[] = [
  { field: 'name', header: 'Nama', sortable: true },
  { field: 'year_start', header: 'Tahun Awal' },
  { field: 'year_end', header: 'Tahun Akhir' },
  { field: 'actions', header: 'Actions' },
]

async function loadData() {
  await masterStore.fetchPeriods(true, { limit: 1000, sort: 'year_start', order: 'desc' })
}

function openCreate() {
  editing.value = null
  Object.assign(form, { name: '', year_start: 2025, year_end: 2029 })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(period: Period) {
  editing.value = period
  Object.assign(form, {
    name: period.name,
    year_start: period.year_start,
    year_end: period.year_end,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = periodSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['name', 'year_start', 'year_end'])
    return
  }

  if (editing.value) {
    await masterStore.updatePeriod(editing.value.id, parsed.data)
    toast.success('Berhasil', 'Period berhasil diperbarui')
  } else {
    await masterStore.createPeriod(parsed.data)
    toast.success('Berhasil', 'Period berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(period: Period) {
  confirm.confirmDelete(`period ${period.name}`, async () => {
    await masterStore.deletePeriod(period.id)
    await loadData()
    toast.success('Berhasil', 'Period berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Period" subtitle="Master periode perencanaan">
      <template #actions>
        <Button v-if="can('period', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <DataTable
      :data="masterStore.periods"
      :columns="columns"
      :loading="false"
      :total="masterStore.periods.length"
      :page="1"
      :limit="1000"
    >
      <template #body-row="{ row, column }">
        <div v-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            v-if="can('period', 'update')"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
            @click="openEdit(row as Period)"
          />
          <Button
            v-if="can('period', 'delete')"
            icon="pi pi-trash"
            label="Hapus"
            size="small"
            severity="danger"
            outlined
            @click="deleteItem(row as Period)"
          />
        </div>
        <span v-else>{{ row[column.field] }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Period' : 'Tambah Period'" class="w-[32rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama</span>
          <InputText v-model="form.name" class="w-full" :invalid="Boolean(errors.name)" />
          <small v-if="errors.name" class="text-red-600">{{ errors.name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tahun Awal</span>
          <InputNumber v-model="form.year_start" class="w-full" :use-grouping="false" :invalid="Boolean(errors.year_start)" />
          <small v-if="errors.year_start" class="text-red-600">{{ errors.year_start }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tahun Akhir</span>
          <InputNumber v-model="form.year_end" class="w-full" :use-grouping="false" :invalid="Boolean(errors.year_end)" />
          <small v-if="errors.year_end" class="text-red-600">{{ errors.year_end }}</small>
        </label>

        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
