import type { PaginationMeta } from '@/types/api.types'
import type { LenderType } from '@/types/master.types'
import type {
  ProjectMasterFundingSummary,
  ProjectMasterRow,
  ProjectMasterSortField,
  ProjectMasterSortOrder,
  ProjectPipelineStatus,
  ProjectStatus,
} from '@/types/project.types'

export type SpatialDistributionLevel = 'province' | 'city'
export type SpatialDistributionMetric = 'count' | 'value'
export type SpatialDistributionRegionType = 'COUNTRY' | 'PROVINCE' | 'CITY'

export interface SpatialDistributionRegionMetric {
  region_id: string
  region_code: string
  region_name: string
  region_type: SpatialDistributionRegionType
  parent_code?: string
  project_count: number
  total_loan_usd: number
}

export interface SpatialDistributionSummary {
  total_regions: number
  active_regions: number
  total_project_count: number
  total_loan_usd: number
  max_project_count: number
  max_loan_usd: number
}

export interface SpatialDistributionChoroplethResponse {
  level: SpatialDistributionLevel
  province_code?: string
  regions: SpatialDistributionRegionMetric[]
  summary: SpatialDistributionSummary
}

export interface SpatialDistributionProjectListResponse {
  level: SpatialDistributionLevel
  region_id: string
  region_code: string
  region_name: string
  region_type: SpatialDistributionRegionType
  parent_code?: string
  data: ProjectMasterRow[]
  meta: PaginationMeta
  summary: ProjectMasterFundingSummary
}

export interface SpatialDistributionParams {
  level?: SpatialDistributionLevel
  province_code?: string
  loan_types?: LenderType[]
  project_statuses?: ProjectStatus[]
  pipeline_statuses?: ProjectPipelineStatus[]
  search?: string
  include_history?: boolean
}

export interface SpatialDistributionProjectParams extends SpatialDistributionParams {
  region_code: string
  page?: number
  limit?: number
  sort?: ProjectMasterSortField
  order?: ProjectMasterSortOrder
}

export interface SpatialMapFeatureProperties {
  REGION_CODE?: string
  PROVINCE_CODE?: string
  name?: string
  provinceName?: string
  typeLabel?: string
}

export interface SpatialMapFeature {
  id?: string | number
  type: 'Feature'
  properties?: SpatialMapFeatureProperties
  geometry: unknown
}

export interface SpatialMapFeatureCollection {
  type: 'FeatureCollection'
  features: SpatialMapFeature[]
}

export interface SpatialCityMapIndexEntry {
  provinceCode: string
  provinceName: string
  featureCount: number
  path: string
}

export interface SpatialCityMapIndex {
  generatedFrom: string
  provinceCount: number
  totalFeatures: number
  provinces: Record<string, SpatialCityMapIndexEntry>
  missing: Array<Record<string, string>>
}
