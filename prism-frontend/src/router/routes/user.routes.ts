import type { RouteRecordRaw } from 'vue-router'

export const userRoutes: RouteRecordRaw[] = [
  {
    path: 'users',
    name: 'users',
    component: () => import('@/pages/user/UserListPage.vue'),
    meta: {
      requiresAuth: true,
      adminOnly: true,
    },
  },
  {
    path: 'users/new',
    name: 'user-create',
    component: () => import('@/pages/user/UserFormPage.vue'),
    meta: {
      requiresAuth: true,
      adminOnly: true,
    },
  },
  {
    path: 'users/:id/edit',
    name: 'user-edit',
    component: () => import('@/pages/user/UserFormPage.vue'),
    meta: {
      requiresAuth: true,
      adminOnly: true,
    },
  },
  {
    path: 'users/:id/permissions',
    name: 'user-permissions',
    component: () => import('@/pages/user/UserPermissionPage.vue'),
    meta: {
      requiresAuth: true,
      adminOnly: true,
    },
  },
]
