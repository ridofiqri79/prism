<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import LenderIndicationTable from '@/components/blue-book/LenderIndicationTable.vue'
import LoITable from '@/components/blue-book/LoITable.vue'
import ProjectCostTable from '@/components/blue-book/ProjectCostTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import ProjectAuditRail from '@/components/common/ProjectAuditRail.vue'
import ProjectInstitutionGrid from '@/components/common/ProjectInstitutionGrid.vue'
import ProjectRevisionHistory from '@/components/common/ProjectRevisionHistory.vue'
import type { RevisionHistoryItem } from '@/components/common/ProjectRevisionHistory.vue'
import ProjectSnapshotHeader from '@/components/common/ProjectSnapshotHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { loiSchema } from '@/schemas/blue-book.schema'
import { useBlueBookStore } from '@/stores/blue-book.store'
import type { BBProjectHistoryItem, LoIPayload } from '@/types/blue-book.types'
import { isRichTextEmpty, sanitizeRichText } from '@/utils/rich-text'
import { toNameList } from '@/utils/formatters'
import { formatBookStatus } from '@/utils/formatters'
import { toFormErrors, type FormErrors } from './blue-book-page-utils'

type LoIField = keyof LoIPayload

const route = useRoute()
const router = useRouter()
const blueBookStore = useBlueBookStore()
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
const programTitleName = computed(() => project.value?.program_title?.title ?? '-')
const executingAgencyNames = computed(() => toNameList(project.value?.executing_agencies))
const implementingAgencyNames = computed(() => toNameList(project.value?.implementing_agencies))
const bappenasPartnerNames = computed(() => toNameList(project.value?.bappenas_partners))
const locationNames = computed(() => toNameList(project.value?.locations))
const nationalPriorityNames = computed(() => toNameList(project.value?.national_priorities))
const allowedLenderIds = computed(
  () => project.value?.lender_indications.map((item) => item.lender.id) ?? [],
)
const allowedLenderOptions = computed(
  () =>
    project.value?.lender_indications.map((item) => ({
      id: item.lender.id,
      label: item.lender.short_name ? `${item.lender.name} (${item.lender.short_name})` : item.lender.name,
      type: item.lender.type,
    })) ?? [],
)

const revisionHistoryItems = computed<RevisionHistoryItem[]>(() =>
  blueBookStore.projectHistory.map((item) => ({
    id: item.id,
    label: historyLabel(item),
    code: item.bb_code,
    book_status: item.book_status,
    status_label: formatBookStatus(item.book_status),
    is_latest: item.is_latest,
    route: { name: 'bb-project-detail', params: { bbId: item.blue_book_id, id: item.id } },
  })),
)

const auditRailItems = computed(() =>
  blueBookStore.projectHistory.flatMap((item) =>
    (item.audit_entries ?? []).map((entry) => ({
      ...entry,
      snapshot_label: historyLabel(item),
    })),
  ),
)

async function loadData() {
  await Promise.all([
    blueBookStore.fetchProject(blueBookId.value, projectId.value),
    blueBookStore.fetchProjectHistory(projectId.value),
    blueBookStore.fetchLoI(projectId.value),
  ])
}

function historyLabel(item: BBProjectHistoryItem) {
  const year = item.revision_year ? ` / ${item.revision_year}` : ''
  return `${item.book_label} - Rev ${item.revision_number}${year}`
}




function hasRichText(value?: string | null) {
  return !isRichTextEmpty(value)
}

function richTextHtml(value?: string | null) {
  return sanitizeRichText(value)
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
      :title="project?.bb_code ?? 'Detail Proyek Blue Book'"
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
          label="Lihat Perjalanan"
          icon="pi pi-share-alt"
          severity="secondary"
        />
      </template>
    </PageHeader>

    <div v-if="project" class="space-y-6">
      <!-- Top card: Judul Program + Status -->
      <ProjectSnapshotHeader
        :program-title-name="programTitleName"
        :status="project.status"
        :status-label="formatBookStatus(project.status)"
        :is-latest="project.is_latest"
        :has-newer-revision="project.has_newer_revision"
      />

      <!-- Rincian Proyek -->
      <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
        <div class="border-b border-surface-100 px-5 py-4">
          <h2 class="text-lg font-semibold text-surface-950">Rincian Proyek</h2>
        </div>
        <ProjectInstitutionGrid
          :executing-agencies="executingAgencyNames"
          :implementing-agencies="implementingAgencyNames"
          :bappenas-partners="bappenasPartnerNames"
          :locations="locationNames"
          :duration="project.duration"
          :national-priorities="nationalPriorityNames"
          class="border-b border-surface-100"
        />
        <div class="divide-y divide-surface-100">
          <div class="grid gap-3 px-5 py-5 lg:grid-cols-[11rem_minmax(0,1fr)]">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Objective</p>
            <div
              v-if="hasRichText(project.objective)"
              class="rich-text-display max-w-5xl text-sm leading-7 text-surface-800"
              v-html="richTextHtml(project.objective)"
            />
            <p v-else class="text-sm text-surface-400">-</p>
          </div>
          <div class="grid gap-3 px-5 py-5 lg:grid-cols-[11rem_minmax(0,1fr)]">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Scope of Work</p>
            <div
              v-if="hasRichText(project.scope_of_work)"
              class="rich-text-display max-w-5xl text-sm leading-7 text-surface-800"
              v-html="richTextHtml(project.scope_of_work)"
            />
            <p v-else class="text-sm text-surface-400">-</p>
          </div>
          <div class="grid gap-3 px-5 py-5 lg:grid-cols-[11rem_minmax(0,1fr)]">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Outputs</p>
            <div
              v-if="hasRichText(project.outputs)"
              class="rich-text-display max-w-5xl text-sm leading-7 text-surface-800"
              v-html="richTextHtml(project.outputs)"
            />
            <p v-else class="text-sm text-surface-400">-</p>
          </div>
          <div class="grid gap-3 px-5 py-5 lg:grid-cols-[11rem_minmax(0,1fr)]">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Outcomes</p>
            <div
              v-if="hasRichText(project.outcomes)"
              class="rich-text-display max-w-5xl text-sm leading-7 text-surface-800"
              v-html="richTextHtml(project.outcomes)"
            />
            <p v-else class="text-sm text-surface-400">-</p>
          </div>
        </div>
      </section>

      <!-- Biaya Proyek -->
      <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
        <div class="border-b border-surface-100 px-5 py-4">
          <h2 class="text-lg font-semibold text-surface-950">Biaya Proyek</h2>
        </div>
        <div class="p-5">
          <ProjectCostTable :rows="project.project_costs" :editable="false" />
        </div>
      </section>

      <!-- Indikasi Lender -->
      <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
        <div class="border-b border-surface-100 px-5 py-4">
          <h2 class="text-lg font-semibold text-surface-950">Indikasi Lender</h2>
        </div>
        <div class="p-5">
          <LenderIndicationTable :rows="project.lender_indications" :editable="false" />
        </div>
      </section>

      <LoITable
        :rows="blueBookStore.lois"
        :can-add="can('bb_project', 'update')"
        @add="openLoIDialog"
      />

      <ProjectRevisionHistory :items="revisionHistoryItems" />

      <ProjectAuditRail :items="auditRailItems" />
    </div>

    <Dialog
      v-model:visible="dialogVisible"
      modal
      header="Tambah LoI"
      class="w-[36rem] max-w-[95vw]"
    >
      <form class="space-y-4" @submit.prevent="saveLoI">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <Select
            v-model="loiForm.lender_id"
            :options="allowedLenderOptions"
            option-label="label"
            option-value="id"
            placeholder="Pilih lender dari indication proyek"
            filter
            append-to="body"
            :overlay-style="{ minWidth: '100%' }"
            class="w-full"
          >
            <template #option="{ option }">
              <div class="flex w-full items-center justify-between gap-3">
                <span>{{ option.label }}</span>
                <Tag :value="option.type" severity="info" rounded />
              </div>
            </template>
          </Select>
          <small v-if="errors.lender_id" class="text-red-600">{{ errors.lender_id }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Perihal</span>
          <InputText v-model="loiForm.subject" class="w-full" :invalid="Boolean(errors.subject)" />
          <small v-if="errors.subject" class="text-red-600">{{ errors.subject }}</small>
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Tanggal</span>
          <InputText
            v-model="loiForm.date"
            type="date"
            class="w-full"
            :invalid="Boolean(errors.date)"
          />
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
