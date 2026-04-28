import { ref } from 'vue'
import { defineStore } from 'pinia'
import { ProjectService } from '@/services/project.service'
import type { ProjectMasterListParams, ProjectMasterRow } from '@/types/project.types'

export const useProjectStore = defineStore('project', () => {
  const projects = ref<ProjectMasterRow[]>([])
  const loading = ref(false)
  const total = ref(0)

  async function fetchProjectMaster(params?: ProjectMasterListParams) {
    loading.value = true
    try {
      const response = await ProjectService.getProjectMaster(params)
      projects.value = response.data
      total.value = response.meta.total
      return response
    } finally {
      loading.value = false
    }
  }

  function $reset() {
    projects.value = []
    loading.value = false
    total.value = 0
  }

  return {
    projects,
    loading,
    total,
    fetchProjectMaster,
    $reset,
  }
})
