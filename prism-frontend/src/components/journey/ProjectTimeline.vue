<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import AbsorptionBar from '@/components/monitoring/AbsorptionBar.vue'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import type {
  DKProjectJourney,
  GBProjectJourney,
  JourneyFundingSource,
  JourneyLoI,
  JourneyResponse,
  LAJourney,
  MonitoringSummaryItem,
} from '@/types/dashboard.types'

const props = defineProps<{
  journey: JourneyResponse
}>()

const openKeys = ref(new Set<string>())

const bbProject = computed(() => props.journey.bb_project)

function isOpen(key: string) {
  return openKeys.value.has(key)
}

function toggle(key: string) {
  const next = new Set(openKeys.value)
  if (next.has(key)) {
    next.delete(key)
  } else {
    next.add(key)
  }
  openKeys.value = next
}

function nodeClass(state: 'completed' | 'pending' | 'extended') {
  if (state === 'extended') return 'text-prism-gold-deep'
  if (state === 'completed') return 'text-prism-green-deep'
  return 'text-surface-400'
}

function formatDate(value?: string | null) {
  if (!value) return '-'
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(new Date(value))
}

function formatFunding(source: JourneyFundingSource) {
  const lender = source.lender?.short_name || source.lender?.name || 'Lender'
  const amount = source.amount_usd ?? source.amount_original ?? 0
  return `${lender} ${new Intl.NumberFormat('en-US', {
    notation: 'compact',
    maximumFractionDigits: 2,
  }).format(amount)}`
}

function loiDate(loi: JourneyLoI) {
  return loi.date ?? loi.tanggal
}

function bbProjectRoute() {
  if (!bbProject.value.blue_book_id) return ''
  return `/blue-books/${bbProject.value.blue_book_id}/projects/${bbProject.value.id}`
}

function gbProjectRoute(project: GBProjectJourney) {
  if (!project.green_book_id) return ''
  return `/green-books/${project.green_book_id}/projects/${project.id}`
}

function dkRoute(project: DKProjectJourney) {
  if (!project.daftar_kegiatan?.id) return ''
  return `/daftar-kegiatan/${project.daftar_kegiatan.id}`
}

function laRoute(loanAgreement: LAJourney) {
  return `/loan-agreements/${loanAgreement.id}`
}

function monitoringRoute(loanAgreement: LAJourney, monitoring?: MonitoringSummaryItem) {
  return `/loan-agreements/${loanAgreement.id}/monitoring${monitoring?.id ? `#${monitoring.id}` : ''}`
}

watch(
  () => props.journey,
  (journey) => {
    openKeys.value = new Set([
      'blue-book',
      'green-book',
      ...journey.gb_projects.map((project) => `gb-${project.id}`),
    ])
  },
  { immediate: true },
)
</script>

<template>
  <div class="rounded-lg border border-surface-200 bg-white p-5">
    <div class="space-y-4 border-l border-surface-200 pl-5">
      <div class="relative">
        <span
          class="absolute -left-[1.82rem] top-1.5 h-3 w-3 rounded-full bg-prism-green ring-4 ring-white"
        />
        <div class="flex items-start gap-3">
          <Button
            icon="pi pi-chevron-down"
            text
            rounded
            size="small"
            :class="{ '-rotate-90': !isOpen('blue-book') }"
            @click="toggle('blue-book')"
          />
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <i class="pi pi-book text-prism-green-deep" />
              <RouterLink
                v-if="bbProjectRoute()"
                :to="bbProjectRoute()"
                class="font-semibold text-prism-green-deep hover:underline"
              >
                Blue Book
              </RouterLink>
              <span v-else class="font-semibold text-prism-green-deep">Blue Book</span>
              <Tag value="Completed" severity="success" rounded />
            </div>
            <p class="mt-1 text-sm text-surface-900">
              {{ bbProject.bb_code }} - {{ bbProject.project_name }}
            </p>

            <div v-if="isOpen('blue-book')" class="mt-3 space-y-3 border-l border-surface-100 pl-5">
              <div class="text-sm">
                <p class="font-medium text-surface-700">Indikasi Lender</p>
                <p
                  v-if="(bbProject.lender_indications?.length ?? 0) === 0"
                  class="italic text-surface-400"
                >
                  Belum ada
                </p>
                <p v-else class="text-surface-600">
                  {{
                    bbProject.lender_indications
                      ?.map((item) => item.lender?.short_name || item.lender?.name)
                      .filter(Boolean)
                      .join(', ')
                  }}
                </p>
              </div>

              <div class="space-y-2">
                <p class="text-sm font-medium text-surface-700">LoI</p>
                <div
                  v-if="journey.loi.length === 0"
                  class="flex items-center gap-2 text-sm italic text-surface-400"
                >
                  <i class="pi pi-circle" />
                  Belum ada
                </div>
                <div
                  v-for="loi in journey.loi"
                  :key="loi.id"
                  class="flex flex-wrap items-center gap-2 text-sm text-prism-green-deep"
                >
                  <i class="pi pi-check-circle" />
                  <span>{{
                    loi.lender?.short_name || loi.lender?.name || loi.subject || 'LoI'
                  }}</span>
                  <span class="text-surface-500">{{ formatDate(loiDate(loi)) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="relative">
        <span
          class="absolute -left-[1.82rem] top-1.5 h-3 w-3 rounded-full ring-4 ring-white"
          :class="journey.gb_projects.length ? 'bg-prism-green' : 'bg-surface-300'"
        />
        <div class="flex items-start gap-3">
          <Button
            icon="pi pi-chevron-down"
            text
            rounded
            size="small"
            :class="{ '-rotate-90': !isOpen('green-book') }"
            @click="toggle('green-book')"
          />
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <i
                class="pi pi-folder"
                :class="nodeClass(journey.gb_projects.length ? 'completed' : 'pending')"
              />
              <span
                class="font-semibold"
                :class="nodeClass(journey.gb_projects.length ? 'completed' : 'pending')"
              >
                Green Book
              </span>
              <Tag
                :value="
                  journey.gb_projects.length ? `${journey.gb_projects.length} project` : 'Belum ada'
                "
                :severity="journey.gb_projects.length ? 'success' : 'secondary'"
                rounded
              />
            </div>

            <div
              v-if="isOpen('green-book')"
              class="mt-3 space-y-4 border-l border-surface-100 pl-5"
            >
              <p v-if="journey.gb_projects.length === 0" class="text-sm italic text-surface-400">
                Belum ada
              </p>

              <div v-for="gbProject in journey.gb_projects" :key="gbProject.id" class="space-y-3">
                <div class="flex items-start gap-3">
                  <Button
                    icon="pi pi-chevron-down"
                    text
                    rounded
                    size="small"
                    :class="{ '-rotate-90': !isOpen(`gb-${gbProject.id}`) }"
                    @click="toggle(`gb-${gbProject.id}`)"
                  />
                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <i class="pi pi-folder-open text-prism-green-deep" />
                      <RouterLink
                        v-if="gbProjectRoute(gbProject)"
                        :to="gbProjectRoute(gbProject)"
                        class="font-semibold text-prism-green-deep hover:underline"
                      >
                        {{ gbProject.gb_code }}
                      </RouterLink>
                      <span v-else class="font-semibold text-prism-green-deep">{{
                        gbProject.gb_code
                      }}</span>
                      <Tag value="Completed" severity="success" rounded />
                    </div>
                    <p class="mt-1 text-sm text-surface-900">{{ gbProject.project_name }}</p>

                    <div
                      v-if="isOpen(`gb-${gbProject.id}`)"
                      class="mt-3 space-y-3 border-l border-surface-100 pl-5"
                    >
                      <div class="text-sm">
                        <p class="font-medium text-surface-700">Funding</p>
                        <p
                          v-if="(gbProject.funding_sources?.length ?? 0) === 0"
                          class="italic text-surface-400"
                        >
                          Belum ada
                        </p>
                        <p v-else class="text-surface-600">
                          {{ gbProject.funding_sources?.map(formatFunding).join(', ') }}
                        </p>
                      </div>

                      <div
                        v-if="gbProject.dk_projects.length === 0"
                        class="text-sm italic text-surface-400"
                      >
                        Daftar Kegiatan: Belum ada
                      </div>

                      <div
                        v-for="dkProject in gbProject.dk_projects"
                        :key="dkProject.id"
                        class="space-y-3"
                      >
                        <div class="flex flex-wrap items-center gap-2">
                          <i class="pi pi-list text-prism-green-deep" />
                          <RouterLink
                            v-if="dkRoute(dkProject)"
                            :to="dkRoute(dkProject)"
                            class="font-semibold text-prism-green-deep hover:underline"
                          >
                            Daftar Kegiatan:
                            {{
                              dkProject.daftar_kegiatan?.subject ||
                              dkProject.objectives ||
                              dkProject.id
                            }}
                          </RouterLink>
                          <span v-else class="font-semibold text-prism-green-deep">
                            Daftar Kegiatan:
                            {{
                              dkProject.daftar_kegiatan?.subject ||
                              dkProject.objectives ||
                              dkProject.id
                            }}
                          </span>
                          <span class="text-sm text-surface-500">
                            {{
                              formatDate(
                                dkProject.daftar_kegiatan?.date ??
                                  dkProject.daftar_kegiatan?.tanggal,
                              )
                            }}
                          </span>
                        </div>

                        <div class="ml-7 border-l border-surface-100 pl-5">
                          <div
                            v-if="!dkProject.loan_agreement"
                            class="flex items-center gap-2 text-sm italic text-surface-400"
                          >
                            <i class="pi pi-file" />
                            Loan Agreement: Belum ada
                          </div>

                          <div v-else class="space-y-3">
                            <div class="flex flex-wrap items-center gap-2">
                              <i
                                class="pi pi-file-edit"
                                :class="
                                  nodeClass(
                                    dkProject.loan_agreement.is_extended ? 'extended' : 'completed',
                                  )
                                "
                              />
                              <RouterLink
                                :to="laRoute(dkProject.loan_agreement)"
                                class="font-semibold hover:underline"
                                :class="
                                  nodeClass(
                                    dkProject.loan_agreement.is_extended ? 'extended' : 'completed',
                                  )
                                "
                              >
                                Loan Agreement: {{ dkProject.loan_agreement.loan_code }}
                              </RouterLink>
                              <Tag
                                :value="
                                  dkProject.loan_agreement.is_extended ? 'Extended' : 'Completed'
                                "
                                :severity="
                                  dkProject.loan_agreement.is_extended ? 'warn' : 'success'
                                "
                                rounded
                              />
                            </div>
                            <div class="grid gap-2 text-sm text-surface-600 md:grid-cols-2">
                              <span
                                >Effective:
                                {{ formatDate(dkProject.loan_agreement.effective_date) }}</span
                              >
                              <span
                                >Closing:
                                {{ formatDate(dkProject.loan_agreement.closing_date) }}</span
                              >
                            </div>

                            <div class="space-y-2 border-l border-surface-100 pl-5">
                              <RouterLink
                                :to="monitoringRoute(dkProject.loan_agreement)"
                                class="inline-flex items-center gap-2 text-sm font-semibold text-prism-green-deep hover:underline"
                              >
                                <i class="pi pi-chart-line" />
                                Monitoring
                              </RouterLink>
                              <p
                                v-if="dkProject.loan_agreement.monitoring.length === 0"
                                class="text-sm italic text-surface-400"
                              >
                                Belum ada
                              </p>
                              <RouterLink
                                v-for="monitoring in dkProject.loan_agreement.monitoring"
                                :key="`${monitoring.budget_year}-${monitoring.quarter}`"
                                :to="monitoringRoute(dkProject.loan_agreement, monitoring)"
                                class="grid gap-2 rounded-lg border border-surface-100 p-3 text-sm hover:border-primary/50 md:grid-cols-[1fr_10rem]"
                              >
                                <span class="font-medium text-prism-green-deep">
                                  {{ monitoring.quarter }} {{ monitoring.budget_year }}
                                </span>
                                <AbsorptionBar :pct="monitoring.absorption_pct" />
                                <span class="text-surface-500">
                                  <CurrencyDisplay
                                    :amount="monitoring.planned_usd"
                                    currency="USD"
                                    compact
                                  />
                                  planned
                                </span>
                                <span class="text-surface-500">
                                  <CurrencyDisplay
                                    :amount="monitoring.realized_usd"
                                    currency="USD"
                                    compact
                                  />
                                  realized
                                </span>
                              </RouterLink>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
