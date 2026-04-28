import type { RouteRecordRaw } from 'vue-router'

export const monitoringRoutes: RouteRecordRaw[] = [
  {
    path: 'loan-agreements/:laId/monitoring',
    name: 'monitoring-list',
    component: () => import('@/pages/monitoring/MonitoringListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Monitoring Disbursement',
      permission: { module: 'monitoring_disbursement', action: 'read' },
    },
  },
  {
    path: 'loan-agreements/:laId/monitoring/new',
    name: 'monitoring-create',
    component: () => import('@/pages/monitoring/MonitoringFormPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Tambah Monitoring Disbursement',
      permission: { module: 'monitoring_disbursement', action: 'create' },
    },
  },
  {
    path: 'loan-agreements/:laId/monitoring/:id/edit',
    name: 'monitoring-edit',
    component: () => import('@/pages/monitoring/MonitoringFormPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Edit Monitoring Disbursement',
      permission: { module: 'monitoring_disbursement', action: 'update' },
    },
  },
]
