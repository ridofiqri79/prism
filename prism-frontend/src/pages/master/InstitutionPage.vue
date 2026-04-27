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
import { institutionSchema } from '@/schemas/master.schema'
import { useMasterStore } from '@/stores/master.store'
import type { Institution, InstitutionLevel, InstitutionPayload } from '@/types/master.types'
import { buildIdTree, toFormErrors, type FormErrors } from './master-page-utils'

type InstitutionField = keyof InstitutionPayload

const masterStore = useMasterStore()
const toast = useToast()
const confirm = useConfirm()
const { can } = usePermission()

const dialogVisible = ref(false)
const editing = ref<Institution | null>(null)
const form = reactive<InstitutionPayload>({
  name: '',
  short_name: '',
  level: 'Kementerian/Badan/Lembaga',
  parent_id: undefined,
})
const errors = ref<FormErrors<InstitutionField>>({})
const levelOptions: InstitutionLevel[] = [
  'Kementerian/Badan/Lembaga',
  'Eselon I',
  'BUMN',
  'Pemerintah Daerah',
  'BUMD',
  'Lainnya',
]
const treeNodes = computed(() => buildIdTree(masterStore.institutions))
const parentOptions = computed(() => masterStore.institutions.filter((item) => item.id !== editing.value?.id))
const showParent = computed(() => form.level !== 'Kementerian/Badan/Lembaga')

async function loadData() {
  await masterStore.fetchInstitutions(true, { limit: 1000, sort: 'name', order: 'asc' })
}

function openCreate() {
  editing.value = null
  Object.assign(form, {
    name: '',
    short_name: '',
    level: 'Kementerian/Badan/Lembaga',
    parent_id: undefined,
  })
  errors.value = {}
  dialogVisible.value = true
}

function openEdit(institution: Institution) {
  editing.value = institution
  Object.assign(form, {
    name: institution.name,
    short_name: institution.short_name ?? '',
    level: institution.level,
    parent_id: institution.parent_id,
  })
  errors.value = {}
  dialogVisible.value = true
}

async function save() {
  const parsed = institutionSchema.safeParse({
    ...form,
    parent_id: showParent.value ? form.parent_id : undefined,
  })

  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['name', 'short_name', 'level', 'parent_id'])
    return
  }

  const payload: InstitutionPayload = {
    ...parsed.data,
    short_name: parsed.data.short_name || undefined,
    parent_id: showParent.value ? parsed.data.parent_id ?? null : null,
  }

  if (editing.value) {
    await masterStore.updateInstitution(editing.value.id, payload)
    toast.success('Berhasil', 'Instansi berhasil diperbarui')
  } else {
    await masterStore.createInstitution(payload)
    toast.success('Berhasil', 'Instansi berhasil dibuat')
  }

  dialogVisible.value = false
  await loadData()
}

function deleteItem(institution: Institution) {
  confirm.confirmDelete(`institution ${institution.name}`, async () => {
    await masterStore.deleteInstitution(institution.id)
    await loadData()
    toast.success('Berhasil', 'Instansi berhasil dihapus')
  })
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Instansi" subtitle="Hierarki instansi sampai 6 level referensi">
      <template #actions>
        <Button v-if="can('institution', 'create')" label="Tambah" icon="pi pi-plus" @click="openCreate" />
      </template>
    </PageHeader>

    <TreeTable :value="treeNodes" class="overflow-hidden rounded-lg border border-surface-200">
      <Column field="name" header="Nama" expander />
      <Column field="short_name" header="Nama Singkat">
        <template #body="{ node }">{{ node.data.short_name || '-' }}</template>
      </Column>
      <Column field="level" header="Level">
        <template #body="{ node }">
          <Tag :value="node.data.level" severity="info" rounded />
        </template>
      </Column>
      <Column header="Aksi">
        <template #body="{ node }">
          <div class="flex flex-wrap gap-2">
            <Button
              v-if="can('institution', 'update')"
              icon="pi pi-pencil"
              label="Edit"
              size="small"
              outlined
              @click="openEdit(node.data as Institution)"
            />
            <Button
              v-if="can('institution', 'delete')"
              icon="pi pi-trash"
              label="Hapus"
              size="small"
              severity="danger"
              outlined
              @click="deleteItem(node.data as Institution)"
            />
          </div>
        </template>
      </Column>
    </TreeTable>

    <Dialog v-model:visible="dialogVisible" modal :header="editing ? 'Edit Instansi' : 'Tambah Instansi'" class="w-[38rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="save">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama</span>
          <InputText v-model="form.name" class="w-full" :invalid="Boolean(errors.name)" />
          <small v-if="errors.name" class="text-red-600">{{ errors.name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Nama Singkat</span>
          <InputText v-model="form.short_name" class="w-full" :invalid="Boolean(errors.short_name)" />
          <small v-if="errors.short_name" class="text-red-600">{{ errors.short_name }}</small>
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Level</span>
          <Select v-model="form.level" :options="levelOptions" class="w-full" />
          <small v-if="errors.level" class="text-red-600">{{ errors.level }}</small>
        </label>

        <label v-if="showParent" class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Parent</span>
          <Select
            v-model="form.parent_id"
            :options="parentOptions"
            option-label="name"
            option-value="id"
            placeholder="Pilih parent"
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
