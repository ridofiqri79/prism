import { ref } from 'vue'
import { defineStore } from 'pinia'
import { isAxiosError } from 'axios'
import { UserService } from '@/services/user.service'
import type { ApiError } from '@/types/api.types'
import type { UserPermission } from '@/types/auth.types'
import type { ListParams } from '@/types/master.types'
import type { AppUser, CreateUserRequest, UpdateUserRequest } from '@/types/user.types'

export const useUserStore = defineStore('user', () => {
  const users = ref<AppUser[]>([])
  const currentUser = ref<AppUser | null>(null)
  const currentPermissions = ref<UserPermission[]>([])
  const loading = ref(false)
  const total = ref(0)
  const error = ref<ApiError | null>(null)

  async function runWithLoading<T>(callback: () => Promise<T>) {
    loading.value = true
    error.value = null

    try {
      return await callback()
    } catch (err) {
      if (isAxiosError<{ error: ApiError }>(err)) {
        error.value = err.response?.data?.error ?? null
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchUsers(params?: ListParams) {
    return runWithLoading(async () => {
      const response = await UserService.getUsers(params)
      users.value = response.data
      total.value = response.meta.total
      return response
    })
  }

  async function fetchUser(id: string) {
    return runWithLoading(async () => {
      currentUser.value = await UserService.getUser(id)
      return currentUser.value
    })
  }

  async function fetchUserPermissions(id: string) {
    return runWithLoading(async () => {
      currentPermissions.value = await UserService.getUserPermissions(id)
      return currentPermissions.value
    })
  }

  async function createUser(data: CreateUserRequest) {
    return runWithLoading(async () => UserService.createUser(data))
  }

  async function updateUser(id: string, data: UpdateUserRequest) {
    return runWithLoading(async () => UserService.updateUser(id, data))
  }

  async function deleteUser(id: string) {
    return runWithLoading(async () => {
      await UserService.deleteUser(id)
      users.value = users.value.filter((user) => user.id !== id)
    })
  }

  async function updatePermissions(id: string, permissions: UserPermission[]) {
    return runWithLoading(async () => {
      currentPermissions.value = await UserService.updatePermissions(id, permissions)
      return currentPermissions.value
    })
  }

  function $reset() {
    users.value = []
    currentUser.value = null
    currentPermissions.value = []
    loading.value = false
    total.value = 0
    error.value = null
  }

  return {
    users,
    currentUser,
    currentPermissions,
    loading,
    total,
    error,
    fetchUsers,
    fetchUser,
    fetchUserPermissions,
    createUser,
    updateUser,
    deleteUser,
    updatePermissions,
    $reset,
  }
})
