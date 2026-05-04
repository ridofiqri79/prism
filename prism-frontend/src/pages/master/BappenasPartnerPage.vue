<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import MultiSelect from '@/components/common/MultiSelectDropdown.vue'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { bappenasPartnerSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { BappenasPartner, BappenasPartnerLevel, BappenasPartnerPayload } from '@/types/master.types'
import MasterTreeTable from './MasterTreeTable.vue'
import { buildLazyIdNodes, toFormErrors, useMasterListControls, type AppTreeNode, type FormErrors } from './master-page-utils'

type PartnerField = keyof BappenasPartnerPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
const controls = useMasterListControls('level', 'asc')

const dialogVisible = ref(false)
const editing = ref<BappenasPartner | null>(null)
const selectedLevels = ref<BappenasPartnerLevel[]>([])
const treeNodes = ref<AppTreeNode<BappenasPartner>[]>([])
const expandedKeys = ref<Record<string, boolean>>({})
const form = reactive<BappenasPartnerPayload>({ name: '', level: 'Eselon I', parent_id: undefined })
const errors = ref<FormErrors<PartnerField>>({})
const levelOptions: BappenasPartnerLevel[] = ['Eselon I', 'Eselon II']
const showParent = computed(() => form.level === 'Eselon II')
const parentOptions = computed(() =>
  masterStore.bappenasPartners.filter((item) => item.level === 'Eselon I' && item.id !== editing.value?.id),
)

async function loadData() {
  controls.loading.value = true
  try {
    const response = await masterStore.fetchBappenasPartnerTree(
      controls.params({ level: selectedLevels.value }),
    )
    controls.syncMeta(response.meta)
    treeNodes.value = buildLazyIdNodes(response.data)
    expandedKeys.value = {}
  } finally {
    controls.loading.value = false
  }
}

async function loadLookupOptions() {
  await masterStore.fetchBappenasPartners(true, { limit: 10000, sort: 'level', order: 'asc' })
}

async function loadChildren(node: AppTreeNode<BappenasPartner>) {
  if (node.leaf || node.children) return

  node.loading = true
  treeNodes.value = [...treeNodes.value]
  try {
    const response = await masterStore.fetchBappenasPartnerTree(
      controls.params({
        level: selectedLevels.value,
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
  Object.assign(form, { name: '', level: 'Eselon I', parent_id: undefined })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(partner: BappenasPartner) {
  editing.value = partner
  Object.assign(form, { name: partner.name, level: partner.level, parent_id: partner.parent_id })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = bappenasPartnerSchema.safeParse({
    ...form,
    parent_id: showParent.value ? form.parent_id : undefined,
  })

  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['name', 'level', 'parent_id'])
    return
  }

  const payload: BappenasPartnerPayload = {
    ...parsed.data,
    parent_id: showParent.value ? parsed.data.parent_id ?? null : null,
  }

  if (editing.value) {
    await masterStore.updateBappenasPartner(editing.value.id, payload)
    toast.success('Berhasil', 'Bappenas partner berhasil diperbarui')
  } else {
    await masterStore.createBappenasPartner(payload)
    toast.success('Berhasil', 'Bappenas partner berhasil dibuat')
  }

  dialogVisible.value = false
  await Promise.all([loadData(), loadLookupOptions()])
}

function deleteItem(partner: BappenasPartner) {
  confirm.confirmDelete(`bappenas partner ${partner.name}`, async () => {
    await masterStore.deleteBappenasPartner(partner.id)
    await Promise.all([loadData(), loadLookupOptions()])
    toast.success('Berhasil', 'Bappenas partner berhasil dihapus')
  })
}

onMounted(() => {
  void Promise.all([loadData(), loadLookupOptions()])
})

watch(controls.search, () => {
  controls.resetAndLoadDebounced(loadData)
})

watch(selectedLevels, () => {
  controls.resetAndLoad(loadData)
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Bappenas Partner" subtitle="Hierarki Eselon I dan Eselon II Bappenas">
      <template #actions>
        <Button v-if="can('bappenas_partner', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <div class="grid gap-4 rounded-lg border border-surface-200 bg-white p-4 md:grid-cols-[minmax(0,1fr)_16rem]">
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Cari Mitra Bappenas</span>
        <span class="relative block">
          <i class="pi pi-search absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400" />
          <InputText v-model="controls.search.value" class="w-full pl-10" placeholder="Nama mitra" />
        </span>
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Filter Level</span>
        <MultiSelect
          v-model="selectedLevels"
          :options="levelOptions"
          placeholder="Semua level"
          display="chip"
          :max-selected-labels="2"
          class="w-full"
        />
      </label>
    </div>

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
      @node-expand="(node) => loadChildren(node as AppTreeNode<BappenasPartner>)"
      @sort="(value) => controls.handleSort(value, loadData)"
    >
      <Column field="name" header="Nama" sortable expander />
      <Column field="level" header="Level" sortable>
        <template #body="{ node }">
          <Tag :value="node.data.level" severity="info" rounded />
        </template>
      </Column>
      <Column header="Aksi">
        <template #body="{ node }">
          <div class="flex flex-wrap gap-2">
            <Button
              v-if="can('bappenas_partner', 'update')"
              icon="pi pi-pencil"
              label="Edit"
              size="small"
              outlined
              @click="openEdit(node.data as BappenasPartner)"
            />
            <Button
              v-if="can('bappenas_partner', 'delete')"
              icon="pi pi-trash"
              label="Hapus"
              size="small"
              severity="danger"
              outlined
              @click="deleteItem(node.data as BappenasPartner)"
            />
          </div>
        </template>
      </Column>
    </MasterTreeTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Bappenas Partner' : 'Tambah Bappenas Partner'" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama</span>
          <InputText v-model="form.name" class="w-full" :invalid="Boolean(errors.name)" />
          <small v-if="errors.name" class="text-red-600">{{ errors.name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Level</span>
          <Select v-model="form.level" :options="levelOptions" class="w-full" />
          <small v-if="errors.level" class="text-red-600">{{ errors.level }}</small>
        </label>

        <label v-if="showParent" class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Eselon I Parent</span>
          <Select
            v-model="form.parent_id"
            :options="parentOptions"
            option-label="name"
            option-value="id"
            placeholder="Pilih Eselon I"
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
