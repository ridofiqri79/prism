<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import AutoComplete, { type AutoCompleteCompleteEvent } from 'primevue/autocomplete'
import Button from 'primevue/button'
import Message from 'primevue/message'
import SelectButton from 'primevue/selectbutton'
import Tag from 'primevue/tag'
import PageHeader from '@/components/common/PageHeader.vue'
import ProjectJourneyFlow from '@/components/journey/ProjectJourneyFlow.vue'
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

type JourneyView = 'summary' | 'flow' | 'detail'

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

const selectedProject = ref<ProjectJourneyOption | null>(null)
const activeView = ref<JourneyView>('summary')
const bbProjectId = computed(() => String(route.params.bbProjectId ?? ''))
const projectSuggestions = computed<ProjectJourneyOption[]>(() =>
  projectOptions.value.map(toOption),
)
const viewOptions: Array<{ label: string; value: JourneyView; icon: string }> = [
  { label: 'Ringkasan', value: 'summary', icon: 'pi pi-chart-bar' },
  { label: 'Alur Visual', value: 'flow', icon: 'pi pi-share-alt' },
  { label: 'Detail Hierarki', value: 'detail', icon: 'pi pi-sitemap' },
]

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

async function searchProjects(event: AutoCompleteCompleteEvent) {
  await journeyStore.searchProjectOptions(event.query)
}

async function loadJourney(projectId: string) {
  const data = await journeyStore.fetchJourney(projectId)
  if (!data) return
  selectedProject.value =
    projectSuggestions.value.find((project) => project.id === projectId) ?? optionFromJourney(data)
}

async function openSelectedProject() {
  if (!selectedProject.value) return
  await router.push({ name: 'project-journey', params: { bbProjectId: selectedProject.value.id } })
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
    selectedProject.value =
      projectSuggestions.value.find((project) => project.id === bbProjectId.value) ??
      optionFromJourney(journeyData.value)
  }
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Perjalanan Proyek" subtitle="Alur proyek dari Blue Book sampai Monitoring" />

    <section
      class="rounded-lg border border-surface-200 bg-white p-4 shadow-sm shadow-surface-200/40"
    >
      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_auto]">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Cari Proyek Blue Book</span>
          <AutoComplete
            v-model="selectedProject"
            :suggestions="projectSuggestions"
            :loading="searching"
            option-label="label"
            placeholder="Cari kode atau nama proyek"
            dropdown
            force-selection
            class="w-full"
            @complete="searchProjects"
            @item-select="openSelectedProject"
          >
            <template #option="{ option }">
              <div class="min-w-0 space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="font-medium text-surface-900">{{ option.bb_code }}</p>
                  <span
                    v-if="option.blue_book_revision_label"
                    class="text-xs font-medium text-surface-500"
                  >
                    {{ option.blue_book_revision_label }}
                  </span>
                </div>
                <p class="text-xs text-surface-500">{{ option.project_name }}</p>
              </div>
            </template>
          </AutoComplete>
        </label>
        <div class="flex items-end">
          <Button
            label="Lihat Perjalanan"
            icon="pi pi-share-alt"
            :disabled="!selectedProject"
            :loading="loading"
            @click="openSelectedProject"
          />
        </div>
        <div
          v-if="selectedProject"
          class="rounded-lg border border-surface-100 bg-surface-50 px-3 py-2 md:col-span-2"
        >
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-sm font-semibold text-surface-900">{{
              selectedProject.bb_code
            }}</span>
            <Tag
              v-if="selectedProject.blue_book_revision_label"
              :value="selectedProject.blue_book_revision_label"
              severity="secondary"
              rounded
            />
            <Tag
              v-if="selectedProject.has_newer_revision"
              value="Ada revisi lebih baru"
              severity="warn"
              rounded
            />
          </div>
          <p class="mt-1 line-clamp-1 text-sm text-surface-500">
            {{ selectedProject.project_name }}
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
      <div class="animate-pulse space-y-4">
        <div class="h-4 w-48 rounded bg-surface-100" />
        <div class="grid gap-3 md:grid-cols-5">
          <div
            v-for="stage in emptyJourneyStages"
            :key="stage.label"
            class="h-20 rounded-lg bg-surface-100"
          />
        </div>
        <div class="h-32 rounded-lg bg-surface-100" />
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
