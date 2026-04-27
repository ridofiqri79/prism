<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import TreeTable from 'primevue/treetable'
import PageHeader from '@/components/common/PageHeader.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { regionSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { Region, RegionPayload, RegionType } from '@/types/master.types'
import { buildCodeTree, toFormErrors, type FormErrors } from './master-page-utils'

type RegionField = keyof RegionPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const dialogVisible = ref(false)
const editing = ref<Region | null>(null)
const form = reactive<RegionPayload>({
  code: '',
  name: '',
  type: 'COUNTRY',
  parent_code: undefined,
})
const errors = ref<FormErrors<RegionField>>({})
const typeOptions: RegionType[] = ['COUNTRY', 'PROVINCE', 'CITY']
const showParent = computed(() => form.type !== 'COUNTRY')
const treeNodes = computed(() => buildCodeTree(masterStore.regions))
const parentOptions = computed(() => {
  const allowedType: RegionType = form.type === 'CITY' ? 'PROVINCE' : 'COUNTRY'
  return masterStore.regions.filter((region) => region.type === allowedType && region.id !== editing.value?.id)
})

async function loadData() {
  await masterStore.fetchRegions(true, { limit: 1000, sort: 'code', order: 'asc' })
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
    toast.success('Berhasil', 'Region berhasil diperbarui')
  } else {
    await masterStore.createRegion(payload)
    toast.success('Berhasil', 'Region berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(region: Region) {
  confirm.confirmDelete(`region ${region.name}`, async () => {
    await masterStore.deleteRegion(region.id)
    await loadData()
    toast.success('Berhasil', 'Region berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Region" subtitle="Hierarki wilayah COUNTRY, PROVINCE, dan CITY">
      <template #actions>
        <Button v-if="can('region', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <TreeTable :value="treeNodes" class="overflow-hidden rounded-lg border border-surface-200">
      <Column field="code" header="Kode" expander />
      <Column field="name" header="Nama" />
      <Column field="type" header="Type">
        <template #body="{ node }">
          <Tag :value="node.data.type" severity="info" rounded />
        </template>
      </Column>
      <Column header="Actions">
        <template #body="{ node }">
          <div class="flex flex-wrap gap-2">
            <Button
              v-if="can('region', 'update')"
              icon="pi pi-pencil"
              label="Edit"
              size="small"
              outlined
              @click="openEdit(node.data as Region)"
            />
            <Button
              v-if="can('region', 'delete')"
              icon="pi pi-trash"
              label="Hapus"
              size="small"
              severity="danger"
              outlined
              @click="deleteItem(node.data as Region)"
            />
          </div>
        </template>
      </Column>
    </TreeTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Region' : 'Tambah Region'" class="w-[36rem] max-w-[95vw]">
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
          <span class="text-sm font-medium text-surface-700">Type</span>
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
