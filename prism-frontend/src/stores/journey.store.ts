import { ref } from 'vue'
import { defineStore } from 'pinia'
import { DashboardService } from '@/services/dashboard.service'
import { ProjectService } from '@/services/project.service'
import type { JourneyResponse } from '@/types/dashboard.types'
import type { ProjectMasterRow } from '@/types/project.types'

export const useJourneyStore = defineStore('journey', () => {
  const journey = ref<JourneyResponse | null>(null)
  const projectOptions = ref<ProjectMasterRow[]>([])
  const loading = ref(false)
  const searching = ref(false)
  const error = ref<string | null>(null)

  async function fetchJourney(bbProjectId: string) {
    const normalized = bbProjectId.trim()
    if (!normalized) {
      journey.value = null
      error.value = null
      return null
    }

    loading.value = true
    error.value = null
    journey.value = null
    try {
      journey.value = await DashboardService.getJourney(normalized)
      return journey.value
    } catch {
      journey.value = null
      error.value = 'Perjalanan proyek tidak dapat dimuat. Periksa pilihan proyek atau coba ulang.'
      return null
    } finally {
      loading.value = false
    }
  }

  async function searchProjectOptions(search?: string) {
    searching.value = true
    try {
      const response = await ProjectService.getProjectMaster({
        page: 1,
        limit: 20,
        sort: 'project_name',
        order: 'asc',
        search: search?.trim() || undefined,
      })
      projectOptions.value = response.data
      return projectOptions.value
    } finally {
      searching.value = false
    }
  }

  function $reset() {
    journey.value = null
    projectOptions.value = []
    loading.value = false
    searching.value = false
    error.value = null
  }

  return {
    journey,
    projectOptions,
    loading,
    searching,
    error,
    fetchJourney,
    searchProjectOptions,
    $reset,
  }
})
