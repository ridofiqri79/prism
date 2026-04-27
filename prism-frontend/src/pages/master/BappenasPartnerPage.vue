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
import { bappenasPartnerSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { BappenasPartner, BappenasPartnerLevel, BappenasPartnerPayload } from '@/types/master.types'
import { buildIdTree, toFormErrors, type FormErrors } from './master-page-utils'

type PartnerField = keyof BappenasPartnerPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const dialogVisible = ref(false)
const editing = ref<BappenasPartner | null>(null)
const form = reactive<BappenasPartnerPayload>({ name: '', level: 'Eselon I', parent_id: undefined })
const errors = ref<FormErrors<PartnerField>>({})
const levelOptions: BappenasPartnerLevel[] = ['Eselon I', 'Eselon II']
const showParent = computed(() => form.level === 'Eselon II')
const treeNodes = computed(() => buildIdTree(masterStore.bappenasPartners))
const parentOptions = computed(() =>
  masterStore.bappenasPartners.filter((item) => item.level === 'Eselon I' && item.id !== editing.value?.id),
)

async function loadData() {
  await masterStore.fetchBappenasPartners(true, { limit: 1000, sort: 'name', order: 'asc' })
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
  await loadData()
}

function deleteItem(partner: BappenasPartner) {
  confirm.confirmDelete(`bappenas partner ${partner.name}`, async () => {
    await masterStore.deleteBappenasPartner(partner.id)
    await loadData()
    toast.success('Berhasil', 'Bappenas partner berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Bappenas Partner" subtitle="Hierarki Eselon I dan Eselon II Bappenas">
      <template #actions>
        <Button v-if="can('bappenas_partner', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <TreeTable :value="treeNodes" class="overflow-hidden rounded-lg border border-surface-200">
      <Column field="name" header="Nama" expander />
      <Column field="level" header="Level">
        <template #body="{ node }">
          <Tag :value="node.data.level" severity="info" rounded />
        </template>
      </Column>
      <Column header="Actions">
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
    </TreeTable>

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
