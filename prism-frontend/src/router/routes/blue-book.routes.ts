import type { RouteRecordRaw } from 'vue-router'

export const blueBookRoutes: RouteRecordRaw[] = [
  {
    path: 'blue-books',
    name: 'blue-books',
    component: () => import('@/pages/blue-book/BlueBookListPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Blue Book',
      permission: { module: 'blue_book', action: 'read' },
    },
  },
  {
    path: 'blue-books/:id',
    name: 'blue-book-detail',
    component: () => import('@/pages/blue-book/BlueBookDetailPage.vue'),
    meta: {
      requiresAuth: true,
      title: 'Detail Blue Book',
      permission: { module: 'blue_book', action: 'read' },
    },
  },
  {
    path: 'blue-books/:bbId/projects/new',
    name: 'bb-project-create',
    component: () => import('@/pages/blue-book/BBProjectFormPage.vue'),
    meta: {
      requiresAuth: true,
        title: 'Tambah Proyek Blue Book',
      permission: { module: 'bb_project', action: 'create' },
    },
  },
  {
    path: 'blue-books/:bbId/projects/:id',
    name: 'bb-project-detail',
    component: () => import('@/pages/blue-book/BBProjectDetailPage.vue'),
    meta: {
      requiresAuth: true,
        title: 'Detail Proyek Blue Book',
      permission: { module: 'bb_project', action: 'read' },
    },
  },
  {
    path: 'blue-books/:bbId/projects/:id/edit',
    name: 'bb-project-edit',
    component: () => import('@/pages/blue-book/BBProjectFormPage.vue'),
    meta: {
      requiresAuth: true,
        title: 'Edit Proyek Blue Book',
      permission: { module: 'bb_project', action: 'update' },
    },
  },
]
