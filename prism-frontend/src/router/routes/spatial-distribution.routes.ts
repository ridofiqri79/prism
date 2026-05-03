import type { RouteRecordRaw } from 'vue-router'

export const spatialDistributionRoutes: RouteRecordRaw[] = [
  {
    path: 'spatial-distribution',
    name: 'spatial-distribution',
    component: () => import('@/pages/spatial/SpatialDistributionPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Sebaran Wilayah',
      permission: { module: 'bb_project', action: 'read' },
    },
  },
]
