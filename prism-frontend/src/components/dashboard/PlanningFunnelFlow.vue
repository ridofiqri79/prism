<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, type RouteLocationRaw } from 'vue-router'
import Tag from 'primevue/tag'
import type {
  DashboardDatum,
  DashboardInsightTarget,
  DashboardMetricTone,
  DashboardPanelSpan,
  DashboardStage,
  DashboardStageKey,
} from '@/types/dashboard-flow.types'

const props = defineProps<{
  stages: DashboardStage[]
}>()

const expandedStages = ref<DashboardStageKey[]>(['BB'])
const activeTabs = ref<Record<DashboardStageKey, string>>({
  BB: 'Status Pipeline',
  GB: 'Lender & Funding',
  DK: 'Lender & Funding',
  LA: 'Kondisi Pinjaman',
})

const countFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 0,
})

const percentFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 1,
  minimumFractionDigits: 1,
})

const funnelTitleId = 'planning-funnel-title'
const funnelSummaryId = 'planning-funnel-summary'

const chartAriaLabel = computed(() =>
  props.stages
    .map((stage) => `${stage.title}: ${countFormatter.format(stage.count)} proyek`)
    .join(', '),
)

function toggleStage(stageKey: DashboardStageKey) {
  if (expandedStages.value.includes(stageKey)) {
    expandedStages.value = expandedStages.value.filter((key) => key !== stageKey)
    return
  }

  expandedStages.value = [...expandedStages.value, stageKey]
}

function isExpanded(stageKey: DashboardStageKey) {
  return expandedStages.value.includes(stageKey)
}

function setActiveTab(stageKey: DashboardStageKey, label: string) {
  activeTabs.value = {
    ...activeTabs.value,
    [stageKey]: label,
  }
}

function slugifyLabel(label: string) {
  const slug = label
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')

  return slug || 'tab'
}

function tabId(stage: DashboardStage, label: string) {
  return `planning-funnel-${stage.key.toLowerCase()}-${slugifyLabel(label)}-tab`
}

function tabPanelId(stage: DashboardStage, label: string) {
  return `planning-funnel-${stage.key.toLowerCase()}-${slugifyLabel(label)}-panel`
}

function moveActiveTab(stage: DashboardStage, direction: number) {
  const labels = stage.details.tabs.map((tab) => tab.label)
  if (!labels.length) return

  const currentIndex = Math.max(labels.indexOf(activeTabLabel(stage)), 0)
  const nextIndex = (currentIndex + direction + labels.length) % labels.length
  setActiveTab(stage.key, labels[nextIndex] ?? labels[0] ?? '')
}

function handleTabKeydown(event: KeyboardEvent, stage: DashboardStage) {
  if (event.key === 'ArrowRight') {
    event.preventDefault()
    moveActiveTab(stage, 1)
    return
  }

  if (event.key === 'ArrowLeft') {
    event.preventDefault()
    moveActiveTab(stage, -1)
    return
  }

  if (event.key === 'Home') {
    event.preventDefault()
    setActiveTab(stage.key, stage.details.tabs[0]?.label ?? '')
    return
  }

  if (event.key === 'End') {
    event.preventDefault()
    setActiveTab(stage.key, stage.details.tabs.at(-1)?.label ?? '')
  }
}

function activeTabLabel(stage: DashboardStage) {
  const current = activeTabs.value[stage.key]
  if (stage.details.tabs.some((tab) => tab.label === current)) return current
  return stage.details.tabs[0]?.label ?? ''
}

function isActiveTab(stage: DashboardStage, label: string) {
  return activeTabLabel(stage) === label
}

function visibleCounters(stage: DashboardStage) {
  const activeLabel = activeTabLabel(stage)
  return stage.details.counters.filter((counter) => counter.tab === activeLabel)
}

function visiblePanels(stage: DashboardStage) {
  const activeLabel = activeTabLabel(stage)
  return stage.details.panels.filter((panel) => panel.tab === activeLabel)
}

function stageBarWidth(stage: DashboardStage) {
  if (stage.count === 0) return '12rem'
  return `${Math.min(Math.max(stage.pipelineShare, 16), 100)}%`
}

function maxValue(data: DashboardDatum[] = []) {
  return Math.max(...data.map((item) => item.value), 1)
}

function sumValue(data: DashboardDatum[] = []) {
  return data.reduce((sum, item) => sum + item.value, 0)
}

function itemPercent(data: DashboardDatum[] = [], item: DashboardDatum) {
  const total = sumValue(data)
  if (total === 0) return 0
  return (item.value / total) * 100
}

function segmentPercent(data: DashboardDatum[] = [], item: DashboardDatum) {
  const total = sumValue(data)
  if (total === 0) return data.length > 0 ? 100 / data.length : 0
  return (item.value / total) * 100
}

function barWidth(data: DashboardDatum[] = [], item: DashboardDatum) {
  return `${Math.max((item.value / maxValue(data)) * 100, item.value > 0 ? 5 : 0)}%`
}

function regionClass(value: number) {
  if (value >= 10) return 'region-xhi'
  if (value >= 7) return 'region-hi'
  if (value >= 4) return 'region-md'
  return 'region-lo'
}

function donutBackground(data: DashboardDatum[] = []) {
  const total = sumValue(data)
  if (total === 0) return 'conic-gradient(#e5e7eb 0deg 360deg)'

  let cursor = 0
  const segments = data.map((item) => {
    const start = cursor
    const degrees = (item.value / total) * 360
    cursor += degrees
    return `${item.color ?? '#64748b'} ${start}deg ${cursor}deg`
  })

  return `conic-gradient(${segments.join(', ')})`
}

function metricToneClass(tone: DashboardMetricTone = 'neutral') {
  const toneClasses: Record<DashboardMetricTone, string> = {
    neutral: 'text-surface-950',
    good: 'text-prism-green-dark',
    warning: 'text-amber-600',
    danger: 'text-red-600',
  }

  return toneClasses[tone]
}

function counterToneClass(tone: DashboardMetricTone = 'neutral') {
  const toneClasses: Record<DashboardMetricTone, string> = {
    neutral: 'border-surface-200 bg-white',
    good: 'border-prism-green-deep/25 bg-prism-green/5',
    warning: 'border-prism-gold/45 bg-prism-gold/10',
    danger: 'border-red-200 bg-red-50/80',
  }

  return toneClasses[tone]
}

function panelSpanClass(span: DashboardPanelSpan = 'medium') {
  const spanClasses: Record<DashboardPanelSpan, string> = {
    small: 'xl:col-span-4',
    medium: 'xl:col-span-6',
    large: 'xl:col-span-8',
    wide: 'xl:col-span-12',
  }

  return spanClasses[span]
}

function normalizedTargetQuery(target: DashboardInsightTarget) {
  const query: Record<string, string | string[]> = {}

  Object.entries(target.query).forEach(([key, value]) => {
    if (value === null || value === undefined) return

    if (Array.isArray(value)) {
      const values = value
        .filter((item) => item !== null && item !== undefined)
        .map((item) => String(item))

      if (values.length > 0) query[key] = values
      return
    }

    query[key] = String(value)
  })

  return query
}

function targetLocation(target: DashboardInsightTarget): RouteLocationRaw {
  return {
    name: target.name,
    query: normalizedTargetQuery(target),
  }
}

function isClickableTarget(target: DashboardInsightTarget | undefined, value: number) {
  return Boolean(target?.exact && value > 0)
}

function isClickableStage(stage: DashboardStage) {
  return isClickableTarget(stage.target, stage.count)
}

function isClickableDatum(item: DashboardDatum) {
  return isClickableTarget(item.target, item.value)
}

function stageLinkProps(stage: DashboardStage) {
  if (!stage.target || !isClickableStage(stage)) return {}
  return {
    to: targetLocation(stage.target),
    'aria-label': stage.target.label ?? `Buka ${stage.title}`,
  }
}

function datumComponent(item: DashboardDatum) {
  return isClickableDatum(item) ? RouterLink : 'div'
}

function datumLinkProps(item: DashboardDatum, fallbackLabel: string) {
  if (!item.target || !isClickableDatum(item)) return {}
  return {
    to: targetLocation(item.target),
    'aria-label': item.target.label ?? `Buka ${fallbackLabel}`,
  }
}

function clickableDatumClass(item: DashboardDatum) {
  if (!isClickableDatum(item)) return ''
  return 'cursor-pointer no-underline transition hover:border-primary-200 hover:bg-primary-50/60 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary'
}

function clickableStageClass(stage: DashboardStage) {
  if (!isClickableStage(stage)) return ''
  return 'no-underline transition hover:ring-2 hover:ring-primary-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary'
}

function clickableRegionClass(item: DashboardDatum) {
  if (!isClickableDatum(item)) return ''
  return 'cursor-pointer no-underline transition hover:ring-2 hover:ring-primary-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary'
}
</script>

<template>
  <section
    class="overflow-hidden rounded-lg border border-surface-200 bg-white shadow-sm"
    :aria-labelledby="funnelTitleId"
    :aria-describedby="funnelSummaryId"
  >
    <header
      class="flex flex-col gap-4 border-b border-surface-100 px-4 py-4 sm:px-5 xl:flex-row xl:items-center xl:justify-between"
    >
      <div class="min-w-0">
        <p class="text-xs font-semibold uppercase text-primary-700">Ringkasan portofolio</p>
        <h2 :id="funnelTitleId" class="mt-1 text-base font-semibold text-surface-950">
          Alur Proyek Pinjaman Luar Negeri
        </h2>
        <p class="mt-1 max-w-3xl text-sm leading-5 text-surface-500">
          Ringkasan tahapan proyek dari perencanaan hingga perjanjian pinjaman.
        </p>
      </div>

      <div class="flex flex-wrap gap-2">
        <span
          v-for="stage in stages"
          :key="stage.key"
          class="inline-flex items-center gap-2 rounded-full border border-surface-200 bg-white px-3 py-1.5 text-xs font-semibold text-surface-700"
        >
          <span
            class="h-2.5 w-2.5 rounded-sm"
            :style="{ backgroundColor: stage.color }"
            aria-hidden="true"
          />
          {{ stage.title }}
        </span>
      </div>
    </header>
    <p :id="funnelSummaryId" class="sr-only">{{ chartAriaLabel }}</p>

    <div class="hidden grid-cols-[12.5rem_minmax(0,1fr)_14rem] gap-0 px-5 py-3 lg:grid">
      <p class="text-[11px] font-semibold uppercase text-surface-400">Tahap</p>
      <p class="text-[11px] font-semibold uppercase text-surface-400">Portofolio</p>
      <p class="text-right text-[11px] font-semibold uppercase text-surface-400">
        Progres berikutnya
      </p>
    </div>

    <div class="divide-y divide-surface-100">
      <template v-for="(stage, stageIndex) in stages" :key="stage.key">
        <article class="bg-white">
          <div
            class="grid w-full text-left transition hover:bg-surface-50 lg:grid-cols-[12.5rem_minmax(0,1fr)_14rem]"
          >
            <div class="border-surface-100 px-4 py-4 sm:px-5 lg:border-r">
              <button
                type="button"
                class="flex w-full items-start justify-between gap-3 rounded-md text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary"
                :aria-expanded="isExpanded(stage.key)"
                @click="toggleStage(stage.key)"
              >
                <span class="min-w-0">
                  <span class="block text-xs font-semibold uppercase text-surface-400">
                    {{ stage.stepLabel }}
                  </span>
                  <span class="mt-2 flex items-center gap-2 text-sm font-semibold text-surface-950">
                    <span
                      class="h-2.5 w-2.5 shrink-0 rounded-sm"
                      :style="{ backgroundColor: stage.color }"
                      aria-hidden="true"
                    />
                    {{ stage.title }}
                  </span>
                </span>
                <i
                  class="pi pi-angle-right mt-0.5 shrink-0 text-xs text-surface-400 transition-transform"
                  :class="isExpanded(stage.key) ? 'rotate-90 text-primary-700' : ''"
                  aria-hidden="true"
                />
              </button>
            </div>

            <div class="px-4 pb-4 sm:px-5 lg:py-4">
              <p class="text-sm leading-5 text-surface-600">{{ stage.subtitle }}</p>
              <component
                :is="isClickableStage(stage) ? RouterLink : 'div'"
                v-bind="stageLinkProps(stage)"
                class="mt-3 block rounded-md bg-surface-50 p-1"
                :class="clickableStageClass(stage)"
              >
                <div
                  class="flex min-h-16 max-w-full items-center justify-between gap-4 rounded-md px-4 text-white shadow-sm transition-[width] duration-500"
                  :style="{
                    width: stageBarWidth(stage),
                    minWidth: '12rem',
                    background: `linear-gradient(90deg, ${stage.color}, ${stage.colorSoft})`,
                  }"
                >
                  <div class="min-w-0">
                    <p class="truncate text-2xl font-semibold leading-none">
                      {{ countFormatter.format(stage.count) }}
                      <span class="text-sm font-medium opacity-85">proyek</span>
                    </p>
                    <p class="mt-1 truncate text-xs font-semibold opacity-90">
                      {{ stage.amountLabel }}
                    </p>
                  </div>
                  <p class="hidden shrink-0 text-xs font-semibold opacity-85 sm:block">
                    {{ percentFormatter.format(stage.pipelineShare) }}%
                  </p>
                </div>
              </component>
            </div>

            <div
              class="border-surface-100 bg-surface-50/70 px-4 py-4 sm:px-5 lg:border-l lg:bg-white lg:text-right"
            >
              <template v-if="stage.nextLabel">
                <p
                  class="text-2xl font-semibold leading-none"
                  :class="stage.conversionLabel === '0,0%' ? 'text-red-600' : 'text-surface-950'"
                >
                  {{ stage.conversionLabel }}
                </p>
                <p class="mt-1 text-xs font-semibold uppercase text-surface-400">
                  ke {{ stage.nextLabel }}
                </p>
                <p class="mt-2 text-xs font-medium leading-5 text-red-600">
                  {{ stage.blockedLabel }}
                </p>
              </template>
              <template v-else>
                <p class="text-xs font-semibold uppercase text-surface-400">Tahap akhir</p>
                <p class="mt-2 text-sm font-semibold leading-5 text-surface-800">
                  {{ stage.finalLabel }}
                </p>
              </template>
            </div>
          </div>

          <Transition name="dashboard-stage">
            <div v-if="isExpanded(stage.key)" class="border-t border-surface-100 bg-surface-50/80">
              <div class="px-4 py-4 sm:px-5">
                <div class="rounded-lg border border-surface-200 bg-white">
                  <div
                    class="flex overflow-x-auto border-b border-surface-100 bg-surface-50/70 px-3"
                    role="tablist"
                    :aria-label="`Detail ${stage.title}`"
                  >
                    <button
                      v-for="tab in stage.details.tabs"
                      :key="tab.label"
                      type="button"
                      role="tab"
                      :id="tabId(stage, tab.label)"
                      :aria-selected="isActiveTab(stage, tab.label)"
                      :aria-controls="tabPanelId(stage, tab.label)"
                      class="shrink-0 border-b-2 px-3 py-3 text-xs font-semibold text-surface-500 transition"
                      :class="
                        isActiveTab(stage, tab.label)
                          ? 'border-primary text-surface-950'
                          : 'border-transparent hover:text-surface-800'
                      "
                      @click.stop="setActiveTab(stage.key, tab.label)"
                      @keydown="handleTabKeydown($event, stage)"
                    >
                      {{ tab.label }}
                    </button>
                  </div>

                  <div
                    class="p-4 sm:p-5"
                    role="tabpanel"
                    :id="tabPanelId(stage, activeTabLabel(stage))"
                    :aria-labelledby="tabId(stage, activeTabLabel(stage))"
                  >
                    <div v-if="visibleCounters(stage).length" class="grid gap-3 md:grid-cols-4">
                      <section
                        v-for="counter in visibleCounters(stage)"
                        :key="counter.label"
                        class="min-h-28 rounded-lg border px-4 py-4"
                        :class="counterToneClass(counter.tone)"
                      >
                        <p class="text-[11px] font-semibold uppercase leading-4 text-surface-500">
                          {{ counter.label }}
                        </p>
                        <p
                          class="mt-2 text-2xl font-semibold leading-none"
                          :class="metricToneClass(counter.tone)"
                        >
                          {{ counter.value }}
                        </p>
                        <p class="mt-2 text-xs leading-5 text-surface-500">{{ counter.meta }}</p>
                      </section>
                    </div>

                    <div
                      v-if="visiblePanels(stage).length"
                      class="mt-4 grid gap-4 xl:grid-cols-12"
                      :class="visibleCounters(stage).length ? '' : 'mt-0'"
                    >
                      <section
                        v-for="panel in visiblePanels(stage)"
                        :key="panel.title"
                        class="min-w-0 rounded-lg border border-surface-200 bg-white p-4"
                        :class="panelSpanClass(panel.span)"
                      >
                        <div class="flex items-baseline justify-between gap-4">
                          <h3 class="text-[11px] font-semibold uppercase text-surface-500">
                            {{ panel.title }}
                          </h3>
                          <p class="shrink-0 text-xs text-surface-400">{{ panel.hint }}</p>
                        </div>

                        <div v-if="panel.kind === 'bars'" class="mt-4 space-y-3">
                          <component
                            :is="datumComponent(item)"
                            v-for="item in panel.data"
                            :key="item.label"
                            v-bind="datumLinkProps(item, item.label)"
                            class="-mx-2 block rounded-md px-2 py-1"
                            :class="clickableDatumClass(item)"
                          >
                            <div class="flex items-center justify-between gap-3 text-sm">
                              <span class="min-w-0 truncate text-surface-700">
                                {{ item.label }}
                              </span>
                              <span class="shrink-0 font-semibold text-surface-950">
                                {{ item.valueLabel ?? countFormatter.format(item.value) }}
                              </span>
                            </div>
                            <div class="mt-1.5 h-2 overflow-hidden rounded-full bg-surface-100">
                              <div
                                class="h-full rounded-full"
                                :style="{
                                  width: barWidth(panel.data, item),
                                  backgroundColor: item.color ?? stage.color,
                                }"
                              />
                            </div>
                          </component>
                        </div>

                        <div
                          v-else-if="panel.kind === 'card'"
                          class="mt-4"
                        >
                          <div class="min-w-0">
                            <p class="text-base font-semibold leading-7 text-surface-900">
                              {{ panel.description }}
                            </p>
                          </div>

                          <div
                            class="mt-5 grid border-y border-surface-100 xl:grid-cols-2 xl:gap-x-8"
                          >
                            <component
                              :is="datumComponent(item)"
                              v-for="item in panel.data"
                              :key="item.label"
                              v-bind="datumLinkProps(item, item.label)"
                              class="grid grid-cols-[auto_minmax(0,1fr)_auto] gap-3 border-b border-surface-100 py-3 last:border-b-0 even:xl:border-l even:xl:pl-8 xl:[&:nth-last-child(-n+2)]:border-b-0"
                              :class="clickableDatumClass(item)"
                            >
                              <span
                                class="mt-1.5 h-2.5 w-2.5 shrink-0 rounded-sm"
                                :style="{ backgroundColor: item.color ?? stage.color }"
                                aria-hidden="true"
                              />
                              <div class="min-w-0">
                                <p class="text-sm font-semibold leading-5 text-surface-800">
                                  {{ item.label }}
                                </p>
                                <p class="mt-1 text-xs leading-5 text-surface-500">
                                  {{ item.description }}
                                </p>
                              </div>
                              <div class="text-right">
                                <p class="text-sm font-semibold leading-5" :class="metricToneClass(item.tone)">
                                  {{ item.valueLabel ?? countFormatter.format(item.value) }}
                                </p>
                                <p class="mt-1 text-[11px] font-semibold text-surface-400">
                                  {{ percentFormatter.format(itemPercent(panel.data, item)) }}%
                                </p>
                              </div>
                            </component>
                          </div>
                        </div>

                        <div
                          v-else-if="panel.kind === 'donut'"
                          class="mt-4 flex flex-col gap-4 sm:flex-row sm:items-center"
                        >
                          <div
                            class="relative h-32 w-32 shrink-0 rounded-full"
                            :style="{ background: donutBackground(panel.data) }"
                          >
                            <div
                              class="absolute inset-5 flex flex-col items-center justify-center rounded-full bg-white text-center shadow-sm"
                            >
                              <span class="text-xl font-semibold text-surface-950">
                                {{ countFormatter.format(sumValue(panel.data)) }}
                              </span>
                              <span class="text-xs text-surface-500">proyek</span>
                            </div>
                          </div>
                          <div class="min-w-0 flex-1 space-y-2">
                            <component
                              :is="datumComponent(item)"
                              v-for="item in panel.data"
                              :key="item.label"
                              v-bind="datumLinkProps(item, item.label)"
                              class="-mx-2 grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 rounded-md px-2 py-1 text-sm"
                              :class="clickableDatumClass(item)"
                            >
                              <span
                                class="h-2.5 w-2.5 rounded-sm"
                                :style="{ backgroundColor: item.color ?? stage.color }"
                                aria-hidden="true"
                              />
                              <span class="truncate text-surface-700">{{ item.label }}</span>
                              <span class="font-semibold text-surface-950">
                                {{ countFormatter.format(item.value) }}
                                <span class="font-medium text-surface-400">
                                  {{ percentFormatter.format(itemPercent(panel.data, item)) }}%
                                </span>
                              </span>
                            </component>
                          </div>
                        </div>

                        <div
                          v-else-if="panel.kind === 'donutbar'"
                          class="mt-4 grid gap-6 sm:items-center"
                          style="grid-template-columns: 1fr 3fr"
                        >
                          <!-- Donut chart (left half, centered) -->
                          <div class="flex items-center justify-center">
                            <div
                              class="relative h-40 w-40 shrink-0 rounded-full"
                              :style="{ background: donutBackground(panel.data) }"
                            >
                              <div
                                class="absolute inset-4 flex flex-col items-center justify-center rounded-full bg-white text-center shadow-sm"
                              >
                                <span class="text-lg font-semibold text-surface-950">
                                  {{ countFormatter.format(sumValue(panel.data)) }}
                                </span>
                                <span class="text-[10px] leading-tight text-surface-500">proyek</span>
                              </div>
                            </div>
                          </div>

                          <!-- Bar list (right half) -->
                          <div class="min-w-0 space-y-2.5">
                            <component
                              :is="datumComponent(item)"
                              v-for="item in panel.data"
                              :key="item.label"
                              v-bind="datumLinkProps(item, item.label)"
                              class="-mx-2 block rounded-md px-2 py-1"
                              :class="clickableDatumClass(item)"
                            >
                              <div class="flex items-center justify-between gap-3 text-sm">
                                <span
                                  class="inline-flex min-w-0 items-center gap-2 truncate text-surface-700"
                                >
                                  <span
                                    class="h-2 w-2 shrink-0 rounded-sm"
                                    :style="{ backgroundColor: item.color ?? stage.color }"
                                    aria-hidden="true"
                                  />
                                  {{ item.label }}
                                </span>
                                <span class="shrink-0 font-semibold text-surface-950">
                                  {{ item.valueLabel ?? countFormatter.format(item.value) }}
                                  <span class="ml-1 font-normal text-surface-400">
                                    {{ percentFormatter.format(itemPercent(panel.data, item)) }}%
                                  </span>
                                </span>
                              </div>
                              <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-surface-100">
                                <div
                                  class="h-full rounded-full transition-[width] duration-500"
                                  :style="{
                                    width: barWidth(panel.data, item),
                                    backgroundColor: item.color ?? stage.color,
                                  }"
                                />
                              </div>
                            </component>
                          </div>
                        </div>

                        <div
                          v-else-if="panel.kind === 'regions'"
                          class="mt-4 grid grid-cols-6 gap-1.5"
                        >
                          <component
                            :is="datumComponent(item)"
                            v-for="item in panel.data"
                            :key="item.label"
                            v-bind="datumLinkProps(item, item.label)"
                            class="flex min-h-16 flex-col justify-end rounded-md px-2 py-2 text-white"
                            :class="[regionClass(item.value), clickableRegionClass(item)]"
                          >
                            <span class="truncate text-xs font-semibold">{{ item.label }}</span>
                            <span class="mt-1 text-base font-semibold">
                              {{ countFormatter.format(item.value) }}
                            </span>
                          </component>
                        </div>

                        <div v-else-if="panel.kind === 'stack'" class="mt-4">
                          <div class="flex h-4 overflow-hidden rounded bg-surface-100">
                            <span
                              v-for="item in panel.data"
                              :key="item.label"
                              class="h-full"
                              :style="{
                                width: `${segmentPercent(panel.data, item)}%`,
                                backgroundColor: item.color ?? stage.color,
                              }"
                            />
                          </div>
                          <div class="mt-3 flex flex-wrap gap-x-5 gap-y-2 text-xs text-surface-600">
                            <component
                              :is="datumComponent(item)"
                              v-for="item in panel.data"
                              :key="item.label"
                              v-bind="datumLinkProps(item, item.label)"
                              class="inline-flex items-center gap-2 rounded-md px-1.5 py-1"
                              :class="clickableDatumClass(item)"
                            >
                              <span
                                class="h-2.5 w-2.5 rounded-sm"
                                :style="{ backgroundColor: item.color ?? stage.color }"
                                aria-hidden="true"
                              />
                              {{ item.label }} ·
                              {{ item.valueLabel ?? countFormatter.format(item.value) }}
                            </component>
                          </div>
                        </div>

                        <div
                          v-else-if="panel.kind === 'table'"
                          class="mt-4 overflow-hidden rounded-md border border-surface-100"
                        >
                          <div
                            class="grid grid-cols-[minmax(0,1fr)_7rem_5rem] gap-3 bg-surface-50 px-3 py-2 text-[11px] font-semibold uppercase text-surface-400"
                          >
                            <span>Proyek</span>
                            <span>Status</span>
                            <span class="text-right">Nilai</span>
                          </div>
                          <div
                            v-for="row in panel.rows"
                            :key="row.name"
                            class="grid grid-cols-[minmax(0,1fr)_7rem_5rem] gap-3 border-t border-surface-100 px-3 py-2 text-sm"
                          >
                            <span class="truncate text-surface-800">{{ row.name }}</span>
                            <span class="font-semibold" :class="metricToneClass(row.tone)">
                              {{ row.status }}
                            </span>
                            <span class="text-right font-semibold text-surface-700">
                              {{ row.value }}
                            </span>
                          </div>
                        </div>

                        <div
                          v-else-if="panel.kind === 'flow'"
                          class="mt-4 grid gap-3 sm:grid-cols-[1fr_auto_1fr]"
                        >
                          <div class="space-y-2">
                            <div
                              v-for="pair in panel.pairs"
                              :key="`${pair.source}-${pair.target}`"
                              class="rounded-md border border-surface-100 bg-surface-50 px-3 py-2 text-sm text-surface-700"
                            >
                              {{ pair.source }}
                            </div>
                          </div>
                          <div
                            class="flex items-center justify-center text-xs font-semibold uppercase text-surface-400"
                          >
                            <i class="pi pi-arrow-right hidden sm:block" aria-hidden="true" />
                            <i class="pi pi-arrow-down sm:hidden" aria-hidden="true" />
                          </div>
                          <div class="space-y-2">
                            <div
                              v-for="pair in panel.pairs"
                              :key="`${pair.target}-${pair.value}`"
                              class="flex items-center justify-between gap-3 rounded-md border border-surface-100 bg-white px-3 py-2 text-sm"
                            >
                              <span class="min-w-0 truncate text-surface-700">{{
                                pair.target
                              }}</span>
                              <span class="shrink-0 font-semibold text-surface-950">{{
                                pair.value
                              }}</span>
                            </div>
                          </div>
                        </div>

                        <div v-else-if="panel.kind === 'status'" class="mt-4">
                          <div class="flex h-8 overflow-hidden rounded-md bg-surface-100">
                            <span
                              v-for="item in panel.data"
                              :key="item.label"
                              class="flex items-center justify-center text-xs font-semibold text-white"
                              :style="{
                                width: `${segmentPercent(panel.data, item)}%`,
                                backgroundColor: item.color ?? stage.color,
                              }"
                            >
                              {{ item.valueLabel }}
                            </span>
                          </div>
                          <div class="mt-3 flex flex-wrap gap-x-5 gap-y-2 text-xs text-surface-600">
                            <component
                              :is="datumComponent(item)"
                              v-for="item in panel.data"
                              :key="item.label"
                              v-bind="datumLinkProps(item, item.label)"
                              class="inline-flex items-center gap-2 rounded-md px-1.5 py-1"
                              :class="clickableDatumClass(item)"
                            >
                              <span
                                class="h-2.5 w-2.5 rounded-sm"
                                :style="{ backgroundColor: item.color ?? stage.color }"
                                aria-hidden="true"
                              />
                              {{ item.label }} · {{ item.amountLabel }}
                            </component>
                          </div>
                        </div>
                      </section>
                    </div>

                    <div
                      v-if="!visibleCounters(stage).length && !visiblePanels(stage).length"
                      class="rounded-lg border border-dashed border-surface-200 bg-surface-50 px-4 py-8 text-center"
                    >
                      <p class="text-sm font-semibold text-surface-700">
                        Tidak ada panel pada tab ini.
                      </p>
                      <p class="mt-1 text-xs text-surface-500">
                        Pilih tab lain untuk melihat detail tahap {{ stage.title }}.
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Transition>
        </article>

        <div v-if="stageIndex < stages.length - 1" class="relative bg-white py-3">
          <div class="mx-4 border-t border-dashed border-surface-200 lg:mx-[13rem]" />
          <Tag
            severity="secondary"
            rounded
            class="absolute left-1/2 top-1/2 max-w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 truncate bg-white"
          >
            {{ stage.blockedLabel }}
          </Tag>
        </div>
      </template>
    </div>

  </section>
</template>

<style scoped>
.dashboard-stage-enter-active,
.dashboard-stage-leave-active {
  overflow: hidden;
  transition:
    max-height 220ms ease,
    opacity 180ms ease;
}

.dashboard-stage-enter-from,
.dashboard-stage-leave-to {
  max-height: 0;
  opacity: 0;
}

.dashboard-stage-enter-to,
.dashboard-stage-leave-from {
  max-height: 1800px;
  opacity: 1;
}

.region-lo {
  background: #ccfbf1;
  color: #0f766e;
}

.region-md {
  background: #5eead4;
  color: #0f766e;
}

.region-hi {
  background: #14b8a6;
}

.region-xhi {
  background: #0b6f73;
}
</style>
