<script setup lang="ts">
import { computed, ref } from 'vue'
import Accordion from 'primevue/accordion'
import AccordionContent from 'primevue/accordioncontent'
import AccordionHeader from 'primevue/accordionheader'
import AccordionPanel from 'primevue/accordionpanel'
import ProgressBar from 'primevue/progressbar'
import Tag from 'primevue/tag'
import type {
  DashboardMetricTone,
  DashboardPrototypeFlowProps,
  DashboardStageKey,
  DashboardStageTone,
} from '@/types/dashboard-prototype.types'

const props = defineProps<DashboardPrototypeFlowProps>()

const expandedStages = ref<string[]>(['BB'])

const countFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 0,
})

const amountFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 2,
  minimumFractionDigits: 0,
})

const percentFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 1,
  minimumFractionDigits: 1,
})

const stageToneClasses: Record<
  DashboardStageTone,
  {
    dot: string
    fill: string
    fillText: string
    progress: string
  }
> = {
  tealDark: {
    dot: 'bg-prism-teal-dark',
    fill: 'bg-prism-teal-dark',
    fillText: 'text-white',
    progress: 'bg-prism-teal-dark',
  },
  teal: {
    dot: 'bg-prism-teal',
    fill: 'bg-prism-teal',
    fillText: 'text-white',
    progress: 'bg-prism-teal',
  },
  green: {
    dot: 'bg-prism-green-deep',
    fill: 'bg-prism-green-deep',
    fillText: 'text-white',
    progress: 'bg-prism-green-deep',
  },
  gold: {
    dot: 'bg-prism-gold',
    fill: 'bg-prism-gold',
    fillText: 'text-surface-950',
    progress: 'bg-prism-gold-deep',
  },
}

const stageRows = computed(() => {
  const baseCount = Math.max(props.stages[0]?.count ?? 0, 1)

  return props.stages.map((stage, index) => {
    const nextStage = props.stages[index + 1]
    const percentOfPipeline = (stage.count / baseCount) * 100
    const conversion = nextStage && stage.count > 0 ? (nextStage.count / stage.count) * 100 : null
    const projectDelta = nextStage ? nextStage.count - stage.count : null
    const amountDelta = nextStage ? nextStage.amountTrillion - stage.amountTrillion : null
    const blockedCount = nextStage ? Math.max(stage.count - nextStage.count, 0) : 0
    const minWidth = stage.count > 0 ? '9rem' : '6.5rem'

    return {
      stage,
      nextStage,
      percentOfPipeline,
      conversion,
      projectDelta,
      amountDelta,
      blockedCount,
      barWidth: `${Math.min(Math.max(percentOfPipeline, stage.count > 0 ? 12 : 8), 100)}%`,
      minWidth,
    }
  })
})

const chartAriaLabel = computed(() =>
  props.stages
    .map((stage) => `${stage.title}: ${countFormatter.format(stage.count)} proyek`)
    .join(', '),
)

const legendItems = computed(() =>
  props.stages.map((stage) => ({
    key: stage.key,
    label: stage.title,
    tone: stage.tone,
  })),
)

function isExpanded(stage: DashboardStageKey) {
  return expandedStages.value.includes(stage)
}

function formatAmount(amountTrillion: number) {
  if (amountTrillion === 0) return 'Rp 0'
  return `Rp ${amountFormatter.format(amountTrillion)} T`
}

function formatSignedAmount(amountTrillion: number | null) {
  if (amountTrillion === null || amountTrillion === 0) return 'Rp 0'
  const sign = amountTrillion > 0 ? '+' : '-'
  return `${sign}Rp ${amountFormatter.format(Math.abs(amountTrillion))} T`
}

function formatPercent(value: number | null) {
  if (value === null) return '-'
  return `${percentFormatter.format(value)}%`
}

function formatProjectDelta(value: number | null) {
  if (value === null || value === 0) return '0'
  return value > 0 ? `+${countFormatter.format(value)}` : `-${countFormatter.format(Math.abs(value))}`
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

function tagSeverity(tone: DashboardStageTone) {
  if (tone === 'green') return 'success'
  if (tone === 'gold') return 'warn'
  return 'info'
}

function toneClass(tone: DashboardStageTone) {
  return stageToneClasses[tone]
}
</script>

<template>
  <section class="overflow-hidden rounded-lg border border-surface-200 bg-white shadow-sm">
    <div
      class="flex flex-col gap-3 border-b border-surface-200 px-4 py-4 lg:flex-row lg:items-center lg:justify-between"
    >
      <div class="min-w-0">
        <h2 class="text-base font-semibold text-surface-950">{{ title }}</h2>
        <p class="mt-1 text-sm leading-5 text-surface-500">
          Klik tahap untuk melihat komposisi, proyek utama, dan catatan operasional.
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <Tag
          v-for="item in legendItems"
          :key="item.key"
          :severity="tagSeverity(item.tone)"
          rounded
        >
          <span class="inline-flex items-center gap-1.5">
            <span class="h-2 w-2 rounded-sm" :class="toneClass(item.tone).dot" />
            {{ item.label }}
          </span>
        </Tag>
      </div>
    </div>

    <div role="img" :aria-label="chartAriaLabel">
      <div
        class="hidden grid-cols-[11rem_minmax(24rem,1fr)_12rem] border-b border-surface-100 bg-surface-50/80 px-4 py-3 text-xs font-semibold uppercase tracking-wide text-surface-500 lg:grid"
      >
        <span>Tahap</span>
        <span>Volume dan nilai</span>
        <span class="text-right">Konversi berikutnya</span>
      </div>

      <Accordion
        v-model:value="expandedStages"
        multiple
        lazy
        :pt="{
          root: { class: 'divide-y divide-surface-100' },
        }"
      >
        <template v-for="row in stageRows" :key="row.stage.key">
          <AccordionPanel
            :value="row.stage.key"
            :pt="{
              root: { class: 'border-0 bg-white' },
            }"
          >
            <AccordionHeader
              :pt="{
                root: {
                  class:
                    'block w-full border-0 bg-white p-0 text-left transition duration-150 hover:bg-primary-50/40 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-primary',
                },
              }"
            >
              <template #toggleicon>
                <span class="hidden" aria-hidden="true" />
              </template>

              <div
                class="grid w-full bg-white text-left transition duration-150 hover:bg-primary-50/40 lg:grid-cols-[11rem_minmax(24rem,1fr)_12rem]"
              >
                <div class="border-surface-100 px-4 py-4 lg:border-r">
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <p class="text-xs font-semibold uppercase tracking-wide text-surface-400">
                        {{ row.stage.stepLabel }}
                      </p>
                      <div class="mt-1 flex items-center gap-2">
                        <span
                          class="h-2.5 w-2.5 shrink-0 rounded-sm"
                          :class="toneClass(row.stage.tone).dot"
                        />
                        <span class="min-w-0 text-sm font-semibold text-surface-950">
                          {{ row.stage.title }}
                        </span>
                      </div>
                      <p class="mt-1 text-xs leading-5 text-surface-500 lg:hidden">
                        {{ row.stage.subtitle }}
                      </p>
                    </div>
                    <i
                      class="pi pi-angle-down mt-0.5 shrink-0 text-xs text-surface-400 transition-transform"
                      :class="isExpanded(row.stage.key) ? 'rotate-180 text-primary-700' : ''"
                      aria-hidden="true"
                    />
                  </div>
                </div>

                <div class="border-surface-100 px-4 pb-4 lg:border-r lg:py-4">
                  <div class="rounded-lg border border-surface-200 bg-surface-50 p-1">
                    <div
                      class="flex h-14 max-w-full items-center justify-between rounded-md px-4 transition-[width] duration-300"
                      :class="[toneClass(row.stage.tone).fill, toneClass(row.stage.tone).fillText]"
                      :style="{
                        width: row.barWidth,
                        minWidth: row.minWidth,
                      }"
                    >
                      <div class="min-w-0">
                        <p class="truncate text-xl font-semibold leading-none">
                          {{ countFormatter.format(row.stage.count) }}
                          <span class="text-xs font-medium">proyek</span>
                        </p>
                        <p class="mt-1 truncate text-xs font-medium">
                          {{ formatAmount(row.stage.amountTrillion) }}
                        </p>
                      </div>
                      <p class="hidden shrink-0 text-xs font-semibold sm:block">
                        {{
                          row.stage.key === 'BB'
                            ? '100%'
                            : formatPercent(row.percentOfPipeline)
                        }}
                      </p>
                    </div>
                  </div>
                </div>

                <div
                  class="flex flex-row items-center justify-between gap-3 bg-surface-50/60 px-4 pb-4 text-left lg:flex-col lg:items-end lg:justify-center lg:bg-white lg:py-4 lg:text-right"
                >
                  <template v-if="row.nextStage">
                    <div>
                      <p
                        class="text-lg font-semibold"
                        :class="row.conversion === 0 ? 'text-red-600' : 'text-surface-950'"
                      >
                        {{ formatPercent(row.conversion) }}
                      </p>
                      <p class="mt-0.5 text-xs font-medium uppercase tracking-wide text-surface-400">
                        ke {{ row.stage.nextLabel }}
                      </p>
                    </div>
                    <p
                      class="text-xs font-medium"
                      :class="row.blockedCount > 0 ? 'text-red-600' : 'text-prism-green-dark'"
                    >
                      {{ countFormatter.format(row.blockedCount) }} tertahan
                    </p>
                  </template>
                  <template v-else>
                    <span class="h-px w-6 bg-surface-700" aria-hidden="true" />
                    <div>
                      <p class="text-xs font-semibold uppercase tracking-wide text-surface-400">
                        Tahap final
                      </p>
                      <p class="mt-1 text-xs font-medium text-surface-700">
                        {{ row.stage.targetLabel }}
                      </p>
                    </div>
                  </template>
                </div>
              </div>
            </AccordionHeader>

            <AccordionContent
              :pt="{
                root: { class: 'border-0 bg-surface-50/70' },
                content: { class: 'border-0 bg-transparent px-4 py-4' },
              }"
            >
              <div
                class="grid overflow-hidden rounded-lg border border-surface-200 bg-white xl:grid-cols-[minmax(0,1.05fr)_minmax(0,0.9fr)_minmax(0,0.8fr)] xl:divide-x xl:divide-surface-200"
              >
                <section class="min-w-0 p-4">
                  <h3 class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                    {{ row.stage.detail.breakdownTitle }}
                  </h3>
                  <div class="mt-3 space-y-3">
                    <div
                      v-for="item in row.stage.detail.breakdownItems"
                      :key="item.label"
                      class="min-w-0"
                    >
                      <div class="flex flex-col gap-1 text-sm sm:flex-row sm:items-center sm:justify-between sm:gap-3">
                        <span class="min-w-0 leading-5 text-surface-800">{{ item.label }}</span>
                        <span class="shrink-0 text-xs font-medium text-surface-700 sm:text-right">
                          {{ countFormatter.format(item.count) }} proyek
                          <template v-if="item.amountDisplay"> - {{ item.amountDisplay }}</template>
                        </span>
                      </div>
                      <ProgressBar
                        :value="Math.min(Math.max(item.percent, 0), 100)"
                        :show-value="false"
                        class="mt-1.5"
                        :pt="{
                          root: { class: 'h-1.5 overflow-hidden rounded-full bg-surface-100' },
                          value: { class: toneClass(row.stage.tone).progress },
                        }"
                      />
                    </div>
                  </div>
                </section>

                <section class="min-w-0 border-t border-surface-200 p-4 xl:border-t-0">
                  <h3 class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                    {{ row.stage.detail.topProjectsTitle }}
                  </h3>
                  <div class="mt-3 divide-y divide-surface-100">
                    <div
                      v-for="project in row.stage.detail.topProjects"
                      :key="project.name"
                      class="py-2 first:pt-0 last:pb-0"
                    >
                      <div class="flex items-start justify-between gap-3">
                        <p class="min-w-0 text-sm font-semibold leading-5 text-surface-950">
                          {{ project.name }}
                        </p>
                        <span class="shrink-0 text-xs font-medium text-surface-700">
                          {{ project.amountDisplay }}
                        </span>
                      </div>
                      <p class="mt-1 truncate text-xs text-surface-500">
                        {{ project.meta }}
                      </p>
                    </div>
                  </div>
                </section>

                <section class="min-w-0 border-t border-surface-200 p-4 xl:border-t-0">
                  <h3 class="text-xs font-semibold uppercase tracking-wide text-surface-500">
                    {{ row.stage.detail.metricsTitle }}
                  </h3>
                  <div
                    class="mt-3 grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-surface-200 bg-surface-200"
                  >
                    <div
                      v-for="metric in row.stage.detail.metrics"
                      :key="metric.label"
                      class="min-h-16 bg-white p-3"
                    >
                      <p class="text-xs font-medium text-surface-500">
                        {{ metric.label }}
                      </p>
                      <p class="mt-1 text-lg font-semibold" :class="metricToneClass(metric.tone)">
                        {{ metric.value }}
                      </p>
                    </div>
                  </div>
                  <div v-if="row.stage.detail.notes?.length" class="mt-3 space-y-1 text-sm text-surface-600">
                    <p v-for="note in row.stage.detail.notes" :key="note" class="leading-5">
                      {{ note }}
                    </p>
                  </div>
                </section>
              </div>
            </AccordionContent>
          </AccordionPanel>

          <div v-if="row.nextStage" class="relative bg-white py-2">
            <div class="mx-4 border-t border-dashed border-surface-200 lg:mx-[11.75rem]" />
            <Tag
              severity="secondary"
              rounded
              class="absolute left-1/2 top-1/2 max-w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 truncate"
            >
              <span :class="row.blockedCount > 0 ? 'text-red-600' : 'text-prism-green-dark'">
                {{ formatProjectDelta(row.projectDelta) }} proyek / {{ formatSignedAmount(row.amountDelta) }}
              </span>
              <span class="text-surface-500">
                {{ row.blockedCount > 0 ? ' tertahan' : ' bertambah' }}
              </span>
            </Tag>
          </div>
        </template>
      </Accordion>
    </div>

    <div
      class="flex flex-wrap items-center justify-end gap-2 border-t border-surface-200 bg-surface-50/70 px-4 py-3 text-xs text-surface-600"
    >
      <span>Klik tahap untuk membuka detail</span>
      <Tag :value="`Snapshot ${lastSyncLabel}`" severity="secondary" rounded />
    </div>
  </section>
</template>
