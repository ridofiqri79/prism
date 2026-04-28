<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Tag from 'primevue/tag'
import LenderIndicationTable from '@/components/blue-book/LenderIndicationTable.vue'
import LoITable from '@/components/blue-book/LoITable.vue'
import ProjectCostTable from '@/components/blue-book/ProjectCostTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { loiSchema } from '@/schemas/blue-book.schema'
import { useBlueBookStore } from '@/stores/blue-book.store'
import { useMasterStore } from '@/stores/master.store'
import type { BBProjectHistoryItem, LoIPayload } from '@/types/blue-book.types'
import type { BappenasPartner } from '@/types/master.types'
import { joinNames, toFormErrors, type FormErrors } from './blue-book-page-utils'

type LoIField = keyof LoIPayload

const route = useRoute()
const router = useRouter()
const blueBookStore = useBlueBookStore()
const masterStore = useMasterStore()
const toast = useToast()
const { can } = usePermission()

const blueBookId = computed(() => String(route.params.bbId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const dialogVisible = ref(false)
const loiForm = reactive<LoIPayload>({
  lender_id: '',
  subject: '',
  date: '',
  letter_number: '',
})
const errors = ref<FormErrors<LoIField>>({})

const project = computed(() => blueBookStore.currentProject)
const programTitleName = computed(
  () =>
    project.value?.program_title?.title ??
    masterStore.programTitles.find((item) => item.id === project.value?.program_title_id)?.title ??
    '-',
)
const bappenasPartner = computed(
  () =>
    project.value?.bappenas_partner ??
    masterStore.bappenasPartners.find((item) => item.id === project.value?.bappenas_partner_id),
)
const bappenasPartnerParent = computed(() => findPartnerParent(bappenasPartner.value))
const allowedLenderIds = computed(
  () => project.value?.lender_indications.map((item) => item.lender.id) ?? [],
)

function findPartnerParent(partner?: BappenasPartner) {
  if (!partner?.parent_id) return partner?.parent?.name ?? '-'
  return masterStore.bappenasPartners.find((item) => item.id === partner.parent_id)?.name ?? partner.parent?.name ?? '-'
}

async function loadData() {
  await Promise.all([
    blueBookStore.fetchProject(blueBookId.value, projectId.value),
    blueBookStore.fetchProjectHistory(projectId.value),
    blueBookStore.fetchLoI(projectId.value),
    masterStore.fetchProgramTitles(true, { limit: 1000 }),
    masterStore.fetchBappenasPartners(true, { limit: 1000 }),
    masterStore.fetchLenders(true, { limit: 1000 }),
  ])
}

function historyRoute(item: BBProjectHistoryItem) {
  return { name: 'bb-project-detail', params: { bbId: item.blue_book_id, id: item.id } }
}

function historyLabel(item: BBProjectHistoryItem) {
  const year = item.revision_year ? ` / ${item.revision_year}` : ''
  return `${item.book_label} - Rev ${item.revision_number}${year}`
}

function openLoIDialog() {
  Object.assign(loiForm, {
    lender_id: allowedLenderIds.value[0] ?? '',
    subject: '',
    date: '',
    letter_number: '',
  })
  errors.value = {}
  dialogVisible.value = true
}

async function saveLoI() {
  const parsed = loiSchema.safeParse(loiForm)

  if (!parsed.success) {
    errors.value = toFormErrors(parsed.error, ['lender_id', 'subject', 'date', 'letter_number'])
    return
  }

  await blueBookStore.createLoI(projectId.value, {
    ...parsed.data,
    letter_number: parsed.data.letter_number ?? null,
  })
  toast.success('Berhasil', 'LoI berhasil dibuat')
  dialogVisible.value = false
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="project?.bb_code ?? 'Detail BB Project'"
      :subtitle="project?.project_name"
    >
      <template #actions>
        <Button
          label="Kembali"
          icon="pi pi-arrow-left"
          outlined
          @click="router.push({ name: 'blue-book-detail', params: { id: blueBookId } })"
        />
        <Button
          v-if="can('bb_project', 'update')"
          as="router-link"
          :to="{ name: 'bb-project-edit', params: { bbId: blueBookId, id: projectId } }"
          label="Edit"
          icon="pi pi-pencil"
          outlined
        />
        <Button
          as="router-link"
          :to="{ name: 'project-journey', params: { bbProjectId: projectId } }"
          label="Lihat Journey"
          icon="pi pi-share-alt"
          severity="secondary"
        />
      </template>
    </PageHeader>

    <div v-if="project" class="space-y-6">
      <div class="rounded-lg border border-surface-200 bg-white p-5">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="space-y-2">
            <p class="text-sm text-surface-500">Status</p>
            <div class="flex flex-wrap items-center gap-2">
              <StatusBadge :status="project.status" />
              <Tag v-if="project.is_latest" value="Latest" severity="success" rounded />
              <Tag v-else-if="project.has_newer_revision" value="Ada revisi lebih baru" severity="warn" rounded />
            </div>
          </div>
          <div class="text-right">
            <p class="text-sm text-surface-500">Judul Program</p>
            <p class="font-semibold text-surface-950">{{ programTitleName }}</p>
          </div>
        </div>
      </div>

      <section class="space-y-3 rounded-lg border border-surface-200 bg-white p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <h2 class="text-lg font-semibold text-surface-950">Histori Revisi</h2>
          <Tag :value="`${blueBookStore.projectHistory.length} snapshot`" severity="secondary" rounded />
        </div>
        <div class="overflow-auto rounded-lg border border-surface-200">
          <table class="w-full min-w-[48rem] text-left text-sm">
            <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
              <tr>
                <th class="px-4 py-3">Blue Book</th>
                <th class="px-4 py-3">Kode</th>
                <th class="px-4 py-3">Status Dokumen</th>
                <th class="px-4 py-3">Snapshot</th>
                <th class="px-4 py-3">Downstream</th>
                <th class="px-4 py-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-surface-100">
              <tr v-for="item in blueBookStore.projectHistory" :key="item.id">
                <td class="px-4 py-3 font-medium text-surface-900">{{ historyLabel(item) }}</td>
                <td class="px-4 py-3 text-surface-700">{{ item.bb_code }}</td>
                <td class="px-4 py-3"><StatusBadge :status="item.book_status" /></td>
                <td class="px-4 py-3">
                  <Tag
                    :value="item.is_latest ? 'Latest' : 'Historical'"
                    :severity="item.is_latest ? 'success' : 'secondary'"
                    rounded
                  />
                </td>
                <td class="px-4 py-3">
                  <Tag
                    :value="item.used_by_downstream ? 'Dipakai downstream' : 'Belum dipakai'"
                    :severity="item.used_by_downstream ? 'info' : 'secondary'"
                    rounded
                  />
                </td>
                <td class="px-4 py-3 text-right">
                  <Button
                    as="router-link"
                    :to="historyRoute(item)"
                    icon="pi pi-eye"
                    severity="secondary"
                    size="small"
                    outlined
                    rounded
                    aria-label="Lihat snapshot"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <div class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-2">
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Executing Agency</p>
          <p class="mt-1 font-medium text-surface-950">{{ joinNames(project.executing_agencies) }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Implementing Agency</p>
          <p class="mt-1 font-medium text-surface-950">{{ joinNames(project.implementing_agencies) }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Bappenas Partner</p>
          <p class="mt-1 font-medium text-surface-950">
            {{ bappenasPartner?.name ?? '-' }}
            <span class="text-surface-500">/ {{ bappenasPartnerParent }}</span>
          </p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Lokasi</p>
          <p class="mt-1 font-medium text-surface-950">{{ joinNames(project.locations) }}</p>
        </div>
        <div class="md:col-span-2">
          <p class="text-xs uppercase tracking-wide text-surface-500">Prioritas Nasional</p>
          <p class="mt-1 font-medium text-surface-950">{{ joinNames(project.national_priorities) }}</p>
        </div>
      </div>

      <div class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-2">
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Duration</p>
          <p class="mt-1 text-surface-950">{{ project.duration || '-' }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Objective</p>
          <p class="mt-1 text-surface-950">{{ project.objective || '-' }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Scope of Work</p>
          <p class="mt-1 text-surface-950">{{ project.scope_of_work || '-' }}</p>
        </div>
        <div>
          <p class="text-xs uppercase tracking-wide text-surface-500">Outputs / Outcomes</p>
          <p class="mt-1 text-surface-950">{{ project.outputs || '-' }} / {{ project.outcomes || '-' }}</p>
        </div>
      </div>

      <section class="space-y-3">
        <h2 class="text-lg font-semibold text-surface-950">Biaya Proyek</h2>
        <ProjectCostTable :rows="project.project_costs" :editable="false" />
      </section>

      <section class="space-y-3">
        <h2 class="text-lg font-semibold text-surface-950">Indikasi Lender</h2>
        <LenderIndicationTable :rows="project.lender_indications" :editable="false" />
      </section>

      <LoITable
        :rows="blueBookStore.lois"
        :can-add="can('bb_project', 'update')"
        @add="openLoIDialog"
      />
    </div>

    <Dialog v-model:visible="dialogVisible" modal header="Tambah LoI" class="w-[36rem] max-w-[95vw]">
      <form class="space-y-4" @submit.prevent="saveLoI">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <LenderSelect
            v-model="loiForm.lender_id"
            :allowed-ids="allowedLenderIds"
            placeholder="Pilih lender dari indication proyek"
          />
          <small v-if="errors.lender_id" class="text-red-600">{{ errors.lender_id }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Subject</span>
          <InputText v-model="loiForm.subject" class="w-full" :invalid="Boolean(errors.subject)" />
          <small v-if="errors.subject" class="text-red-600">{{ errors.subject }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tanggal</span>
          <InputText v-model="loiForm.date" type="date" class="w-full" :invalid="Boolean(errors.date)" />
          <small v-if="errors.date" class="text-red-600">{{ errors.date }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Letter Number</span>
          <InputText v-model="loiForm.letter_number" class="w-full" />
        </label>
        <div class="flex justify-end gap-2 border-t border-surface-200 pt-4">
          <Button label="Batal" severity="secondary" outlined @click="dialogVisible = false" />
          <Button type="submit" label="Simpan" icon="pi pi-save" />
        </div>
      </form>
    </Dialog>
  </section>
</template>
