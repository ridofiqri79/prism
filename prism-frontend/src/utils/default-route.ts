import type { RouteLocationRaw } from 'vue-router'

import type { AuthUser, PermissionAction, UserPermission } from '@/types/auth.types'
import { readStoredPermissions, readStoredUser } from '@/utils/auth-session'

interface RouteCandidate {
  route: RouteLocationRaw
  module?: string
  action?: PermissionAction
  adminOnly?: boolean
}

const authenticatedRouteCandidates: RouteCandidate[] = [
  { route: { name: 'project-master' }, module: 'bb_project', action: 'read' },
  { route: { name: 'blue-books' }, module: 'blue_book', action: 'read' },
  { route: { name: 'green-books' }, module: 'green_book', action: 'read' },
  { route: { name: 'daftar-kegiatan' }, module: 'daftar_kegiatan', action: 'read' },
  { route: { name: 'loan-agreements' }, module: 'loan_agreement', action: 'read' },
  { route: { name: 'spatial-distribution' }, module: 'bb_project', action: 'read' },
  { route: { name: 'project-journey-search' }, module: 'bb_project', action: 'read' },
  { route: { name: 'master-countries' }, module: 'country', action: 'read' },
  { route: { name: 'master-lenders' }, module: 'lender', action: 'read' },
  { route: { name: 'master-currencies' }, module: 'currency', action: 'read' },
  { route: { name: 'master-institutions' }, module: 'institution', action: 'read' },
  { route: { name: 'master-regions' }, module: 'region', action: 'read' },
  { route: { name: 'master-program-titles' }, module: 'program_title', action: 'read' },
  { route: { name: 'master-bappenas-partners' }, module: 'bappenas_partner', action: 'read' },
  { route: { name: 'master-periods' }, module: 'period', action: 'read' },
  { route: { name: 'master-national-priorities' }, module: 'national_priority', action: 'read' },
  { route: { name: 'users' }, adminOnly: true },
]

function canAccessRoute(user: AuthUser | null, permissions: UserPermission[], candidate: RouteCandidate) {
  if (!user) {
    return false
  }

  if (user.role === 'ADMIN') {
    return true
  }

  if (candidate.adminOnly) {
    return false
  }

  if (!candidate.module || !candidate.action) {
    return true
  }

  const permission = permissions.find((item) => item.module === candidate.module)
  if (!permission) {
    return false
  }

  return permission[`can_${candidate.action}`]
}

export function resolveDefaultAuthenticatedRoute(input?: {
  user?: AuthUser | null
  permissions?: UserPermission[] | null
}): RouteLocationRaw {
  const user = input?.user ?? (typeof window !== 'undefined' ? readStoredUser() : null)
  const permissions =
    input?.permissions ?? (typeof window !== 'undefined' ? readStoredPermissions() : [])

  if (!user) {
    return { name: 'login' }
  }

  for (const candidate of authenticatedRouteCandidates) {
    if (canAccessRoute(user, permissions, candidate)) {
      return candidate.route
    }
  }

  return { name: 'forbidden' }
}
