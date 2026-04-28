import http from '@/services/http'
import type { PaginatedResponse } from '@/types/api.types'
import type { ProjectMasterListParams, ProjectMasterRow } from '@/types/project.types'

export const ProjectService = {
  async getProjectMaster(params?: ProjectMasterListParams) {
    const response = await http.get<PaginatedResponse<ProjectMasterRow>>('/projects', { params })

    return response.data
  },
}
