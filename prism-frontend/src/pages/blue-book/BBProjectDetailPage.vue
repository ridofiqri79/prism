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
import StatusBadge from '@/components/common/StatusBadge.vue'
import ValueChipList from '@/components/common/ValueChipList.vue'
import { usePermission } from '@/composables/usePermission'
import { useToast } from '@/composables/useToast'
import { loiSchema } from '@/schemas/blue-book.schema'
import { useBlueBookStore } from '@/stores/blue-book.store'
import type { BBProjectHistoryItem, LoIPayload } from '@/types/blue-book.types'
import { isRichTextEmpty, sanitizeRichText } from '@/utils/rich-text'
import { formatDateTime } from '@/utils/formatters'
import { formatBlueBookStatus, toFormErrors, type FormErrors } from './blue-book-page-utils'

type LoIField = keyof LoIPayload

const route = useRoute()
const router = useRouter()
const blueBookStore = useBlueBookStore()
const toast = useToast()
const { can } = usePermission()

const blueBookId = computed(() => String(route.params.bbId ?? ''))
const projectId = computed(() => String(route.params.id ?? ''))
const dialogVisible = ref(false)
const isRevisionHistoryOpen = ref(false)
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

function historyRoute(item: BBProjectHistoryItem) {
  return { name: 'bb-project-detail', params: { bbId: item.blue_book_id, id: item.id } }
}

function historyLabel(item: BBProjectHistoryItem) {
  const year = item.revision_year ? ` / ${item.revision_year}` : ''
  return `${item.book_label} - Rev ${item.revision_number}${year}`
}

function formatProjectStatus(status: string) {
  if (status === 'active') return 'Berlaku'
  return status
}

function revisionSnapshotLabel(item: BBProjectHistoryItem) {
  return item.is_latest ? 'Versi terbaru' : 'Historis'
}

function revisionChangeMeta(item: BBProjectHistoryItem) {
  const parts = []
  if (item.last_changed_by) parts.push(item.last_changed_by)
  if (item.last_changed_at) parts.push(formatDateTime(item.last_changed_at))

  return parts.join(' - ') || 'Belum ada catatan perubahan'
}

function hasRichText(value?: string | null) {
  return !isRichTextEmpty(value)
}

function richTextHtml(value?: string | null) {
  return sanitizeRichText(value)
}

function toNameList(items?: { name?: string; title?: string }[]) {
  return items?.map((item) => item.name ?? item.title).filter((item): item is string => Boolean(item)) ?? []
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
      <div class="rounded-lg border border-surface-200 bg-white p-5">
        <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center">
          <div class="min-w-0 space-y-1.5">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Judul Program</p>
            <p class="truncate text-base font-semibold text-surface-950">{{ programTitleName }}</p>
          </div>
          <div class="flex flex-col gap-2 lg:items-end">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Status Snapshot</p>
            <div class="flex flex-wrap items-center gap-2 lg:justify-end">
              <StatusBadge :status="project.status" :label="formatProjectStatus(project.status)" />
              <Tag v-if="project.is_latest" value="Versi terbaru" severity="success" rounded />
              <Tag
                v-else-if="project.has_newer_revision"
                value="Ada revisi lebih baru"
                severity="warn"
                rounded
              />
            </div>
          </div>
        </div>
      </div>

      <section class="space-y-3 rounded-lg border border-surface-200 bg-white p-5">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2">
            <h2 class="text-lg font-semibold text-surface-950">Histori Revisi</h2>
            <Tag
              :value="`${blueBookStore.projectHistory.length} snapshot`"
              severity="secondary"
              rounded
            />
          </div>
          <Button
            :label="isRevisionHistoryOpen ? 'Tutup' : 'Detail'"
            :icon="isRevisionHistoryOpen ? 'pi pi-chevron-up' : 'pi pi-chevron-down'"
            severity="secondary"
            size="small"
            outlined
            @click="isRevisionHistoryOpen = !isRevisionHistoryOpen"
          />
        </div>
        <ol
          v-if="isRevisionHistoryOpen"
          class="mt-4 overflow-hidden rounded-lg border border-surface-200 bg-surface-0"
        >
          <li
            v-for="item in blueBookStore.projectHistory"
            :key="item.id"
            class="grid gap-4 border-b border-surface-100 px-4 py-4 last:border-b-0 lg:grid-cols-[minmax(0,1fr)_auto]"
          >
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="min-w-0 font-semibold text-surface-950">{{ historyLabel(item) }}</p>
                <span
                  class="rounded border border-surface-200 bg-surface-50 px-2 py-0.5 font-mono text-xs font-semibold text-surface-700"
                >
                  {{ item.bb_code }}
                </span>
                <StatusBadge
                  :status="item.book_status"
                  :label="formatBlueBookStatus(item.book_status)"
                />
                <Tag
                  :value="revisionSnapshotLabel(item)"
                  :severity="item.is_latest ? 'success' : 'secondary'"
                  rounded
                />
              </div>

              <div class="mt-3 space-y-1.5">
                <p v-if="item.last_change_summary" class="text-sm font-medium text-surface-900">
                  {{ item.last_change_summary }}
                </p>
                <p v-else class="text-sm text-surface-500">Belum ada catatan perubahan terakhir.</p>
                <div class="flex flex-wrap items-center gap-2 text-xs text-surface-500">
                  <span
                    class="inline-flex items-center gap-1 rounded-md border px-2 py-1 font-medium"
                    :class="
                      item.used_by_downstream
                        ? 'border-primary-100 bg-primary-50 text-primary-700'
                        : 'border-surface-200 bg-surface-50 text-surface-600'
                    "
                  >
                    <i class="pi pi-link text-[0.7rem]" />
                    {{ item.used_by_downstream ? 'Dipakai tahap berikutnya' : 'Belum dipakai downstream' }}
                  </span>
                  <span>{{ revisionChangeMeta(item) }}</span>
                </div>
              </div>
            </div>

            <div class="flex items-start justify-end">
              <Button
                v-tooltip.top="'Lihat detail snapshot'"
                as="router-link"
                :to="historyRoute(item)"
                icon="pi pi-eye"
                severity="secondary"
                size="small"
                outlined
                rounded
                aria-label="Lihat detail snapshot"
              />
            </div>
          </li>
        </ol>
      </section>

      <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
        <div class="border-b border-surface-100 px-5 py-4">
          <h2 class="text-lg font-semibold text-surface-950">Profil Kelembagaan</h2>
        </div>
        <div class="grid gap-x-8 gap-y-5 p-5 md:grid-cols-2">
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Executing Agency
            </p>
            <ValueChipList :items="executingAgencyNames" />
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Implementing Agency
            </p>
            <ValueChipList :items="implementingAgencyNames" />
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Mitra Kerja Bappenas
            </p>
            <ValueChipList :items="bappenasPartnerNames" />
          </div>
          <div class="min-w-0 space-y-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Lokasi</p>
            <ValueChipList :items="locationNames" />
          </div>
          <div class="min-w-0 space-y-2 md:col-span-2">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Prioritas Nasional
            </p>
            <ValueChipList :items="nationalPriorityNames" />
          </div>
        </div>
      </section>

      <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
        <div class="border-b border-surface-100 px-5 py-4">
          <h2 class="text-lg font-semibold text-surface-950">Rincian Proyek</h2>
        </div>
        <div class="border-b border-surface-100 px-5 py-4">
          <div class="grid gap-2 sm:grid-cols-[11rem_minmax(0,1fr)]">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">Durasi</p>
            <p class="text-sm font-semibold text-surface-950">
              {{ project.duration ? `${project.duration} bulan` : '-' }}
            </p>
          </div>
        </div>
        <div class="divide-y divide-surface-100">
          <div class="grid gap-3 px-5 py-5 lg:grid-cols-[11rem_minmax(0,1fr)]">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Objective
            </p>
            <div
              v-if="hasRichText(project.objective)"
              class="rich-text-display max-w-5xl text-sm leading-7 text-surface-800"
              v-html="richTextHtml(project.objective)"
            />
            <p v-else class="text-sm text-surface-400">-</p>
          </div>
          <div class="grid gap-3 px-5 py-5 lg:grid-cols-[11rem_minmax(0,1fr)]">
            <p class="text-xs font-semibold uppercase tracking-wide text-surface-500">
              Scope of Work
            </p>
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
          <span class="text-sm font-medium text-surface-700">Subject</span>
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

<style scoped>
.rich-text-display :deep(p) {
  margin: 0 0 0.5rem;
}

.rich-text-display :deep(p:last-child) {
  margin-bottom: 0;
}

.rich-text-display :deep(ol),
.rich-text-display :deep(ul) {
  margin: 0.25rem 0;
  padding-left: 1.25rem;
}
</style>
