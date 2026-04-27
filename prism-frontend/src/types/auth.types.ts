export type UserRole = 'ADMIN' | 'STAFF'
export type PermissionAction = 'create' | 'read' | 'update' | 'delete'

export interface AuthUser {
  id: string
  username: string
  email: string
  role: UserRole
}

export interface UserPermission {
  module: string
  can_create: boolean
  can_read: boolean
  can_update: boolean
  can_delete: boolean
}

export interface LoginResponse {
  access_token: string
  expires_in: number
  user: AuthUser
}

export interface AuthSession {
  user: AuthUser
  permissions: UserPermission[]
}
