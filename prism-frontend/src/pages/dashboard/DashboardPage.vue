<script setup lang="ts">
import DashboardKpiGrid from '@/components/dashboard/DashboardKpiGrid.vue'
import DashboardPageHeader from '@/components/dashboard/DashboardPageHeader.vue'
import PlanningFunnelFlow from '@/components/dashboard/PlanningFunnelFlow.vue'
import type {
  DashboardHeaderAction,
  DashboardKpi,
  DashboardStage,
} from '@/types/dashboard-flow.types'

const kpis: DashboardKpi[] = [
  {
    label: 'Total Proyek Aktif',
    value: '167',
    delta: '+12 vs bulan lalu',
    tone: 'good',
  },
  {
    label: 'Nilai Pipeline',
    value: 'Rp 65,18 T',
    delta: '+4,2% vs bulan lalu',
    tone: 'good',
  },
  {
    label: 'Konversi BB -> LA',
    value: '0,0%',
    delta: 'Bottleneck di tahap DK',
    tone: 'danger',
  },
  {
    label: 'Avg. Cycle Time',
    value: '187',
    unit: 'hari',
    delta: 'Stabil',
  },
]

const headerActions: DashboardHeaderAction[] = [
  { key: 'filter', label: 'Filter', icon: 'pi pi-filter', severity: 'secondary', outlined: true },
  {
    key: 'export',
    label: 'Ekspor',
    icon: 'pi pi-download',
    severity: 'secondary',
    outlined: true,
  },
  { key: 'share', label: 'Bagikan', icon: 'pi pi-share-alt' },
]

const stages: DashboardStage[] = [
  {
    key: 'BB',
    stepLabel: 'Tahap 01',
    title: 'Blue Book',
    subtitle: 'Usulan proyek yang menjadi dasar rencana pinjaman luar negeri.',
    count: 96,
    amountLabel: 'Rp 29,28 T',
    pipelineShare: 100,
    color: '#0b6f73',
    colorSoft: '#0f8f8c',
    nextLabel: 'Green Book',
    conversionLabel: '37,5%',
    blockedLabel: '60 proyek belum lanjut',
    details: {
      tabs: [
        { label: 'Status & Lender Readiness' },
        { label: 'Distribusi K/L & Wilayah' },
        { label: 'Program' },
      ],
      counters: [
        {
          tab: 'Status & Lender Readiness',
          label: 'Belum ada indikasi lender',
          value: '42 proyek',
          meta: '44% dari total Blue Book · Rp 12,8 T',
          tone: 'warning',
        },
        {
          tab: 'Status & Lender Readiness',
          label: 'Sudah indikasi, belum LoI',
          value: '18 proyek',
          meta: '19% dari total Blue Book · Rp 5,9 T',
          tone: 'warning',
        },
        {
          tab: 'Status & Lender Readiness',
          label: 'Sudah lanjut ke Green Book',
          value: '36 proyek',
          meta: '37,5% konversi dari Blue Book',
          tone: 'good',
        },
      ],
      panels: [
        {
          tab: 'Distribusi K/L & Wilayah',
          title: 'Distribusi jenis K/L',
          hint: '96 proyek',
          kind: 'donut',
          span: 'medium',
          data: [
            { label: 'Pemerintah Pusat', value: 52, color: '#0b6f73' },
            { label: 'Pemerintah Daerah', value: 22, color: '#1fb5b2' },
            { label: 'BUMN', value: 14, color: '#3ccb6b' },
            { label: 'BUMD', value: 8, color: '#fdb813' },
          ],
        },
        {
          tab: 'Distribusi K/L & Wilayah',
          title: 'Top K/L',
          hint: 'Top 5',
          kind: 'bars',
          span: 'medium',
          data: [
            { label: 'Kementerian PUPR', value: 24, valueLabel: '24 · Rp 9,8 T', color: '#0b6f73' },
            {
              label: 'Kementerian Perhubungan',
              value: 18,
              valueLabel: '18 · Rp 7,2 T',
              color: '#1fb5b2',
            },
            { label: 'Kementerian ESDM', value: 9, valueLabel: '9 · Rp 3,3 T', color: '#1fa06f' },
            {
              label: 'Kementerian Kesehatan',
              value: 6,
              valueLabel: '6 · Rp 1,9 T',
              color: '#fdb813',
            },
            { label: 'Kemendikbudristek', value: 5, valueLabel: '5 · Rp 1,4 T', color: '#64748b' },
          ],
        },
        {
          tab: 'Distribusi K/L & Wilayah',
          title: 'Wilayah',
          hint: 'Jumlah proyek',
          kind: 'regions',
          span: 'small',
          data: [
            { label: 'Jawa', value: 28 },
            { label: 'Sumatera', value: 17 },
            { label: 'Sulawesi', value: 12 },
            { label: 'Kalimantan', value: 9 },
            { label: 'Papua', value: 6 },
            { label: 'Nasional', value: 24 },
          ],
        },
        {
          tab: 'Program',
          title: 'Program prioritas',
          hint: '8 program',
          kind: 'bars',
          span: 'large',
          data: [
            {
              label: 'Konektivitas & Transportasi',
              value: 38,
              valueLabel: '38 · Rp 12,1 T',
              color: '#0b6f73',
            },
            {
              label: 'Sumber Daya Air & Pangan',
              value: 19,
              valueLabel: '19 · Rp 6,2 T',
              color: '#1fb5b2',
            },
            { label: 'Ketahanan Energi', value: 16, valueLabel: '16 · Rp 5,1 T', color: '#1fa06f' },
            {
              label: 'Transformasi Digital',
              value: 12,
              valueLabel: '12 · Rp 3,9 T',
              color: '#fdb813',
            },
            {
              label: 'Kesehatan & Pendidikan',
              value: 11,
              valueLabel: '11 · Rp 1,98 T',
              color: '#64748b',
            },
          ],
        },
      ],
    },
  },
  {
    key: 'GB',
    stepLabel: 'Tahap 02',
    title: 'Green Book',
    subtitle: 'Proyek prioritas yang sudah masuk rencana pendanaan dan kegiatan.',
    count: 36,
    amountLabel: 'Rp 20,00 T',
    pipelineShare: 37.5,
    color: '#1fb5b2',
    colorSoft: '#26b7a5',
    nextLabel: 'Daftar Kegiatan',
    conversionLabel: '97,2%',
    blockedLabel: '1 proyek belum lanjut',
    details: {
      tabs: [
        { label: 'Relasi BB -> GB' },
        { label: 'Lender & Funding' },
        { label: 'Distribusi K/L' },
      ],
      counters: [
        {
          tab: 'Relasi BB -> GB',
          label: '1 BB -> >1 GB',
          value: '7 BB · 16 GB',
          meta: 'Paket Blue Book yang dipecah menjadi beberapa proyek Green Book',
          tone: 'warning',
        },
        {
          tab: 'Relasi BB -> GB',
          label: '>1 BB -> 1 GB',
          value: '5 GB · 14 BB',
          meta: 'Gabungan tetap harus berasal dari header Blue Book yang sama',
          tone: 'warning',
        },
        {
          tab: 'Relasi BB -> GB',
          label: 'Belum masuk DK',
          value: '1 proyek',
          meta: 'Antrian terakhir sebelum surat Daftar Kegiatan',
          tone: 'danger',
        },
      ],
      panels: [
        {
          tab: 'Relasi BB -> GB',
          title: 'Alur relasi BB ke GB',
          hint: 'Contoh relasi',
          kind: 'flow',
          span: 'medium',
          pairs: [
            { source: 'BB Transportasi Jawa', target: 'GB MRT East-West', value: '1 -> 1' },
            { source: 'BB Air Minum Regional', target: 'GB SPAM Jatiluhur II', value: '3 -> 1' },
            { source: 'BB Energi Terbarukan', target: 'GB PLTS Singkarak', value: '1 -> 2' },
          ],
        },
        {
          tab: 'Lender & Funding',
          title: 'Komposisi lender',
          hint: '36 proyek',
          kind: 'stack',
          span: 'medium',
          data: [
            { label: 'Bilateral', value: 16, valueLabel: '16 · Rp 8,7 T', color: '#0b6f73' },
            { label: 'Multilateral', value: 14, valueLabel: '14 · Rp 7,6 T', color: '#1fb5b2' },
            { label: 'KSA', value: 6, valueLabel: '6 · Rp 1,1 T', color: '#fdb813' },
          ],
        },
        {
          tab: 'Lender & Funding',
          title: 'Top lender',
          hint: 'Top 5',
          kind: 'bars',
          span: 'medium',
          data: [
            { label: 'JICA', value: 10, valueLabel: '10 · Rp 4,1 T', color: '#0b6f73' },
            { label: 'World Bank', value: 7, valueLabel: '7 · Rp 3,0 T', color: '#1fb5b2' },
            { label: 'ADB', value: 6, valueLabel: '6 · Rp 2,4 T', color: '#1fa06f' },
            { label: 'KfW', value: 4, valueLabel: '4 · Rp 1,5 T', color: '#fdb813' },
            { label: 'AIIB', value: 3, valueLabel: '3 · Rp 1,1 T', color: '#64748b' },
          ],
        },
        {
          tab: 'Relasi BB -> GB',
          title: 'Kesiapan dokumen',
          hint: 'Readiness',
          kind: 'bars',
          span: 'medium',
          data: [
            {
              label: 'Funding source lengkap',
              value: 36,
              valueLabel: '36 proyek',
              color: '#1fa06f',
            },
            {
              label: 'Activity & allocation lengkap',
              value: 31,
              valueLabel: '31 proyek',
              color: '#1fb5b2',
            },
            {
              label: 'Disbursement plan lengkap',
              value: 29,
              valueLabel: '29 proyek',
              color: '#fdb813',
            },
            { label: 'Perlu cek BB terbaru', value: 2, valueLabel: '2 proyek', color: '#dc2626' },
          ],
        },
        {
          tab: 'Distribusi K/L',
          title: 'Top K/L dan program',
          hint: '36 proyek',
          kind: 'bars',
          span: 'medium',
          data: [
            { label: 'PUPR', value: 11, valueLabel: '11 · Rp 5,1 T', color: '#0b6f73' },
            { label: 'Kemenhub', value: 8, valueLabel: '8 · Rp 3,4 T', color: '#1fb5b2' },
            { label: 'ESDM', value: 5, valueLabel: '5 · Rp 2,0 T', color: '#1fa06f' },
            { label: 'SDA dan pangan', value: 6, valueLabel: '6 · Rp 2,8 T', color: '#fdb813' },
            { label: 'Digital', value: 4, valueLabel: '4 · Rp 1,2 T', color: '#64748b' },
          ],
        },
        {
          tab: 'Distribusi K/L',
          title: 'Wilayah Green Book',
          hint: 'Jumlah proyek',
          kind: 'regions',
          span: 'small',
          data: [
            { label: 'Jawa', value: 10 },
            { label: 'Sumatera', value: 7 },
            { label: 'Sulawesi', value: 5 },
            { label: 'Kalimantan', value: 4 },
            { label: 'Papua', value: 3 },
            { label: 'Nasional', value: 7 },
          ],
        },
      ],
    },
  },
  {
    key: 'DK',
    stepLabel: 'Tahap 03',
    title: 'Daftar Kegiatan',
    subtitle: 'Proyek yang sudah dicatat dalam surat dan rincian pembiayaan.',
    count: 35,
    amountLabel: 'Rp 13,90 T',
    pipelineShare: 36.4,
    color: '#1fa06f',
    colorSoft: '#3ccb6b',
    nextLabel: 'Loan Agreement',
    conversionLabel: '0,0%',
    blockedLabel: '35 proyek belum lanjut',
    details: {
      tabs: [{ label: 'Relasi GB -> DK' }, { label: 'Lender' }, { label: 'Bottleneck' }],
      counters: [
        {
          tab: 'Relasi GB -> DK',
          label: '1 GB -> >1 DK',
          value: '5 GB · 12 DK',
          meta: 'Paket Green Book yang dipecah menjadi beberapa kegiatan',
          tone: 'warning',
        },
        {
          tab: 'Bottleneck',
          label: 'Belum menjadi LA',
          value: '35 proyek',
          meta: '100% dari Daftar Kegiatan · Rp 13,90 T',
          tone: 'danger',
        },
        {
          tab: 'Bottleneck',
          label: 'Stuck > 180 hari',
          value: '9 proyek',
          meta: 'Perlu eskalasi rapat portofolio',
          tone: 'warning',
        },
      ],
      panels: [
        {
          tab: 'Bottleneck',
          title: 'Penyebab tertahan',
          hint: '35 proyek',
          kind: 'bars',
          span: 'medium',
          data: [
            {
              label: 'Negosiasi term-sheet belum tuntas',
              value: 14,
              valueLabel: '14 · 40%',
              color: '#f6a800',
            },
            {
              label: 'Readiness criteria K/L belum lengkap',
              value: 11,
              valueLabel: '11 · 31%',
              color: '#fdb813',
            },
            {
              label: 'Persetujuan DPR / Komisi XI',
              value: 7,
              valueLabel: '7 · 20%',
              color: '#f59e0b',
            },
            {
              label: 'Lahan, AMDAL, izin lingkungan',
              value: 3,
              valueLabel: '3 · 9%',
              color: '#dc2626',
            },
          ],
        },
        {
          tab: 'Lender',
          title: 'Distribusi lender',
          hint: '35 proyek',
          kind: 'stack',
          span: 'medium',
          data: [
            { label: 'Bilateral', value: 19, valueLabel: '19 · Rp 7,5 T', color: '#0b6f73' },
            { label: 'Multilateral', value: 11, valueLabel: '11 · Rp 4,8 T', color: '#1fb5b2' },
            { label: 'KSA', value: 5, valueLabel: '5 · Rp 1,6 T', color: '#fdb813' },
          ],
        },
        {
          tab: 'Relasi GB -> DK',
          title: 'Top K/L',
          hint: 'Top 5',
          kind: 'bars',
          span: 'small',
          data: [
            { label: 'Kemenhub', value: 9, valueLabel: '9 · Rp 4,1 T', color: '#0b6f73' },
            { label: 'PUPR', value: 8, valueLabel: '8 · Rp 3,3 T', color: '#1fb5b2' },
            { label: 'ESDM', value: 6, valueLabel: '6 · Rp 2,4 T', color: '#1fa06f' },
            { label: 'Kominfo', value: 4, valueLabel: '4 · Rp 1,4 T', color: '#fdb813' },
            { label: 'Kemenkes', value: 3, valueLabel: '3 · Rp 1,1 T', color: '#64748b' },
          ],
        },
        {
          tab: 'Relasi GB -> DK',
          title: 'Wilayah',
          hint: 'Jumlah proyek',
          kind: 'regions',
          span: 'small',
          data: [
            { label: 'Jawa', value: 11 },
            { label: 'Sumatera', value: 8 },
            { label: 'Sulawesi', value: 6 },
            { label: 'Kalimantan', value: 5 },
            { label: 'Papua', value: 3 },
            { label: 'Bali, NT, Maluku', value: 2 },
          ],
        },
        {
          tab: 'Relasi GB -> DK',
          title: 'Program',
          hint: '8 program prioritas',
          kind: 'bars',
          span: 'medium',
          data: [
            {
              label: 'Konektivitas & Transportasi',
              value: 10,
              valueLabel: '10 · Rp 4,4 T',
              color: '#0b6f73',
            },
            { label: 'Ketahanan Energi', value: 7, valueLabel: '7 · Rp 3,0 T', color: '#1fb5b2' },
            {
              label: 'Sumber Daya Air & Pangan',
              value: 6,
              valueLabel: '6 · Rp 2,2 T',
              color: '#1fa06f',
            },
            {
              label: 'Transformasi Digital',
              value: 4,
              valueLabel: '4 · Rp 1,5 T',
              color: '#fdb813',
            },
          ],
        },
      ],
    },
  },
  {
    key: 'LA',
    stepLabel: 'Tahap 04',
    title: 'Loan Agreement',
    subtitle: 'Perjanjian pinjaman yang sudah memiliki dasar legal.',
    count: 0,
    amountLabel: 'Rp 0',
    pipelineShare: 0,
    color: '#f6a800',
    colorSoft: '#fdb813',
    finalLabel: 'Loan Agreement belum tercatat',
    details: {
      tabs: [
        { label: 'Status & Effectiveness' },
        { label: 'Schedule Health' },
        { label: 'Relasi DK -> LA' },
      ],
      counters: [
        {
          tab: 'Status & Effectiveness',
          label: 'LA tercatat',
          value: '0 LA',
          meta: 'Belum ada loan code pada snapshot contoh',
          tone: 'warning',
        },
        {
          tab: 'Status & Effectiveness',
          label: 'Kandidat dari DK',
          value: '35 proyek',
          meta: 'Semua DK masih menjadi kandidat perjanjian pinjaman',
          tone: 'danger',
        },
        {
          tab: 'Schedule Health',
          label: 'Pantau closing',
          value: '0 LA',
          meta: 'Akan terisi setelah tanggal effective dan closing tersedia',
        },
      ],
      panels: [
        {
          tab: 'Schedule Health',
          title: 'Health status',
          hint: 'LA',
          kind: 'status',
          span: 'large',
          data: [
            {
              label: 'On Schedule',
              value: 0,
              valueLabel: '0',
              amountLabel: 'Rp 0',
              color: '#10a36b',
            },
            { label: 'Behind', value: 0, valueLabel: '0', amountLabel: 'Rp 0', color: '#d97706' },
            { label: 'At Risk', value: 0, valueLabel: '0', amountLabel: 'Rp 0', color: '#dc2626' },
          ],
        },
        {
          tab: 'Relasi DK -> LA',
          title: 'Kandidat LA terbesar',
          hint: 'Dari DK',
          kind: 'table',
          span: 'wide',
          rows: [
            {
              name: 'MRT Jakarta Fase 3 East-West',
              status: 'Kandidat',
              value: 'Rp 3,9 T',
              tone: 'warning',
            },
            {
              name: 'SPAM Regional Jatiluhur II',
              status: 'Kandidat',
              value: 'Rp 2,2 T',
              tone: 'warning',
            },
            { name: 'Bandara Singkawang', status: 'Kandidat', value: 'Rp 1,8 T', tone: 'warning' },
            {
              name: 'Tol Lingkar Pekanbaru Selatan',
              status: 'Kandidat',
              value: 'Rp 1,7 T',
              tone: 'warning',
            },
          ],
        },
      ],
    },
  },
]

function handleHeaderAction(_key: string) {
  // Reserved for wiring filters, export, and share once dashboard data is live.
}
</script>

<template>
  <section class="mx-auto max-w-[1320px] space-y-5">
    <DashboardPageHeader
      eyebrow="Dashboard / Perencanaan / Funnel"
      title="Alur Proyek Perencanaan"
      subtitle="Funnel BB -> GB -> DK -> LA - snapshot 06 Mei 2026, 09:14 WIB."
      status-label="Data contoh"
      :actions="headerActions"
      @action="handleHeaderAction"
    />

    <DashboardKpiGrid :items="kpis" />

    <PlanningFunnelFlow :stages="stages" last-sync-label="09:14 WIB" />
  </section>
</template>
