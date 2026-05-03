import { reactive, ref, watch } from 'vue'
import type { ZodError } from 'zod'
import { gbProjectSchema } from '@/schemas/green-book.schema'
import type { BBProject } from '@/types/blue-book.types'
import type {
  GBActivityPayload,
  GBAllocationValues,
  GBDisbursementPlanPayload,
  GBFundingSourcePayload,
  GBProject,
  GBProjectPayload,
} from '@/types/green-book.types'

export interface GBProjectFormValues {
  program_title_id: string
  gb_code: string
  project_name: string
  duration: number | null
  objective: string
  scope_of_project: string
  bb_project_ids: string[]
  bappenas_partner_ids: string[]
  executing_agency_ids: string[]
  implementing_agency_ids: string[]
  location_ids: string[]
}

type GBProjectFormErrors = Partial<Record<keyof GBProjectFormValues, string>>

function defaultValues(): GBProjectFormValues {
  return {
    program_title_id: '',
    gb_code: '',
    project_name: '',
    duration: null,
    objective: '',
    scope_of_project: '',
    bb_project_ids: [],
    bappenas_partner_ids: [],
    executing_agency_ids: [],
    implementing_agency_ids: [],
    location_ids: [],
  }
}

function emptyAllocation(): GBAllocationValues {
  return {
    services: 0,
    constructions: 0,
    goods: 0,
    trainings: 0,
    other: 0,
  }
}

function normalizeCurrency(value?: string | null) {
  return (value || 'USD').trim().toUpperCase()
}

function normalizeUSDAmount(currency: string, original: number, usd: number) {
  if (normalizeCurrency(currency) !== 'USD') return usd
  return original === 0 && usd !== 0 ? usd : original
}

function fromProject(project?: GBProject | null): Partial<GBProjectFormValues> {
  if (!project) return {}

  return {
    program_title_id: project.program_title_id ?? project.program_title?.id ?? '',
    gb_code: project.gb_code,
    project_name: project.project_name,
    duration: project.duration ?? null,
    objective: project.objective ?? '',
    scope_of_project: project.scope_of_project ?? '',
    bb_project_ids: project.bb_projects.map((item) => item.id),
    bappenas_partner_ids: project.bappenas_partners.map((item) => item.id),
    executing_agency_ids: project.executing_agencies.map((item) => item.id),
    implementing_agency_ids: project.implementing_agencies.map((item) => item.id),
    location_ids: project.locations.map((item) => item.id),
  }
}

function assignErrors(target: GBProjectFormErrors, error: ZodError) {
  Object.keys(target).forEach((key) => {
    delete target[key as keyof GBProjectFormValues]
  })

  for (const issue of error.issues) {
    const field = String(issue.path[0]) as keyof GBProjectFormValues
    if (!target[field]) {
      target[field] = issue.message
    }
  }
}

function normalizeActivitySort(rows: GBActivityPayload[]) {
  rows.forEach((row, index) => {
    row.sort_order = index
  })
}

export type GBProjectSourceMode = 'new' | 'existing'

export function useGBProjectForm(initialData?: Partial<GBProjectFormValues> | GBProject | null) {
  const initialValues: Partial<GBProjectFormValues> =
    initialData && 'id' in initialData ? fromProject(initialData) : (initialData ?? {})
  const values = reactive<GBProjectFormValues>({
    ...defaultValues(),
    ...initialValues,
  })
  const errors = reactive<GBProjectFormErrors>({})
  const disbursementError = ref('')

  const activities = ref<GBActivityPayload[]>(
    initialData && 'id' in initialData
      ? initialData.activities.map((item, index) => ({
          activity_name: item.activity_name,
          implementation_location: item.implementation_location ?? '',
          piu: item.piu ?? '',
          sort_order: item.sort_order ?? index,
        }))
      : [],
  )
  const fundingSources = ref<GBFundingSourcePayload[]>(
    initialData && 'id' in initialData
      ? initialData.funding_sources.map((item) => ({
          lender_id: item.lender.id,
          institution_id: item.institution?.id ?? null,
          currency: normalizeCurrency(item.currency),
          loan_original: item.loan_original ?? item.loan_usd,
          grant_original: item.grant_original ?? item.grant_usd,
          local_original: item.local_original ?? item.local_usd,
          loan_usd: item.loan_usd,
          grant_usd: item.grant_usd,
          local_usd: item.local_usd,
        }))
      : [],
  )
  const disbursementPlan = ref<GBDisbursementPlanPayload[]>(
    initialData && 'id' in initialData
      ? initialData.disbursement_plan.map((item) => ({
          year: item.year,
          amount_usd: item.amount_usd,
        }))
      : [],
  )
  const allocationValues = ref<GBAllocationValues[]>(
    initialData && 'id' in initialData
      ? initialData.activities.map((activity) => {
          const allocation = initialData.funding_allocations.find(
            (item) => item.gb_activity_id === activity.id,
          )
          return allocation
            ? {
                services: allocation.services,
                constructions: allocation.constructions,
                goods: allocation.goods,
                trainings: allocation.trainings,
                other: allocation.other,
              }
            : emptyAllocation()
        })
      : [],
  )

  watch(
    activities,
    (newActivities) => {
      while (allocationValues.value.length < newActivities.length) {
        allocationValues.value.push(emptyAllocation())
      }
      allocationValues.value.length = newActivities.length
    },
    { deep: true, immediate: true },
  )

  function addActivity() {
    activities.value.push({
      activity_name: '',
      implementation_location: '',
      piu: '',
      sort_order: activities.value.length,
    })
  }

  function removeActivity(index: number) {
    activities.value.splice(index, 1)
    allocationValues.value.splice(index, 1)
    normalizeActivitySort(activities.value)
  }

  function reorderActivities(from: number, to: number) {
    if (to < 0 || to >= activities.value.length) return
    const [activity] = activities.value.splice(from, 1)
    const [allocation] = allocationValues.value.splice(from, 1)
    if (!activity || !allocation) return
    activities.value.splice(to, 0, activity)
    allocationValues.value.splice(to, 0, allocation)
    normalizeActivitySort(activities.value)
  }

  function addFundingSource() {
    fundingSources.value.push({
      lender_id: '',
      institution_id: null,
      currency: 'USD',
      loan_original: 0,
      grant_original: 0,
      local_original: 0,
      loan_usd: 0,
      grant_usd: 0,
      local_usd: 0,
    })
  }

  function removeFundingSource(index: number) {
    fundingSources.value.splice(index, 1)
  }

  function addDisbursementYear(year: number) {
    if (!Number.isInteger(year) || year <= 0) {
      disbursementError.value = 'Tahun wajib diisi'
      return false
    }

    if (disbursementPlan.value.some((row) => row.year === year)) {
      disbursementError.value = `Tahun ${year} sudah ada`
      return false
    }

    disbursementPlan.value.push({ year, amount_usd: 0 })
    disbursementPlan.value.sort((a, b) => a.year - b.year)
    disbursementError.value = ''
    return true
  }

  function updateDisbursementYear(index: number, year: number) {
    if (disbursementPlan.value.some((row, rowIndex) => row.year === year && rowIndex !== index)) {
      disbursementError.value = `Tahun ${year} sudah ada`
      return false
    }

    const target = disbursementPlan.value[index]
    if (!target) return false
    target.year = year
    disbursementPlan.value.sort((a, b) => a.year - b.year)
    disbursementError.value = ''
    return true
  }

  function removeDisbursementYear(index: number) {
    disbursementPlan.value.splice(index, 1)
    disbursementError.value = ''
  }

  function toPayload(): GBProjectPayload {
    return {
      ...values,
      duration: values.duration ?? null,
      objective: values.objective || null,
      scope_of_project: values.scope_of_project || null,
      activities: activities.value.map((item, index) => ({
        activity_name: item.activity_name,
        implementation_location: item.implementation_location || null,
        piu: item.piu || null,
        sort_order: index,
      })),
      funding_sources: fundingSources.value.map((item) => ({
        lender_id: item.lender_id,
        institution_id: item.institution_id || null,
        currency: normalizeCurrency(item.currency),
        loan_original: item.loan_original ?? item.loan_usd,
        grant_original: item.grant_original ?? item.grant_usd,
        local_original: item.local_original ?? item.local_usd,
        loan_usd: normalizeUSDAmount(item.currency, item.loan_original, item.loan_usd),
        grant_usd: normalizeUSDAmount(item.currency, item.grant_original, item.grant_usd),
        local_usd: normalizeUSDAmount(item.currency, item.local_original, item.local_usd),
      })),
      disbursement_plan: disbursementPlan.value,
      funding_allocations: activities.value.map((_, index) => ({
        activity_index: index,
        ...(allocationValues.value[index] ?? emptyAllocation()),
      })),
    }
  }

  function submit(callback: (payload: GBProjectPayload) => unknown | Promise<unknown>) {
    return async () => {
      const parsed = gbProjectSchema.safeParse(values)
      if (!parsed.success) {
        assignErrors(errors, parsed.error)
        return
      }

      Object.keys(errors).forEach((key) => {
        delete errors[key as keyof GBProjectFormValues]
      })
      await callback(toPayload())
    }
  }

  function applyProject(project: GBProject) {
    Object.assign(values, { ...defaultValues(), ...fromProject(project) })
    activities.value = project.activities.map((item, index) => ({
      activity_name: item.activity_name,
      implementation_location: item.implementation_location ?? '',
      piu: item.piu ?? '',
      sort_order: item.sort_order ?? index,
    }))
    fundingSources.value = project.funding_sources.map((item) => ({
      lender_id: item.lender.id,
      institution_id: item.institution?.id ?? null,
      currency: normalizeCurrency(item.currency),
      loan_original: item.loan_original,
      grant_original: item.grant_original,
      local_original: item.local_original,
      loan_usd: item.loan_usd,
      grant_usd: item.grant_usd,
      local_usd: item.local_usd,
    }))
    disbursementPlan.value = project.disbursement_plan.map((item) => ({
      year: item.year,
      amount_usd: item.amount_usd,
    }))
    allocationValues.value = project.activities.map((activity) => {
      const allocation = project.funding_allocations.find(
        (item) => item.gb_activity_id === activity.id,
      )
      return allocation
        ? {
            services: allocation.services,
            constructions: allocation.constructions,
            goods: allocation.goods,
            trainings: allocation.trainings,
            other: allocation.other,
          }
        : emptyAllocation()
    })
  }

  function applyBBProjectSource(project: BBProject, mode: GBProjectSourceMode) {
    const base: Partial<GBProjectFormValues> = {
      bb_project_ids: [project.id],
      gb_code: project.bb_code,
    }

    if (mode === 'existing') {
      Object.assign(base, {
        program_title_id: project.program_title_id ?? project.program_title?.id ?? '',
        project_name: project.project_name,
        duration: project.duration ?? null,
        objective: project.objective ?? '',
        scope_of_project: project.scope_of_work ?? '',
        bappenas_partner_ids: project.bappenas_partners.map((item) => item.id),
        executing_agency_ids: project.executing_agencies.map((item) => item.id),
        implementing_agency_ids: project.implementing_agencies.map((item) => item.id),
        location_ids: project.locations.map((item) => item.id),
      })
    }

    Object.assign(values, base)
  }

  return {
    values,
    errors,
    disbursementError,
    activities,
    addActivity,
    removeActivity,
    reorderActivities,
    fundingSources,
    addFundingSource,
    removeFundingSource,
    disbursementPlan,
    addDisbursementYear,
    updateDisbursementYear,
    removeDisbursementYear,
    allocationValues,
    submit,
    applyProject,
    applyBBProjectSource,
  }
}
