<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Accordion from 'primevue/accordion'
import AccordionContent from 'primevue/accordioncontent'
import AccordionHeader from 'primevue/accordionheader'
import AccordionPanel from 'primevue/accordionpanel'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { usePermission } from '@/composables/usePermission'
import { useDaftarKegiatanStore } from '@/stores/daftar-kegiatan.store'
import type { DKProject, GBProjectSummary } from '@/types/daftar-kegiatan.types'
import { formatDate, joinNames } from './daftar-kegiatan-page-utils'

const route = useRoute()
const router = useRouter()
const dkStore = useDaftarKegiatanStore()
const { can } = usePermission()

const dkId = computed(() => String(route.params.id ?? ''))

async function loadData() {
  await Promise.all([
    dkStore.fetchDK(dkId.value),
    dkStore.fetchProjects(dkId.value, { limit: 1000 }),
  ])
}

function gbProjectRoute(project: GBProjectSummary) {
  return project.green_book_id
    ? { name: 'gb-project-detail', params: { gbId: project.green_book_id, id: project.id } }
    : { name: 'green-books' }
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      :title="dkStore.currentDK?.subject ?? 'Detail Daftar Kegiatan'"
      :subtitle="dkStore.currentDK ? formatDate(dkStore.currentDK.date) : undefined"
    >
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'daftar-kegiatan' })" />
        <Button
          v-if="can('daftar_kegiatan', 'create')"
          as="router-link"
          :to="{ name: 'dk-project-create', params: { dkId } }"
          label="Tambah Proyek ke Surat"
          icon="pi pi-plus"
        />
      </template>
    </PageHeader>

    <div v-if="dkStore.currentDK" class="grid gap-4 rounded-lg border border-surface-200 bg-white p-5 md:grid-cols-3">
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Perihal</p>
        <p class="font-semibold text-surface-950">{{ dkStore.currentDK.subject }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Tanggal</p>
        <p class="font-semibold text-surface-950">{{ formatDate(dkStore.currentDK.date) }}</p>
      </div>
      <div>
        <p class="text-xs uppercase tracking-wide text-surface-500">Nomor Surat</p>
        <p class="font-semibold text-surface-950">{{ dkStore.currentDK.letter_number || '-' }}</p>
      </div>
    </div>

    <Accordion v-if="dkStore.projects.length" value="0" class="space-y-3">
      <AccordionPanel
        v-for="(project, index) in dkStore.projects"
        :key="project.id"
        :value="String(index)"
        class="overflow-hidden rounded-lg border border-surface-200 bg-white"
      >
        <AccordionHeader>
          <div class="flex w-full flex-wrap items-center justify-between gap-3 pr-4">
            <span>{{ joinNames((project as DKProject).gb_projects) }}</span>
            <span class="text-sm font-normal text-surface-500">{{ project.duration || '-' }}</span>
          </div>
        </AccordionHeader>
        <AccordionContent>
          <div class="space-y-5 p-2">
            <div class="grid gap-4 md:grid-cols-2">
              <div>
                <p class="text-xs uppercase tracking-wide text-surface-500">Executing Agency</p>
                <p class="font-medium text-surface-900">{{ project.institution?.name ?? project.institution_id ?? '-' }}</p>
              </div>
              <div>
                <p class="text-xs uppercase tracking-wide text-surface-500">Lokasi</p>
                <p class="font-medium text-surface-900">{{ joinNames(project.locations) }}</p>
              </div>
              <div class="md:col-span-2">
                <p class="text-xs uppercase tracking-wide text-surface-500">Objectives</p>
                <p class="text-surface-800">{{ project.objectives || '-' }}</p>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">GB Project Snapshot yang Dipakai</h3>
              <div class="flex flex-wrap gap-2">
                <RouterLink
                  v-for="gbProject in project.gb_projects"
                  :key="gbProject.id"
                  :to="gbProjectRoute(gbProject)"
                  class="inline-flex items-center gap-2 rounded-full border border-surface-200 px-3 py-1.5 text-sm font-medium text-primary"
                >
                  <span>{{ gbProject.gb_code }} - {{ gbProject.project_name }}</span>
                  <Tag
                    v-if="gbProject.has_newer_revision"
                    value="Ada revisi lebih baru"
                    severity="warn"
                    rounded
                  />
                  <Tag v-else-if="gbProject.is_latest" value="Latest" severity="success" rounded />
                </RouterLink>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">Rincian Pembiayaan</h3>
              <div class="overflow-auto rounded-lg border border-surface-200">
                <table class="w-full min-w-[56rem] text-left text-sm">
                  <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
                    <tr>
                      <th class="px-4 py-3">Lender</th>
                      <th class="px-4 py-3">Currency</th>
                      <th class="px-4 py-3">Original</th>
                      <th class="px-4 py-3">USD</th>
                      <th class="px-4 py-3">Remarks</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-surface-100">
                    <tr v-for="row in project.financing_details" :key="row.id">
                      <td class="px-4 py-3">{{ row.lender?.name ?? '-' }}</td>
                      <td class="px-4 py-3">{{ row.currency }}</td>
                      <td class="px-4 py-3"><CurrencyDisplay :amount="row.amount_original" :currency="row.currency" /></td>
                      <td class="px-4 py-3"><CurrencyDisplay :amount="row.amount_usd" /></td>
                      <td class="px-4 py-3">{{ row.remarks || '-' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">Alokasi Pinjaman</h3>
              <div class="overflow-auto rounded-lg border border-surface-200">
                <table class="w-full min-w-[56rem] text-left text-sm">
                  <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
                    <tr>
                      <th class="px-4 py-3">Instansi</th>
                      <th class="px-4 py-3">Currency</th>
                      <th class="px-4 py-3">Original</th>
                      <th class="px-4 py-3">USD</th>
                      <th class="px-4 py-3">Remarks</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-surface-100">
                    <tr v-for="row in project.loan_allocations" :key="row.id">
                      <td class="px-4 py-3">{{ row.institution?.name ?? '-' }}</td>
                      <td class="px-4 py-3">{{ row.currency }}</td>
                      <td class="px-4 py-3"><CurrencyDisplay :amount="row.amount_original" :currency="row.currency" /></td>
                      <td class="px-4 py-3"><CurrencyDisplay :amount="row.amount_usd" /></td>
                      <td class="px-4 py-3">{{ row.remarks || '-' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="space-y-2">
              <h3 class="font-semibold text-surface-950">Rincian Kegiatan</h3>
              <ol class="space-y-2">
                <li
                  v-for="activity in project.activity_details"
                  :key="activity.id"
                  class="rounded-lg border border-surface-200 p-3 text-sm"
                >
                  <span class="font-semibold">{{ activity.activity_number }}.</span>
                  {{ activity.activity_name }}
                </li>
              </ol>
            </div>
          </div>
        </AccordionContent>
      </AccordionPanel>
    </Accordion>

    <div v-else class="rounded-lg border border-dashed border-surface-300 bg-white p-8 text-center text-surface-500">
      Belum ada proyek dalam surat ini.
    </div>
  </section>
</template>
