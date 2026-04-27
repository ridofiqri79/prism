import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { LoanAgreementService } from '@/services/loan-agreement.service'
import type {
  DKProjectLoanOption,
  LoanAgreement,
  LoanAgreementListParams,
  LoanAgreementPayload,
} from '@/types/loan-agreement.types'

export const useLoanAgreementStore = defineStore('loanAgreement', () => {
  const loanAgreements = ref<LoanAgreement[]>([])
  const currentLoanAgreement = ref<LoanAgreement | null>(null)
  const dkProjectOptions = ref<DKProjectLoanOption[]>([])
  const loading = ref(false)
  const total = ref(0)

  const dkProjectOptionMap = computed(
    () => new Map(dkProjectOptions.value.map((project) => [project.id, project])),
  )

  async function withLoading<T>(action: () => Promise<T>) {
    loading.value = true
    try {
      return await action()
    } finally {
      loading.value = false
    }
  }

  async function fetchLoanAgreements(params?: LoanAgreementListParams) {
    return withLoading(async () => {
      const response = await LoanAgreementService.getLoanAgreements(params)
      loanAgreements.value = response.data
      total.value = response.meta.total
      return response
    })
  }

  async function fetchLoanAgreement(id: string) {
    return withLoading(async () => {
      currentLoanAgreement.value = await LoanAgreementService.getLoanAgreement(id)
      return currentLoanAgreement.value
    })
  }

  async function createLoanAgreement(data: LoanAgreementPayload) {
    return LoanAgreementService.createLoanAgreement(data)
  }

  async function updateLoanAgreement(id: string, data: LoanAgreementPayload) {
    const result = await LoanAgreementService.updateLoanAgreement(id, data)
    currentLoanAgreement.value = result
    return result
  }

  async function deleteLoanAgreement(id: string) {
    await LoanAgreementService.deleteLoanAgreement(id)
  }

  async function fetchDKProjectOptions(search?: string) {
    return withLoading(async () => {
      dkProjectOptions.value = await LoanAgreementService.getDKProjectOptions(search)
      return dkProjectOptions.value
    })
  }

  function $reset() {
    loanAgreements.value = []
    currentLoanAgreement.value = null
    dkProjectOptions.value = []
    loading.value = false
    total.value = 0
  }

  return {
    loanAgreements,
    currentLoanAgreement,
    dkProjectOptions,
    dkProjectOptionMap,
    loading,
    total,
    fetchLoanAgreements,
    fetchLoanAgreement,
    createLoanAgreement,
    updateLoanAgreement,
    deleteLoanAgreement,
    fetchDKProjectOptions,
    $reset,
  }
})
