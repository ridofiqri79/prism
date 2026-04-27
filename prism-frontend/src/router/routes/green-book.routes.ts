import type { RouteRecordRaw } from 'vue-router'

export const greenBookRoutes: RouteRecordRaw[] = [
  {
    path: 'green-books',
    name: 'green-books',
    component: () => import('@/pages/green-book/GreenBookListPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'green_book', action: 'read' } },
  },
  {
    path: 'green-books/:id',
    name: 'green-book-detail',
    component: () => import('@/pages/green-book/GreenBookDetailPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'green_book', action: 'read' } },
  },
  {
    path: 'green-books/:gbId/projects/new',
    name: 'gb-project-create',
    component: () => import('@/pages/green-book/GBProjectFormPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'gb_project', action: 'create' } },
  },
  {
    path: 'green-books/:gbId/projects/:id',
    name: 'gb-project-detail',
    component: () => import('@/pages/green-book/GBProjectDetailPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'gb_project', action: 'read' } },
  },
  {
    path: 'green-books/:gbId/projects/:id/edit',
    name: 'gb-project-edit',
    component: () => import('@/pages/green-book/GBProjectFormPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'gb_project', action: 'update' } },
  },
]
