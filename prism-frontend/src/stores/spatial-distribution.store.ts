import { defineStore } from 'pinia'
import { ref } from 'vue'
import { SpatialDistributionService } from '@/services/spatial-distribution.service'
import type {
  SpatialDistributionChoroplethResponse,
  SpatialDistributionParams,
  SpatialDistributionProjectListResponse,
  SpatialDistributionProjectParams,
} from '@/types/spatial-distribution.types'

function emptyChoropleth(): SpatialDistributionChoroplethResponse {
  return {
    level: 'province',
    regions: [],
    summary: {
      total_regions: 0,
      active_regions: 0,
      total_project_count: 0,
      total_loan_usd: 0,
      max_project_count: 0,
      max_loan_usd: 0,
    },
  }
}

export const useSpatialDistributionStore = defineStore('spatial-distribution', () => {
  const choropleth = ref<SpatialDistributionChoroplethResponse>(emptyChoropleth())
  const projectList = ref<SpatialDistributionProjectListResponse | null>(null)
  const loadingMap = ref(false)
  const loadingProjects = ref(false)
  const error = ref<string | null>(null)
  const projectError = ref<string | null>(null)

  async function fetchChoropleth(params?: SpatialDistributionParams) {
    loadingMap.value = true
    error.value = null
    try {
      choropleth.value = await SpatialDistributionService.getChoropleth(params)
      return choropleth.value
    } catch (err) {
      error.value = 'Gagal memuat data sebaran wilayah.'
      throw err
    } finally {
      loadingMap.value = false
    }
  }

  async function fetchRegionProjects(params: SpatialDistributionProjectParams) {
    loadingProjects.value = true
    projectError.value = null
    try {
      projectList.value = await SpatialDistributionService.getRegionProjects(params)
      return projectList.value
    } catch (err) {
      projectError.value = 'Gagal memuat daftar proyek wilayah.'
      throw err
    } finally {
      loadingProjects.value = false
    }
  }

  function clearProjectList() {
    projectList.value = null
    projectError.value = null
  }

  function $reset() {
    choropleth.value = emptyChoropleth()
    projectList.value = null
    loadingMap.value = false
    loadingProjects.value = false
    error.value = null
    projectError.value = null
  }

  return {
    choropleth,
    projectList,
    loadingMap,
    loadingProjects,
    error,
    projectError,
    fetchChoropleth,
    fetchRegionProjects,
    clearProjectList,
    $reset,
  }
})
