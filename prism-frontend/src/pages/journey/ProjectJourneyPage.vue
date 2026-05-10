<script setup lang="ts">
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import SelectButton from 'primevue/selectbutton'
import Skeleton from 'primevue/skeleton'
import Tag from 'primevue/tag'
import PageHeader from '@/components/common/PageHeader.vue'
import ProjectJourneySummary from '@/components/journey/ProjectJourneySummary.vue'
import ProjectTimeline from '@/components/journey/ProjectTimeline.vue'
import { useJourneyStore } from '@/stores/journey.store'
import type { JourneyResponse } from '@/types/journey.types'
import type { ProjectMasterRow } from '@/types/project.types'

type ProjectJourneyOption = Pick<
  ProjectMasterRow,
  | 'id'
  | 'blue_book_id'
  | 'project_identity_id'
  | 'bb_code'
  | 'project_name'
  | 'blue_book_revision_label'
  | 'is_latest'
  | 'has_newer_revision'
> & {
  label: string
}
type ProjectJourneySelection = ProjectJourneyOption | string | null

type JourneyView = 'summary' | 'flow' | 'detail'

const ProjectJourneyFlow = defineAsyncComponent(
  () => import('@/components/journey/ProjectJourneyFlow.vue'),
)

const emptyJourneyStages = [
  { label: 'Blue Book', icon: 'pi pi-book', state: 'Aktif' },
  { label: 'Green Book', icon: 'pi pi-folder', state: 'Terkait' },
  { label: 'Daftar Kegiatan', icon: 'pi pi-list', state: 'Snapshot' },
  { label: 'Loan Agreement', icon: 'pi pi-file-edit', state: 'Legal' },
  { label: 'Monitoring', icon: 'pi pi-chart-line', state: 'Realisasi' },
]

const route = useRoute()
const router = useRouter()
const journeyStore = useJourneyStore()
const {
  journey: journeyData,
  projectOptions,
  loading,
  searching,
  error,
} = storeToRefs(journeyStore)

const selectedProject = ref<ProjectJourneySelection>(null)
const searchQuery = ref('')
const searchPanelOpen = ref(false)
const activeView = ref<JourneyView>('summary')
const bbProjectId = computed(() => String(route.params.bbProjectId ?? ''))
const projectSuggestions = computed<ProjectJourneyOption[]>(() =>
  projectOptions.value.map(toOption),
)
const selectedProjectOption = computed(() =>
  isProjectJourneyOption(selectedProject.value) ? selectedProject.value : null,
)
const viewOptions: Array<{ label: string; value: JourneyView; icon: string }> = [
  { label: 'Ringkasan', value: 'summary', icon: 'pi pi-chart-bar' },
  { label: 'Alur Visual', value: 'flow', icon: 'pi pi-share-alt' },
  { label: 'Detail Hierarki', value: 'detail', icon: 'pi pi-sitemap' },
]
let searchTimer: ReturnType<typeof window.setTimeout> | undefined

function toOption(project: ProjectMasterRow): ProjectJourneyOption {
  return {
    id: project.id,
    blue_book_id: project.blue_book_id,
    project_identity_id: project.project_identity_id,
    bb_code: project.bb_code,
    project_name: project.project_name,
    blue_book_revision_label: project.blue_book_revision_label,
    is_latest: project.is_latest,
    has_newer_revision: project.has_newer_revision,
    label: `${project.bb_code} - ${project.project_name}`,
  }
}

function optionFromJourney(journey: JourneyResponse): ProjectJourneyOption {
  return {
    id: journey.bb_project.id,
    blue_book_id: journey.bb_project.blue_book_id ?? '',
    project_identity_id: journey.bb_project.project_identity_id ?? '',
    bb_code: journey.bb_project.bb_code,
    project_name: journey.bb_project.project_name,
    blue_book_revision_label: journey.bb_project.blue_book_revision_label ?? '',
    is_latest: journey.bb_project.is_latest ?? false,
    has_newer_revision: journey.bb_project.has_newer_revision ?? false,
    label: `${journey.bb_project.bb_code} - ${journey.bb_project.project_name}`,
  }
}

function isProjectJourneyOption(value: ProjectJourneySelection): value is ProjectJourneyOption {
  return typeof value === 'object' && value !== null && typeof value.id === 'string' && value.id !== ''
}

function scheduleSearch() {
  selectedProject.value = null
  searchPanelOpen.value = true

  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }

  searchTimer = window.setTimeout(() => {
    void journeyStore.searchProjectOptions(searchQuery.value)
  }, 250)
}

async function openSearchPanel() {
  searchPanelOpen.value = true
  if (projectOptions.value.length === 0) {
    await journeyStore.searchProjectOptions(searchQuery.value)
  }
}

async function loadJourney(projectId: string) {
  const data = await journeyStore.fetchJourney(projectId)
  if (!data) return
  const option = projectSuggestions.value.find((project) => project.id === projectId) ?? optionFromJourney(data)
  selectedProject.value = option
  searchQuery.value = option.label
}

async function openSelectedProject() {
  if (!selectedProjectOption.value) return
  await router.push({ name: 'project-journey', params: { bbProjectId: selectedProjectOption.value.id } })
}

async function selectProject(option: ProjectJourneyOption) {
  selectedProject.value = option
  searchQuery.value = option.label
  searchPanelOpen.value = false
  await router.push({ name: 'project-journey', params: { bbProjectId: option.id } })
}

async function selectFirstSuggestion() {
  if (selectedProjectOption.value) {
    await openSelectedProject()
    return
  }

  const firstProject = projectSuggestions.value[0]
  if (firstProject) {
    await selectProject(firstProject)
  }
}

function closeSearchPanel() {
  searchPanelOpen.value = false
}

async function retryLoad() {
  if (!bbProjectId.value) return
  await loadJourney(bbProjectId.value)
}

watch(
  bbProjectId,
  (projectId) => {
    void loadJourney(projectId)
  },
  { immediate: true },
)

onMounted(async () => {
  await journeyStore.searchProjectOptions('')
  if (bbProjectId.value && !selectedProject.value && journeyData.value) {
    const option =
      projectSuggestions.value.find((project) => project.id === bbProjectId.value) ??
      optionFromJourney(journeyData.value)
    selectedProject.value = option
    searchQuery.value = option.label
  }
})

onBeforeUnmount(() => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Perjalanan Proyek" subtitle="Alur proyek dari Blue Book sampai Monitoring" />

    <section class="rounded-lg border border-surface-200 bg-white p-4 shadow-sm shadow-surface-200/40">
      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto]">
        <div class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Cari Proyek Blue Book</span>
          <div class="relative">
            <i
              class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-surface-400"
            />
            <InputText
              v-model="searchQuery"
              placeholder="Cari kode atau nama proyek"
              class="w-full pl-10"
              autocomplete="off"
              @focus="openSearchPanel"
              @input="scheduleSearch"
              @keydown.enter.prevent="selectFirstSuggestion"
              @keydown.escape="closeSearchPanel"
            />
            <div
              v-if="searchPanelOpen"
              class="absolute left-0 right-0 top-full z-50 mt-2 max-h-80 overflow-auto rounded-lg border border-surface-200 bg-white py-1 shadow-xl shadow-surface-900/10"
            >
              <div v-if="searching" class="px-3 py-3 text-sm text-surface-500">
                Memuat hasil...
              </div>
              <button
                v-for="option in projectSuggestions"
                :key="option.id"
                type="button"
                class="block w-full px-3 py-2 text-left hover:bg-surface-50 focus:bg-surface-50 focus:outline-none"
                @mousedown.prevent="selectProject(option)"
              >
                <span class="flex min-w-0 flex-wrap items-center gap-2">
                  <span class="font-medium text-surface-900">{{ option.bb_code }}</span>
                  <span
                    v-if="option.blue_book_revision_label"
                    class="text-xs font-medium text-surface-500"
                  >
                    {{ option.blue_book_revision_label }}
                  </span>
                </span>
                <span class="mt-1 block truncate text-xs text-surface-500">
                  {{ option.project_name }}
                </span>
              </button>
              <div
                v-if="!searching && projectSuggestions.length === 0"
                class="px-3 py-3 text-sm text-surface-500"
              >
                Tidak ada hasil.
              </div>
            </div>
          </div>
        </div>
        <div class="flex items-end">
          <Button
            label="Lihat Perjalanan"
            icon="pi pi-share-alt"
            :disabled="!selectedProjectOption"
            :loading="loading"
            @click="openSelectedProject"
          />
        </div>
        <div
          v-if="selectedProjectOption"
          class="rounded-lg border border-surface-100 bg-surface-50 px-3 py-2 md:col-span-2"
        >
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-sm font-semibold text-surface-900">{{
              selectedProjectOption.bb_code
            }}</span>
            <Tag
              v-if="selectedProjectOption.blue_book_revision_label"
              :value="selectedProjectOption.blue_book_revision_label"
              severity="secondary"
              rounded
            />
            <Tag
              v-if="selectedProjectOption.has_newer_revision"
              value="Ada revisi lebih baru"
              severity="warn"
              rounded
            />
          </div>
          <p class="mt-1 line-clamp-1 text-sm text-surface-500">
            {{ selectedProjectOption.project_name }}
          </p>
        </div>
      </div>
    </section>

    <Message v-if="error" severity="error" :closable="false">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <span>{{ error }}</span>
        <Button
          v-if="bbProjectId"
          label="Coba ulang"
          icon="pi pi-refresh"
          text
          size="small"
          :loading="loading"
          @click="retryLoad"
        />
      </div>
    </Message>

    <section
      v-else-if="loading && !journeyData"
      class="rounded-lg border border-surface-200 bg-white p-5"
    >
      <div class="space-y-4">
        <Skeleton width="12rem" height="1rem" />
        <div class="grid gap-3 md:grid-cols-5">
          <Skeleton
            v-for="stage in emptyJourneyStages"
            :key="stage.label"
            height="5rem"
            border-radius="0.5rem"
          />
        </div>
        <Skeleton height="8rem" border-radius="0.5rem" />
      </div>
    </section>

    <section
      v-else-if="!bbProjectId && !journeyData"
      class="rounded-lg border border-surface-200 bg-white p-6 shadow-sm shadow-surface-200/40"
    >
      <div class="grid gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(22rem,1.05fr)]">
        <div class="flex min-h-64 flex-col justify-center">
          <span
            class="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-full bg-teal-50 text-prism-teal-deep"
          >
            <i class="pi pi-sitemap text-lg" />
          </span>
          <h2 class="text-lg font-semibold text-surface-950">Pilih Proyek Blue Book</h2>
          <p class="mt-2 max-w-xl text-sm leading-6 text-surface-500">
            Setelah proyek dipilih, halaman ini menampilkan jalur konkret dari Blue Book sampai
            Monitoring Disbursement, termasuk status tahap dan indikator revisi.
          </p>
        </div>

        <div class="grid gap-3 sm:grid-cols-5 xl:self-center">
          <div
            v-for="(stage, index) in emptyJourneyStages"
            :key="stage.label"
            class="relative rounded-lg border border-surface-200 bg-surface-50 p-3"
          >
            <span
              v-if="index < emptyJourneyStages.length - 1"
              class="absolute right-[-0.9rem] top-1/2 hidden h-px w-4 bg-surface-200 sm:block"
            />
            <span
              class="inline-flex h-9 w-9 items-center justify-center rounded-full bg-white text-prism-teal-deep shadow-sm shadow-surface-200/70"
            >
              <i :class="stage.icon" />
            </span>
            <p class="mt-3 text-sm font-semibold text-surface-900">{{ stage.label }}</p>
            <p class="mt-1 text-xs text-surface-500">{{ stage.state }}</p>
          </div>
        </div>
      </div>
    </section>

    <section v-if="journeyData" class="space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-base font-semibold text-surface-950">Visualisasi Perjalanan</h2>
          <p class="text-sm text-surface-500">
            Pilih ringkasan, flow, atau detail hierarki sesuai kebutuhan baca.
          </p>
        </div>
        <SelectButton
          v-model="activeView"
          :options="viewOptions"
          option-label="label"
          option-value="value"
          :allow-empty="false"
          data-key="value"
        >
          <template #option="{ option }">
            <span class="inline-flex items-center gap-2">
              <i :class="option.icon" />
              <span>{{ option.label }}</span>
            </span>
          </template>
        </SelectButton>
      </div>

      <ProjectJourneySummary v-if="activeView === 'summary'" :journey="journeyData" />
      <ProjectJourneyFlow v-else-if="activeView === 'flow'" :journey="journeyData" />
      <ProjectTimeline v-else :journey="journeyData" />
    </section>
  </section>
</template>
