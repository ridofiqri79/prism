import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  BappenasPartner,
  BappenasPartnerPayload,
  Country,
  CountryPayload,
  Institution,
  InstitutionPayload,
  Lender,
  LenderPayload,
  ListParams,
  MasterImportSummary,
  NationalPriority,
  NationalPriorityPayload,
  Period,
  PeriodPayload,
  ProgramTitle,
  ProgramTitlePayload,
  Region,
  RegionPayload,
} from '@/types/master.types'

type MasterCollection =
  | Country
  | Lender
  | Institution
  | Region
  | ProgramTitle
  | BappenasPartner
  | Period
  | NationalPriority

async function getList<T extends MasterCollection>(endpoint: string, params?: ListParams) {
  const response = await http.get<PaginatedResponse<T>>(endpoint, { params })

  return response.data
}

async function createItem<T extends MasterCollection, TPayload>(endpoint: string, data: TPayload) {
  const response = await http.post<ApiResponse<T>>(endpoint, data)

  return response.data.data
}

async function updateItem<T extends MasterCollection, TPayload>(
  endpoint: string,
  id: string,
  data: Partial<TPayload>,
) {
  const response = await http.put<ApiResponse<T>>(`${endpoint}/${id}`, data)

  return response.data.data
}

async function deleteItem(endpoint: string, id: string) {
  await http.delete(`${endpoint}/${id}`)
}

export const MasterService = {
  async downloadImportTemplate() {
    const response = await http.get<Blob>('/master/import-data/template', {
      responseType: 'blob',
    })

    return response.data
  },

  async previewImportData(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const response = await http.post<ApiResponse<MasterImportSummary>>(
      '/master/import-data/preview',
      formData,
    )

    return response.data.data
  },

  async executeImportData(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const response = await http.post<ApiResponse<MasterImportSummary>>(
      '/master/import-data/execute',
      formData,
    )

    return response.data.data
  },

  getCountries(params?: ListParams) {
    return getList<Country>('/master/countries', params)
  },
  createCountry(data: CountryPayload) {
    return createItem<Country, CountryPayload>('/master/countries', data)
  },
  updateCountry(id: string, data: Partial<CountryPayload>) {
    return updateItem<Country, CountryPayload>('/master/countries', id, data)
  },
  deleteCountry(id: string) {
    return deleteItem('/master/countries', id)
  },

  getLenders(params?: ListParams) {
    return getList<Lender>('/master/lenders', params)
  },
  createLender(data: LenderPayload) {
    return createItem<Lender, LenderPayload>('/master/lenders', data)
  },
  updateLender(id: string, data: Partial<LenderPayload>) {
    return updateItem<Lender, LenderPayload>('/master/lenders', id, data)
  },
  deleteLender(id: string) {
    return deleteItem('/master/lenders', id)
  },

  getInstitutions(params?: ListParams) {
    return getList<Institution>('/master/institutions', params)
  },
  createInstitution(data: InstitutionPayload) {
    return createItem<Institution, InstitutionPayload>('/master/institutions', data)
  },
  updateInstitution(id: string, data: Partial<InstitutionPayload>) {
    return updateItem<Institution, InstitutionPayload>('/master/institutions', id, data)
  },
  deleteInstitution(id: string) {
    return deleteItem('/master/institutions', id)
  },

  getRegions(params?: ListParams) {
    return getList<Region>('/master/regions', params)
  },
  createRegion(data: RegionPayload) {
    return createItem<Region, RegionPayload>('/master/regions', data)
  },
  updateRegion(id: string, data: Partial<RegionPayload>) {
    return updateItem<Region, RegionPayload>('/master/regions', id, data)
  },
  deleteRegion(id: string) {
    return deleteItem('/master/regions', id)
  },

  getProgramTitles(params?: ListParams) {
    return getList<ProgramTitle>('/master/program-titles', params)
  },
  createProgramTitle(data: ProgramTitlePayload) {
    return createItem<ProgramTitle, ProgramTitlePayload>('/master/program-titles', data)
  },
  updateProgramTitle(id: string, data: Partial<ProgramTitlePayload>) {
    return updateItem<ProgramTitle, ProgramTitlePayload>('/master/program-titles', id, data)
  },
  deleteProgramTitle(id: string) {
    return deleteItem('/master/program-titles', id)
  },

  getBappenasPartners(params?: ListParams) {
    return getList<BappenasPartner>('/master/bappenas-partners', params)
  },
  createBappenasPartner(data: BappenasPartnerPayload) {
    return createItem<BappenasPartner, BappenasPartnerPayload>('/master/bappenas-partners', data)
  },
  updateBappenasPartner(id: string, data: Partial<BappenasPartnerPayload>) {
    return updateItem<BappenasPartner, BappenasPartnerPayload>('/master/bappenas-partners', id, data)
  },
  deleteBappenasPartner(id: string) {
    return deleteItem('/master/bappenas-partners', id)
  },

  getPeriods(params?: ListParams) {
    return getList<Period>('/master/periods', params)
  },
  createPeriod(data: PeriodPayload) {
    return createItem<Period, PeriodPayload>('/master/periods', data)
  },
  updatePeriod(id: string, data: Partial<PeriodPayload>) {
    return updateItem<Period, PeriodPayload>('/master/periods', id, data)
  },
  deletePeriod(id: string) {
    return deleteItem('/master/periods', id)
  },

  getNationalPriorities(params?: ListParams) {
    return getList<NationalPriority>('/master/national-priorities', params)
  },
  createNationalPriority(data: NationalPriorityPayload) {
    return createItem<NationalPriority, NationalPriorityPayload>('/master/national-priorities', data)
  },
  updateNationalPriority(id: string, data: Partial<NationalPriorityPayload>) {
    return updateItem<NationalPriority, NationalPriorityPayload>('/master/national-priorities', id, data)
  },
  deleteNationalPriority(id: string) {
    return deleteItem('/master/national-priorities', id)
  },
}
