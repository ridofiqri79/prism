import { computed, reactive, watch } from 'vue'
import { assignFormErrors } from '@/utils/form-errors'
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
    cumulative_disbursement: 0,
  }
}



function daysBetween(start: string, end: string) {
  if (!start || !end) return 0
  const startTime = new Date(start).getTime()
  const endTime = new Date(end).getTime()
  if (Number.isNaN(startTime) || Number.isNaN(endTime)) return 0
  return Math.round((endTime - startTime) / 86_400_000)
}

function normalizeCurrency(value?: string | null) {
  return (value || 'USD').trim().toUpperCase()
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
          original_closing_date: initialData.original_closing_date ?? '',
          closing_date: initialData.closing_date,
          currency: initialData.currency,
          amount_original: initialData.amount_original,
          amount_usd: initialData.amount_usd,
          cumulative_disbursement: initialData.cumulative_disbursement,
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
  const isUSD = computed(() => normalizeCurrency(values.currency) === 'USD')

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
      values.currency = normalizeCurrency(currency).slice(0, 3)
      if (values.currency === 'USD') {
        values.amount_usd = values.amount_original
      }
    },
  )

  watch(
    () => values.amount_original,
    (amount) => {
      if (isUSD.value) {
        values.amount_usd = amount
      }
    },
  )

  function applyLoanAgreement(data: LoanAgreement) {
    Object.assign(values, {
      dk_project_id: data.dk_project?.id ?? data.dk_project_id ?? '',
      lender_id: data.lender.id,
      loan_code: data.loan_code,
      agreement_date: data.agreement_date,
      effective_date: data.effective_date,
      original_closing_date: data.original_closing_date ?? '',
      closing_date: data.closing_date,
      currency: data.currency,
      amount_original: data.amount_original,
      amount_usd: data.amount_usd,
      cumulative_disbursement: data.cumulative_disbursement,
    })
  }

  function submit(callback: (payload: LoanAgreementPayload) => unknown | Promise<unknown>) {
    return async () => {
      const parsed = loanAgreementSchema.safeParse(values)
      if (!parsed.success) {
        assignFormErrors(errors, parsed.error)
        return
      }

      Object.keys(errors).forEach((key) => {
        delete errors[key as keyof LAFormErrors]
      })
      const payload = {
        ...parsed.data,
        original_closing_date: parsed.data.original_closing_date?.trim() ?? '',
        currency: normalizeCurrency(parsed.data.currency),
        amount_usd:
          normalizeCurrency(parsed.data.currency) === 'USD'
            ? parsed.data.amount_original === 0 && parsed.data.amount_usd !== 0
              ? parsed.data.amount_usd
              : parsed.data.amount_original
            : parsed.data.amount_usd,
        cumulative_disbursement: parsed.data.cumulative_disbursement,
      }
      await callback(payload)
    }
  }

  function reset(preserve: Partial<LoanAgreementPayload> = {}) {
    Object.assign(values, defaultValues(), preserve)
    Object.keys(errors).forEach((key) => {
      delete errors[key as keyof LAFormErrors]
    })
  }

  return {
    values,
    errors,
    selectedDKProject,
    allowedLenderIds,
    isExtended,
    isUSD,
    extensionDays,
    submit,
    applyLoanAgreement,
    reset,
  }
}
