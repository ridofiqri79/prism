import type { RouteRecordRaw } from 'vue-router'

export const dashboardRoutes: RouteRecordRaw[] = [
  {
    path: 'dashboard',
    name: 'dashboard',
    component: () => import('@/pages/dashboard/DashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Dashboard',
      permission: { module: 'bb_project', action: 'read' },
    },
  },
]
