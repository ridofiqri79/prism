import http from '@/services/http'
import type { ProjectMasterListParams, ProjectMasterListResponse } from '@/types/project.types'

export const ProjectService = {
  async getProjectMaster(params?: ProjectMasterListParams) {
    const response = await http.get<ProjectMasterListResponse>('/projects', { params })

    return response.data
  },

  async downloadProjectMasterExport(params?: ProjectMasterListParams) {
    const response = await http.get<Blob>('/projects/export', {
      params,
      responseType: 'blob',
    })

    return response.data
  },
}
