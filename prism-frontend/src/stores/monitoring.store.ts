import { ref } from 'vue'
import { defineStore } from 'pinia'
import { LoanAgreementService } from '@/services/loan-agreement.service'
import { MonitoringService } from '@/services/monitoring.service'
import type { LoanAgreement } from '@/types/loan-agreement.types'
import type {
  MonitoringDisbursement,
  MonitoringListParams,
  MonitoringPayload,
} from '@/types/monitoring.types'

export const useMonitoringStore = defineStore('monitoring', () => {
  const monitorings = ref<MonitoringDisbursement[]>([])
  const currentMonitoring = ref<MonitoringDisbursement | null>(null)
  const currentLA = ref<LoanAgreement | null>(null)
  const loading = ref(false)
  const total = ref(0)

  async function withLoading<T>(action: () => Promise<T>) {
    loading.value = true
    try {
      return await action()
    } finally {
      loading.value = false
    }
  }

  async function fetchLoanAgreement(loanAgreementId: string) {
    return withLoading(async () => {
      currentLA.value = await LoanAgreementService.getLoanAgreement(loanAgreementId)
      return currentLA.value
    })
  }

  async function fetchMonitorings(loanAgreementId: string, params?: MonitoringListParams) {
    return withLoading(async () => {
      const response = await MonitoringService.getMonitorings(loanAgreementId, params)
      monitorings.value = response.data
      total.value = response.meta.total
      return response
    })
  }

  async function fetchMonitoring(loanAgreementId: string, id: string) {
    return withLoading(async () => {
      currentMonitoring.value = await MonitoringService.getMonitoring(loanAgreementId, id)
      return currentMonitoring.value
    })
  }

  async function createMonitoring(loanAgreementId: string, data: MonitoringPayload) {
    return withLoading(async () => {
      const created = await MonitoringService.createMonitoring(loanAgreementId, data)
      monitorings.value = [created, ...monitorings.value]
      total.value += 1
      return created
    })
  }

  async function updateMonitoring(loanAgreementId: string, id: string, data: MonitoringPayload) {
    return withLoading(async () => {
      const updated = await MonitoringService.updateMonitoring(loanAgreementId, id, data)
      monitorings.value = monitorings.value.map((item) => (item.id === id ? updated : item))
      currentMonitoring.value = updated
      return updated
    })
  }

  async function deleteMonitoring(loanAgreementId: string, id: string) {
    return withLoading(async () => {
      await MonitoringService.deleteMonitoring(loanAgreementId, id)
      monitorings.value = monitorings.value.filter((item) => item.id !== id)
      total.value = Math.max(0, total.value - 1)
    })
  }

  function $reset() {
    monitorings.value = []
    currentMonitoring.value = null
    currentLA.value = null
    loading.value = false
    total.value = 0
  }

  return {
    monitorings,
    currentMonitoring,
    currentLA,
    loading,
    total,
    fetchLoanAgreement,
    fetchMonitorings,
    fetchMonitoring,
    createMonitoring,
    updateMonitoring,
    deleteMonitoring,
    $reset,
  }
})
