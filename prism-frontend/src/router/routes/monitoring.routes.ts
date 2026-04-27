import type { RouteRecordRaw } from 'vue-router'

export const monitoringRoutes: RouteRecordRaw[] = [
  {
    path: 'loan-agreements/:laId/monitoring',
    name: 'monitoring-list',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'loan-agreements/:laId/monitoring/new',
    name: 'monitoring-create',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
]
