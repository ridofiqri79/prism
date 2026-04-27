<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { nationalPrioritySchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { NationalPriority, NationalPriorityPayload } from '@/types/master.types'
import { toFormErrors, type FormErrors } from './master-page-utils'

type NationalPriorityField = keyof NationalPriorityPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const dialogVisible = ref(false)
const editing = ref<NationalPriority | null>(null)
const selectedPeriodId = ref<string | null>(null)
const form = reactive<NationalPriorityPayload>({ period_id: '', title: '' })
const errors = ref<FormErrors<NationalPriorityField>>({})
const columns: ColumnDef[] = [
  { field: 'title', header: 'Judul', sortable: true },
  { field: 'period', header: 'Period' },
  { field: 'actions', header: 'Actions' },
]
const filteredPriorities = computed(() => {
  if (!selectedPeriodId.value) return masterStore.nationalPriorities
  return masterStore.nationalPriorities.filter((priority) => priority.period_id === selectedPeriodId.value)
})

async function loadData() {
  await Promise.all([
    masterStore.fetchPeriods(true, { limit: 1000, sort: 'year_start', order: 'desc' }),
    masterStore.fetchNationalPriorities(true, {
      limit: 1000,
      sort: 'title',
      order: 'asc',
      ...(selectedPeriodId.value ? { period_id: selectedPeriodId.value } : {}),
    }),
  ])
}

function periodName(priority: NationalPriority) {
  return priority.period?.name ?? masterStore.periods.find((period) => period.id === priority.period_id)?.name ?? '-'
}

function openCreate() {
  editing.value = null
  Object.assign(form, { period_id: selectedPeriodId.value ?? '', title: '' })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(priority: NationalPriority) {
  editing.value = priority
  Object.assign(form, { period_id: priority.period_id, title: priority.title })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = nationalPrioritySchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['period_id', 'title'])
    return
  }

  if (editing.value) {
    await masterStore.updateNationalPriority(editing.value.id, parsed.data)
    toast.success('Berhasil', 'National priority berhasil diperbarui')
  } else {
    await masterStore.createNationalPriority(parsed.data)
    toast.success('Berhasil', 'National priority berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(priority: NationalPriority) {
  confirm.confirmDelete(`national priority ${priority.title}`, async () => {
    await masterStore.deleteNationalPriority(priority.id)
    await loadData()
    toast.success('Berhasil', 'National priority berhasil dihapus')
  })
}

watch(selectedPeriodId, () => {
  void loadData()
})

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="National Priority" subtitle="Master prioritas nasional per period">
      <template #actions>
        <Button v-if="can('national_priority', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <div class="rounded-lg border border-surface-200 bg-white p-4">
      <label class="block max-w-sm space-y-2">
        <span class="text-sm font-medium text-surface-700">Filter Period</span>
        <Select
          v-model="selectedPeriodId"
          :options="masterStore.periods"
          option-label="name"
          option-value="id"
          placeholder="Semua period"
          show-clear
          class="w-full"
        />
      </label>
    </div>

    <DataTable
      :data="filteredPriorities"
      :columns="columns"
      :loading="false"
      :total="filteredPriorities.length"
      :page="1"
      :limit="1000"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'period'">{{ periodName(row as NationalPriority) }}</span>
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            v-if="can('national_priority', 'update')"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
            @click="openEdit(row as NationalPriority)"
          />
          <Button
            v-if="can('national_priority', 'delete')"
            icon="pi pi-trash"
            label="Hapus"
            size="small"
            severity="danger"
            outlined
            @click="deleteItem(row as NationalPriority)"
          />
        </div>
        <span v-else>{{ row[column.field] }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit National Priority' : 'Tambah National Priority'" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Period</span>
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
          <span class="text-sm font-medium text-surface-700">Judul</span>
          <InputText v-model="form.title" class="w-full" :invalid="Boolean(errors.title)" />
          <small v-if="errors.title" class="text-red-600">{{ errors.title }}</small>
        </label>

        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
