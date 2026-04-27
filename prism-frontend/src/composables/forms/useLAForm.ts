import { computed, reactive, watch } from 'vue'
import type { ZodError } from 'zod'
import { loanAgreementSchema } from '@/schemas/loan-agreement.schema'
import { LoanAgreementService } from '@/services/loan-agreement.service'
import type { DKProjectLoanOption, LoanAgreement, LoanAgreementPayload } from '@/types/loan-agreement.types'

export type LAFormErrors = Partial<Record<keyof LoanAgreementPayload, string>>

function defaultValues(): LoanAgreementPayload {
  return {
    dk_project_id: '',
    lender_id: '',
    loan_code: '',
    agreement_date: '',
    effective_date: '',
    original_closing_date: '',
    closing_date: '',
    currency: 'USD',
    amount_original: 0,
    amount_usd: 0,
  }
}

function assignErrors(target: LAFormErrors, error: ZodError) {
  Object.keys(target).forEach((key) => {
    delete target[key as keyof LAFormErrors]
  })

  for (const issue of error.issues) {
    const field = String(issue.path[0]) as keyof LoanAgreementPayload
    if (!target[field]) {
      target[field] = issue.message
    }
  }
}

function daysBetween(start: string, end: string) {
  if (!start || !end) return 0
  const startTime = new Date(start).getTime()
  const endTime = new Date(end).getTime()
  if (Number.isNaN(startTime) || Number.isNaN(endTime)) return 0
  return Math.round((endTime - startTime) / 86_400_000)
}

export function useLAForm(
  initialData?: LoanAgreement | null,
  options?: {
    dkProjects?: () => DKProjectLoanOption[]
  },
) {
  const values = reactive<LoanAgreementPayload>({
    ...defaultValues(),
    ...(initialData
      ? {
          dk_project_id: initialData.dk_project?.id ?? initialData.dk_project_id ?? '',
          lender_id: initialData.lender.id,
          loan_code: initialData.loan_code,
          agreement_date: initialData.agreement_date,
          effective_date: initialData.effective_date,
          original_closing_date: initialData.original_closing_date,
          closing_date: initialData.closing_date,
          currency: initialData.currency,
          amount_original: initialData.amount_original,
          amount_usd: initialData.amount_usd,
        }
      : {}),
  })
  const errors = reactive<LAFormErrors>({})

  const selectedDKProject = computed(
    () => options?.dkProjects?.().find((project) => project.id === values.dk_project_id) ?? null,
  )
  const allowedLenderIds = computed(() => LoanAgreementService.getAllowedLenderIds(selectedDKProject.value))
  const extensionDays = computed(() =>
    Math.max(0, daysBetween(values.original_closing_date, values.closing_date)),
  )
  const isExtended = computed(() => extensionDays.value > 0)

  watch(
    allowedLenderIds,
    (ids) => {
      if (values.lender_id && !ids.includes(values.lender_id)) {
        values.lender_id = ''
      }
    },
    { deep: true },
  )

  watch(
    () => values.currency,
    (currency) => {
      values.currency = currency.toUpperCase().slice(0, 3)
    },
  )

  function applyLoanAgreement(data: LoanAgreement) {
    Object.assign(values, {
      dk_project_id: data.dk_project?.id ?? data.dk_project_id ?? '',
      lender_id: data.lender.id,
      loan_code: data.loan_code,
      agreement_date: data.agreement_date,
      effective_date: data.effective_date,
      original_closing_date: data.original_closing_date,
      closing_date: data.closing_date,
      currency: data.currency,
      amount_original: data.amount_original,
      amount_usd: data.amount_usd,
    })
  }

  function submit(callback: (payload: LoanAgreementPayload) => unknown | Promise<unknown>) {
    return async () => {
      const parsed = loanAgreementSchema.safeParse(values)
      if (!parsed.success) {
        assignErrors(errors, parsed.error)
        return
      }

      Object.keys(errors).forEach((key) => {
        delete errors[key as keyof LAFormErrors]
      })
      await callback(parsed.data)
    }
  }

  return {
    values,
    errors,
    selectedDKProject,
    allowedLenderIds,
    isExtended,
    extensionDays,
    submit,
    applyLoanAgreement,
  }
}
