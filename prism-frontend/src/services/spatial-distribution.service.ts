import http from '@/services/http'
import type { ApiResponse } from '@/types/api.types'
import type {
  SpatialDistributionChoroplethResponse,
  SpatialDistributionParams,
  SpatialDistributionProjectListResponse,
  SpatialDistributionProjectParams,
} from '@/types/spatial-distribution.types'

export const SpatialDistributionService = {
  async getChoropleth(params?: SpatialDistributionParams) {
    const response = await http.get<ApiResponse<SpatialDistributionChoroplethResponse>>(
      '/spatial-distribution/choropleth',
      { params },
    )

    return response.data.data
  },

  async getRegionProjects(params: SpatialDistributionProjectParams) {
    const response = await http.get<SpatialDistributionProjectListResponse>(
      '/spatial-distribution/projects',
      { params },
    )

    return response.data
  },
}
