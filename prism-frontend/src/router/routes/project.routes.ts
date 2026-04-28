import type { RouteRecordRaw } from 'vue-router'

export const projectRoutes: RouteRecordRaw[] = [
  {
    path: 'projects',
    name: 'project-master',
    component: () => import('@/pages/project/ProjectMasterPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Project',
      permission: { module: 'bb_project', action: 'read' },
    },
  },
]
