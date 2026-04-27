import type { RouteRecordRaw } from 'vue-router'

export const journeyRoutes: RouteRecordRaw[] = [
  {
    path: 'journey',
    name: 'project-journey-search',
    component: () => import('@/pages/journey/ProjectJourneyPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'bb_project', action: 'read' } },
  },
  {
    path: 'journey/:bbProjectId',
    name: 'project-journey',
    component: () => import('@/pages/journey/ProjectJourneyPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'bb_project', action: 'read' } },
  },
]
