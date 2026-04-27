import { computed, reactive, ref, watch } from 'vue'
import type { ZodError } from 'zod'
import { dkProjectSchema } from '@/schemas/daftar-kegiatan.schema'
import type { BBProject } from '@/types/blue-book.types'
import type {
  DKActivityDetailPayload,
  DKFinancingDetailPayload,
  DKLoanAllocationPayload,
  DKProject,
  DKProjectPayload,
  GBProjectOption,
} from '@/types/daftar-kegiatan.types'

export interface DKProjectFormValues {
  program_title_id: string
  institution_id: string
  duration: string
  objectives: string
  gb_project_ids: string[]
  location_ids: string[]
}

type DKProjectFormErrors = Partial<
  Record<keyof DKProjectFormValues | 'financing_details' | 'loan_allocations' | 'activity_details', string>
>

function defaultValues(): DKProjectFormValues {
  return {
    program_title_id: '',
    institution_id: '',
    duration: '',
    objectives: '',
    gb_project_ids: [],
    location_ids: [],
  }
}

function emptyFinancing(): DKFinancingDetailPayload {
  return {
    lender_id: '',
    currency: 'USD',
    amount_original: 0,
    grant_original: 0,
    counterpart_original: 0,
    amount_usd: 0,
    grant_usd: 0,
    counterpart_usd: 0,
    remarks: null,
  }
}

function emptyAllocation(): DKLoanAllocationPayload {
  return {
    institution_id: '',
    currency: 'USD',
    amount_original: 0,
    grant_original: 0,
    counterpart_original: 0,
    amount_usd: 0,
    grant_usd: 0,
    counterpart_usd: 0,
    remarks: null,
  }
}

function fromProject(project?: DKProject | null): Partial<DKProjectFormValues> {
  if (!project) return {}

  return {
    program_title_id: project.program_title_id ?? project.program_title?.id ?? '',
    institution_id: project.institution_id ?? project.institution?.id ?? '',
    duration: project.duration ?? '',
    objectives: project.objectives ?? '',
    gb_project_ids: project.gb_projects.map((item) => item.id),
    location_ids: project.locations.map((item) => item.id),
  }
}

function assignErrors(target: DKProjectFormErrors, error: ZodError) {
  Object.keys(target).forEach((key) => {
    delete target[key as keyof DKProjectFormErrors]
  })

  for (const issue of error.issues) {
    const field = String(issue.path[0]) as keyof DKProjectFormErrors
    if (!target[field]) {
      target[field] = issue.message
    }
  }
}

function normalizeActivities(rows: DKActivityDetailPayload[]) {
  rows.forEach((row, index) => {
    row.activity_number = index + 1
  })
}

export function useDKProjectForm(
  initialData?: Partial<DKProjectFormValues> | DKProject | null,
  options?: {
    gbProjects?: () => GBProjectOption[]
    bbProjects?: () => BBProject[]
  },
) {
  const initialValues: Partial<DKProjectFormValues> =
    initialData && 'id' in initialData ? fromProject(initialData) : (initialData ?? {})

  const values = reactive<DKProjectFormValues>({
    ...defaultValues(),
    ...initialValues,
  })
  const errors = reactive<DKProjectFormErrors>({})

  const financingDetails = ref<DKFinancingDetailPayload[]>(
    initialData && 'id' in initialData
      ? initialData.financing_details.map((item) => ({
          lender_id: item.lender?.id ?? '',
          currency: item.currency,
          amount_original: item.amount_original,
          grant_original: item.grant_original,
          counterpart_original: item.counterpart_original,
          amount_usd: item.amount_usd,
          grant_usd: item.grant_usd,
          counterpart_usd: item.counterpart_usd,
          remarks: item.remarks ?? null,
        }))
      : [],
  )
  const loanAllocations = ref<DKLoanAllocationPayload[]>(
    initialData && 'id' in initialData
      ? initialData.loan_allocations.map((item) => ({
          institution_id: item.institution?.id ?? '',
          currency: item.currency,
          amount_original: item.amount_original,
          grant_original: item.grant_original,
          counterpart_original: item.counterpart_original,
          amount_usd: item.amount_usd,
          grant_usd: item.grant_usd,
          counterpart_usd: item.counterpart_usd,
          remarks: item.remarks ?? null,
        }))
      : [],
  )
  const activityDetails = ref<DKActivityDetailPayload[]>(
    initialData && 'id' in initialData
      ? initialData.activity_details.map((item, index) => ({
          activity_number: item.activity_number || index + 1,
          activity_name: item.activity_name,
        }))
      : [],
  )

  const selectedGBProjects = computed(() => {
    const gbProjects = options?.gbProjects?.() ?? []
    const selected = new Set(values.gb_project_ids)
    return gbProjects.filter((project) => selected.has(project.id))
  })

  const allowedLenderIds = computed(() => {
    const allowed = new Set<string>()
    const bbProjects = options?.bbProjects?.() ?? []

    selectedGBProjects.value.forEach((project) => {
      project.funding_sources.forEach((source) => allowed.add(source.lender.id))

      project.bb_projects.forEach((summary) => {
        const bbProject = bbProjects.find((item) => item.id === summary.id)
        bbProject?.lender_indications.forEach((indication) => allowed.add(indication.lender.id))
      })
    })

    return [...allowed]
  })

  watch(
    allowedLenderIds,
    (ids) => {
      const allowed = new Set(ids)
      financingDetails.value.forEach((row) => {
        if (row.lender_id && !allowed.has(row.lender_id)) {
          row.lender_id = ''
        }
      })
    },
    { deep: true },
  )

  function addFinancing() {
    financingDetails.value.push(emptyFinancing())
  }

  function removeFinancing(index: number) {
    financingDetails.value.splice(index, 1)
  }

  function addAllocation() {
    loanAllocations.value.push(emptyAllocation())
  }

  function removeAllocation(index: number) {
    loanAllocations.value.splice(index, 1)
  }

  function addActivity() {
    activityDetails.value.push({
      activity_number: activityDetails.value.length + 1,
      activity_name: '',
    })
  }

  function removeActivity(index: number) {
    activityDetails.value.splice(index, 1)
    normalizeActivities(activityDetails.value)
  }

  function toPayload(): DKProjectPayload {
    return {
      program_title_id: values.program_title_id || null,
      institution_id: values.institution_id,
      duration: values.duration || null,
      objectives: values.objectives || null,
      gb_project_ids: values.gb_project_ids,
      location_ids: values.location_ids,
      financing_details: financingDetails.value.map((row) => ({
        ...row,
        remarks: row.remarks || null,
      })),
      loan_allocations: loanAllocations.value.map((row) => ({
        ...row,
        remarks: row.remarks || null,
      })),
      activity_details: activityDetails.value.map((row, index) => ({
        activity_number: index + 1,
        activity_name: row.activity_name,
      })),
    }
  }

  function submit(callback: (payload: DKProjectPayload) => unknown | Promise<unknown>) {
    return async () => {
      normalizeActivities(activityDetails.value)
      const parsed = dkProjectSchema.safeParse(toPayload())
      if (!parsed.success) {
        assignErrors(errors, parsed.error)
        return
      }

      Object.keys(errors).forEach((key) => {
        delete errors[key as keyof DKProjectFormErrors]
      })
      await callback(parsed.data)
    }
  }

  function applyProject(project: DKProject) {
    Object.assign(values, { ...defaultValues(), ...fromProject(project) })
    financingDetails.value = project.financing_details.map((item) => ({
      lender_id: item.lender?.id ?? '',
      currency: item.currency,
      amount_original: item.amount_original,
      grant_original: item.grant_original,
      counterpart_original: item.counterpart_original,
      amount_usd: item.amount_usd,
      grant_usd: item.grant_usd,
      counterpart_usd: item.counterpart_usd,
      remarks: item.remarks ?? null,
    }))
    loanAllocations.value = project.loan_allocations.map((item) => ({
      institution_id: item.institution?.id ?? '',
      currency: item.currency,
      amount_original: item.amount_original,
      grant_original: item.grant_original,
      counterpart_original: item.counterpart_original,
      amount_usd: item.amount_usd,
      grant_usd: item.grant_usd,
      counterpart_usd: item.counterpart_usd,
      remarks: item.remarks ?? null,
    }))
    activityDetails.value = project.activity_details.map((item, index) => ({
      activity_number: item.activity_number || index + 1,
      activity_name: item.activity_name,
    }))
    normalizeActivities(activityDetails.value)
  }

  return {
    values,
    errors,
    financingDetails,
    addFinancing,
    removeFinancing,
    loanAllocations,
    addAllocation,
    removeAllocation,
    activityDetails,
    addActivity,
    removeActivity,
    allowedLenderIds,
    selectedGBProjects,
    submit,
    applyProject,
  }
}
