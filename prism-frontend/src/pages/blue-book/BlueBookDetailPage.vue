<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { blueBookSchema } from '@/schemas/blue-book.schema'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useMasterStore } from '@/stores/master.store'
import type { BBProject, BlueBookPayload } from '@/types/blue-book.types'
import { formatRevision, joinNames, toFormErrors, type FormErrors } from './blue-book-page-utils'

type BlueBookField = keyof BlueBookPayload

const route = useRoute()
const router = useRouter()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const blueBookId = computed(() => String(route.params.id ?? ''))
const dialogVisible = ref(false)
const form = reactive<BlueBookPayload>({
  period_id: '',
  publish_date: '',
  revision_number: 0,
  revision_year: null,
})
const errors = ref<FormErrors<BlueBookField>>({})
const columns: ColumnDef[] = [
  { field: 'bb_code', header: 'BB Code' },
  { field: 'project_name', header: 'Nama Proyek' },
  { field: 'executing_agency', header: 'Executing Agency' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Aksi' },
]

async function loadData() {
  await Promise.all([
    masterStore.fetchPeriods(true, { limit: 1000 }),
    blueBookStore.fetchBlueBook(blueBookId.value),
    blueBookStore.fetchProjects(blueBookId.value, { limit: 1000 }),
  ])
}

function openEdit() {
  const current = blueBookStore.currentBlueBook
  if (!current) return
  Object.assign(form, {
    period_id: current.period.id,
    publish_date: current.publish_date,
    revision_number: current.revision_number,
    revision_year: current.revision_year ?? null,
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

  await blueBookStore.updateBlueBook(blueBookId.value, {
    ...parsed.data,
    revision_year: parsed.data.revision_year ?? null,
  })
  toast.success('Berhasil', 'Blue Book berhasil diperbarui')
  dialogVisible.value = false
  await loadData()
}

function deleteProject(project: BBProject) {
  confirm.confirmDelete(`BB Project ${project.bb_code}`, async () => {
    await blueBookStore.deleteProject(blueBookId.value, project.id)
    toast.success('Berhasil', 'BB Project berhasil dihapus')
    await blueBookStore.fetchProjects(blueBookId.value, { limit: 1000 })
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="blueBookStore.currentBlueBook?.period.name ?? 'Blue Book Detail'"
      :subtitle="blueBookStore.currentBlueBook ? `Terbit ${blueBookStore.currentBlueBook.publish_date}` : undefined"
    >
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'blue-books' })" />
        <Button v-if="can('blue_book', 'update')" label="Edit BB" icon="pi pi-pencil" outlined @click="openEdit" />
        <Button
          v-if="can('bb_project', 'create')"
          as="router-link"
          :to="{ name: 'bb-project-create', params: { bbId: blueBookId } }"
          label="Tambah Proyek"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <div v-if="blueBookStore.currentBlueBook" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-4">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Periode</p>
        <p class="font-semibold text-surface-950">{{ blueBookStore.currentBlueBook.period.name }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Tanggal Terbit</p>
        <p class="font-semibold text-surface-950">{{ blueBookStore.currentBlueBook.publish_date }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Revision</p>
        <p class="font-semibold text-surface-950">
          {{ formatRevision(blueBookStore.currentBlueBook.revision_number, blueBookStore.currentBlueBook.revision_year) }}
        </p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Status</p>
        <StatusBadge :status="blueBookStore.currentBlueBook.status" />
      </div>
    </div>

    <DataTable
      :data="blueBookStore.projects"
      :columns="columns"
      :loading="blueBookStore.loading"
      :total="blueBookStore.projectTotal"
      :page="1"
      :limit="1000"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'executing_agency'">
          {{ (row as BBProject).executing_agencies[0]?.name ?? '-' }}
        </span>
        <StatusBadge v-else-if="column.field === 'status'" :status="String(row.status)" />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'bb-project-detail', params: { bbId: blueBookId, id: row.id } }"
            icon="pi pi-eye"
            label="View"
            size="small"
            outlined
          />
          <Button
            v-if="can('bb_project', 'update')"
            as="router-link"
            :to="{ name: 'bb-project-edit', params: { bbId: blueBookId, id: row.id } }"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
          />
          <Button
            v-if="can('bb_project', 'delete')"
            icon="pi pi-trash"
            label="Delete"
            size="small"
            severity="danger"
            outlined
            @click="deleteProject(row as BBProject)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Edit Blue Book" class="w-[36rem] max-w-[95vw]">
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
            disabled
          />
          <small class="text-surface-500">Periode tidak diubah saat edit Blue Book.</small>
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
