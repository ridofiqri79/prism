import { ref } from 'vue'
import { defineStore } from 'pinia'
import { DashboardService } from '@/services/dashboard.service'
import type {
  DashboardBlueBookDistribution,
  DashboardDaftarKegiatanDistribution,
  DashboardGreenBookDistribution,
  DashboardLoanAgreementDistribution,
} from '@/types/dashboard.types'

function emptyBlueBookDistribution(): DashboardBlueBookDistribution {
  return {
    agency_groups: [],
    top_agencies: [],
    programs: [],
  }
}

function emptyGreenBookDistribution(): DashboardGreenBookDistribution {
  return {
    lender_types: [],
    top_lenders: [],
    top_agencies: [],
  }
}

export const useDashboardStore = defineStore('dashboard', () => {
  const blueBookDistribution = ref<DashboardBlueBookDistribution>(emptyBlueBookDistribution())
  const greenBookDistribution = ref<DashboardGreenBookDistribution>(emptyGreenBookDistribution())
  const daftarKegiatanDistribution = ref<DashboardDaftarKegiatanDistribution>(emptyGreenBookDistribution())
  const loanAgreementDistribution = ref<DashboardLoanAgreementDistribution>(emptyGreenBookDistribution())
  const loadingBlueBookDistribution = ref(false)
  const loadingGreenBookDistribution = ref(false)
  const loadingDaftarKegiatanDistribution = ref(false)
  const loadingLoanAgreementDistribution = ref(false)
  const error = ref<string | null>(null)

  async function fetchBlueBookDistribution(periodIds?: string[]) {
    loadingBlueBookDistribution.value = true
    error.value = null

    try {
      blueBookDistribution.value = await DashboardService.getBlueBookDistribution(
        periodIds?.length ? { period_ids: periodIds } : undefined,
      )
      return blueBookDistribution.value
    } catch (err) {
      error.value = 'Gagal mengambil distribusi Blue Book'
      throw err
    } finally {
      loadingBlueBookDistribution.value = false
    }
  }

  async function fetchGreenBookDistribution(periodIds?: string[]) {
    loadingGreenBookDistribution.value = true
    error.value = null

    try {
      greenBookDistribution.value = await DashboardService.getGreenBookDistribution(
        periodIds?.length ? { period_ids: periodIds } : undefined,
      )
      return greenBookDistribution.value
    } catch (err) {
      error.value = 'Gagal mengambil distribusi Green Book'
      throw err
    } finally {
      loadingGreenBookDistribution.value = false
    }
  }

  async function fetchDaftarKegiatanDistribution(periodIds?: string[]) {
    loadingDaftarKegiatanDistribution.value = true
    error.value = null

    try {
      daftarKegiatanDistribution.value = await DashboardService.getDaftarKegiatanDistribution(
        periodIds?.length ? { period_ids: periodIds } : undefined,
      )
      return daftarKegiatanDistribution.value
    } catch (err) {
      error.value = 'Gagal mengambil distribusi Daftar Kegiatan'
      throw err
    } finally {
      loadingDaftarKegiatanDistribution.value = false
    }
  }

  async function fetchLoanAgreementDistribution(periodIds?: string[]) {
    loadingLoanAgreementDistribution.value = true
    error.value = null

    try {
      loanAgreementDistribution.value = await DashboardService.getLoanAgreementDistribution(
        periodIds?.length ? { period_ids: periodIds } : undefined,
      )
      return loanAgreementDistribution.value
    } catch (err) {
      error.value = 'Gagal mengambil distribusi Loan Agreement'
      throw err
    } finally {
      loadingLoanAgreementDistribution.value = false
    }
  }

  function $reset() {
    blueBookDistribution.value = emptyBlueBookDistribution()
    greenBookDistribution.value = emptyGreenBookDistribution()
    daftarKegiatanDistribution.value = emptyGreenBookDistribution()
    loanAgreementDistribution.value = emptyGreenBookDistribution()
    loadingBlueBookDistribution.value = false
    loadingGreenBookDistribution.value = false
    loadingDaftarKegiatanDistribution.value = false
    loadingLoanAgreementDistribution.value = false
    error.value = null
  }

  return {
    blueBookDistribution,
    greenBookDistribution,
    daftarKegiatanDistribution,
    loanAgreementDistribution,
    loadingBlueBookDistribution,
    loadingGreenBookDistribution,
    loadingDaftarKegiatanDistribution,
    loadingLoanAgreementDistribution,
    error,
    fetchBlueBookDistribution,
    fetchGreenBookDistribution,
    fetchDaftarKegiatanDistribution,
    fetchLoanAgreementDistribution,
    $reset,
  }
})
