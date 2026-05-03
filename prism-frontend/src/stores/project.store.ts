import { ref } from 'vue'
import { defineStore } from 'pinia'
import { ProjectService } from '@/services/project.service'
import type {
  ProjectMasterFundingSummary,
  ProjectMasterListParams,
  ProjectMasterRow,
} from '@/types/project.types'

function emptyFundingSummary(): ProjectMasterFundingSummary {
  return {
    total_loan_usd: 0,
    total_grant_usd: 0,
    total_counterpart_usd: 0,
  }
}

export const useProjectStore = defineStore('project', () => {
  const projects = ref<ProjectMasterRow[]>([])
  const loading = ref(false)
  const exporting = ref(false)
  const total = ref(0)
  const fundingSummary = ref<ProjectMasterFundingSummary>(emptyFundingSummary())

  async function fetchProjectMaster(params?: ProjectMasterListParams) {
    loading.value = true
    try {
      const response = await ProjectService.getProjectMaster(params)
      projects.value = response.data
      total.value = response.meta.total
      fundingSummary.value = response.summary ?? emptyFundingSummary()
      return response
    } finally {
      loading.value = false
    }
  }

  async function downloadProjectMasterExport(params?: ProjectMasterListParams): Promise<Blob> {
    exporting.value = true
    try {
      return await ProjectService.downloadProjectMasterExport(params)
    } finally {
      exporting.value = false
    }
  }

  function $reset() {
    projects.value = []
    loading.value = false
    exporting.value = false
    total.value = 0
    fundingSummary.value = emptyFundingSummary()
  }

  return {
    projects,
    loading,
    exporting,
    total,
    fundingSummary,
    fetchProjectMaster,
    downloadProjectMasterExport,
    $reset,
  }
})
