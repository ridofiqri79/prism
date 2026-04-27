<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputNumber from 'primevue/inputnumber'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { greenBookSchema } from '@/schemas/green-book.schema'
import { useGreenBookStore } from '@/stores/green-book.store'
import type { GreenBook, GreenBookPayload } from '@/types/green-book.types'
import { formatGBRevision, toFormErrors, type FormErrors } from './green-book-page-utils'

type GreenBookField = keyof GreenBookPayload

const greenBookStore = useGreenBookStore()
const toast = useToast()
const { can } = usePermission()

const page = ref(1)
const limit = ref(20)
const dialogVisible = ref(false)
const form = reactive<GreenBookPayload>({
  publish_year: new Date().getFullYear(),
  revision_number: 0,
})
const errors = ref<FormErrors<GreenBookField>>({})
const columns: ColumnDef[] = [
  { field: 'publish_year', header: 'Publish Year' },
  { field: 'revision', header: 'Revision' },
  { field: 'status', header: 'Status' },
  { field: 'actions', header: 'Actions' },
]

async function loadData() {
  await greenBookStore.fetchGreenBooks({ page: page.value, limit: limit.value })
}

function openCreate() {
  Object.assign(form, { publish_year: new Date().getFullYear(), revision_number: 0 })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = greenBookSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['publish_year', 'revision_number'])
    return
  }

  await greenBookStore.createGreenBook(parsed.data)
  toast.success('Berhasil', 'Green Book berhasil dibuat')
  dialogVisible.value = false
  await loadData()
}

watch([page, limit], () => {
  void loadData()
})

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Green Book" subtitle="Header Green Book per publish year">
      <template #actions>
        <Button
          v-if="can('green_book', 'create')"
          label="Buat Green Book"
          icon="pi pi-plus"
          @click="openCreate"
        />
      </template>
    </PageHeader>

    <DataTable
      v-model:page="page"
      v-model:limit="limit"
      :data="greenBookStore.greenBooks"
      :columns="columns"
      :loading="greenBookStore.loading"
      :total="greenBookStore.total"
    >
      <template #body-row="{ row, column }">
        <span v-if="column.field === 'revision'">
          {{ formatGBRevision((row as GreenBook).revision_number) }}
        </span>
        <StatusBadge v-else-if="column.field === 'status'" :status="String(row.status)" />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button
            as="router-link"
            :to="{ name: 'green-book-detail', params: { id: row.id } }"
            icon="pi pi-eye"
            label="Detail"
            size="small"
            outlined
          />
        </div>
        <span v-else>{{ row[column.field] || '-' }}</span>
      </template>
    </DataTable>

    <Dialog v-model:visible="dialogVisible" modal header="Buat Green Book" class="w-[32rem] max-w-[95vw]">
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

