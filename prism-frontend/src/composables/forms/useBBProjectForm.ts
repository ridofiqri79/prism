import { reactive, ref } from 'vue'
import { assignFormErrors } from '@/utils/form-errors'
import { bbProjectSchema } from '@/schemas/blue-book.schema'
import type {
  BBProject,
  BBProjectPayload,
  FundingType,
  LenderIndicationPayload,
  ProjectCostPayload,
} from '@/types/blue-book.types'
import { isRichTextEmpty, sanitizeRichText } from '@/utils/rich-text'

export interface BBProjectFormValues {
  project_identity_id: string
  program_title_id: string
  bappenas_partner_ids: string[]
  bb_code: string
  project_name: string
  duration: number | null
  objective: string
  scope_of_work: string
  outputs: string
  outcomes: string
  executing_agency_ids: string[]
  implementing_agency_ids: string[]
  location_ids: string[]
  national_priority_ids: string[]
}

type BBProjectFormErrors = Partial<Record<keyof BBProjectFormValues, string>>

export const foreignFundingCategories = ['Loan', 'Grant']
export const counterpartFundingCategories = [
  'Central Government',
  'Regional Government',
  'State-Owned Enterprise',
  'Others',
]

export function categoriesForFundingType(type: FundingType) {
  return type === 'Foreign' ? foreignFundingCategories : counterpartFundingCategories
}

function defaultValues(): BBProjectFormValues {
  return {
    project_identity_id: '',
    program_title_id: '',
    bappenas_partner_ids: [],
    bb_code: '',
    project_name: '',
    duration: null,
    objective: '',
    scope_of_work: '',
    outputs: '',
    outcomes: '',
    executing_agency_ids: [],
    implementing_agency_ids: [],
    location_ids: [],
    national_priority_ids: [],
  }
}

function fromProject(project?: BBProject | null): Partial<BBProjectFormValues> {
  if (!project) return {}

  return {
    project_identity_id: project.project_identity_id,
    program_title_id: project.program_title_id ?? project.program_title?.id ?? '',
    bappenas_partner_ids: project.bappenas_partners.map((item) => item.id),
    bb_code: project.bb_code,
    project_name: project.project_name,
    duration: project.duration ?? null,
    objective: project.objective ?? '',
    scope_of_work: project.scope_of_work ?? '',
    outputs: project.outputs ?? '',
    outcomes: project.outcomes ?? '',
    executing_agency_ids: project.executing_agencies.map((item) => item.id),
    implementing_agency_ids: project.implementing_agencies.map((item) => item.id),
    location_ids: project.locations.map((item) => item.id),
    national_priority_ids: project.national_priorities.map((item) => item.id),
  }
}



function richTextPayload(value: string) {
  const sanitized = sanitizeRichText(value)
  return isRichTextEmpty(sanitized) ? null : sanitized
}

export function useBBProjectForm(initialData?: Partial<BBProjectFormValues> | BBProject | null) {
  const initialValues: Partial<BBProjectFormValues> =
    initialData && 'id' in initialData ? fromProject(initialData) : (initialData ?? {})
  const values = reactive<BBProjectFormValues>({
    ...defaultValues(),
    ...initialValues,
  })
  const errors = reactive<BBProjectFormErrors>({})
  const projectCosts = ref<ProjectCostPayload[]>(
    initialData && 'project_costs' in initialData
      ? initialData.project_costs.map((item) => ({
          funding_type: item.funding_type,
          funding_category: item.funding_category,
          amount_usd: item.amount_usd,
        }))
      : [],
  )
  const lenderIndications = ref<LenderIndicationPayload[]>(
    initialData && 'id' in initialData
      ? initialData.lender_indications.map((item) => ({
          lender_id: item.lender.id,
          remarks: item.remarks ?? '',
        }))
      : [],
  )

  function addCost() {
    projectCosts.value.push({ funding_type: 'Foreign', funding_category: 'Loan', amount_usd: 0 })
  }

  function removeCost(index: number) {
    projectCosts.value.splice(index, 1)
  }

  function addIndication() {
    lenderIndications.value.push({ lender_id: '', remarks: '' })
  }

  function removeIndication(index: number) {
    lenderIndications.value.splice(index, 1)
  }

  function toPayload(): BBProjectPayload {
    return {
      ...values,
      project_identity_id: values.project_identity_id || null,
      duration: values.duration ?? null,
      objective: richTextPayload(values.objective),
      scope_of_work: richTextPayload(values.scope_of_work),
      outputs: richTextPayload(values.outputs),
      outcomes: richTextPayload(values.outcomes),
      project_costs: projectCosts.value,
      lender_indications: lenderIndications.value.map((item) => ({
        lender_id: item.lender_id,
        remarks: item.remarks || null,
      })),
    }
  }

  function submit(callback: (payload: BBProjectPayload) => unknown | Promise<unknown>) {
    return async () => {
      const parsed = bbProjectSchema.safeParse(values)
      if (!parsed.success) {
        assignFormErrors(errors, parsed.error)
        return
      }

      Object.keys(errors).forEach((key) => {
        delete errors[key as keyof BBProjectFormValues]
      })
      await callback(toPayload())
    }
  }

  function applyProject(project: BBProject) {
    Object.assign(values, { ...defaultValues(), ...fromProject(project) })
    projectCosts.value = project.project_costs.map((item) => ({
      funding_type: item.funding_type,
      funding_category: item.funding_category,
      amount_usd: item.amount_usd,
    }))
    lenderIndications.value = project.lender_indications.map((item) => ({
      lender_id: item.lender.id,
      remarks: item.remarks ?? '',
    }))
  }

  function reset() {
    Object.assign(values, defaultValues())
    projectCosts.value = []
    lenderIndications.value = []
  }

  return {
    values,
    errors,
    projectCosts,
    lenderIndications,
    addCost,
    removeCost,
    addIndication,
    removeIndication,
    submit,
    applyProject,
    reset,
  }
}
