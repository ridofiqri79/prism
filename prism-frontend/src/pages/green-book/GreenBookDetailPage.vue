<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { greenBookSchema } from '@/schemas/green-book.schema'
import { useGreenBookStore } from '@/stores/green-book.store'
import type { GBProject, GreenBookPayload } from '@/types/green-book.types'
import { formatGBRevision, joinNames, toFormErrors, type FormErrors } from './green-book-page-utils'

type GreenBookField = keyof GreenBookPayload

const route = useRoute()
const router = useRouter()
const greenBookStore = useGreenBookStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const greenBookId = computed(() => String(route.params.id ?? ''))
const dialogVisible = ref(false)
const form = reactive<GreenBookPayload>({
  publish_year: new Date().getFullYear(),
  revision_number: 0,
})
const errors = ref<FormErrors<GreenBookField>>({})
const columns: ColumnDef[] = [
  { field: 'gb_code', header: 'GB Code' },
  { field: 'project_name', header: 'Nama Proyek' },
  { field: 'bb_projects', header: 'BB Projects' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Actions' },
]

async function loadData() {
  await Promise.all([
    greenBookStore.fetchGreenBook(greenBookId.value),
    greenBookStore.fetchProjects(greenBookId.value, { limit: 1000 }),
  ])
}

function openEdit() {
  const current = greenBookStore.currentGreenBook
  if (!current) return
  Object.assign(form, {
    publish_year: current.publish_year,
    revision_number: current.revision_number,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = greenBookSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['publish_year', 'revision_number'])
    return
  }

  await greenBookStore.updateGreenBook(greenBookId.value, parsed.data)
  toast.success('Berhasil', 'Green Book berhasil diperbarui')
  dialogVisible.value = false
  await loadData()
}

function deleteProject(project: GBProject) {
  confirm.confirmDelete(`GB Project ${project.gb_code}`, async () => {
    await greenBookStore.deleteProject(greenBookId.value, project.id)
    toast.success('Berhasil', 'GB Project berhasil dihapus')
    await greenBookStore.fetchProjects(greenBookId.value, { limit: 1000 })
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="greenBookStore.currentGreenBook ? `GB ${greenBookStore.currentGreenBook.publish_year}` : 'Green Book Detail'"
      :subtitle="greenBookStore.currentGreenBook ? formatGBRevision(greenBookStore.currentGreenBook.revision_number) : undefined"
    >
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'green-books' })" />
        <Button v-if="can('green_book', 'update')" label="Edit GB" icon="pi pi-pencil" outlined @click="openEdit" />
        <Button
          v-if="can('gb_project', 'create')"
          as="router-link"
          :to="{ name: 'gb-project-create', params: { gbId: greenBookId } }"
          label="Tambah Proyek"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <div v-if="greenBookStore.currentGreenBook" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-3">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Publish Year</p>
        <p class="font-semibold text-surface-950">{{ greenBookStore.currentGreenBook.publish_year }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Revision</p>
        <p class="font-semibold text-surface-950">{{ formatGBRevision(greenBookStore.currentGreenBook.revision_number) }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Status</p>
        <StatusBadge :status="greenBookStore.currentGreenBook.status" />
      </div>
    </div>

    <DataTable
      :data="greenBookStore.projects"
      :columns="columns"
      :loading="greenBookStore.loading"
      :total="greenBookStore.projectTotal"
      :page="1"
      :limit="1000"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'bb_projects'">{{ joinNames((row as GBProject).bb_projects) }}</span>
        <StatusBadge v-else-if="column.field === 'status'" :status="String(row.status)" />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'gb-project-detail', params: { gbId: greenBookId, id: row.id } }"
            icon="pi pi-eye"
            label="View"
            size="small"
            outlined
          />
          <Button
            v-if="can('gb_project', 'update')"
            as="router-link"
            :to="{ name: 'gb-project-edit', params: { gbId: greenBookId, id: row.id } }"
            icon="pi pi-pencil"
            label="Edit"
            size="small"
            outlined
          />
          <Button
            v-if="can('gb_project', 'delete')"
            icon="pi pi-trash"
            label="Delete"
            size="small"
            severity="danger"
            outlined
            @click="deleteProject(row as GBProject)"
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Edit Green Book" class="w-[32rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Publish Year</span>
          <InputNumber v-model="form.publish_year" :use-grouping="false" class="w-full" />
          <small v-if="errors.publish_year" class="text-red-600">{{ errors.publish_year }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Revision Number</span>
          <InputNumber v-model="form.revision_number" :min="0" class="w-full" />
          <small v-if="errors.revision_number" class="text-red-600">{{ errors.revision_number }}</small>
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>

