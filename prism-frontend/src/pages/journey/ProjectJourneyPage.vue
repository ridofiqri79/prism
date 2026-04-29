<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AutoComplete, { type AutoCompleteCompleteEvent } from 'primevue/autocomplete'
import Button from 'primevue/button'
import Message from 'primevue/message'
import PageHeader from '@/components/common/PageHeader.vue'
import ProjectTimeline from '@/components/journey/ProjectTimeline.vue'
import { DashboardService } from '@/services/dashboard.service'
import { useBlueBookStore } from '@/stores/blue-book.store'
import type { BBProject } from '@/types/blue-book.types'
import type { JourneyResponse } from '@/types/dashboard.types'

type BBProjectOption = BBProject & { label: string }

const route = useRoute()
const router = useRouter()
const blueBookStore = useBlueBookStore()

const selectedProject = ref<BBProjectOption | null>(null)
const filteredProjects = ref<BBProjectOption[]>([])
const journeyData = ref<JourneyResponse | null>(null)
const loading = ref(false)

const bbProjectId = computed(() => String(route.params.bbProjectId ?? ''))
const projectOptions = computed<BBProjectOption[]>(() =>
  blueBookStore.projectOptions
    .filter((project) => project.is_latest !== false)
    .map((project) => ({
      ...project,
      label: `${project.bb_code} - ${project.project_name}`,
    })),
)

function filterProjects(keyword: string) {
  const normalized = keyword.trim().toLowerCase()
  const options = projectOptions.value

  if (!normalized) {
    return options.slice(0, 20)
  }

  return options
    .filter((project) =>
      [project.bb_code, project.project_name]
        .join(' ')
        .toLowerCase()
        .includes(normalized),
    )
    .slice(0, 20)
}

async function searchProjects(event: AutoCompleteCompleteEvent) {
  if (blueBookStore.projectOptions.length === 0) {
    await blueBookStore.fetchProjectOptions()
  }
  filteredProjects.value = filterProjects(event.query)
}

async function loadJourney(projectId: string) {
  if (!projectId) {
    journeyData.value = null
    return
  }

  loading.value = true
  try {
    journeyData.value = await DashboardService.getJourney(projectId)
    selectedProject.value =
      projectOptions.value.find((project) => project.id === projectId) ?? selectedProject.value
  } finally {
    loading.value = false
  }
}

async function openSelectedProject() {
  if (!selectedProject.value) return
  await router.push({ name: 'project-journey', params: { bbProjectId: selectedProject.value.id } })
}

watch(
  bbProjectId,
  (projectId) => {
    void loadJourney(projectId)
  },
  { immediate: true },
)

onMounted(async () => {
  await blueBookStore.fetchProjectOptions()
  filteredProjects.value = filterProjects('')

  if (bbProjectId.value) {
    selectedProject.value =
      projectOptions.value.find((project) => project.id === bbProjectId.value) ?? null
  }
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Project Journey" subtitle="Alur proyek dari Blue Book sampai Monitoring" />

    <section class="rounded-lg border border-surface-200 bg-white p-4">
      <div class="grid gap-3 md:grid-cols-[1fr_auto]">
        <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Cari Proyek Blue Book</span>
          <AutoComplete
            v-model="selectedProject"
            :suggestions="filteredProjects"
            option-label="label"
            placeholder="Cari bb_code atau project_name"
            dropdown
            force-selection
            class="w-full"
            @complete="searchProjects"
            @item-select="openSelectedProject"
          >
            <template #option="{ option }">
              <div class="space-y-1">
                <p class="font-medium text-surface-900">{{ option.bb_code }}</p>
                <p class="text-xs text-surface-500">{{ option.project_name }}</p>
              </div>
            </template>
          </AutoComplete>
        </label>
        <div class="flex items-end">
          <Button
            label="Tampilkan Journey"
            icon="pi pi-share-alt"
            :disabled="!selectedProject"
            :loading="loading"
            @click="openSelectedProject"
          />
        </div>
      </div>
    </section>

    <Message v-if="!bbProjectId && !journeyData" severity="info" :closable="false">
        Pilih Proyek Blue Book untuk melihat timeline perjalanan proyek.
    </Message>

    <ProjectTimeline v-if="journeyData" :journey="journeyData" />
  </section>
</template>
