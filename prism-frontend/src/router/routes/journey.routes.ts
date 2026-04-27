import type { RouteRecordRaw } from 'vue-router'

export const journeyRoutes: RouteRecordRaw[] = [
  {
    path: 'journey/:bbProjectId',
    name: 'project-journey',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'bb_project', action: 'read' } },
  },
]
