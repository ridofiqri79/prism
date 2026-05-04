import { ref } from 'vue'
import { defineStore } from 'pinia'
import { BlueBookService } from '@/services/blue-book.service'
import type {
  BBProject,
  BBProjectHistoryItem,
  BBProjectListParams,
  BBProjectPayload,
  BBProjectRevisionSourceOption,
  BlueBook,
  BlueBookListParams,
  BlueBookPayload,
  ImportBBProjectsFromBlueBookPayload,
  LoI,
  LoIPayload,
} from '@/types/blue-book.types'
import type { MasterImportSummary } from '@/types/master.types'

function isRevisionBlueBook(blueBook: BlueBook) {
  return Boolean(blueBook.replaces_blue_book_id) || blueBook.revision_number > 0
}

function isSourceBlueBook(blueBook: BlueBook, targetBlueBook: BlueBook) {
  if (blueBook.id === targetBlueBook.id || blueBook.period.id !== targetBlueBook.period.id) {
    return false
  }

  if (blueBook.id === targetBlueBook.replaces_blue_book_id) {
    return true
  }

  if (blueBook.revision_number < targetBlueBook.revision_number) {
    return true
  }

  return (
    blueBook.revision_number === targetBlueBook.revision_number &&
    Boolean(blueBook.created_at) &&
    Boolean(targetBlueBook.created_at) &&
    String(blueBook.created_at) < String(targetBlueBook.created_at)
  )
}

function compareBlueBookVersionDesc(left: BlueBook, right: BlueBook) {
  if (left.revision_number !== right.revision_number) {
    return right.revision_number - left.revision_number
  }

  const leftYear = left.revision_year ?? 0
  const rightYear = right.revision_year ?? 0
  if (leftYear !== rightYear) {
    return rightYear - leftYear
  }

  return String(right.created_at ?? '').localeCompare(String(left.created_at ?? ''))
}

function normalizeProjectCode(code: string) {
  return code.trim().toLowerCase()
}

function formatBlueBookSourceLabel(blueBook: BlueBook) {
  if (blueBook.revision_number <= 0) {
    return `${blueBook.period.name} - Awal`
  }

  const year = blueBook.revision_year ? ` Tahun ${blueBook.revision_year}` : ''
  return `${blueBook.period.name} - Revisi ke-${blueBook.revision_number}${year}`
}

export const useBlueBookStore = defineStore('blueBook', () => {
  const blueBooks = ref<BlueBook[]>([])
  const currentBlueBook = ref<BlueBook | null>(null)
  const projects = ref<BBProject[]>([])
  const projectOptions = ref<BBProject[]>([])
  const revisionSourceProjectOptions = ref<BBProjectRevisionSourceOption[]>([])
  const currentProject = ref<BBProject | null>(null)
  const projectHistory = ref<BBProjectHistoryItem[]>([])
  const lois = ref<LoI[]>([])
  const loading = ref(false)
  const historyLoading = ref(false)
  const revisionSourceProjectLoading = ref(false)
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

  async function fetchBlueBooks(params?: BlueBookListParams) {
    return withLoading(async () => {
      const response = await BlueBookService.getBlueBooks(params)
      blueBooks.value = response.data
      total.value = response.meta.total
      return response
    })
  }

  async function getBlueBooksByPeriod(periodId: string) {
    const response = await BlueBookService.getBlueBooks({
      period_id: [periodId],
      limit: 1000,
      sort: 'revision_number',
      order: 'desc',
    })
    return response.data
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

  async function fetchProjects(blueBookId: string, params?: BBProjectListParams) {
    return withLoading(async () => {
      const response = await BlueBookService.getProjects(blueBookId, params)
      projects.value = response.data
      projectTotal.value = response.meta.total
      return response
    })
  }

  async function getProjectsByBlueBook(blueBookId: string) {
    const response = await BlueBookService.getProjects(blueBookId, {
      limit: 1000,
      sort: 'bb_code',
      order: 'asc',
    })
    return response.data
  }

  async function fetchProjectOptions() {
    const response = await BlueBookService.getBlueBooks({ limit: 1000 })
    const nested = await Promise.all(
      response.data.map((blueBook) => BlueBookService.getProjects(blueBook.id, { limit: 1000 })),
    )
    projectOptions.value = nested.flatMap((item) => item.data)
    return projectOptions.value
  }

  async function fetchRevisionSourceProjectOptions(targetBlueBook: BlueBook) {
    if (!isRevisionBlueBook(targetBlueBook)) {
      revisionSourceProjectOptions.value = []
      return revisionSourceProjectOptions.value
    }

    revisionSourceProjectLoading.value = true
    try {
      const blueBooksResponse = await BlueBookService.getBlueBooks({
        period_id: [targetBlueBook.period.id],
        limit: 1000,
      })
      const sourceBlueBooks = blueBooksResponse.data
        .filter((blueBook) => isSourceBlueBook(blueBook, targetBlueBook))
        .sort(compareBlueBookVersionDesc)
      const currentProjectsResponse = await BlueBookService.getProjects(targetBlueBook.id, {
        limit: 1000,
      })
      const usedIdentityIds = new Set(
        currentProjectsResponse.data.map((project) => project.project_identity_id),
      )
      const usedCodes = new Set(
        currentProjectsResponse.data.map((project) => normalizeProjectCode(project.bb_code)),
      )
      const sourceProjectsByBook = await Promise.all(
        sourceBlueBooks.map(async (blueBook) => {
          const response = await BlueBookService.getProjects(blueBook.id, { limit: 1000 })
          const sourceBlueBookLabel = formatBlueBookSourceLabel(blueBook)

          return response.data.map((project) => ({
            ...project,
            source_blue_book_id: blueBook.id,
            source_blue_book_label: sourceBlueBookLabel,
          }))
        }),
      )
      const seenIdentityIds = new Set<string>()
      const seenCodes = new Set<string>()

      revisionSourceProjectOptions.value = sourceProjectsByBook
        .flat()
        .filter((project) => {
          const code = normalizeProjectCode(project.bb_code)

          if (usedIdentityIds.has(project.project_identity_id) || usedCodes.has(code)) {
            return false
          }
          if (seenIdentityIds.has(project.project_identity_id) || seenCodes.has(code)) {
            return false
          }

          seenIdentityIds.add(project.project_identity_id)
          seenCodes.add(code)
          return true
        })

      return revisionSourceProjectOptions.value
    } finally {
      revisionSourceProjectLoading.value = false
    }
  }

  function clearRevisionSourceProjectOptions() {
    revisionSourceProjectOptions.value = []
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

  async function importProjectsFromBlueBook(
    blueBookId: string,
    data: ImportBBProjectsFromBlueBookPayload,
  ) {
    return BlueBookService.importProjectsFromBlueBook(blueBookId, data)
  }

  async function fetchLoI(projectId: string) {
    lois.value = await BlueBookService.getLoI(projectId)
    return lois.value
  }

  async function createLoI(projectId: string, data: LoIPayload) {
    const result = await BlueBookService.createLoI(projectId, data)
    await fetchLoI(projectId)
    await fetchProjectHistory(projectId)
    return result
  }

  async function deleteLoI(projectId: string, id: string) {
    await BlueBookService.deleteLoI(projectId, id)
    await fetchLoI(projectId)
    await fetchProjectHistory(projectId)
  }

  function $reset() {
    blueBooks.value = []
    currentBlueBook.value = null
    projects.value = []
    projectOptions.value = []
    revisionSourceProjectOptions.value = []
    currentProject.value = null
    projectHistory.value = []
    lois.value = []
    loading.value = false
    historyLoading.value = false
    revisionSourceProjectLoading.value = false
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
    revisionSourceProjectOptions,
    currentProject,
    projectHistory,
    lois,
    loading,
    historyLoading,
    revisionSourceProjectLoading,
    templateDownloading,
    importPreviewing,
    importExecuting,
    total,
    projectTotal,
    fetchBlueBooks,
    getBlueBooksByPeriod,
    fetchBlueBook,
    createBlueBook,
    updateBlueBook,
    deleteBlueBook,
    fetchProjects,
    getProjectsByBlueBook,
    fetchProjectOptions,
    fetchRevisionSourceProjectOptions,
    clearRevisionSourceProjectOptions,
    fetchProject,
    fetchProjectHistory,
    createProject,
    updateProject,
    deleteProject,
    downloadProjectImportTemplate,
    previewProjectImport,
    importProjects,
    importProjectsFromBlueBook,
    fetchLoI,
    createLoI,
    deleteLoI,
    $reset,
  }
})
