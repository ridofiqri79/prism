import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { AuthService } from '@/services/auth.service'
import type {
  AuthUser,
  PermissionAction,
  UserPermission,
} from '@/types/auth.types'
import {
  clearStoredSession,
  readStoredPermissions,
  readStoredToken,
  readStoredUser,
  storeSession,
} from '@/utils/auth-session'
import { emitLoginRedirect } from '@/utils/app-events'

type PermissionKey = `can_${PermissionAction}`

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(readStoredUser())
  const token = ref<string | null>(readStoredToken())
  const permissions = ref<UserPermission[]>(readStoredPermissions())
  const loading = ref(false)

  const isAuthenticated = computed(() => token.value !== null)

  function clearSession() {
    user.value = null
    token.value = null
    permissions.value = []
    clearStoredSession()
  }

  async function login(payload: { username: string; password: string }) {
    loading.value = true

    try {
      const result = await AuthService.login(payload)

      user.value = result.user
      token.value = result.access_token
      permissions.value = []

      storeSession({
        token: result.access_token,
        user: result.user,
        permissions: [],
      })

      await fetchMe()

      return result
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      if (token.value) {
        await AuthService.logout()
      }
    } catch {
      // Local session cleanup must still happen if the server is unreachable.
    }

    clearSession()
    emitLoginRedirect()
  }

  async function fetchMe() {
    if (!token.value) {
      return
    }

    loading.value = true

    try {
      const session = await AuthService.getMe()

      user.value = {
        id: session.id,
        username: session.username,
        email: session.email,
        role: session.role,
      }
      permissions.value = session.permissions

      storeSession({
        token: token.value,
        user: user.value,
        permissions: session.permissions,
      })
    } finally {
      loading.value = false
    }
  }

  function can(module: string, action: PermissionAction) {
    if (user.value?.role === 'ADMIN') {
      return true
    }

    const permission = permissions.value.find((item) => item.module === module)

    if (!permission) {
      return false
    }

    const permissionKey: PermissionKey = `can_${action}`
    return permission[permissionKey]
  }

  async function restoreSession() {
    const storedToken = readStoredToken()

    if (!storedToken) {
      clearSession()
      return
    }

    token.value = storedToken
    user.value = readStoredUser()
    permissions.value = readStoredPermissions()

    try {
      await fetchMe()
    } catch {
      clearSession()
    }
  }

  return {
    user,
    token,
    permissions,
    loading,
    isAuthenticated,
    login,
    logout,
    fetchMe,
    can,
    restoreSession,
    clearSession,
  }
})
