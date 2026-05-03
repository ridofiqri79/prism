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
  {
    path: 'dashboard/green-book-readiness',
    name: 'dashboard-green-book-readiness',
    component: () => import('@/pages/dashboard/GreenBookReadinessDashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Green Book Readiness',
    },
  },
  {
    path: 'dashboard/lender-financing-mix',
    name: 'dashboard-lender-financing-mix',
    component: () => import('@/pages/dashboard/LenderFinancingMixDashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Lender & Financing Mix',
    },
  },
  {
    path: 'dashboard/kl-portfolio-performance',
    name: 'dashboard-kl-portfolio-performance',
    component: () => import('@/pages/dashboard/KLPortfolioPerformanceDashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'K/L Portfolio Performance',
    },
  },
  {
    path: 'dashboard/la-disbursement',
    name: 'dashboard-la-disbursement',
    component: () => import('@/pages/dashboard/LADisbursementDashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Loan Agreement & Disbursement',
    },
  },
  {
    path: 'dashboard/data-quality-governance',
    name: 'dashboard-data-quality-governance',
    component: () => import('@/pages/dashboard/DataQualityGovernanceDashboardPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Data Quality & Governance',
    },
  },
]
