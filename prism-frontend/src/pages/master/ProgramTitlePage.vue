<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import TreeTable from 'primevue/treetable'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { programTitleSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { ProgramTitle, ProgramTitlePayload } from '@/types/master.types'
import { buildIdTree, toFormErrors, type FormErrors } from './master-page-utils'

type ProgramTitleField = keyof ProgramTitlePayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const dialogVisible = ref(false)
const editing = ref<ProgramTitle | null>(null)
const form = reactive<ProgramTitlePayload>({ title: '', parent_id: undefined })
const errors = ref<FormErrors<ProgramTitleField>>({})
const treeNodes = computed(() => buildIdTree(masterStore.programTitles))
const parentOptions = computed(() => masterStore.programTitles.filter((item) => item.id !== editing.value?.id))

async function loadData() {
  await masterStore.fetchProgramTitles(true, { limit: 1000, sort: 'title', order: 'asc' })
}

function openCreate() {
  editing.value = null
  Object.assign(form, { title: '', parent_id: undefined })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(programTitle: ProgramTitle) {
  editing.value = programTitle
  Object.assign(form, { title: programTitle.title, parent_id: programTitle.parent_id })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = programTitleSchema.safeParse(form)
  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['title', 'parent_id'])
    return
  }

  const payload: ProgramTitlePayload = {
    title: parsed.data.title,
    parent_id: parsed.data.parent_id ?? null,
  }

  if (editing.value) {
    await masterStore.updateProgramTitle(editing.value.id, payload)
    toast.success('Berhasil', 'Program title berhasil diperbarui')
  } else {
    await masterStore.createProgramTitle(payload)
    toast.success('Berhasil', 'Program title berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(programTitle: ProgramTitle) {
  confirm.confirmDelete(`program title ${programTitle.title}`, async () => {
    await masterStore.deleteProgramTitle(programTitle.id)
    await loadData()
    toast.success('Berhasil', 'Program title berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Program Title" subtitle="Master program title parent dan child">
      <template #actions>
        <Button v-if="can('program_title', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <TreeTable :value="treeNodes" class="overflow-hidden rounded-lg border border-surface-200">
      <Column field="title" header="Judul" expander />
      <Column header="Actions">
        <template #body="{ node }">
          <div class="flex flex-wrap gap-2">
            <Button
              v-if="can('program_title', 'update')"
              icon="pi pi-pencil"
              label="Edit"
              size="small"
              outlined
              @click="openEdit(node.data as ProgramTitle)"
            />
            <Button
              v-if="can('program_title', 'delete')"
              icon="pi pi-trash"
              label="Hapus"
              size="small"
              severity="danger"
              outlined
              @click="deleteItem(node.data as ProgramTitle)"
            />
          </div>
        </template>
      </Column>
    </TreeTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Program Title' : 'Tambah Program Title'" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Judul</span>
          <InputText v-model="form.title" class="w-full" :invalid="Boolean(errors.title)" />
          <small v-if="errors.title" class="text-red-600">{{ errors.title }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Parent</span>
          <Select
            v-model="form.parent_id"
            :options="parentOptions"
            option-label="title"
            option-value="id"
            placeholder="Opsional"
            filter
            show-clear
            class="w-full"
            :invalid="Boolean(errors.parent_id)"
          />
          <small v-if="errors.parent_id" class="text-red-600">{{ errors.parent_id }}</small>
        </label>

        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
