import type { RouteRecordRaw } from 'vue-router'

export const dashboardRoutes: RouteRecordRaw[] = [
  {
    path: '',
    redirect: { name: 'dashboard' },
  },
  {
    path: 'dashboard',
    name: 'dashboard',
    component: () => import('@/pages/dashboard/DashboardHomePage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Dasbor',
    },
  },
  {
    path: 'dashboard/executive-portfolio',
    redirect: { name: 'dashboard', query: { tab: 'executive' } },
  },
  {
    path: 'dashboard/pipeline-bottleneck',
    redirect: { name: 'dashboard', query: { tab: 'pipeline' } },
  },
  {
    path: 'dashboard/green-book-readiness',
    redirect: { name: 'dashboard', query: { tab: 'readiness' } },
  },
  {
    path: 'dashboard/lender-financing-mix',
    redirect: { name: 'dashboard', query: { tab: 'financing' } },
  },
  {
    path: 'dashboard/kl-portfolio-performance',
    redirect: { name: 'dashboard', query: { tab: 'institution' } },
  },
  {
    path: 'dashboard/data-quality-governance',
    redirect: { name: 'dashboard', query: { tab: 'quality' } },
  },
]
