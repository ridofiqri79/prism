import http from '@/services/http'

import type { ApiResponse } from '@/types/api.types'
import type { JourneyResponse } from '@/types/journey.types'

export const JourneyService = {
  async getJourney(bbProjectId: string) {
    const response = await http.get<ApiResponse<JourneyResponse>>(`/projects/${bbProjectId}/journey`)
    return response.data.data
  },
}
