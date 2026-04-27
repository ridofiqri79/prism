import type { UserPermission, UserRole } from '@/types/auth.types'

export interface AppUser {
  id: string
  username: string
  email: string
  role: UserRole
  is_active: boolean
}

export interface CreateUserRequest {
  username: string
  email: string
  password: string
  role: UserRole
}

export interface UpdateUserRequest {
  username: string
  email: string
  role: UserRole
  is_active: boolean
}

export interface UpdatePermissionsRequest {
  permissions: UserPermission[]
}

export interface PermissionModule {
  module: string
  label: string
}

export const permissionModules: PermissionModule[] = [
  { module: 'bb_project', label: 'BB Project' },
  { module: 'gb_project', label: 'GB Project' },
  { module: 'daftar_kegiatan', label: 'Daftar Kegiatan' },
  { module: 'loan_agreement', label: 'Loan Agreement' },
  { module: 'monitoring_disbursement', label: 'Monitoring Disbursement' },
  { module: 'institution', label: 'Institution' },
  { module: 'lender', label: 'Lender' },
  { module: 'region', label: 'Region' },
  { module: 'national_priority', label: 'National Priority' },
  { module: 'program_title', label: 'Program Title' },
  { module: 'bappenas_partner', label: 'Bappenas Partner' },
  { module: 'period', label: 'Period' },
  { module: 'country', label: 'Country' },
]
