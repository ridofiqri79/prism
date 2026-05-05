<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { nationalPrioritySchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { NationalPriority, NationalPriorityPayload } from '@/types/master.types'
import { toFormErrors, useMasterListControls, type FormErrors } from './master-page-utils'

type NationalPriorityField = keyof NationalPriorityPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
const controls = useMasterListControls('title', 'asc')

const dialogVisible = ref(false)
const editing = ref<NationalPriority | null>(null)
const selectedPeriodIds = ref<string[]>([])
const form = reactive<NationalPriorityPayload>({ period_id: '', title: '' })
const errors = ref<FormErrors<NationalPriorityField>>({})
const columns: ColumnDef[] = [
  { field: 'title', header: 'Judul', sortable: true },
  { field: 'period', header: 'Periode', sortable: true },
  { field: 'actions', header: 'Aksi' },
]

async function loadData() {
  controls.loading.value = true
  try {
    const [, priorityResponse] = await Promise.all([
      masterStore.fetchPeriods(true, { limit: 100, sort: 'year_start', order: 'desc' }),
      masterStore.fetchNationalPriorities(
        true,
        controls.params({ period_id: selectedPeriodIds.value }),
      ),
    ])
    if (priorityResponse) controls.syncMeta(priorityResponse.meta)
  } finally {
    controls.loading.value = false
  }
}

function periodName(priority: NationalPriority) {
  return priority.period?.name ?? masterStore.periods.find((period) => period.id === priority.period_id)?.name ?? '-'
}

function openCreate() {
  editing.value = null
  Object.assign(form, {
    period_id: selectedPeriodIds.value.length === 1 ? selectedPeriodIds.value[0] : '',
    title: '',
  })
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
    toast.success('Berhasil', 'Prioritas nasional berhasil diperbarui')
  } else {
    await masterStore.createNationalPriority(parsed.data)
    toast.success('Berhasil', 'Prioritas nasional berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(priority: NationalPriority) {
  confirm.confirmDelete(`national priority ${priority.title}`, async () => {
    await masterStore.deleteNationalPriority(priority.id)
    await loadData()
    toast.success('Berhasil', 'Prioritas nasional berhasil dihapus')
  })
}

watch(controls.search, () => {
  controls.resetAndLoadDebounced(loadData)
})

watch(selectedPeriodIds, () => {
  controls.resetAndLoad(loadData)
})

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Prioritas Nasional" subtitle="Master prioritas nasional per periode">
      <template #actions>
        <Button v-if="can('national_priority', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="controls.search.value"
      search-placeholder="Cari judul prioritas nasional"
      :filter-count="selectedPeriodIds.length"
      @reset="selectedPeriodIds = []; controls.resetAndLoad(loadData)"
      @apply="controls.resetAndLoad(loadData)"
    >
      <template #filters>
        <label class="col-span-2 block space-y-1 md:col-span-1">
          <span class="text-sm font-medium text-surface-700">Periode</span>
          <MultiSelect
            v-model="selectedPeriodIds"
            :options="masterStore.periods"
            option-label="name"
            option-value="id"
            placeholder="Semua period"
            display="chip"
            :max-selected-labels="2"
            filter
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

    <DataTable
      :data="masterStore.nationalPriorities"
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
        <span v-if="column.field === 'period'">{{ periodName(row as NationalPriority) }}</span>
        <span
          v-else-if="column.field === 'title'"
          class="block max-w-[42rem] whitespace-normal break-words leading-relaxed"
        >
          {{ (row as NationalPriority).title }}
        </span>
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            v-if="can('national_priority', 'update')"
            icon="pi pi-pencil"
            rounded
            outlined
            aria-label="Edit"
            @click="openEdit(row as NationalPriority)"
          />
          <Button
            v-if="can('national_priority', 'delete')"
            icon="pi pi-trash"
            rounded
            outlined
            severity="danger"
            aria-label="Hapus"
            @click="deleteItem(row as NationalPriority)"
          />
        </div>
        <span v-else>{{ row[column.field] }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Prioritas Nasional' : 'Tambah Prioritas Nasional'" class="w-[36rem] max-w-[95vw]">
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
