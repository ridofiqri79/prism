import { ref } from 'vue'
import { defineStore } from 'pinia'
import { DaftarKegiatanService } from '@/services/daftar-kegiatan.service'
import type {
  DaftarKegiatan,
  DaftarKegiatanPayload,
  DKProject,
  DKProjectPayload,
} from '@/types/daftar-kegiatan.types'
import type { ListParams } from '@/types/master.types'

export const useDaftarKegiatanStore = defineStore('daftarKegiatan', () => {
  const daftarKegiatan = ref<DaftarKegiatan[]>([])
  const currentDK = ref<DaftarKegiatan | null>(null)
  const projects = ref<DKProject[]>([])
  const currentProject = ref<DKProject | null>(null)
  const loading = ref(false)
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
    total.value = 0
    projectTotal.value = 0
  }

  return {
    daftarKegiatan,
    currentDK,
    projects,
    currentProject,
    loading,
    total,
    projectTotal,
    fetchDaftarKegiatan,
    fetchDK,
    createDK,
    updateDK,
    deleteDK,
    fetchProjects,
    fetchProject,
    createProject,
    updateProject,
    deleteProject,
    $reset,
  }
})
