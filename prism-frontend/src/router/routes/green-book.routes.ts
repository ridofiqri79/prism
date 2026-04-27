import type { RouteRecordRaw } from 'vue-router'

export const greenBookRoutes: RouteRecordRaw[] = [
  {
    path: 'green-books',
    name: 'green-books',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'green-books/:id',
    name: 'green-book-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'green-books/:gbId/projects/new',
    name: 'gb-project-create',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'green-books/:gbId/projects/:id',
    name: 'gb-project-detail',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: 'green-books/:gbId/projects/:id/edit',
    name: 'gb-project-edit',
    component: () => import('@/pages/common/RoutePlaceholderPage.vue'),
    meta: { requiresAuth: true },
  },
]
