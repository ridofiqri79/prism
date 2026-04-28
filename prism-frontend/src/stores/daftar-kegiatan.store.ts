import { ref } from 'vue'
import { defineStore } from 'pinia'
import { DaftarKegiatanService } from '@/services/daftar-kegiatan.service'
import type {
  DaftarKegiatan,
  DaftarKegiatanPayload,
  DKProject,
  DKProjectPayload,
} from '@/types/daftar-kegiatan.types'
import type { ListParams, MasterImportSummary } from '@/types/master.types'

export const useDaftarKegiatanStore = defineStore('daftarKegiatan', () => {
  const daftarKegiatan = ref<DaftarKegiatan[]>([])
  const currentDK = ref<DaftarKegiatan | null>(null)
  const projects = ref<DKProject[]>([])
  const currentProject = ref<DKProject | null>(null)
  const loading = ref(false)
  const templateDownloading = ref(false)
  const importPreviewing = ref(false)
  const importExecuting = ref(false)
  const total = ref(0)
  const projectTotal = ref(0)

  async function withLoading<T>(action: () => Promise<T>) {
    loading.value = true
    try {
      return await action()
    } finally {
      loading.value = false
    }
  }

  async function fetchDaftarKegiatan(params?: ListParams) {
    return withLoading(async () => {
      const response = await DaftarKegiatanService.getDaftarKegiatan(params)
      daftarKegiatan.value = response.data
      total.value = response.meta.total
      return response
    })
  }

  async function fetchDK(id: string) {
    return withLoading(async () => {
      currentDK.value = await DaftarKegiatanService.getDK(id)
      return currentDK.value
    })
  }

  async function createDK(data: DaftarKegiatanPayload) {
    return DaftarKegiatanService.createDK(data)
  }

  async function updateDK(id: string, data: DaftarKegiatanPayload) {
    const result = await DaftarKegiatanService.updateDK(id, data)
    currentDK.value = result
    return result
  }

  async function deleteDK(id: string) {
    await DaftarKegiatanService.deleteDK(id)
  }

  async function downloadImportTemplate(): Promise<Blob> {
    templateDownloading.value = true
    try {
      return await DaftarKegiatanService.downloadImportTemplate()
    } finally {
      templateDownloading.value = false
    }
  }

  async function previewImport(file: File): Promise<MasterImportSummary> {
    importPreviewing.value = true
    try {
      return await DaftarKegiatanService.previewImport(file)
    } finally {
      importPreviewing.value = false
    }
  }

  async function executeImport(file: File): Promise<MasterImportSummary> {
    importExecuting.value = true
    try {
      const result = await DaftarKegiatanService.executeImport(file)
      daftarKegiatan.value = []
      projects.value = []
      return result
    } finally {
      importExecuting.value = false
    }
  }

  async function fetchProjects(dkId: string, params?: ListParams) {
    return withLoading(async () => {
      const response = await DaftarKegiatanService.getProjects(dkId, params)
      projects.value = response.data
      projectTotal.value = response.meta.total
      return response
    })
  }

  async function fetchProject(dkId: string, id: string) {
    return withLoading(async () => {
      currentProject.value = await DaftarKegiatanService.getProject(dkId, id)
      return currentProject.value
    })
  }

  async function createProject(dkId: string, data: DKProjectPayload) {
    return DaftarKegiatanService.createProject(dkId, data)
  }

  async function updateProject(dkId: string, id: string, data: DKProjectPayload) {
    const result = await DaftarKegiatanService.updateProject(dkId, id, data)
    currentProject.value = result
    return result
  }

  async function deleteProject(dkId: string, id: string) {
    await DaftarKegiatanService.deleteProject(dkId, id)
  }

  function $reset() {
    daftarKegiatan.value = []
    currentDK.value = null
    projects.value = []
    currentProject.value = null
    loading.value = false
    templateDownloading.value = false
    importPreviewing.value = false
    importExecuting.value = false
    total.value = 0
    projectTotal.value = 0
  }

  return {
    daftarKegiatan,
    currentDK,
    projects,
    currentProject,
    loading,
    templateDownloading,
    importPreviewing,
    importExecuting,
    total,
    projectTotal,
    fetchDaftarKegiatan,
    fetchDK,
    createDK,
    updateDK,
    deleteDK,
    downloadImportTemplate,
    previewImport,
    executeImport,
    fetchProjects,
    fetchProject,
    createProject,
    updateProject,
    deleteProject,
    $reset,
  }
})
