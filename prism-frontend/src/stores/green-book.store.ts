import { ref } from 'vue'
import { defineStore } from 'pinia'
import { GreenBookService } from '@/services/green-book.service'
import type {
  GBProject,
  GBProjectPayload,
  GreenBook,
  GreenBookPayload,
} from '@/types/green-book.types'
import type { ListParams } from '@/types/master.types'

export const useGreenBookStore = defineStore('greenBook', () => {
  const greenBooks = ref<GreenBook[]>([])
  const currentGreenBook = ref<GreenBook | null>(null)
  const projects = ref<GBProject[]>([])
  const projectOptions = ref<GBProject[]>([])
  const currentProject = ref<GBProject | null>(null)
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

  async function fetchGreenBooks(params?: ListParams) {
    return withLoading(async () => {
      const response = await GreenBookService.getGreenBooks(params)
      greenBooks.value = response.data
      total.value = response.meta.total
      return response
    })
  }

  async function fetchGreenBook(id: string) {
    return withLoading(async () => {
      currentGreenBook.value = await GreenBookService.getGreenBook(id)
      return currentGreenBook.value
    })
  }

  async function createGreenBook(data: GreenBookPayload) {
    return GreenBookService.createGreenBook(data)
  }

  async function updateGreenBook(id: string, data: GreenBookPayload) {
    const result = await GreenBookService.updateGreenBook(id, data)
    currentGreenBook.value = result
    return result
  }

  async function deleteGreenBook(id: string) {
    await GreenBookService.deleteGreenBook(id)
  }

  async function fetchProjects(greenBookId: string, params?: ListParams) {
    return withLoading(async () => {
      const response = await GreenBookService.getProjects(greenBookId, params)
      projects.value = response.data
      projectTotal.value = response.meta.total
      return response
    })
  }

  async function fetchProjectOptions() {
    const response = await GreenBookService.getGreenBooks({ limit: 1000 })
    const nested = await Promise.all(
      response.data.map((greenBook) =>
        GreenBookService.getProjects(greenBook.id, { limit: 1000 }),
      ),
    )
    projectOptions.value = nested.flatMap((item) => item.data)
    return projectOptions.value
  }

  async function fetchProject(greenBookId: string, id: string) {
    return withLoading(async () => {
      currentProject.value = await GreenBookService.getProject(greenBookId, id)
      return currentProject.value
    })
  }

  async function createProject(greenBookId: string, data: GBProjectPayload) {
    return GreenBookService.createProject(greenBookId, data)
  }

  async function updateProject(greenBookId: string, id: string, data: GBProjectPayload) {
    const result = await GreenBookService.updateProject(greenBookId, id, data)
    currentProject.value = result
    return result
  }

  async function deleteProject(greenBookId: string, id: string) {
    await GreenBookService.deleteProject(greenBookId, id)
  }

  function $reset() {
    greenBooks.value = []
    currentGreenBook.value = null
    projects.value = []
    projectOptions.value = []
    currentProject.value = null
    loading.value = false
    total.value = 0
    projectTotal.value = 0
  }

  return {
    greenBooks,
    currentGreenBook,
    projects,
    projectOptions,
    currentProject,
    loading,
    total,
    projectTotal,
    fetchGreenBooks,
    fetchGreenBook,
    createGreenBook,
    updateGreenBook,
    deleteGreenBook,
    fetchProjects,
    fetchProjectOptions,
    fetchProject,
    createProject,
    updateProject,
    deleteProject,
    $reset,
  }
})
