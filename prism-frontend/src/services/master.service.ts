import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  BappenasPartner,
  Country,
  CreatePayload,
  Institution,
  Lender,
  ListParams,
  NationalPriority,
  Period,
  ProgramTitle,
  Region,
  UpdatePayload,
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

async function createItem<T extends MasterCollection>(endpoint: string, data: CreatePayload<T>) {
  const response = await http.post<ApiResponse<T>>(endpoint, data)

  return response.data.data
}

async function updateItem<T extends MasterCollection>(
  endpoint: string,
  id: string,
  data: UpdatePayload<T>,
) {
  const response = await http.put<ApiResponse<T>>(`${endpoint}/${id}`, data)

  return response.data.data
}

async function deleteItem(endpoint: string, id: string) {
  await http.delete(`${endpoint}/${id}`)
}

export const MasterService = {
  getCountries(params?: ListParams) {
    return getList<Country>('/master/countries', params)
  },
  createCountry(data: CreatePayload<Country>) {
    return createItem<Country>('/master/countries', data)
  },
  updateCountry(id: string, data: UpdatePayload<Country>) {
    return updateItem<Country>('/master/countries', id, data)
  },
  deleteCountry(id: string) {
    return deleteItem('/master/countries', id)
  },

  getLenders(params?: ListParams) {
    return getList<Lender>('/master/lenders', params)
  },
  createLender(data: CreatePayload<Lender>) {
    return createItem<Lender>('/master/lenders', data)
  },
  updateLender(id: string, data: UpdatePayload<Lender>) {
    return updateItem<Lender>('/master/lenders', id, data)
  },
  deleteLender(id: string) {
    return deleteItem('/master/lenders', id)
  },

  getInstitutions(params?: ListParams) {
    return getList<Institution>('/master/institutions', params)
  },
  createInstitution(data: CreatePayload<Institution>) {
    return createItem<Institution>('/master/institutions', data)
  },
  updateInstitution(id: string, data: UpdatePayload<Institution>) {
    return updateItem<Institution>('/master/institutions', id, data)
  },
  deleteInstitution(id: string) {
    return deleteItem('/master/institutions', id)
  },

  getRegions(params?: ListParams) {
    return getList<Region>('/master/regions', params)
  },
  createRegion(data: CreatePayload<Region>) {
    return createItem<Region>('/master/regions', data)
  },
  updateRegion(id: string, data: UpdatePayload<Region>) {
    return updateItem<Region>('/master/regions', id, data)
  },
  deleteRegion(id: string) {
    return deleteItem('/master/regions', id)
  },

  getProgramTitles(params?: ListParams) {
    return getList<ProgramTitle>('/master/program-titles', params)
  },
  createProgramTitle(data: CreatePayload<ProgramTitle>) {
    return createItem<ProgramTitle>('/master/program-titles', data)
  },
  updateProgramTitle(id: string, data: UpdatePayload<ProgramTitle>) {
    return updateItem<ProgramTitle>('/master/program-titles', id, data)
  },
  deleteProgramTitle(id: string) {
    return deleteItem('/master/program-titles', id)
  },

  getBappenasPartners(params?: ListParams) {
    return getList<BappenasPartner>('/master/bappenas-partners', params)
  },
  createBappenasPartner(data: CreatePayload<BappenasPartner>) {
    return createItem<BappenasPartner>('/master/bappenas-partners', data)
  },
  updateBappenasPartner(id: string, data: UpdatePayload<BappenasPartner>) {
    return updateItem<BappenasPartner>('/master/bappenas-partners', id, data)
  },
  deleteBappenasPartner(id: string) {
    return deleteItem('/master/bappenas-partners', id)
  },

  getPeriods(params?: ListParams) {
    return getList<Period>('/master/periods', params)
  },
  createPeriod(data: CreatePayload<Period>) {
    return createItem<Period>('/master/periods', data)
  },
  updatePeriod(id: string, data: UpdatePayload<Period>) {
    return updateItem<Period>('/master/periods', id, data)
  },
  deletePeriod(id: string) {
    return deleteItem('/master/periods', id)
  },

  getNationalPriorities(params?: ListParams) {
    return getList<NationalPriority>('/master/national-priorities', params)
  },
  createNationalPriority(data: CreatePayload<NationalPriority>) {
    return createItem<NationalPriority>('/master/national-priorities', data)
  },
  updateNationalPriority(id: string, data: UpdatePayload<NationalPriority>) {
    return updateItem<NationalPriority>('/master/national-priorities', id, data)
  },
  deleteNationalPriority(id: string) {
    return deleteItem('/master/national-priorities', id)
  },
}
