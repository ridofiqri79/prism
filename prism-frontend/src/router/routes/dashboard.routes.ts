import type { RouteRecordRaw } from 'vue-router'

export const dashboardRoutes: RouteRecordRaw[] = [
  {
    path: '',
    redirect: { name: 'dashboard' },
  },
  {
    path: 'dashboard',
    name: 'dashboard',
    component: () => import('@/pages/dashboard/DashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Dashboard',
    },
  },
  {
    path: 'dashboard/executive-portfolio',
    name: 'dashboard-executive-portfolio',
    component: () => import('@/pages/dashboard/ExecutivePortfolioDashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Executive Portfolio',
    },
  },
  {
    path: 'dashboard/pipeline-bottleneck',
    name: 'dashboard-pipeline-bottleneck',
    component: () => import('@/pages/dashboard/PipelineBottleneckDashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Pipeline & Bottleneck',
    },
  },
]
