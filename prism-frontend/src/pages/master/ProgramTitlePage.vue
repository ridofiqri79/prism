<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { programTitleSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { ProgramTitle, ProgramTitlePayload } from '@/types/master.types'
import MasterTreeTable from './MasterTreeTable.vue'
import { buildLazyIdNodes, toFormErrors, useMasterListControls, type AppTreeNode, type FormErrors } from './master-page-utils'

type ProgramTitleField = keyof ProgramTitlePayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
const controls = useMasterListControls('title', 'asc')

const dialogVisible = ref(false)
const editing = ref<ProgramTitle | null>(null)
const treeNodes = ref<AppTreeNode<ProgramTitle>[]>([])
const expandedKeys = ref<Record<string, boolean>>({})
const form = reactive<ProgramTitlePayload>({ title: '', parent_id: undefined })
const errors = ref<FormErrors<ProgramTitleField>>({})
const parentOptions = computed(() => masterStore.programTitles.filter((item) => item.id !== editing.value?.id))

async function loadData() {
  controls.loading.value = true
  try {
    const response = await masterStore.fetchProgramTitleTree(controls.params())
    controls.syncMeta(response.meta)
    treeNodes.value = buildLazyIdNodes(response.data)
    expandedKeys.value = {}
  } finally {
    controls.loading.value = false
  }
}

async function loadLookupOptions() {
  await masterStore.fetchProgramTitles(true, { limit: 10000, sort: 'title', order: 'asc' })
}

async function loadChildren(node: AppTreeNode<ProgramTitle>) {
  if (node.leaf || node.children) return

  node.loading = true
  treeNodes.value = [...treeNodes.value]
  try {
    const response = await masterStore.fetchProgramTitleTree(
      controls.params({
        parent_id: node.data.id,
        page: 1,
        limit: 10000,
      }),
    )
    node.children = buildLazyIdNodes(response.data)
    node.leaf = node.children.length === 0
  } finally {
    node.loading = false
    treeNodes.value = [...treeNodes.value]
  }
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
    toast.success('Berhasil', 'Judul program berhasil diperbarui')
  } else {
    await masterStore.createProgramTitle(payload)
    toast.success('Berhasil', 'Judul program berhasil dibuat')
  }

  dialogVisible.value = false
  await Promise.all([loadData(), loadLookupOptions()])
}

function deleteItem(programTitle: ProgramTitle) {
  confirm.confirmDelete(`program title ${programTitle.title}`, async () => {
    await masterStore.deleteProgramTitle(programTitle.id)
    await Promise.all([loadData(), loadLookupOptions()])
    toast.success('Berhasil', 'Judul program berhasil dihapus')
  })
}

onMounted(() => {
  void Promise.all([loadData(), loadLookupOptions()])
})

watch(controls.search, () => {
  controls.resetAndLoadDebounced(loadData)
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Judul Program" subtitle="Master judul program induk dan turunan">
      <template #actions>
        <Button v-if="can('program_title', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="controls.search.value"
      search-placeholder="Judul program"
      :hide-filter-button="true"
    />

    <MasterTreeTable
      :value="treeNodes"
      :loading="controls.loading.value"
      :total="controls.total.value"
      :page="controls.pagination.page.value"
      :limit="controls.pagination.limit.value"
      :sort-field="controls.pagination.sort.value"
      :sort-order="controls.pagination.order.value"
      v-model:expanded-keys="expandedKeys"
      @update:page="(value) => controls.handlePage(value, loadData)"
      @update:limit="(value) => controls.handleLimit(value, loadData)"
      @node-expand="(node) => loadChildren(node as AppTreeNode<ProgramTitle>)"
      @sort="(value) => controls.handleSort(value, loadData)"
    >
      <Column field="title" header="Judul" sortable expander />
      <Column header="Aksi">
        <template #body="{ node }">
          <div class="flex flex-wrap gap-2">
            <Button
              v-if="can('program_title', 'update')"
              icon="pi pi-pencil"
              rounded
              outlined
              aria-label="Edit"
              @click="openEdit(node.data as ProgramTitle)"
            />
            <Button
              v-if="can('program_title', 'delete')"
              icon="pi pi-trash"
              rounded
              outlined
              severity="danger"
              aria-label="Hapus"
              @click="deleteItem(node.data as ProgramTitle)"
            />
          </div>
        </template>
      </Column>
    </MasterTreeTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Judul Program' : 'Tambah Judul Program'" class="w-[36rem] max-w-[95vw]">
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
