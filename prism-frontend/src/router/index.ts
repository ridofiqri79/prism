import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { authRoutes } from '@/router/routes/auth.routes'
import { blueBookRoutes } from '@/router/routes/blue-book.routes'
import { daftarKegiatanRoutes } from '@/router/routes/daftar-kegiatan.routes'
import { greenBookRoutes } from '@/router/routes/green-book.routes'
import { homeRoutes } from '@/router/routes/home.routes'
import { journeyRoutes } from '@/router/routes/journey.routes'
import { loanAgreementRoutes } from '@/router/routes/loan-agreement.routes'
import { masterRoutes } from '@/router/routes/master.routes'
import { projectRoutes } from '@/router/routes/project.routes'
import { spatialDistributionRoutes } from '@/router/routes/spatial-distribution.routes'
import { userRoutes } from '@/router/routes/user.routes'
import { resolveDefaultAuthenticatedRoute } from '@/utils/default-route'
import { resolveRouteTitle } from '@/utils/route-title'

const appRoutes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/AppLayout.vue'),
    children: [
      ...homeRoutes,
      ...spatialDistributionRoutes,
      ...projectRoutes,
      ...masterRoutes,
      ...blueBookRoutes,
      ...greenBookRoutes,
      ...daftarKegiatanRoutes,
      ...loanAgreementRoutes,
      ...journeyRoutes,
      {
        path: 'forbidden',
        name: 'forbidden',
        component: () => import('@/pages/common/ForbiddenPage.vue'),
        meta: {
          requiresAuth: true,
          title: 'Akses Ditolak',
        },
      },
      ...userRoutes,
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes: [
    ...authRoutes,
    ...appRoutes,
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/pages/common/NotFoundPage.vue'),
      meta: {
        title: 'Halaman Tidak Ditemukan',
      },
    },
  ],
})

let sessionRestorePromise: Promise<void> | null = null
let hasAttemptedRestore = false

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  if (!hasAttemptedRestore) {
    sessionRestorePromise ??= auth.restoreSession().finally(() => {
      hasAttemptedRestore = true
      sessionRestorePromise = null
    })

    await sessionRestorePromise
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return {
      name: 'login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  if (to.meta.adminOnly && auth.user?.role !== 'ADMIN') {
    return { name: 'forbidden' }
  }

  if (to.meta.permission) {
    const permission = to.meta.permission

    if (!auth.can(permission.module, permission.action)) {
      return { name: 'forbidden' }
    }
  }

  if (to.name === 'login' && auth.isAuthenticated) {
    return resolveDefaultAuthenticatedRoute({
      user: auth.user,
      permissions: auth.permissions,
    })
  }

  return true
})

router.afterEach((to) => {
  document.title = resolveRouteTitle(to)
})

export default router
