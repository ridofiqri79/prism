export interface Country {
  id: string
  name: string
  code: string
}

export type CountryPayload = Omit<Country, 'id'>

export interface Currency {
  id: string
  code: string
  name: string
  symbol?: string
  is_active: boolean
  sort_order: number
}

export type CurrencyPayload = Omit<Currency, 'id'>

export type LenderType = 'Bilateral' | 'Multilateral' | 'KSA'

export interface Lender {
  id: string
  name: string
  short_name?: string
  type: LenderType
  country_id?: string
  country?: Country
}

export interface LenderPayload {
  name: string
  short_name?: string
  type: LenderType
  country_id?: string | null
}

export type InstitutionLevel =
  | 'Kementerian/Badan/Lembaga'
  | 'Eselon I'
  | 'Eselon II'
  | 'BUMN'
  | 'Pemerintah Daerah Tk. I'
  | 'Pemerintah Daerah Tk. II'
  | 'BUMD'
  | 'Lainya'

export interface Institution {
  id: string
  name: string
  short_name?: string
  level: InstitutionLevel
  parent_id?: string
  parent?: Institution
  has_children?: boolean
}

export interface InstitutionPayload {
  name: string
  short_name?: string
  level: InstitutionLevel
  parent_id?: string | null
}

export type RegionType = 'COUNTRY' | 'PROVINCE' | 'CITY'

export interface Region {
  id: string
  code: string
  name: string
  type: RegionType
  parent_code?: string
  has_children?: boolean
}

export interface RegionPayload {
  code: string
  name: string
  type: RegionType
  parent_code?: string | null
}

export interface ProgramTitle {
  id: string
  title: string
  parent_id?: string
  parent?: ProgramTitle
  has_children?: boolean
}

export interface ProgramTitlePayload {
  title: string
  parent_id?: string | null
}

export type BappenasPartnerLevel = 'Eselon I' | 'Eselon II'

export interface BappenasPartner {
  id: string
  name: string
  level: BappenasPartnerLevel
  parent_id?: string
  parent?: BappenasPartner
  has_children?: boolean
}

export interface BappenasPartnerPayload {
  name: string
  level: BappenasPartnerLevel
  parent_id?: string | null
}

export interface Period {
  id: string
  name: string
  year_start: number
  year_end: number
}

export type PeriodPayload = Omit<Period, 'id'>

export interface NationalPriority {
  id: string
  title: string
  period_id: string
  period?: Period
}

export interface NationalPriorityPayload {
  period_id: string
  title: string
}

export interface MasterImportRowError {
  row: number
  message: string
}

export type MasterImportRowStatus = 'create' | 'skip' | 'failed'

export interface MasterImportRowResult {
  row: number
  status: MasterImportRowStatus
  label: string
  message?: string
}

export interface MasterImportSheetResult {
  sheet: string
  inserted: number
  skipped: number
  failed: number
  rows?: MasterImportRowResult[]
  errors?: MasterImportRowError[]
}

export interface MasterImportSummary {
  file_name: string
  total_inserted: number
  total_skipped: number
  total_failed: number
  sheets: MasterImportSheetResult[]
}

export interface ListParams {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
  [key: string]: string | string[] | number | boolean | undefined
}

export type CreatePayload<T extends { id: string }> = Omit<T, 'id'>
export type UpdatePayload<T extends { id: string }> = Partial<CreatePayload<T>>
