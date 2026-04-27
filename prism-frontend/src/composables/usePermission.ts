import { useAuthStore } from '@/stores/auth.store'
import type { PermissionAction, UserPermission } from '@/types/auth.types'

export function usePermission() {
  const auth = useAuthStore()

  const can = (module: string, action: PermissionAction) => {
    if (auth.user?.role === 'ADMIN') {
      return true
    }

    const permission = auth.permissions.find((item) => item.module === module)

    return permission ? (permission[`can_${action}` as keyof UserPermission] as boolean) : false
  }

  return { can }
}
