import type { RouteRecordRaw } from 'vue-router'

export const loanAgreementRoutes: RouteRecordRaw[] = [
  {
    path: 'loan-agreements',
    name: 'loan-agreements',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'loan-agreements/new',
    name: 'loan-agreement-create',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'loan-agreements/:id',
    name: 'loan-agreement-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'loan-agreements/:id/edit',
    name: 'loan-agreement-edit',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
]
