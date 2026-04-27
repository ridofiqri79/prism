import type { RouteRecordRaw } from 'vue-router'

export const masterRoutes: RouteRecordRaw[] = [
  {
    path: 'master/countries',
    name: 'master-countries',
    component: () => import('@/pages/master/CountryPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'country', action: 'read' } },
  },
  {
    path: 'master/lenders',
    name: 'master-lenders',
    component: () => import('@/pages/master/LenderPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'lender', action: 'read' } },
  },
  {
    path: 'master/institutions',
    name: 'master-institutions',
    component: () => import('@/pages/master/InstitutionPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'institution', action: 'read' } },
  },
  {
    path: 'master/regions',
    name: 'master-regions',
    component: () => import('@/pages/master/RegionPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'region', action: 'read' } },
  },
  {
    path: 'master/program-titles',
    name: 'master-program-titles',
    component: () => import('@/pages/master/ProgramTitlePage.vue'),
    meta: { requiresAuth: true, permission: { module: 'program_title', action: 'read' } },
  },
  {
    path: 'master/bappenas-partners',
    name: 'master-bappenas-partners',
    component: () => import('@/pages/master/BappenasPartnerPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'bappenas_partner', action: 'read' } },
  },
  {
    path: 'master/periods',
    name: 'master-periods',
    component: () => import('@/pages/master/PeriodPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'period', action: 'read' } },
  },
  {
    path: 'master/national-priorities',
    name: 'master-national-priorities',
    component: () => import('@/pages/master/NationalPriorityPage.vue'),
    meta: { requiresAuth: true, permission: { module: 'national_priority', action: 'read' } },
  },
]
