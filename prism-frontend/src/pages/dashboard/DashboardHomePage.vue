<script setup lang="ts">
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import EmptyInsightState from '@/components/dashboard/EmptyInsightState.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import type { DashboardNavigationItem } from '@/types/dashboard.types'

const dashboardItems: DashboardNavigationItem[] = [
  {
    key: 'executive',
    title: 'Executive Portfolio',
    description: 'Pipeline, legal commitment, disbursement, funnel, top K/L, top lenders, dan high-risk items.',
    route_name: 'dashboard-executive-portfolio',
    icon: 'pi pi-briefcase',
    accent: 'portfolio',
  },
  {
    key: 'pipeline',
    title: 'Pipeline & Bottleneck',
    description: 'Worklist proyek yang tertahan dari Blue Book sampai monitoring dengan server-side pagination.',
    route_name: 'dashboard-pipeline-bottleneck',
    icon: 'pi pi-sort-amount-down',
    accent: 'pipeline',
  },
  {
    key: 'readiness',
    title: 'Green Book Readiness',
    description: 'Kelengkapan funding source, activities, disbursement plan, allocation, dan cofinancing.',
    route_name: 'dashboard-green-book-readiness',
    icon: 'pi pi-check-square',
    accent: 'readiness',
  },
  {
    key: 'financing',
    title: 'Lender & Financing Mix',
    description: 'Profil lender, certainty ladder, conversion, cofinancing, dan currency exposure.',
    route_name: 'dashboard-lender-financing-mix',
    icon: 'pi pi-building-columns',
    accent: 'financing',
  },
  {
    key: 'institution',
    title: 'K/L Portfolio Performance',
    description: 'Perbandingan portfolio dan kinerja K/L/Badan lintas pipeline, LA, dan monitoring.',
    route_name: 'dashboard-kl-portfolio-performance',
    icon: 'pi pi-sitemap',
    accent: 'institution',
  },
  {
    key: 'disbursement',
    title: 'Loan Agreement & Disbursement',
    description: 'Closing risk, undisbursed balance, planned vs realized, dan under-disbursement warnings.',
    route_name: 'dashboard-la-disbursement',
    icon: 'pi pi-chart-line',
    accent: 'disbursement',
  },
  {
    key: 'quality',
    title: 'Data Quality & Governance',
    description: 'Issue detection, relationship integrity, monitoring compliance, dan audit summary ADMIN.',
    route_name: 'dashboard-data-quality-governance',
    icon: 'pi pi-shield',
    accent: 'quality',
  },
]

function accentClass(accent: DashboardNavigationItem['accent']) {
  return {
    portfolio: 'bg-primary text-primary-contrast',
    pipeline: 'bg-orange-500 text-white',
    readiness: 'bg-emerald-600 text-white',
    financing: 'bg-cyan-700 text-white',
    institution: 'bg-indigo-600 text-white',
    disbursement: 'bg-sky-700 text-white',
    quality: 'bg-rose-700 text-white',
  }[accent]
}
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Dashboard"
      subtitle="Pilih dashboard operasional sesuai kebutuhan kontrol portfolio, pipeline, pembiayaan, monitoring, dan governance."
    />

    <section v-if="dashboardItems.length" class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <RouterLink
        v-for="item in dashboardItems"
        :key="item.key"
        :to="{ name: item.route_name }"
        class="group rounded-lg border border-surface-200 bg-white p-5 transition hover:-translate-y-0.5 hover:border-primary hover:shadow-sm"
      >
        <div class="flex items-start justify-between gap-4">
          <div :class="[accentClass(item.accent), 'flex h-11 w-11 shrink-0 items-center justify-center rounded-lg']">
            <i :class="[item.icon, 'text-lg']" />
          </div>
          <Tag value="Dashboard" severity="secondary" />
        </div>

        <h2 class="mt-5 text-lg font-semibold text-surface-950">{{ item.title }}</h2>
        <p class="mt-2 min-h-16 text-sm leading-6 text-surface-600">{{ item.description }}</p>
        <div class="mt-5 flex items-center justify-between gap-3">
          <span class="text-sm font-medium text-primary">Buka dashboard</span>
          <Button icon="pi pi-arrow-right" text rounded aria-label="Buka dashboard" />
        </div>
      </RouterLink>
    </section>

    <EmptyInsightState
      v-else
      title="Belum ada dashboard"
      message="Route dashboard belum tersedia untuk sesi ini."
      icon="pi pi-chart-bar"
    />
  </section>
</template>
