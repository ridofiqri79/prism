<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import AutoComplete, { type AutoCompleteCompleteEvent } from 'primevue/autocomplete'
import Button from 'primevue/button'
import Message from 'primevue/message'
import PageHeader from '@/components/common/PageHeader.vue'
import ProjectTimeline from '@/components/journey/ProjectTimeline.vue'
import { useJourneyStore } from '@/stores/journey.store'
import type { JourneyResponse } from '@/types/dashboard.types'
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
const bbProjectId = computed(() => String(route.params.bbProjectId ?? ''))
const projectSuggestions = computed<ProjectJourneyOption[]>(() => projectOptions.value.map(toOption))

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
    <PageHeader
      title="Perjalanan Proyek"
      subtitle="Alur proyek dari Blue Book sampai Monitoring"
    />

    <section class="rounded-lg border border-surface-200 bg-white p-4">
      <div class="grid gap-3 md:grid-cols-[1fr_auto]">
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
            label="Tampilkan Perjalanan"
            icon="pi pi-share-alt"
            :disabled="!selectedProject"
            :loading="loading"
            @click="openSelectedProject"
          />
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

    <Message v-else-if="loading && !journeyData" severity="info" :closable="false">
      Memuat perjalanan proyek.
    </Message>

    <Message v-else-if="!bbProjectId && !journeyData" severity="info" :closable="false">
      Pilih Proyek Blue Book untuk melihat timeline perjalanan proyek.
    </Message>

    <ProjectTimeline v-if="journeyData" :journey="journeyData" />
  </section>
</template>
