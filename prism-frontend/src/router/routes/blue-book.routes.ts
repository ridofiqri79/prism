import type { RouteRecordRaw } from 'vue-router'

export const blueBookRoutes: RouteRecordRaw[] = [
  {
    path: 'blue-books',
    name: 'blue-books',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'blue-books/:id',
    name: 'blue-book-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'blue-books/:bbId/projects/new',
    name: 'bb-project-create',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'blue-books/:bbId/projects/:id',
    name: 'bb-project-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'blue-books/:bbId/projects/:id/edit',
    name: 'bb-project-edit',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
]
