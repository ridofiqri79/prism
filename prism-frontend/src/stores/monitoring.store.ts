import { ref } from 'vue'
import { defineStore } from 'pinia'
import { LoanAgreementService } from '@/services/loan-agreement.service'
import { MonitoringService } from '@/services/monitoring.service'
import type { LoanAgreement } from '@/types/loan-agreement.types'
import type { MasterImportSummary } from '@/types/master.types'
import type {
  MonitoringDisbursement,
  MonitoringLoanAgreementListParams,
  MonitoringLoanAgreementReference,
  MonitoringListParams,
  MonitoringPayload,
} from '@/types/monitoring.types'

export const useMonitoringStore = defineStore('monitoring', () => {
  const loanAgreementReferences = ref<MonitoringLoanAgreementReference[]>([])
  const monitorings = ref<MonitoringDisbursement[]>([])
  const currentMonitoring = ref<MonitoringDisbursement | null>(null)
  const currentLA = ref<LoanAgreement | null>(null)
  const loading = ref(false)
  const templateDownloading = ref(false)
  const importPreviewing = ref(false)
  const importExecuting = ref(false)
  const total = ref(0)
  const referenceTotal = ref(0)

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

  async function fetchLoanAgreementReferences(params?: MonitoringLoanAgreementListParams) {
    return withLoading(async () => {
      const response = await MonitoringService.getLoanAgreementReferences(params)
      loanAgreementReferences.value = response.data
      referenceTotal.value = response.meta.total
      return response
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

  async function downloadImportTemplate(): Promise<Blob> {
    templateDownloading.value = true
    try {
      return await MonitoringService.downloadImportTemplate()
    } finally {
      templateDownloading.value = false
    }
  }

  async function previewImport(file: File): Promise<MasterImportSummary> {
    importPreviewing.value = true
    try {
      return await MonitoringService.previewImport(file)
    } finally {
      importPreviewing.value = false
    }
  }

  async function executeImport(file: File): Promise<MasterImportSummary> {
    importExecuting.value = true
    try {
      const result = await MonitoringService.executeImport(file)
      await fetchLoanAgreementReferences()
      return result
    } finally {
      importExecuting.value = false
    }
  }

  function $reset() {
    loanAgreementReferences.value = []
    monitorings.value = []
    currentMonitoring.value = null
    currentLA.value = null
    loading.value = false
    templateDownloading.value = false
    importPreviewing.value = false
    importExecuting.value = false
    total.value = 0
    referenceTotal.value = 0
  }

  return {
    loanAgreementReferences,
    monitorings,
    currentMonitoring,
    currentLA,
    loading,
    templateDownloading,
    importPreviewing,
    importExecuting,
    total,
    referenceTotal,
    fetchLoanAgreement,
    fetchLoanAgreementReferences,
    fetchMonitorings,
    fetchMonitoring,
    createMonitoring,
    updateMonitoring,
    deleteMonitoring,
    downloadImportTemplate,
    previewImport,
    executeImport,
    $reset,
  }
})
