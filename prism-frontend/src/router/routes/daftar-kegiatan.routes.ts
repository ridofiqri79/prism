import type { RouteRecordRaw } from 'vue-router'

export const daftarKegiatanRoutes: RouteRecordRaw[] = [
  {
    path: 'daftar-kegiatan',
    name: 'daftar-kegiatan',
    component: () => import('@/pages/daftar-kegiatan/DKListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Daftar Kegiatan',
      permission: { module: 'daftar_kegiatan', action: 'read' },
    },
  },
  {
    path: 'daftar-kegiatan/:id',
    name: 'daftar-kegiatan-detail',
    component: () => import('@/pages/daftar-kegiatan/DKDetailPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Detail Daftar Kegiatan',
      permission: { module: 'daftar_kegiatan', action: 'read' },
    },
  },
  {
    path: 'daftar-kegiatan/:dkId/projects/new',
    name: 'dk-project-create',
    component: () => import('@/pages/daftar-kegiatan/DKProjectFormPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Tambah DK Project',
      permission: { module: 'daftar_kegiatan', action: 'create' },
    },
  },
  {
    path: 'daftar-kegiatan/:dkId/projects/:id',
    name: 'dk-project-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Detail DK Project',
      permission: { module: 'daftar_kegiatan', action: 'read' },
    },
  },
  {
    path: 'daftar-kegiatan/:dkId/projects/:id/edit',
    name: 'dk-project-edit',
    component: () => import('@/pages/daftar-kegiatan/DKProjectFormPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Edit DK Project',
      permission: { module: 'daftar_kegiatan', action: 'update' },
    },
  },
]
