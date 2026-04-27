import type { AuthUser, UserPermission } from '@/types/auth.types'

export const AUTH_STORAGE_KEYS = {
  token: 'prism.access_token',
  user: 'prism.user',
  permissions: 'prism.permissions',
} as const

function parseStoredJson<T>(value: string | null): T | null {
  if (!value) {
    return null
  }

  try {
    return JSON.parse(value) as T
  } catch {
    return null
  }
}

export function readStoredToken(): string | null {
  return window.localStorage.getItem(AUTH_STORAGE_KEYS.token)
}

export function readStoredUser(): AuthUser | null {
  return parseStoredJson<AuthUser>(window.localStorage.getItem(AUTH_STORAGE_KEYS.user))
}

export function readStoredPermissions(): UserPermission[] {
  return parseStoredJson<UserPermission[]>(window.localStorage.getItem(AUTH_STORAGE_KEYS.permissions)) ?? []
}

export function storeSession(payload: {
  token: string
  user: AuthUser | null
  permissions: UserPermission[]
}) {
  window.localStorage.setItem(AUTH_STORAGE_KEYS.token, payload.token)

  if (payload.user) {
    window.localStorage.setItem(AUTH_STORAGE_KEYS.user, JSON.stringify(payload.user))
  } else {
    window.localStorage.removeItem(AUTH_STORAGE_KEYS.user)
  }

  window.localStorage.setItem(AUTH_STORAGE_KEYS.permissions, JSON.stringify(payload.permissions))
}

export function clearStoredSession() {
  window.localStorage.removeItem(AUTH_STORAGE_KEYS.token)
  window.localStorage.removeItem(AUTH_STORAGE_KEYS.user)
  window.localStorage.removeItem(AUTH_STORAGE_KEYS.permissions)
}
