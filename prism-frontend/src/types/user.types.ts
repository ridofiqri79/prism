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
  { module: 'institution', label: 'Instansi' },
  { module: 'lender', label: 'Lender' },
  { module: 'region', label: 'Wilayah' },
  { module: 'national_priority', label: 'Prioritas Nasional' },
  { module: 'program_title', label: 'Judul Program' },
  { module: 'bappenas_partner', label: 'Mitra Bappenas' },
  { module: 'period', label: 'Periode' },
  { module: 'country', label: 'Negara' },
]
