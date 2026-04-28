import { ref } from 'vue'
import { defineStore } from 'pinia'
import { BlueBookService } from '@/services/blue-book.service'
import type {
  BBProject,
  BBProjectHistoryItem,
  BBProjectPayload,
  BlueBook,
  BlueBookPayload,
  LoI,
  LoIPayload,
} from '@/types/blue-book.types'
import type { ListParams, MasterImportSummary } from '@/types/master.types'

export const useBlueBookStore = defineStore('blueBook', () => {
  const blueBooks = ref<BlueBook[]>([])
  const currentBlueBook = ref<BlueBook | null>(null)
  const projects = ref<BBProject[]>([])
  const projectOptions = ref<BBProject[]>([])
  const currentProject = ref<BBProject | null>(null)
  const projectHistory = ref<BBProjectHistoryItem[]>([])
  const lois = ref<LoI[]>([])
  const loading = ref(false)
  const historyLoading = ref(false)
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

  async function fetchBlueBooks(params?: ListParams) {
    return withLoading(async () => {
      const response = await BlueBookService.getBlueBooks(params)
      blueBooks.value = response.data
      total.value = response.meta.total
      return response
    })
  }

  async function fetchBlueBook(id: string) {
    return withLoading(async () => {
      currentBlueBook.value = await BlueBookService.getBlueBook(id)
      return currentBlueBook.value
    })
  }

  async function createBlueBook(data: BlueBookPayload) {
    return BlueBookService.createBlueBook(data)
  }

  async function updateBlueBook(id: string, data: BlueBookPayload) {
    const result = await BlueBookService.updateBlueBook(id, data)
    currentBlueBook.value = result
    return result
  }

  async function deleteBlueBook(id: string) {
    await BlueBookService.deleteBlueBook(id)
  }

  async function fetchProjects(blueBookId: string, params?: ListParams) {
    return withLoading(async () => {
      const response = await BlueBookService.getProjects(blueBookId, params)
      projects.value = response.data
      projectTotal.value = response.meta.total
      return response
    })
  }

  async function fetchProjectOptions() {
    const response = await BlueBookService.getBlueBooks({ limit: 1000 })
    const nested = await Promise.all(
      response.data.map((blueBook) => BlueBookService.getProjects(blueBook.id, { limit: 1000 })),
    )
    projectOptions.value = nested.flatMap((item) => item.data)
    return projectOptions.value
  }

  async function fetchProject(blueBookId: string, id: string) {
    return withLoading(async () => {
      currentProject.value = await BlueBookService.getProject(blueBookId, id)
      return currentProject.value
    })
  }

  async function fetchProjectHistory(id: string) {
    historyLoading.value = true
    try {
      projectHistory.value = await BlueBookService.getBBProjectHistory(id)
      return projectHistory.value
    } finally {
      historyLoading.value = false
    }
  }

  async function createProject(blueBookId: string, data: BBProjectPayload) {
    return BlueBookService.createProject(blueBookId, data)
  }

  async function updateProject(blueBookId: string, id: string, data: BBProjectPayload) {
    const result = await BlueBookService.updateProject(blueBookId, id, data)
    currentProject.value = result
    return result
  }

  async function deleteProject(blueBookId: string, id: string) {
    await BlueBookService.deleteProject(blueBookId, id)
  }

  async function downloadProjectImportTemplate(blueBookId: string): Promise<Blob> {
    templateDownloading.value = true
    try {
      return await BlueBookService.downloadImportTemplate(blueBookId)
    } finally {
      templateDownloading.value = false
    }
  }

  async function previewProjectImport(
    blueBookId: string,
    file: File,
  ): Promise<MasterImportSummary> {
    importPreviewing.value = true
    try {
      return await BlueBookService.previewImportProjects(blueBookId, file)
    } finally {
      importPreviewing.value = false
    }
  }

  async function importProjects(blueBookId: string, file: File): Promise<MasterImportSummary> {
    importExecuting.value = true
    try {
      const result = await BlueBookService.executeImportProjects(blueBookId, file)
      projectOptions.value = []
      return result
    } finally {
      importExecuting.value = false
    }
  }

  async function fetchLoI(projectId: string) {
    lois.value = await BlueBookService.getLoI(projectId)
    return lois.value
  }

  async function createLoI(projectId: string, data: LoIPayload) {
    const result = await BlueBookService.createLoI(projectId, data)
    await fetchLoI(projectId)
    return result
  }

  async function deleteLoI(projectId: string, id: string) {
    await BlueBookService.deleteLoI(projectId, id)
    await fetchLoI(projectId)
  }

  function $reset() {
    blueBooks.value = []
    currentBlueBook.value = null
    projects.value = []
    projectOptions.value = []
    currentProject.value = null
    projectHistory.value = []
    lois.value = []
    loading.value = false
    historyLoading.value = false
    templateDownloading.value = false
    importPreviewing.value = false
    importExecuting.value = false
    total.value = 0
    projectTotal.value = 0
  }

  return {
    blueBooks,
    currentBlueBook,
    projects,
    projectOptions,
    currentProject,
    projectHistory,
    lois,
    loading,
    historyLoading,
    templateDownloading,
    importPreviewing,
    importExecuting,
    total,
    projectTotal,
    fetchBlueBooks,
    fetchBlueBook,
    createBlueBook,
    updateBlueBook,
    deleteBlueBook,
    fetchProjects,
    fetchProjectOptions,
    fetchProject,
    fetchProjectHistory,
    createProject,
    updateProject,
    deleteProject,
    downloadProjectImportTemplate,
    previewProjectImport,
    importProjects,
    fetchLoI,
    createLoI,
    deleteLoI,
    $reset,
  }
})
