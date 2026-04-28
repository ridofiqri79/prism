import type { RouteTitle } from '@/utils/route-title'

import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    adminOnly?: boolean
    permission?: {
      module: string
      action: 'create' | 'read' | 'update' | 'delete'
    }
    requiresAuth?: boolean
    title?: RouteTitle
  }
}
