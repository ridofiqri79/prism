import type { RouteRecordRaw } from 'vue-router'

export const loanAgreementRoutes: RouteRecordRaw[] = [
  {
    path: 'loan-agreements',
    name: 'loan-agreements',
    component: () => import('@/pages/loan-agreement/LAListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Loan Agreement',
      permission: { module: 'loan_agreement', action: 'read' },
    },
  },
  {
    path: 'loan-agreements/new',
    name: 'loan-agreement-create',
    component: () => import('@/pages/loan-agreement/LAFormPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Buat Loan Agreement',
      permission: { module: 'loan_agreement', action: 'create' },
    },
  },
  {
    path: 'loan-agreements/:id',
    name: 'loan-agreement-detail',
    component: () => import('@/pages/loan-agreement/LADetailPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Detail Loan Agreement',
      permission: { module: 'loan_agreement', action: 'read' },
    },
  },
  {
    path: 'loan-agreements/:id/edit',
    name: 'loan-agreement-edit',
    component: () => import('@/pages/loan-agreement/LAFormPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Edit Loan Agreement',
      permission: { module: 'loan_agreement', action: 'update' },
    },
  },
]
