import { computed, reactive, watch } from 'vue'
import { assignFormErrors } from '@/utils/form-errors'
import { loanAgreementSchema } from '@/schemas/loan-agreement.schema'
import { LoanAgreementService } from '@/services/loan-agreement.service'
import type {
  DKProjectLoanOption,
  LoanAgreement,
  LoanAgreementDKProjectAllocationPayload,
  LoanAgreementPayload,
} from '@/types/loan-agreement.types'
import type { KursTengah } from '@/types/master.types'

export type LAFormErrors = Partial<Record<keyof LoanAgreementPayload, string>>

function defaultValues(): LoanAgreementPayload {
  return {
    dk_project_id: '',
    dk_project_allocations: [],
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

function latestKursForCurrency(items: KursTengah[], currency: string) {
  return (
    [...items]
      .filter((item) => normalizeCurrency(item.currency.code) === normalizeCurrency(currency))
      .sort((a, b) => b.cut_off_date.localeCompare(a.cut_off_date))[0] ?? null
  )
}

function latestGlobalKursCutOffDate(items: KursTengah[]) {
  return (
    [...items].sort((a, b) => b.cut_off_date.localeCompare(a.cut_off_date))[0]?.cut_off_date ?? null
  )
}

function convertToUSD(amount: number, currency: string, kurs: KursTengah | null) {
  if (normalizeCurrency(currency) === 'USD') return amount
  const rate = kurs?.kurs_tengah_bi ?? 0
  if (rate <= 0) return null
  return amount / rate
}

function todayString() {
  const today = new Date()
  const year = today.getFullYear()
  const month = String(today.getMonth() + 1).padStart(2, '0')
  const day = String(today.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function calculateDisbursementRatio(amountUSD: number | null, cumulativeUSD: number | null) {
  if (!amountUSD || cumulativeUSD === null) return null
  return Math.max(0, Math.min(100, (cumulativeUSD / amountUSD) * 100))
}

function calculateEstimatedTimeRatio(
  effectiveDate: string,
  closingDate: string,
  cutOffDate: string | null,
) {
  if (!effectiveDate || !closingDate || !cutOffDate) return null
  const totalDays = daysBetween(effectiveDate, closingDate)
  if (totalDays <= 0) return null
  return (daysBetween(effectiveDate, cutOffDate) / totalDays) * 100
}

function calculatePerformanceValue(
  disbursementRatio: number | null,
  estimatedTimeRatio: number | null,
) {
  if (disbursementRatio === null || estimatedTimeRatio === null || estimatedTimeRatio === 0)
    return null
  return disbursementRatio / estimatedTimeRatio
}

function calculatePerformanceStatus(
  disbursementRatio: number | null,
  estimatedTimeRatio: number | null,
  performanceValue: number | null,
) {
  if (disbursementRatio === null || estimatedTimeRatio === null) return null
  if (disbursementRatio !== 0) {
    if (performanceValue === null) return null
    if (performanceValue <= 0.3) return 'At-Risk'
    if (performanceValue < 1) return 'Behind Schedule'
    return 'On Schedule'
  }
  return estimatedTimeRatio > 71 ? 'At-Risk' : 'Behind Schedule'
}

export function useLAForm(
  initialData?: LoanAgreement | null,
  options?: {
    dkProjects?: () => DKProjectLoanOption[]
    kursTengah?: () => KursTengah[]
  },
) {
  const initialAllocations = initialData ? loanAgreementAllocationsFromData(initialData) : []
  const values = reactive<LoanAgreementPayload>({
    ...defaultValues(),
    ...(initialData
      ? {
          dk_project_id: initialAllocations[0]?.dk_project_id ?? '',
          dk_project_allocations: initialAllocations,
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
  const selectedDKProjects = computed(() => {
    const optionMap = new Map(
      (options?.dkProjects?.() ?? []).map((project) => [project.id, project]),
    )
    return values.dk_project_allocations
      .map((allocation) => optionMap.get(allocation.dk_project_id))
      .filter((project): project is DKProjectLoanOption => Boolean(project))
  })
  const allowedLenderIds = computed(() =>
    LoanAgreementService.getAllowedLenderIdsForProjects(selectedDKProjects.value),
  )
  const allocationTotal = computed(() =>
    values.dk_project_allocations.reduce((sum, item) => sum + item.allocation_original, 0),
  )
  const allocationDifference = computed(() => values.amount_original - allocationTotal.value)
  const extensionDays = computed(() =>
    Math.max(0, daysBetween(values.original_closing_date, values.closing_date)),
  )
  const isExtended = computed(() => extensionDays.value > 0)
  const isUSD = computed(() => normalizeCurrency(values.currency) === 'USD')
  const latestGlobalCutOffDate = computed(() =>
    latestGlobalKursCutOffDate(options?.kursTengah?.() ?? []),
  )
  const latestKursTengah = computed(() =>
    isUSD.value ? null : latestKursForCurrency(options?.kursTengah?.() ?? [], values.currency),
  )
  const kursCutOffDate = computed(
    () =>
      latestKursTengah.value?.cut_off_date ??
      (isUSD.value ? (latestGlobalCutOffDate.value ?? todayString()) : null),
  )
  const calculatedAmountUSD = computed(() =>
    convertToUSD(values.amount_original, values.currency, latestKursTengah.value),
  )
  const calculatedCumulativeDisbursementUSD = computed(() =>
    convertToUSD(values.cumulative_disbursement, values.currency, latestKursTengah.value),
  )
  const disbursementRatio = computed(() =>
    calculateDisbursementRatio(
      calculatedAmountUSD.value,
      calculatedCumulativeDisbursementUSD.value,
    ),
  )
  const estimatedTimeRatio = computed(() =>
    calculateEstimatedTimeRatio(values.effective_date, values.closing_date, kursCutOffDate.value),
  )
  const performanceValue = computed(() =>
    calculatePerformanceValue(disbursementRatio.value, estimatedTimeRatio.value),
  )
  const performanceStatus = computed(() =>
    calculatePerformanceStatus(
      disbursementRatio.value,
      estimatedTimeRatio.value,
      performanceValue.value,
    ),
  )

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
    () => values.amount_original,
    (amount) => {
      const allocation = values.dk_project_allocations[0]
      if (values.dk_project_allocations.length === 1 && allocation) {
        allocation.allocation_original = amount
      }
    },
  )

  watch(
    () => values.currency,
    (currency) => {
      const normalized = normalizeCurrency(currency).slice(0, 3)
      if (values.currency !== normalized) {
        values.currency = normalized
      }
      values.amount_usd = calculatedAmountUSD.value ?? 0
    },
  )

  watch([() => values.amount_original, latestKursTengah, isUSD], () => {
    values.amount_usd = calculatedAmountUSD.value ?? 0
  })

  function applyLoanAgreement(data: LoanAgreement) {
    const allocations = loanAgreementAllocationsFromData(data)
    Object.assign(values, {
      dk_project_id: allocations[0]?.dk_project_id ?? '',
      dk_project_allocations: allocations,
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

  function setDKProjectIDs(ids: string[]) {
    const existing = new Map(
      values.dk_project_allocations.map((allocation) => [allocation.dk_project_id, allocation]),
    )
    values.dk_project_allocations = ids.map((id) => ({
      dk_project_id: id,
      allocation_original:
        ids.length === 1 ? values.amount_original : (existing.get(id)?.allocation_original ?? 0),
    }))
    values.dk_project_id = values.dk_project_allocations[0]?.dk_project_id ?? ''
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
        dk_project_id: parsed.data.dk_project_allocations[0]?.dk_project_id ?? '',
        dk_project_allocations: parsed.data.dk_project_allocations.map((allocation) => ({
          dk_project_id: allocation.dk_project_id,
          allocation_original: allocation.allocation_original,
        })),
        original_closing_date: parsed.data.original_closing_date?.trim() ?? '',
        currency: normalizeCurrency(parsed.data.currency),
        amount_usd:
          convertToUSD(parsed.data.amount_original, parsed.data.currency, latestKursTengah.value) ??
          0,
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
    selectedDKProjects,
    allowedLenderIds,
    allocationTotal,
    allocationDifference,
    isExtended,
    isUSD,
    extensionDays,
    latestKursTengah,
    kursCutOffDate,
    calculatedAmountUSD,
    calculatedCumulativeDisbursementUSD,
    disbursementRatio,
    estimatedTimeRatio,
    performanceValue,
    performanceStatus,
    submit,
    applyLoanAgreement,
    setDKProjectIDs,
    reset,
  }
}

function loanAgreementAllocationsFromData(
  data: LoanAgreement,
): LoanAgreementDKProjectAllocationPayload[] {
  if (data.dk_projects?.length) {
    return data.dk_projects.map((project) => ({
      dk_project_id: project.id,
      allocation_original: project.allocation_original,
    }))
  }
  const fallbackID = data.dk_project?.id ?? data.dk_project_id ?? ''
  if (!fallbackID) return []
  return [{ dk_project_id: fallbackID, allocation_original: data.amount_original }]
}
