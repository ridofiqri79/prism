<script setup lang="ts">
import Tag from 'primevue/tag'
import PageHeader from '@/components/common/PageHeader.vue'
import DashboardPrototypeFlow from '@/components/dashboard/DashboardPrototypeFlow.vue'
import type { DashboardFlowStage } from '@/types/dashboard-prototype.types'

const prototypeStages: DashboardFlowStage[] = [
  {
    key: 'BB',
    stepLabel: 'Tahap 01',
    title: 'Blue Book',
    subtitle: 'Baseline pipeline proyek perencanaan.',
    amountTrillion: 29.28,
    count: 96,
    tone: 'tealDark',
    nextLabel: 'Green Book',
    detail: {
      breakdownTitle: 'Komposisi program',
      topProjectsTitle: 'Proyek nilai terbesar',
      metricsTitle: 'Kelengkapan usulan',
      breakdownItems: [
        { label: 'Infrastruktur transportasi', count: 38, amountDisplay: 'Rp 12,1 T', percent: 96 },
        { label: 'Energi dan ketenagalistrikan', count: 22, amountDisplay: 'Rp 8,4 T', percent: 64 },
        { label: 'Sumber daya air', count: 19, amountDisplay: 'Rp 5,7 T', percent: 45 },
        { label: 'Konektivitas digital', count: 17, amountDisplay: 'Rp 3,1 T', percent: 28 },
      ],
      topProjects: [
        {
          name: 'Jalur Kereta Trans-Sulawesi Fase II',
          meta: 'BB-2025-014 - EA: Kemenhub - 60 bulan',
          amountDisplay: 'Rp 4,8 T',
        },
        {
          name: 'PLTA Mamberamo Hulu',
          meta: 'BB-2025-023 - EA: ESDM - 48 bulan',
          amountDisplay: 'Rp 3,2 T',
        },
        {
          name: 'Bendungan Way Apu',
          meta: 'BB-2025-031 - EA: PUPR - 54 bulan',
          amountDisplay: 'Rp 2,1 T',
        },
      ],
      metrics: [
        { label: 'Project cost terisi', value: '96 proyek', tone: 'good' },
        { label: 'Lender indication', value: '84 proyek', tone: 'warning' },
        { label: 'Lokasi terisi', value: '92 proyek', tone: 'good' },
        { label: 'Narasi lengkap', value: '78 proyek', tone: 'warning' },
      ],
    },
  },
  {
    key: 'GB',
    stepLabel: 'Tahap 02',
    title: 'Green Book',
    subtitle: 'Proyek yang sudah masuk daftar rencana prioritas.',
    amountTrillion: 20,
    count: 36,
    tone: 'teal',
    nextLabel: 'Daftar Kegiatan',
    detail: {
      breakdownTitle: 'Komposisi pendanaan GB',
      topProjectsTitle: 'Proyek nilai pendanaan terbesar',
      metricsTitle: 'Kelengkapan data GB',
      breakdownItems: [
        {
          label: 'Pinjaman bilateral/KSA (JICA, KfW)',
          count: 16,
          amountDisplay: 'Rp 8,7 T',
          percent: 72,
        },
        {
          label: 'Pinjaman multilateral (ADB, WB, AIIB)',
          count: 14,
          amountDisplay: 'Rp 7,6 T',
          percent: 64,
        },
        { label: 'Hibah luar negeri', count: 6, amountDisplay: 'Rp 1,1 T', percent: 30 },
        { label: 'Dana pendamping/local component', count: 22, amountDisplay: 'Rp 2,6 T', percent: 45 },
      ],
      topProjects: [
        {
          name: 'MRT Jakarta Fase 3 East-West',
          meta: 'GB-2025-006 - Funding: JICA, ADB - 6 kegiatan - plan 2026-2030',
          amountDisplay: 'Rp 3,9 T',
        },
        {
          name: 'SPAM Regional Jatiluhur II',
          meta: 'GB-2025-011 - Funding: WB - 4 kegiatan - plan 2026-2029',
          amountDisplay: 'Rp 2,2 T',
        },
        {
          name: 'Bandara Singkawang',
          meta: 'GB-2025-018 - Funding: KfW - 3 kegiatan - plan 2026-2028',
          amountDisplay: 'Rp 1,8 T',
        },
      ],
      metrics: [
        { label: 'Relasi BB latest', value: '34/36 proyek', tone: 'warning' },
        { label: 'Funding source', value: '36 proyek', tone: 'good' },
        { label: 'Activity + allocation', value: '31 proyek', tone: 'warning' },
        { label: 'Disbursement plan', value: '29 proyek', tone: 'warning' },
      ],
      notes: [
        'Prioritas tindak lanjut: 2 proyek perlu memakai snapshot BB terbaru.',
        '5 proyek belum punya rencana disbursement tahunan yang lengkap.',
      ],
    },
  },
  {
    key: 'DK',
    stepLabel: 'Tahap 03',
    title: 'Daftar Kegiatan',
    subtitle: 'Snapshot proyek dalam surat dan rincian pembiayaan.',
    amountTrillion: 13.9,
    count: 35,
    tone: 'green',
    nextLabel: 'Loan Agreement',
    detail: {
      breakdownTitle: 'Komposisi financing detail',
      topProjectsTitle: 'Proyek nilai DK terbesar',
      metricsTitle: 'Kelengkapan data DK',
      breakdownItems: [
        { label: 'Loan dari lender funding source', count: 26, amountDisplay: 'Rp 10,8 T', percent: 78 },
        { label: 'Grant dalam financing detail', count: 7, amountDisplay: 'Rp 1,1 T', percent: 34 },
        { label: 'Counterpart / dana pendamping', count: 22, amountDisplay: 'Rp 2,0 T', percent: 54 },
        { label: 'Multi-lender dalam satu proyek DK', count: 9, amountDisplay: 'Rp 3,6 T', percent: 42 },
      ],
      topProjects: [
        {
          name: 'Tol Lingkar Pekanbaru Selatan',
          meta: 'Surat: B-214/D.8/PP.01.02/03/2026 - GB: GB-2025-021 - 5 activities',
          amountDisplay: 'Rp 1,7 T',
        },
        {
          name: 'Pelabuhan Patimban Fase 2',
          meta: 'Surat: B-214/D.8/PP.01.02/03/2026 - GB: GB-2025-013 - 4 activities',
          amountDisplay: 'Rp 1,4 T',
        },
        {
          name: 'PLTS Terapung Singkarak',
          meta: 'Surat: B-198/D.8/PP.01.02/02/2026 - GB: GB-2025-028 - 3 activities',
          amountDisplay: 'Rp 0,9 T',
        },
      ],
      metrics: [
        { label: 'Relasi GB concrete', value: '35 proyek', tone: 'good' },
        { label: 'Financing detail', value: '35 proyek', tone: 'good' },
        { label: 'Loan allocation', value: '32 proyek', tone: 'warning' },
        { label: 'Activity detail', value: '35 proyek', tone: 'good' },
      ],
      notes: [
        '3 proyek perlu review alokasi pinjaman ke institution sebelum dibuatkan LA.',
        'Relasi DK tetap menunjuk snapshot GB yang tersimpan dan tidak auto-pindah saat ada revisi.',
      ],
    },
  },
  {
    key: 'LA',
    stepLabel: 'Tahap 04',
    title: 'Loan Agreement',
    subtitle: 'Perjanjian legal yang dibuat dari proyek DK.',
    amountTrillion: 0,
    count: 0,
    tone: 'gold',
    targetLabel: 'Belum ada LA tercatat',
    detail: {
      breakdownTitle: 'Status pencatatan LA',
      topProjectsTitle: 'Kandidat dari DK terbesar',
      metricsTitle: 'Kelengkapan field LA',
      breakdownItems: [
        { label: 'Loan Agreement tercatat', count: 0, amountDisplay: 'Rp 0', percent: 8 },
        { label: 'DK dengan financing detail', count: 35, amountDisplay: 'Rp 13,9 T', percent: 86 },
        { label: 'DK dengan calon lender lebih dari satu', count: 8, amountDisplay: 'Rp 3,4 T', percent: 38 },
      ],
      topProjects: [
        {
          name: 'MRT Jakarta Fase 3 East-West',
          meta: 'DK Project - lender: JICA, ADB - currency: JPY/USD',
          amountDisplay: 'Rp 3,9 T',
        },
        {
          name: 'SPAM Regional Jatiluhur II',
          meta: 'DK Project - lender: WB - currency: USD',
          amountDisplay: 'Rp 2,2 T',
        },
        {
          name: 'Bandara Singkawang',
          meta: 'DK Project - lender: KfW - currency: EUR',
          amountDisplay: 'Rp 1,8 T',
        },
      ],
      metrics: [
        { label: 'Loan code', value: '0 record' },
        { label: 'Effective date', value: '0 record' },
        { label: 'Closing date', value: '0 record' },
        { label: 'Cumulative disb.', value: 'Rp 0', tone: 'warning' },
      ],
      notes: [
        'LA masih 0 pada snapshot contoh; daftar kandidat diambil dari DK Project dengan financing detail.',
        'Cumulative disbursement akan dicatat manual mengikuti currency LA, tanpa konversi otomatis.',
      ],
    },
  },
]
</script>

<template>
  <section class="space-y-5">
    <PageHeader
      title="Dashboard"
      subtitle="Alur ringkas proyek pinjaman luar negeri dari Blue Book sampai Perjanjian Pinjaman."
    >
      <template #actions>
        <Tag value="Data contoh" severity="warn" />
      </template>
    </PageHeader>

    <section class="rounded-lg border border-surface-200 bg-white p-4 shadow-sm">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div class="min-w-0">
          <p class="text-xs font-semibold uppercase tracking-wide text-primary-700">
            Snapshot contoh
          </p>
          <p class="mt-1 max-w-4xl text-sm leading-6 text-surface-600">
            Angka di halaman ini dipakai untuk validasi desain. Belum ada endpoint dashboard
            backend yang dipanggil.
          </p>
        </div>
        <div class="flex flex-wrap gap-2 text-xs text-surface-600">
          <span class="rounded-full border border-surface-200 bg-surface-50 px-3 py-1">
            BB 96
          </span>
          <span class="rounded-full border border-surface-200 bg-surface-50 px-3 py-1">
            GB 36
          </span>
          <span class="rounded-full border border-surface-200 bg-surface-50 px-3 py-1">
            DK 35
          </span>
          <span class="rounded-full border border-prism-gold/40 bg-prism-gold/10 px-3 py-1">
            LA 0
          </span>
        </div>
      </div>
    </section>

    <DashboardPrototypeFlow
      title="Alur Pipeline Perencanaan"
      last-sync-label="09:14 WIB"
      :stages="prototypeStages"
    />
  </section>
</template>
