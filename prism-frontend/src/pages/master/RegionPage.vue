<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import MultiSelect from 'primevue/multiselect'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import PageHeader from '@/components/common/PageHeader.vue'
import SearchFilterBar from '@/components/common/SearchFilterBar.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { regionSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { Region, RegionPayload, RegionType } from '@/types/master.types'
import MasterTreeTable from './MasterTreeTable.vue'
import { buildLazyCodeNodes, toFormErrors, useMasterListControls, type AppTreeNode, type FormErrors } from './master-page-utils'

type RegionField = keyof RegionPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()
const controls = useMasterListControls('type', 'asc')

const dialogVisible = ref(false)
const editing = ref<Region | null>(null)
const selectedTypes = ref<RegionType[]>([])
const treeNodes = ref<AppTreeNode<Region>[]>([])
const expandedKeys = ref<Record<string, boolean>>({})
const form = reactive<RegionPayload>({
  code: '',
  name: '',
  type: 'COUNTRY',
  parent_code: undefined,
})
const errors = ref<FormErrors<RegionField>>({})
const typeOptions: RegionType[] = ['COUNTRY', 'PROVINCE', 'CITY']
const showParent = computed(() => form.type !== 'COUNTRY')
const parentOptions = computed(() => {
  const allowedType: RegionType = form.type === 'CITY' ? 'PROVINCE' : 'COUNTRY'
  return masterStore.regions.filter((region) => region.type === allowedType && region.id !== editing.value?.id)
})

async function loadData() {
  controls.loading.value = true
  try {
    const response = await masterStore.fetchRegionTree(
      controls.params({ type: selectedTypes.value }),
    )
    controls.syncMeta(response.meta)
    treeNodes.value = buildLazyCodeNodes(response.data)
    expandedKeys.value = {}
  } finally {
    controls.loading.value = false
  }
}

async function loadLookupOptions() {
  await masterStore.fetchRegions(true, { limit: 10000, sort: 'type', order: 'asc' })
}

async function loadChildren(node: AppTreeNode<Region>) {
  if (node.leaf || node.children) return

  node.loading = true
  treeNodes.value = [...treeNodes.value]
  try {
    const response = await masterStore.fetchRegionTree(
      controls.params({
        type: selectedTypes.value,
        parent_code: node.data.code,
        page: 1,
        limit: 10000,
      }),
    )
    node.children = buildLazyCodeNodes(response.data)
    node.leaf = node.children.length === 0
  } finally {
    node.loading = false
    treeNodes.value = [...treeNodes.value]
  }
}

function openCreate() {
  editing.value = null
  Object.assign(form, { code: '', name: '', type: 'COUNTRY', parent_code: undefined })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(region: Region) {
  editing.value = region
  Object.assign(form, {
    code: region.code,
    name: region.name,
    type: region.type,
    parent_code: region.parent_code,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = regionSchema.safeParse({
    ...form,
    parent_code: showParent.value ? form.parent_code : undefined,
  })

  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['code', 'name', 'type', 'parent_code'])
    return
  }

  const payload: RegionPayload = {
    ...parsed.data,
    parent_code: showParent.value ? parsed.data.parent_code ?? null : null,
  }

  if (editing.value) {
    await masterStore.updateRegion(editing.value.id, payload)
    toast.success('Berhasil', 'Wilayah berhasil diperbarui')
  } else {
    await masterStore.createRegion(payload)
    toast.success('Berhasil', 'Wilayah berhasil dibuat')
  }

  dialogVisible.value = false
  await Promise.all([loadData(), loadLookupOptions()])
}

function deleteItem(region: Region) {
  confirm.confirmDelete(`region ${region.name}`, async () => {
    await masterStore.deleteRegion(region.id)
    await Promise.all([loadData(), loadLookupOptions()])
    toast.success('Berhasil', 'Wilayah berhasil dihapus')
  })
}

onMounted(() => {
  void Promise.all([loadData(), loadLookupOptions()])
})

watch(controls.search, () => {
  controls.resetAndLoadDebounced(loadData)
})

watch(selectedTypes, () => {
  controls.resetAndLoad(loadData)
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Wilayah" subtitle="Hierarki wilayah nasional, provinsi, dan kota/kabupaten">
      <template #actions>
        <Button v-if="can('region', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <SearchFilterBar
      v-model:search="controls.search.value"
      search-placeholder="Nama wilayah"
      :filter-count="selectedTypes.length"
      @apply="loadData"
      @reset="selectedTypes = []; loadData()"
    >
      <template #filters>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tipe Wilayah</span>
          <MultiSelect
            v-model="selectedTypes"
            :options="typeOptions"
            placeholder="Semua level"
            display="chip"
            :max-selected-labels="2"
            class="w-full"
          />
        </label>
      </template>
    </SearchFilterBar>

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
      @node-expand="(node) => loadChildren(node as AppTreeNode<Region>)"
      @sort="(value) => controls.handleSort(value, loadData)"
    >
      <Column field="code" header="Kode" sortable expander />
      <Column field="name" header="Nama" sortable />
      <Column field="type" header="Tipe" sortable>
        <template #body="{ node }">
          <Tag :value="node.data.type" severity="info" rounded />
        </template>
      </Column>
      <Column header="Aksi">
        <template #body="{ node }">
          <div class="flex flex-wrap gap-2">
            <Button
              v-if="can('region', 'update')"
              icon="pi pi-pencil"
              rounded
              outlined
              aria-label="Edit"
              @click="openEdit(node.data as Region)"
            />
            <Button
              v-if="can('region', 'delete')"
              icon="pi pi-trash"
              rounded
              outlined
              severity="danger"
              aria-label="Hapus"
              @click="deleteItem(node.data as Region)"
            />
          </div>
        </template>
      </Column>
    </MasterTreeTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Wilayah' : 'Tambah Wilayah'" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Kode</span>
          <InputText v-model="form.code" class="w-full uppercase" :invalid="Boolean(errors.code)" />
          <small v-if="errors.code" class="text-red-600">{{ errors.code }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama</span>
          <InputText v-model="form.name" class="w-full" :invalid="Boolean(errors.name)" />
          <small v-if="errors.name" class="text-red-600">{{ errors.name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tipe</span>
          <Select v-model="form.type" :options="typeOptions" class="w-full" />
          <small v-if="errors.type" class="text-red-600">{{ errors.type }}</small>
        </label>

        <label v-if="showParent" class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Parent</span>
          <Select
            v-model="form.parent_code"
            :options="parentOptions"
            option-label="name"
            option-value="code"
            placeholder="Pilih parent"
            filter
            show-clear
            class="w-full"
            :invalid="Boolean(errors.parent_code)"
          />
          <small v-if="errors.parent_code" class="text-red-600">{{ errors.parent_code }}</small>
        </label>

        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
