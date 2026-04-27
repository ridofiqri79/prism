export interface Country {
  id: string
  name: string
  code: string
}

export type LenderType = 'Bilateral' | 'Multilateral' | 'KSA'

export interface Lender {
  id: string
  name: string
  short_name?: string
  type: LenderType
  country?: Country
}

export interface Institution {
  id: string
  name: string
  short_name?: string
  level: string
  parent_id?: string
  parent?: Institution
}

export type RegionType = 'COUNTRY' | 'PROVINCE' | 'CITY'

export interface Region {
  id: string
  code: string
  name: string
  type: RegionType
  parent_code?: string
}

export interface ProgramTitle {
  id: string
  title: string
  parent_id?: string
  parent?: ProgramTitle
}

export type BappenasPartnerLevel = 'Eselon I' | 'Eselon II'

export interface BappenasPartner {
  id: string
  name: string
  level: BappenasPartnerLevel
  parent_id?: string
  parent?: BappenasPartner
}

export interface Period {
  id: string
  name: string
  year_start: number
  year_end: number
}

export interface NationalPriority {
  id: string
  title: string
  period_id: string
  period?: Period
}

export interface ListParams {
  page?: number
  limit?: number
  sort?: string
  order?: 'asc' | 'desc'
  [key: string]: string | number | boolean | undefined
}

export type CreatePayload<T extends { id: string }> = Omit<T, 'id'>
export type UpdatePayload<T extends { id: string }> = Partial<CreatePayload<T>>
