import { ref } from 'vue'
import { defineStore } from 'pinia'
import { MasterService } from '@/services/master.service'
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

type MasterKey =
  | 'countries'
  | 'lenders'
  | 'institutions'
  | 'regions'
  | 'programTitles'
  | 'bappenasPartners'
  | 'periods'
  | 'nationalPriorities'

export const useMasterStore = defineStore('master', () => {
  const countries = ref<Country[]>([])
  const lenders = ref<Lender[]>([])
  const institutions = ref<Institution[]>([])
  const regions = ref<Region[]>([])
  const programTitles = ref<ProgramTitle[]>([])
  const bappenasPartners = ref<BappenasPartner[]>([])
  const periods = ref<Period[]>([])
  const nationalPriorities = ref<NationalPriority[]>([])
  const loaded = ref<Record<string, boolean>>({})

  async function fetchCountries(force = false, params?: ListParams) {
    if (loaded.value.countries && !force) return
    countries.value = (await MasterService.getCountries(params)).data
    loaded.value.countries = true
  }

  async function fetchLenders(force = false, params?: ListParams) {
    if (loaded.value.lenders && !force) return
    lenders.value = (await MasterService.getLenders(params)).data
    loaded.value.lenders = true
  }

  async function fetchInstitutions(force = false, params?: ListParams) {
    if (loaded.value.institutions && !force) return
    institutions.value = (await MasterService.getInstitutions(params)).data
    loaded.value.institutions = true
  }

  async function fetchRegions(force = false, params?: ListParams) {
    if (loaded.value.regions && !force) return
    regions.value = (await MasterService.getRegions(params)).data
    loaded.value.regions = true
  }

  async function fetchProgramTitles(force = false, params?: ListParams) {
    if (loaded.value.programTitles && !force) return
    programTitles.value = (await MasterService.getProgramTitles(params)).data
    loaded.value.programTitles = true
  }

  async function fetchBappenasPartners(force = false, params?: ListParams) {
    if (loaded.value.bappenasPartners && !force) return
    bappenasPartners.value = (await MasterService.getBappenasPartners(params)).data
    loaded.value.bappenasPartners = true
  }

  async function fetchPeriods(force = false, params?: ListParams) {
    if (loaded.value.periods && !force) return
    periods.value = (await MasterService.getPeriods(params)).data
    loaded.value.periods = true
  }

  async function fetchNationalPriorities(force = false, params?: ListParams) {
    if (loaded.value.nationalPriorities && !force) return
    nationalPriorities.value = (await MasterService.getNationalPriorities(params)).data
    loaded.value.nationalPriorities = true
  }

  function invalidate(key: MasterKey) {
    loaded.value[key] = false
  }

  async function createCountry(data: CreatePayload<Country>) {
    const result = await MasterService.createCountry(data)
    invalidate('countries')
    return result
  }

  async function updateCountry(id: string, data: UpdatePayload<Country>) {
    const result = await MasterService.updateCountry(id, data)
    invalidate('countries')
    return result
  }

  async function deleteCountry(id: string) {
    await MasterService.deleteCountry(id)
    invalidate('countries')
  }

  async function createLender(data: CreatePayload<Lender>) {
    const result = await MasterService.createLender(data)
    invalidate('lenders')
    return result
  }

  async function updateLender(id: string, data: UpdatePayload<Lender>) {
    const result = await MasterService.updateLender(id, data)
    invalidate('lenders')
    return result
  }

  async function deleteLender(id: string) {
    await MasterService.deleteLender(id)
    invalidate('lenders')
  }

  async function createInstitution(data: CreatePayload<Institution>) {
    const result = await MasterService.createInstitution(data)
    invalidate('institutions')
    return result
  }

  async function updateInstitution(id: string, data: UpdatePayload<Institution>) {
    const result = await MasterService.updateInstitution(id, data)
    invalidate('institutions')
    return result
  }

  async function deleteInstitution(id: string) {
    await MasterService.deleteInstitution(id)
    invalidate('institutions')
  }

  async function createRegion(data: CreatePayload<Region>) {
    const result = await MasterService.createRegion(data)
    invalidate('regions')
    return result
  }

  async function updateRegion(id: string, data: UpdatePayload<Region>) {
    const result = await MasterService.updateRegion(id, data)
    invalidate('regions')
    return result
  }

  async function deleteRegion(id: string) {
    await MasterService.deleteRegion(id)
    invalidate('regions')
  }

  async function createProgramTitle(data: CreatePayload<ProgramTitle>) {
    const result = await MasterService.createProgramTitle(data)
    invalidate('programTitles')
    return result
  }

  async function updateProgramTitle(id: string, data: UpdatePayload<ProgramTitle>) {
    const result = await MasterService.updateProgramTitle(id, data)
    invalidate('programTitles')
    return result
  }

  async function deleteProgramTitle(id: string) {
    await MasterService.deleteProgramTitle(id)
    invalidate('programTitles')
  }

  async function createBappenasPartner(data: CreatePayload<BappenasPartner>) {
    const result = await MasterService.createBappenasPartner(data)
    invalidate('bappenasPartners')
    return result
  }

  async function updateBappenasPartner(id: string, data: UpdatePayload<BappenasPartner>) {
    const result = await MasterService.updateBappenasPartner(id, data)
    invalidate('bappenasPartners')
    return result
  }

  async function deleteBappenasPartner(id: string) {
    await MasterService.deleteBappenasPartner(id)
    invalidate('bappenasPartners')
  }

  async function createPeriod(data: CreatePayload<Period>) {
    const result = await MasterService.createPeriod(data)
    invalidate('periods')
    return result
  }

  async function updatePeriod(id: string, data: UpdatePayload<Period>) {
    const result = await MasterService.updatePeriod(id, data)
    invalidate('periods')
    return result
  }

  async function deletePeriod(id: string) {
    await MasterService.deletePeriod(id)
    invalidate('periods')
  }

  async function createNationalPriority(data: CreatePayload<NationalPriority>) {
    const result = await MasterService.createNationalPriority(data)
    invalidate('nationalPriorities')
    return result
  }

  async function updateNationalPriority(id: string, data: UpdatePayload<NationalPriority>) {
    const result = await MasterService.updateNationalPriority(id, data)
    invalidate('nationalPriorities')
    return result
  }

  async function deleteNationalPriority(id: string) {
    await MasterService.deleteNationalPriority(id)
    invalidate('nationalPriorities')
  }

  function $reset() {
    countries.value = []
    lenders.value = []
    institutions.value = []
    regions.value = []
    programTitles.value = []
    bappenasPartners.value = []
    periods.value = []
    nationalPriorities.value = []
    loaded.value = {}
  }

  return {
    countries,
    lenders,
    institutions,
    regions,
    programTitles,
    bappenasPartners,
    periods,
    nationalPriorities,
    loaded,
    fetchCountries,
    fetchLenders,
    fetchInstitutions,
    fetchRegions,
    fetchProgramTitles,
    fetchBappenasPartners,
    fetchPeriods,
    fetchNationalPriorities,
    createCountry,
    updateCountry,
    deleteCountry,
    createLender,
    updateLender,
    deleteLender,
    createInstitution,
    updateInstitution,
    deleteInstitution,
    createRegion,
    updateRegion,
    deleteRegion,
    createProgramTitle,
    updateProgramTitle,
    deleteProgramTitle,
    createBappenasPartner,
    updateBappenasPartner,
    deleteBappenasPartner,
    createPeriod,
    updatePeriod,
    deletePeriod,
    createNationalPriority,
    updateNationalPriority,
    deleteNationalPriority,
    $reset,
  }
})
