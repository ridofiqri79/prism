import type { RouteRecordRaw } from 'vue-router'

export const daftarKegiatanRoutes: RouteRecordRaw[] = [
  {
    path: 'daftar-kegiatan',
    name: 'daftar-kegiatan',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'daftar-kegiatan/:id',
    name: 'daftar-kegiatan-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'daftar-kegiatan/:dkId/projects/new',
    name: 'dk-project-create',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'daftar-kegiatan/:dkId/projects/:id',
    name: 'dk-project-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'daftar-kegiatan/:dkId/projects/:id/edit',
    name: 'dk-project-edit',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
]
