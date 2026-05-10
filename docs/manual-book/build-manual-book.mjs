import fs from 'node:fs/promises'
import path from 'node:path'
import { execFile } from 'node:child_process'
import { createRequire } from 'node:module'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { promisify } from 'node:util'

const require = createRequire(import.meta.url)
const { chromium } = require('playwright')
const { PDFDocument, StandardFonts, rgb } = require('pdf-lib')
const execFileAsync = promisify(execFile)

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)
const repoRoot = path.resolve(__dirname, '..', '..')
const outputDir = __dirname
const assetsDir = path.join(outputDir, 'assets')
const htmlPath = path.join(outputDir, 'PRISM_Manual_Book.html')
const pdfPath = path.join(outputDir, 'PRISM_Manual_Book.pdf')
const frontendUrl = process.env.PRISM_FRONTEND_URL ?? 'http://localhost:5173'
const apiUrl = process.env.PRISM_API_URL ?? 'http://localhost:8080/api/v1'
const htmlOnly = process.argv.includes('--html-only')

const screenshotTargets = [
  {
    key: 'login',
    route: '/login',
    title: 'Halaman Masuk',
    caption: 'Pintu masuk PRISM untuk ADMIN dan STAFF.',
    public: true,
  },
  {
    key: 'dashboard',
    route: '/dashboard',
    title: 'Dashboard Pinjaman Luar Negeri',
    caption: 'Ikhtisar pipeline dari Blue Book sampai Perjanjian Pinjaman.',
  },
  {
    key: 'project-master',
    route: '/projects',
    title: 'Proyek',
    caption: 'Tabel gabungan proyek dengan pencarian, penyaring data, dan unduh Excel.',
  },
  {
    key: 'blue-book',
    route: '/blue-books',
    title: 'Blue Book',
    caption: 'Daftar dokumen Blue Book dan proyek indikatif.',
  },
  {
    key: 'green-book',
    route: '/green-books',
    title: 'Green Book',
    caption: 'Daftar Green Book dan proyek prioritas pendanaan.',
  },
  {
    key: 'daftar-kegiatan',
    route: '/daftar-kegiatan',
    title: 'Daftar Kegiatan',
    caption: 'Surat Daftar Kegiatan dan proyek yang siap masuk pembiayaan.',
  },
  {
    key: 'loan-agreements',
    route: '/loan-agreements',
    title: 'Perjanjian Pinjaman',
    caption: 'Kontrak pinjaman, alokasi komitmen, dan indikator kinerja.',
  },
  {
    key: 'spatial',
    route: '/spatial-distribution',
    title: 'Sebaran Wilayah',
    caption: 'Peta choropleth dan daftar proyek berdasarkan wilayah.',
  },
  {
    key: 'import-data',
    route: '/master/import-data',
    title: 'Impor Data',
    caption: 'Template, pemeriksaan awal, dan eksekusi impor file Excel.',
  },
]

const stageFlowSteps = [
  {
    number: 'Tahap 01',
    title: 'Blue Book',
    description: 'Usulan dan indikasi awal proyek.',
    imageKey: 'blue-book',
    color: 'blue',
  },
  {
    number: 'Tahap 02',
    title: 'Green Book',
    description: 'Prioritas pendanaan dan kegiatan.',
    imageKey: 'green-book',
    color: 'green',
  },
  {
    number: 'Tahap 03',
    title: 'Daftar Kegiatan',
    description: 'Surat kegiatan dan pembiayaan.',
    imageKey: 'daftar-kegiatan',
    color: 'orange',
  },
  {
    number: 'Tahap 04',
    title: 'Perjanjian Pinjaman',
    description: 'Komitmen legal dan kinerja pinjaman.',
    imageKey: 'loan-agreements',
    color: 'violet',
  },
]

const modules = [
  {
    title: 'Dashboard',
    owner: 'Semua pengguna berizin baca proyek',
    purpose:
      'Memberikan gambaran cepat mengenai portofolio pinjaman luar negeri: jumlah proyek pada setiap tahap, nilai pinjaman, sebaran pemberi pinjaman, instansi, wilayah, program, dan kondisi Perjanjian Pinjaman.',
    steps: [
      'Buka menu Dashboard.',
      'Pilih periode bila tersedia untuk membatasi konteks portofolio.',
      'Gunakan funnel tahap untuk melihat posisi proyek pada Blue Book, Green Book, Daftar Kegiatan, dan Perjanjian Pinjaman.',
      'Klik kartu atau panel analitik untuk membuka daftar Proyek dengan penyaring data yang sesuai.',
    ],
    notes: [
      'Angka pada Dashboard adalah ringkasan dari data yang sudah tercatat di aplikasi.',
      'Jika ada kartu yang kosong, periksa kembali data proyek dan pilihan periode sebelum menyimpulkan bahwa proyek belum tersedia.',
    ],
  },
  {
    title: 'Proyek',
    owner: 'Perencana dan peninjau portofolio',
    purpose:
      'Menjadi meja kerja untuk mencari seluruh Proyek Blue Book beserta tahapan proses, pemberi pinjaman, instansi, lokasi, nilai pinjaman, dan riwayat revisinya.',
    steps: [
      'Gunakan pencarian untuk menemukan kode proyek, nama proyek, pemberi pinjaman, instansi pelaksana, atau program.',
      'Buka panel penyaring data untuk memilih jenis pinjaman, tahap proses, status proyek, instansi, wilayah, nilai pinjaman, dan tanggal Daftar Kegiatan.',
      'Aktifkan riwayat revisi bila perlu melihat data versi lama, bukan hanya versi terkini.',
      'Klik Unduh Excel untuk mengunduh seluruh data sesuai penyaring yang sedang aktif.',
    ],
    notes: [
      'Halaman Proyek berguna sebagai titik pemeriksaan sebelum menentukan tindak lanjut.',
      'Dashboard dan Sebaran Wilayah dapat membawa pengguna langsung ke daftar Proyek dengan konteks data yang sama.',
    ],
  },
  {
    title: 'Perjalanan Proyek',
    owner: 'Peninjau proyek dan pengendali alur kerja',
    purpose:
      'Menelusuri perjalanan satu proyek dari Blue Book, Green Book, Daftar Kegiatan, sampai Perjanjian Pinjaman, termasuk tanda bila tersedia revisi yang lebih baru.',
    steps: [
      'Buka menu Perjalanan Proyek.',
      'Cari Proyek Blue Book berdasarkan kode atau nama proyek.',
      'Klik Lihat Perjalanan.',
      'Gunakan tab ringkasan, visualisasi alur, atau timeline sesuai kebutuhan peninjauan.',
    ],
    notes: [
      'Label revisi membantu membedakan data versi lama dan versi terbaru.',
      'Jika muncul tanda ada revisi lebih baru, cek versi terkini sebelum memakai data untuk keputusan baru.',
    ],
  },
  {
    title: 'Sebaran Wilayah',
    owner: 'Peninjau wilayah dan portofolio',
    purpose:
      'Menganalisis persebaran proyek menurut provinsi atau kabupaten/kota dengan metrik jumlah proyek maupun nilai pinjaman.',
    steps: [
      'Buka menu Sebaran Wilayah.',
      'Pilih tampilan jumlah proyek atau nilai pinjaman.',
      'Gunakan penyaring tahap proses, status proyek, jenis pinjaman, LoI, indikasi pemberi pinjaman, dan riwayat revisi.',
      'Klik wilayah pada peta untuk melihat daftar proyek di wilayah tersebut.',
      'Telusuri dari provinsi ke kabupaten/kota bila fitur tersebut tersedia.',
    ],
    notes: [
      'Data nasional atau provinsi tidak otomatis digandakan ke kota. Penelusuran ke kota hanya menghitung lokasi kota yang memang tercatat secara jelas.',
      'Daftar proyek mengikuti filter peta dan wilayah fokus.',
    ],
  },
]

const filterGuides = [
  {
    title: 'Dashboard',
    scope: 'Menyaring ringkasan portofolio sebelum membaca angka dan grafik.',
    filters: [
      'Periode perencanaan: pilih satu periode, beberapa periode, atau semua periode.',
      'Pencarian di daftar periode: gunakan saat pilihan periode cukup banyak.',
      'Kartu dan panel analitik: dapat diklik untuk membuka daftar Proyek atau Sebaran Wilayah dengan konteks periode yang sama.',
    ],
    steps: [
      'Buka Dashboard.',
      'Pada bagian kanan atas, klik pilihan periode.',
      'Centang periode yang ingin dianalisis. Jika semua periode dipilih, Dashboard membaca seluruh data portofolio.',
      'Tunggu angka dan panel Dashboard selesai diperbarui.',
      'Klik funnel, kartu, atau panel analitik bila ingin melihat daftar proyek pembentuk angka tersebut.',
    ],
    tips: [
      'Gunakan satu periode untuk membaca kondisi satu siklus perencanaan.',
      'Gunakan beberapa periode untuk membandingkan portofolio lintas periode.',
      'Jika angka tidak sesuai ekspektasi, cek apakah periode yang dipilih sudah benar sebelum menyimpulkan datanya salah.',
    ],
  },
  {
    title: 'Proyek',
    scope: 'Mencari dan menyaring daftar proyek lintas Blue Book, Green Book, Daftar Kegiatan, dan Perjanjian Pinjaman.',
    filters: [
      'Pencarian cepat: nama proyek, pemberi pinjaman, instansi pelaksana, atau program.',
      'Klasifikasi proyek: jenis pinjaman, status proyek, status tahapan, sudah mencapai tahap, dan belum mencapai tahap.',
      'Pemberi pinjaman dan instansi: indikasi pemberi pinjaman, pemberi pinjaman tetap di Green Book, pemberi pinjaman Daftar Kegiatan, pemberi pinjaman Perjanjian Pinjaman, instansi pelaksana, dan instansi pelaksana Daftar Kegiatan.',
      'Program dan wilayah: judul program serta wilayah atau lokasi.',
      'Nilai dan waktu: batas nilai pinjaman minimum/maksimum serta rentang tanggal Daftar Kegiatan.',
      'Riwayat revisi: aktifkan bila perlu melihat data versi lama, bukan hanya versi terbaru.',
    ],
    steps: [
      'Buka menu Proyek.',
      'Gunakan kolom pencarian untuk mencari proyek secara cepat.',
      'Klik tombol Filter untuk membuka Filter lanjutan.',
      'Isi satu atau beberapa pilihan filter sesuai kebutuhan analisis.',
      'Klik Terapkan agar daftar proyek diperbarui.',
      'Perhatikan label Filter aktif di bawah pencarian. Label ini menunjukkan filter apa saja yang sedang digunakan.',
      'Klik tanda silang pada label filter aktif untuk melepas satu filter, atau klik Reset untuk menghapus semua filter.',
      'Gunakan tombol atur kolom bila perlu menampilkan kolom tambahan seperti pemberi pinjaman, lokasi, nilai pinjaman, atau tanggal Daftar Kegiatan.',
    ],
    tips: [
      'Untuk mencari proyek yang berhenti di tahap tertentu, gunakan filter Status Tahapan atau Belum Mencapai Tahap.',
      'Untuk analisis pemberi pinjaman, bedakan indikasi awal, pemberi pinjaman Green Book, pemberi pinjaman Daftar Kegiatan, dan pemberi pinjaman Perjanjian Pinjaman karena masing-masing berasal dari tahap berbeda.',
      'Unduh Excel akan mengikuti filter yang sedang aktif, sehingga hasil unduhan sama dengan konteks daftar yang sedang ditinjau.',
    ],
  },
  {
    title: 'Perjalanan Proyek',
    scope: 'Mencari satu Proyek Blue Book lalu melihat alur dokumennya dari awal sampai perjanjian pinjaman.',
    filters: [
      'Pencarian proyek: masukkan kode Proyek Blue Book atau nama proyek.',
      'Daftar saran: sistem menampilkan proyek yang cocok dengan kata kunci.',
      'Label revisi: membantu memilih versi proyek yang benar bila proyek memiliki revisi.',
      'Tampilan hasil: Ringkasan, Alur, dan Timeline dapat dipilih setelah proyek dibuka.',
    ],
    steps: [
      'Buka menu Perjalanan Proyek.',
      'Klik kotak Cari Proyek Blue Book.',
      'Ketik sebagian kode proyek atau nama proyek.',
      'Pilih proyek dari daftar saran yang muncul. Pastikan kode, nama proyek, dan label revisi sesuai.',
      'Klik Lihat Perjalanan.',
      'Gunakan tab Ringkasan untuk membaca status umum, Alur untuk melihat hubungan antar dokumen, dan Timeline untuk membaca urutan kejadian.',
    ],
    tips: [
      'Perjalanan Proyek memakai pencarian proyek sebagai filter utama, bukan panel filter besar seperti menu Proyek.',
      'Jika muncul tanda ada revisi lebih baru, buka versi terbaru sebelum memakai data untuk keputusan lanjutan.',
      'Jika proyek tidak muncul, coba gunakan kata kunci yang lebih spesifik seperti kode Blue Book atau potongan nama proyek utama.',
    ],
  },
  {
    title: 'Sebaran Wilayah',
    scope: 'Menyaring peta dan daftar proyek berdasarkan wilayah, tahap proses, status proyek, pemberi pinjaman, dan jenis pinjaman.',
    filters: [
      'Metrik peta: Jumlah Proyek atau Nilai Pinjaman.',
      'Tingkat wilayah: provinsi atau kabupaten/kota setelah memilih provinsi.',
      'Pencarian proyek: kata kunci proyek pada daftar wilayah fokus.',
      'Tahap proses: Status Tahapan, Sudah Mencapai Tahap, dan Belum Mencapai Tahap.',
      'Status dan sumber pendanaan: status proyek, LoI, indikasi pemberi pinjaman, dan tipe pinjaman.',
      'Riwayat revisi: aktifkan bila perlu memasukkan data versi lama dalam analisis wilayah.',
    ],
    steps: [
      'Buka menu Sebaran Wilayah.',
      'Pilih metrik peta: Jumlah Proyek untuk melihat banyaknya proyek, atau Nilai Pinjaman untuk melihat besaran pinjaman.',
      'Gunakan kolom pencarian bila ingin mempersempit daftar proyek di wilayah yang sedang dipilih.',
      'Klik tombol Filter untuk membuka Filter lanjutan.',
      'Pilih tahap, status proyek, keberadaan LoI, indikasi pemberi pinjaman, tipe pinjaman, atau riwayat revisi sesuai kebutuhan.',
      'Klik Terapkan agar peta, ringkasan wilayah, dan daftar proyek diperbarui.',
      'Klik provinsi pada peta untuk memfokuskan daftar proyek pada wilayah tersebut.',
      'Klik Lihat Kab/Kota bila ingin menelusuri wilayah sampai kabupaten/kota.',
      'Gunakan Kembali ke Indonesia atau Fokus Indonesia untuk kembali ke cakupan nasional.',
    ],
    tips: [
      'Daftar proyek di bawah peta selalu mengikuti wilayah fokus dan filter yang sedang aktif.',
      'Jika suatu wilayah bernilai nol, cek pilihan metrik dan filter sebelum menyimpulkan tidak ada proyek.',
      'Data tingkat nasional atau provinsi tidak otomatis dipecah menjadi seluruh kabupaten/kota; daftar kabupaten/kota hanya menampilkan data yang memang tercatat pada tingkat tersebut.',
    ],
  },
]

const planningModules = [
  {
    title: 'Blue Book',
    subtitle: 'Padanan aplikasi untuk DRPLN-JM, yaitu daftar rencana kegiatan jangka menengah yang layak dibiayai pinjaman luar negeri.',
    context:
      'Pada tahap ini, PRISM membantu pengguna menata usulan kegiatan sejak awal: siapa pengusul dan pelaksananya, di mana lokasi kegiatan, apa keluaran yang diharapkan, berapa kebutuhan pembiayaan, serta pemberi pinjaman mana yang baru menjadi indikasi. Data pada Blue Book belum berarti pinjaman sudah disetujui, tetapi menjadi dasar penilaian dan penyiapan tahap berikutnya.',
    imageKey: 'blue-book',
    fields: ['Dokumen Blue Book', 'Proyek Blue Book', 'Instansi Pengusul', 'Instansi Pelaksana', 'Lokasi', 'Prioritas Nasional', 'Biaya Proyek', 'Indikasi Pemberi Pinjaman', 'LoI'],
    process: [
      'Buat atau pilih header Blue Book sesuai periode dan status Berlaku atau Tidak Berlaku.',
      'Tambahkan Proyek Blue Book dari detail dokumen.',
      'Lengkapi identitas proyek, instansi, lokasi, prioritas, biaya, indikasi pemberi pinjaman, dan LoI bila sudah tersedia.',
      'Untuk revisi, gunakan impor dari Blue Book lain agar proyek revisi tetap terbaca sebagai kelanjutan dari proyek yang sama.',
    ],
    rules: [
      'Blue Book mengacu pada daftar rencana jangka menengah, sehingga periodenya harus jelas dan sesuai siklus perencanaan.',
      'Kode proyek unik hanya di dalam Blue Book yang sama.',
      'Proyek yang sama pada revisi berbeda tetap dapat ditelusuri melalui riwayat proyek.',
      'Mitra Kerja Bappenas opsional dan boleh lebih dari satu.',
    ],
  },
  {
    title: 'Green Book',
    subtitle: 'Padanan aplikasi untuk DRPPLN, yaitu daftar prioritas tahunan yang sudah memiliki indikasi pendanaan dan kesiapan lebih lanjut.',
    context:
      'Green Book meneruskan usulan dari Blue Book yang sudah diprioritaskan. Di sini pengguna memperjelas kegiatan, sumber pendanaan, rencana pencairan, dan alokasi pendanaan. Hubungan ke Proyek Blue Book dijaga agar pembaca tetap dapat melihat asal usulan sebelum proyek masuk ke tahap Daftar Kegiatan.',
    imageKey: 'green-book',
    fields: ['Dokumen Green Book', 'Proyek Green Book', 'Relasi Proyek Blue Book', 'Kegiatan', 'Sumber Pendanaan', 'Rencana Pencairan', 'Alokasi Pendanaan'],
    process: [
      'Buat atau pilih dokumen Green Book berdasarkan tahun terbit dan nomor revisi.',
      'Tambahkan Proyek Green Book dan hubungkan ke minimal satu Proyek Blue Book.',
      'Isi kegiatan, sumber pendanaan, rencana pencairan, dan alokasi pendanaan per kegiatan.',
      'Jika membawa proyek dari dokumen lain, gunakan aksi Tambahkan Proyek dari Green Book Lain pada detail Green Book.',
    ],
    rules: [
      'Green Book berangkat dari rencana yang sudah tercantum dalam Blue Book dan sudah lebih siap untuk pendanaan tahunan.',
      'Proyek Green Book wajib terhubung dengan minimal satu Proyek Blue Book.',
      'Relasi ke Proyek Blue Book memakai versi terbaru saat proyek dibuat.',
      'Alokasi pendanaan selalu mengikuti daftar kegiatan.',
    ],
  },
  {
    title: 'Daftar Kegiatan',
    subtitle: 'Daftar rencana kegiatan yang siap diusulkan atau dirundingkan dengan calon pemberi pinjaman.',
    context:
      'Daftar Kegiatan menandai bahwa proyek dari Green Book sudah masuk daftar yang lebih siap untuk proses pembiayaan. PRISM mencatat surat, proyek di dalam surat, rincian pembiayaan, alokasi, dan rincian kegiatan agar data yang dibawa ke perjanjian pinjaman tetap jelas sumbernya.',
    imageKey: 'daftar-kegiatan',
    fields: ['Dokumen surat', 'Proyek Daftar Kegiatan', 'Relasi Proyek Green Book', 'Rincian Pembiayaan', 'Alokasi Pinjaman', 'Rincian Kegiatan'],
    process: [
      'Buat atau pilih header Daftar Kegiatan.',
      'Saat menambah proyek, pilih Proyek Green Book terlebih dahulu agar data dasar bisa terisi otomatis.',
      'Periksa ulang nama proyek, mitra, pembiayaan multi-mata uang, alokasi, dan rincian kegiatan.',
      'Simpan setelah data pembiayaan dan relasi proyek sudah konsisten.',
    ],
    rules: [
      'Daftar Kegiatan dipakai untuk kegiatan yang sudah tercantum dalam Green Book dan siap masuk proses usulan atau perundingan.',
      'Daftar Kegiatan final setelah diterbitkan dan perubahan dibatasi untuk ADMIN.',
      'Hubungan data memakai versi terbaru Proyek Green Book saat Daftar Kegiatan dibuat.',
      'Tahapan setelahnya tetap memakai data proyek yang dipilih saat Daftar Kegiatan dibuat dan tidak otomatis berpindah ketika ada revisi baru.',
    ],
  },
  {
    title: 'Perjanjian Pinjaman',
    subtitle: 'Kesepakatan tertulis yang mengikat Pemerintah dan pemberi pinjaman luar negeri setelah proses perundingan.',
    context:
      'Perjanjian Pinjaman adalah tahap ketika komitmen pembiayaan sudah memiliki dasar tertulis. Aplikasi mencatat identitas pinjaman, pemberi pinjaman, tanggal perjanjian, tanggal efektif, tanggal penutupan, nilai komitmen, pembagian alokasi ke Proyek Daftar Kegiatan, realisasi kumulatif, serta indikator kinerja untuk pemantauan.',
    imageKey: 'loan-agreements',
    fields: ['Kode pinjaman', 'Tanggal perjanjian', 'Tanggal efektif', 'Tanggal penutupan', 'Mata uang', 'Nilai pinjaman utama', 'Realisasi kumulatif', 'Alokasi per Proyek Daftar Kegiatan'],
    process: [
      'Buka menu Perjanjian Pinjaman dan klik buat baru.',
      'Pilih satu atau lebih Proyek Daftar Kegiatan yang sudah memenuhi syarat.',
      'Isi pemberi pinjaman, mata uang, nilai pinjaman, tanggal penting, dan alokasi komitmen per proyek.',
      'Pastikan total alokasi per proyek sama dengan nilai pinjaman utama.',
    ],
    rules: [
      'Perjanjian Pinjaman perlu menjelaskan jumlah, peruntukan, hak dan kewajiban, serta ketentuan dan persyaratan pinjaman.',
      'Kode pinjaman harus unik dan tidak boleh sama dengan Perjanjian Pinjaman lain.',
      'Pemberi pinjaman pada Perjanjian Pinjaman harus berasal dari rincian pembiayaan Proyek Daftar Kegiatan yang dipilih.',
      'Realisasi kumulatif diisi manual mengikuti mata uang Perjanjian Pinjaman.',
      'Status kinerja dihitung dari Kurs Tengah BI terbaru dan tanggal efektif atau closing.',
    ],
  },
]

const masterData = [
  ['Negara', 'Referensi negara untuk pemberi pinjaman Bilateral dan kebutuhan referensi lain.'],
  ['Pemberi Pinjaman', 'Daftar pemberi pinjaman dengan tipe Bilateral, Multilateral, atau KSA. Bilateral wajib memiliki negara, Multilateral tidak memakai negara, dan KSA bersifat opsional.'],
  ['Mata Uang', 'Daftar mata uang resmi yang dapat dipakai pada Green Book, Daftar Kegiatan, dan Perjanjian Pinjaman.'],
  ['Kurs Tengah BI', 'Nilai kurs per mata uang dan tanggal acuan, dipakai untuk tampilan konversi dan indikator Perjanjian Pinjaman.'],
  ['Instansi', 'Hierarki kementerian/lembaga, eselon, BUMN, pemerintah daerah, dan entitas lain.'],
  ['Wilayah', 'Hierarki nasional, provinsi, dan kabupaten/kota untuk lokasi proyek dan analitik spasial.'],
  ['Judul Program', 'Referensi program untuk klasifikasi proyek.'],
  ['Mitra Kerja Bappenas', 'Unit mitra Bappenas yang dapat dikaitkan dengan proyek.'],
  ['Periode', 'Periode perencanaan Blue Book dan referensi prioritas.'],
  ['Prioritas Nasional', 'Master prioritas nasional yang bisa dikaitkan dengan proyek.'],
]

const importKinds = [
  ['Master Data', 'Judul program, instansi, wilayah, periode, prioritas, pemberi pinjaman, mata uang, dan kurs tengah.'],
  ['Blue Book', 'Beberapa Blue Book sekaligus beserta proyek, instansi, lokasi, biaya, dan indikasi pemberi pinjaman.'],
  ['Green Book', 'Proyek Green Book, relasi Proyek Blue Book, kegiatan, sumber pendanaan, dan alokasi.'],
  ['Daftar Kegiatan', 'Dokumen surat, proyek, relasi Proyek Green Book, pembiayaan, alokasi, dan rincian kegiatan.'],
  ['Perjanjian Pinjaman', 'Membuat Perjanjian Pinjaman dari Proyek Daftar Kegiatan yang sudah memenuhi syarat.'],
]

const tocEntries = [
  ['section-orientasi', 'Orientasi', 'Pengantar, tujuan aplikasi, dan urutan kerja utama.'],
  ['section-daftar-isi', 'Daftar Isi', 'Peta isi dokumen yang bisa diklik.'],
  ['section-akses', 'Akses dan navigasi', 'Login, sidebar, pencarian menu, dan peran pengguna.'],
  ['section-dashboard', 'Dashboard', 'Membaca funnel, panel analitik, dan filter periode.'],
  ['section-detail-dashboard', 'Detail Dashboard', 'Cara membaca setiap panel dashboard.'],
  ['section-modul-peninjauan', 'Modul Peninjauan', 'Proyek, Perjalanan Proyek, dan Sebaran Wilayah.'],
  ['section-proyek-ekspor', 'Proyek dan Ekspor', 'Tabel, filter aktif, aksi baris, dan hasil Excel.'],
  ['section-penggunaan-filter', 'Penggunaan Filter', 'Cara memakai filter pada Dashboard dan Proyek.'],
  ['section-filter-lanjutan', 'Filter Lanjutan Peninjauan', 'Pencarian Perjalanan Proyek dan filter Sebaran Wilayah.'],
  ['section-tampilan-peninjauan', 'Tampilan Peninjauan', 'Contoh layar Proyek dan Sebaran Wilayah.'],
  ['section-master-data', 'Master Data', 'Referensi dasar sebelum pengisian dokumen.'],
  ['section-panduan-referensi', 'Panduan Referensi', 'Cara memakai setiap master data dengan bahasa operasional.'],
  ['section-dokumen-perencanaan', 'Dokumen Perencanaan', 'Padanan istilah hukum dan alur Blue Book sampai Perjanjian Pinjaman.'],
  ['section-panduan-formulir', 'Panduan Isi Formulir', 'Bagian formulir penting pada setiap dokumen.'],
  ['section-contoh-layar-dokumen', 'Contoh Layar Dokumen', 'Contoh tampilan daftar dokumen perencanaan.'],
  ['section-blue-book', 'Blue Book', 'Dokumen, proyek, LoI, revisi, dan riwayat proyek.'],
  ['section-green-book', 'Green Book', 'Relasi Blue Book, kegiatan, sumber pendanaan, dan alokasi.'],
  ['section-daftar-kegiatan-perjanjian', 'Daftar Kegiatan dan Perjanjian Pinjaman', 'Pembiayaan, rincian kegiatan, kontrak, kurs, dan indikator kinerja.'],
  ['section-alur-operasional', 'Alur Operasional Khusus', 'Impor antar dokumen, kurs massal, draft perjanjian, kolom tabel, dan hierarki.'],
  ['section-pesan-sistem', 'Pesan Sistem dan Kondisi Layar', 'Akses ditolak, data kosong, validasi formulir, konfirmasi hapus, dan hasil preview impor.'],
  ['section-riwayat-revisi', 'Riwayat Revisi', 'Membedakan data lama, data terbaru, dan jejak perjalanan proyek.'],
  ['section-impor-data', 'Impor Data', 'Template, preview, eksekusi, dan perbaikan error.'],
  ['section-detail-file-excel', 'Detail File Excel', 'Sheet penting dan pemeriksaan per jenis impor.'],
  ['section-admin', 'Admin', 'Manajemen pengguna, hak akses, dan checklist operasional.'],
  ['section-hak-akses-detail', 'Hak Akses Detail', 'Alur pengguna dan prinsip pemberian akses.'],
  ['section-troubleshooting', 'Troubleshooting', 'Masalah umum dan langkah perbaikannya.'],
]

const sectionPageMarkers = [
  ['section-orientasi', '01 - Orientasi'],
  ['section-daftar-isi', '02 - Daftar Isi'],
  ['section-akses', '03 - Akses'],
  ['section-dashboard', '04 - Dashboard dan Analitik'],
  ['section-detail-dashboard', '05 - Detail Dashboard'],
  ['section-modul-peninjauan', '06 - Modul Peninjauan'],
  ['section-proyek-ekspor', '07 - Proyek dan Ekspor'],
  ['section-penggunaan-filter', '08 - Penggunaan Filter'],
  ['section-filter-lanjutan', '09 - Filter Lanjutan Peninjauan'],
  ['section-tampilan-peninjauan', '10 - Tampilan Peninjauan'],
  ['section-master-data', '11 - Master Data'],
  ['section-panduan-referensi', '12 - Panduan Referensi'],
  ['section-dokumen-perencanaan', '13 - Dokumen Perencanaan'],
  ['section-panduan-formulir', '14 - Panduan Isi Formulir'],
  ['section-contoh-layar-dokumen', '15 - Contoh Layar Dokumen'],
  ['section-blue-book', '16 - Mengelola Blue Book'],
  ['section-green-book', '17 - Mengelola Green Book'],
  ['section-daftar-kegiatan-perjanjian', '18 - Mengelola Daftar Kegiatan dan Perjanjian Pinjaman'],
  ['section-alur-operasional', '19 - Alur Operasional Khusus'],
  ['section-pesan-sistem', '20 - Pesan Sistem dan Kondisi Layar'],
  ['section-riwayat-revisi', '21 - Riwayat Revisi'],
  ['section-impor-data', '22 - Impor Data'],
  ['section-detail-file-excel', '23 - Detail File Excel Impor'],
  ['section-admin', '24 - Admin'],
  ['section-hak-akses-detail', '25 - Hak Akses Detail'],
  ['section-troubleshooting', '26 - Troubleshooting'],
].map(([id, marker]) => ({ id, marker }))

const legalReferenceRows = [
  [
    'Pinjaman Luar Negeri',
    'Pembiayaan melalui utang dari pemberi pinjaman luar negeri yang dituangkan dalam perjanjian pinjaman dan wajib dibayar kembali sesuai persyaratan.',
    'PRISM mencatat perjalanan pinjaman kegiatan dari rencana proyek sampai perjanjian dan indikator pelaksanaan.',
  ],
  [
    'Blue Book / DRPLN-JM',
    'Daftar Rencana Pinjaman Luar Negeri Jangka Menengah berisi rencana kegiatan yang layak dibiayai pinjaman luar negeri untuk periode jangka menengah.',
    'Di aplikasi, Blue Book menyimpan dokumen per periode, revisi, Proyek Blue Book, instansi, lokasi, biaya, indikasi pemberi pinjaman, dan LoI.',
  ],
  [
    'Green Book / DRPPLN',
    'Daftar Rencana Prioritas Pinjaman Luar Negeri berisi rencana kegiatan yang telah memiliki indikasi pendanaan dan siap dibiayai untuk jangka tahunan.',
    'Di aplikasi, Green Book menyimpan prioritas pendanaan tahunan, relasi ke Proyek Blue Book, kegiatan, sumber pendanaan, rencana pencairan, dan alokasi.',
  ],
  [
    'Daftar Kegiatan',
    'Daftar rencana kegiatan yang sudah tercantum dalam DRPPLN dan siap diusulkan atau dirundingkan dengan calon pemberi pinjaman.',
    'Di aplikasi, Daftar Kegiatan menyimpan surat, Proyek Daftar Kegiatan, relasi ke Proyek Green Book, rincian pembiayaan, alokasi, dan rincian kegiatan.',
  ],
  [
    'Perjanjian Pinjaman',
    'Kesepakatan tertulis hasil perundingan antara Pemerintah dan pemberi pinjaman luar negeri, paling sedikit memuat jumlah, peruntukan, hak dan kewajiban, serta ketentuan dan persyaratan.',
    'Di aplikasi, Perjanjian Pinjaman menyimpan kode pinjaman, pemberi pinjaman, tanggal penting, mata uang, nilai komitmen, alokasi per proyek, realisasi kumulatif, dan indikator kinerja.',
  ],
  [
    'Instansi Pengusul dan Instansi Pelaksana',
    'Instansi pengusul mengajukan usulan kegiatan, sedangkan instansi pelaksana melaksanakan kegiatan yang dibiayai pinjaman luar negeri.',
    'Di aplikasi, kedua peran ini dicatat pada Proyek Blue Book, Green Book, dan Daftar Kegiatan agar tanggung jawab perencanaan dan pelaksanaan tetap jelas.',
  ],
]

const dashboardDetailPanels = [
  {
    title: 'Funnel Tahapan',
    focus: 'Membandingkan jumlah proyek dari Blue Book sampai Perjanjian Pinjaman.',
    usage: [
      'Baca dari kiri ke kanan sesuai urutan proses bisnis.',
      'Perhatikan selisih jumlah antar tahap untuk menemukan proyek yang belum lanjut.',
      'Klik tahap yang ingin diperiksa untuk membuka daftar proyek terkait bila tautan tersedia.',
    ],
  },
  {
    title: 'Ringkasan Nilai',
    focus: 'Melihat gambaran nilai pinjaman dan nilai pembiayaan portofolio.',
    usage: [
      'Gunakan sebagai indikator awal, bukan pengganti pemeriksaan detail proyek.',
      'Pastikan filter periode sudah sesuai sebelum membandingkan angka.',
      'Jika nilai terlihat kosong, cek apakah proyek sudah memiliki nilai pada tahap yang dibaca.',
    ],
  },
  {
    title: 'Sebaran Pemberi Pinjaman',
    focus: 'Melihat pemberi pinjaman yang paling sering muncul pada portofolio.',
    usage: [
      'Bedakan pemberi pinjaman awal, pemberi pinjaman Green Book, pemberi pinjaman Daftar Kegiatan, dan pemberi pinjaman Perjanjian Pinjaman.',
      'Gunakan panel ini untuk melihat konsentrasi pendanaan pada pihak tertentu.',
      'Lanjutkan ke halaman Proyek bila perlu melihat proyek pembentuk angka tersebut.',
    ],
  },
  {
    title: 'Sebaran Instansi',
    focus: 'Membaca instansi pengusul atau pelaksana yang paling banyak terlibat.',
    usage: [
      'Gunakan untuk melihat beban portofolio per instansi.',
      'Cocok dipakai sebelum melakukan validasi data instansi di detail proyek.',
      'Jika nama instansi terasa dobel, periksa Master Instansi karena kemungkinan ada penulisan yang belum seragam.',
    ],
  },
  {
    title: 'Sebaran Wilayah',
    focus: 'Membaca konsentrasi proyek menurut lokasi yang dicatat pada proyek.',
    usage: [
      'Gunakan untuk mencari wilayah dengan jumlah proyek atau nilai terbesar.',
      'Klik tautan lanjutan ke Sebaran Wilayah bila perlu membaca peta.',
      'Ingat bahwa data nasional atau provinsi tidak otomatis dipecah menjadi seluruh kabupaten/kota.',
    ],
  },
  {
    title: 'Kondisi Perjanjian Pinjaman',
    focus: 'Melihat pinjaman yang mendekati penutupan, diperpanjang, atau memiliki indikator kinerja tertentu.',
    usage: [
      'Gunakan sebagai daftar prioritas pemeriksaan kontrak.',
      'Buka detail Perjanjian Pinjaman untuk melihat tanggal efektif, tanggal penutupan, dan realisasi kumulatif.',
      'Pastikan Kurs Tengah BI terbaru sudah tersedia agar indikator tampilan terbaca wajar.',
    ],
  },
]

const projectMasterDetails = [
  ['Pencarian cepat', 'Mencari proyek berdasarkan nama, kode, pemberi pinjaman, instansi, program, atau kata kunci lain.', 'Gunakan ketika sudah mengetahui kata kunci. Untuk hasil yang lebih spesifik, kombinasikan dengan Filter.'],
  ['Filter lanjutan', 'Membatasi daftar berdasarkan tahap proses, status proyek, jenis pinjaman, pemberi pinjaman, instansi, wilayah, nilai, tanggal, dan riwayat revisi.', 'Klik Terapkan setelah memilih filter. Gunakan Reset bila ingin kembali ke seluruh data.'],
  ['Label filter aktif', 'Menunjukkan filter yang sedang digunakan pada daftar proyek.', 'Hapus satu label bila hanya ingin melepas satu filter tanpa mengulang seluruh pencarian.'],
  ['Atur kolom', 'Menampilkan atau menyembunyikan kolom agar tabel sesuai kebutuhan peninjauan.', 'Tampilkan kolom pemberi pinjaman, lokasi, nilai, dan tanggal bila sedang menyiapkan bahan rapat atau ekspor.'],
  ['Aksi baris', 'Membuka detail proyek, melihat perjalanan proyek, atau masuk ke modul terkait.', 'Gunakan aksi ini untuk berpindah dari daftar gabungan ke dokumen sumber.'],
  ['Unduh Excel', 'Mengunduh data sesuai pencarian dan filter yang sedang aktif.', 'Sebelum unduh, cek label filter aktif agar file Excel tidak berisi cakupan yang salah.'],
]

const masterDataGuides = [
  {
    title: 'Negara',
    purpose: 'Menjadi referensi negara untuk pemberi pinjaman Bilateral.',
    fill: ['Isi nama negara secara resmi dan konsisten.', 'Gunakan satu baris untuk satu negara.', 'Periksa data lama sebelum menambah negara baru.'],
    caution: 'Negara yang sudah dipakai oleh pemberi pinjaman sebaiknya tidak dihapus sembarangan.',
  },
  {
    title: 'Pemberi Pinjaman',
    purpose: 'Menjadi daftar pemberi pinjaman yang dipilih pada proyek, pembiayaan, dan Perjanjian Pinjaman.',
    fill: ['Isi nama pemberi pinjaman dan nama singkat bila ada.', 'Pilih tipe Bilateral, Multilateral, atau KSA.', 'Untuk Bilateral, pilih negara. Untuk Multilateral, negara tidak perlu diisi. KSA boleh tanpa negara.'],
    caution: 'Kesalahan tipe pemberi pinjaman dapat membuat impor atau formulir pembiayaan gagal diperiksa.',
  },
  {
    title: 'Mata Uang',
    purpose: 'Menentukan mata uang yang dapat dipilih pada Green Book, Daftar Kegiatan, dan Perjanjian Pinjaman.',
    fill: ['Gunakan kode mata uang resmi seperti USD, EUR, JPY, atau IDR.', 'Aktifkan hanya mata uang yang boleh dipakai operator.', 'Nonaktifkan mata uang yang hanya menjadi referensi.'],
    caution: 'Mata uang nonaktif tidak muncul pada pilihan pengisian baru.',
  },
  {
    title: 'Kurs Tengah BI',
    purpose: 'Dipakai untuk tampilan nilai USD dan indikator kinerja Perjanjian Pinjaman.',
    fill: ['Pilih mata uang.', 'Isi kurs dan tanggal cut off.', 'Pastikan satu mata uang tidak memiliki dua kurs pada tanggal cut off yang sama.'],
    caution: 'Jika kurs terbaru belum tersedia, nilai tampilan dan indikator Perjanjian Pinjaman bisa tidak terbaca.',
  },
  {
    title: 'Instansi',
    purpose: 'Dipakai sebagai instansi pengusul, instansi pelaksana, dan pihak terkait proyek.',
    fill: ['Isi nama instansi sesuai hierarki.', 'Pilih tingkat yang benar, misalnya kementerian, eselon, BUMN, atau pemerintah daerah.', 'Isi induk organisasi untuk unit yang berada di bawah instansi lain.'],
    caution: 'Nama yang sama pada induk berbeda bisa saja terjadi, tetapi saat impor lebih aman memakai rujukan yang tidak ambigu.',
  },
  {
    title: 'Wilayah',
    purpose: 'Dipakai sebagai lokasi proyek dan dasar analisis Sebaran Wilayah.',
    fill: ['Isi nasional, provinsi, dan kabupaten/kota sesuai hirarki.', 'Pastikan parent wilayah benar.', 'Gunakan kode wilayah yang konsisten.'],
    caution: 'Memilih nasional berarti proyek berlaku nasional; jangan menduplikasi ke semua provinsi bila sumber datanya memang nasional.',
  },
  {
    title: 'Judul Program',
    purpose: 'Mengelompokkan proyek berdasarkan program.',
    fill: ['Isi nama program utama.', 'Gunakan parent bila ada struktur program bertingkat.', 'Hindari variasi penulisan untuk program yang sama.'],
    caution: 'Program yang tidak konsisten membuat filter dan laporan sulit dibaca.',
  },
  {
    title: 'Mitra Kerja Bappenas',
    purpose: 'Mencatat unit Bappenas yang terkait dengan proyek.',
    fill: ['Isi unit eselon sesuai daftar organisasi.', 'Eselon II dapat berada di bawah Eselon I.', 'Pilih lebih dari satu mitra pada proyek bila memang ada beberapa unit terkait.'],
    caution: 'Pada proyek, sistem menyimpan unit yang dipilih; pastikan unit tersebut sudah benar sebelum dipakai luas.',
  },
  {
    title: 'Periode',
    purpose: 'Menjadi acuan periode perencanaan Blue Book dan prioritas nasional.',
    fill: ['Isi nama periode dan rentang tahun.', 'Gunakan format nama yang mudah dikenali pengguna.', 'Pastikan periode belum ada sebelum menambah.'],
    caution: 'Periode yang salah akan membuat Blue Book dan prioritas nasional sulit dicocokkan.',
  },
  {
    title: 'Prioritas Nasional',
    purpose: 'Menandai keterkaitan proyek dengan prioritas nasional.',
    fill: ['Pilih periode referensi.', 'Isi nama prioritas sesuai dokumen sumber.', 'Gunakan prioritas dari periode mana pun bila memang perlu dikaitkan pada Proyek Blue Book.'],
    caution: 'Pastikan prioritas tidak tertukar dengan program title; keduanya dipakai untuk kebutuhan berbeda.',
  },
]

const formFieldGuides = [
  {
    title: 'Blue Book dan Proyek Blue Book',
    rows: [
      ['Kedudukan dokumen', 'Blue Book dipakai sebagai daftar rencana jangka menengah yang sepadan dengan DRPLN-JM pada dasar hukum.', 'Gunakan bagian ini sebagai tahap kelayakan awal, bukan sebagai bukti pinjaman sudah pasti.'],
      ['Header Blue Book', 'Periode, tanggal terbit, nomor revisi, tahun revisi, status Berlaku atau Tidak Berlaku.', 'Gunakan status Berlaku hanya untuk dokumen yang masih menjadi acuan.'],
      ['Identitas proyek', 'Kode proyek, nama proyek, judul program, durasi, tujuan, ruang lingkup, keluaran, dan manfaat.', 'Kode proyek unik di dalam Blue Book yang sama. Kode boleh muncul lagi pada revisi lain bila itu proyek yang sama.'],
      ['Pihak terkait', 'Instansi pengusul, instansi pelaksana, dan Mitra Kerja Bappenas.', 'Satu instansi bisa menjadi pengusul sekaligus pelaksana bila memang sesuai dokumen sumber.'],
      ['Lokasi dan prioritas', 'Lokasi nasional, provinsi, kabupaten/kota, serta prioritas nasional.', 'Jika lokasi nasional dipilih, jangan menambah semua provinsi hanya untuk menggandakan cakupan nasional.'],
      ['Biaya proyek', 'Jenis pendanaan, kategori pendanaan, dan nilai biaya.', 'Periksa angka sebelum menyimpan karena data ini menjadi dasar peninjauan tahap awal.'],
      ['Indikasi pemberi pinjaman dan LoI', 'Calon pemberi pinjaman, jenis pinjaman, catatan, surat minat, tanggal surat, dan nomor surat.', 'Indikasi pemberi pinjaman belum berarti pinjaman disetujui; LoI menunjukkan minat tertulis.'],
    ],
  },
  {
    title: 'Green Book dan Proyek Green Book',
    rows: [
      ['Kedudukan dokumen', 'Green Book dipakai sebagai daftar prioritas pinjaman tahunan yang sepadan dengan DRPPLN pada dasar hukum.', 'Gunakan bagian ini untuk proyek yang sudah memiliki indikasi pendanaan dan kesiapan lebih lanjut.'],
      ['Header Green Book', 'Tahun terbit, nomor revisi, dan status dokumen.', 'Satu kombinasi tahun terbit dan revisi tidak boleh dibuat dua kali.'],
      ['Relasi Proyek Blue Book', 'Pilih minimal satu Proyek Blue Book sebagai asal proyek.', 'Jika memilih lebih dari satu Proyek Blue Book, semuanya harus berasal dari dokumen Blue Book yang sama.'],
      ['Identitas dan pihak proyek', 'Kode, nama proyek, instansi, lokasi, mitra Bappenas, durasi, dan informasi proyek.', 'Saat revisi, sistem menjaga hubungan proyek lama dan baru agar riwayat tetap terbaca.'],
      ['Kegiatan', 'Daftar kegiatan proyek Green Book dan urutannya.', 'Alokasi pendanaan mengikuti daftar kegiatan; periksa kembali setelah menambah atau menghapus kegiatan.'],
      ['Sumber pendanaan', 'Pemberi pinjaman, mata uang, nilai sesuai dokumen, dan nilai USD.', 'Untuk USD, nilai sesuai dokumen dan nilai USD sama. Untuk mata uang lain, isi nilai sesuai dokumen dan aturan yang berlaku.'],
      ['Rencana pencairan dan alokasi', 'Rencana pencairan per tahun serta pembagian pendanaan ke kegiatan.', 'Rencana pencairan dibaca sebagai total proyek per tahun, bukan per pemberi pinjaman.'],
    ],
  },
  {
    title: 'Daftar Kegiatan dan Proyek Daftar Kegiatan',
    rows: [
      ['Kedudukan dokumen', 'Daftar Kegiatan berisi kegiatan yang sudah tercantum dalam Green Book dan siap diusulkan atau dirundingkan.', 'Data pada tahap ini menjadi dasar sebelum masuk Perjanjian Pinjaman.'],
      ['Header surat', 'Perihal, tanggal surat, nomor surat, dan informasi surat lainnya.', 'Nomor surat membantu membedakan dokumen dan mencegah duplikasi.'],
      ['Pemilihan Proyek Green Book', 'Pilih Proyek Green Book terlebih dahulu sebelum mengisi detail proyek.', 'Data dasar seperti nama proyek, instansi, lokasi, dan mitra dapat terisi otomatis, tetapi tetap perlu diperiksa.'],
      ['Identitas proyek', 'Nama proyek Daftar Kegiatan, durasi, tujuan, instansi, lokasi, dan mitra.', 'Nama proyek boleh disesuaikan dari Green Book karena Daftar Kegiatan menyimpan catatan sesuai surat.'],
      ['Rincian pembiayaan', 'Pemberi pinjaman, mata uang, nilai sesuai dokumen, nilai USD, dan informasi pembiayaan.', 'Pemberi pinjaman dipilih dari Master Pemberi Pinjaman dan boleh berbeda dari indikasi awal pada Blue Book.'],
      ['Alokasi pinjaman', 'Pembagian pembiayaan pada proyek Daftar Kegiatan.', 'Pastikan angka sesuai dokumen sumber sebelum proyek masuk Perjanjian Pinjaman.'],
      ['Rincian kegiatan', 'Nomor kegiatan dan nama kegiatan pada tahap Daftar Kegiatan.', 'Bagian ini berupa catatan bebas dan tidak harus sama persis dengan kegiatan di Green Book.'],
    ],
  },
  {
    title: 'Perjanjian Pinjaman',
    rows: [
      ['Kedudukan dokumen', 'Perjanjian Pinjaman adalah kesepakatan tertulis hasil perundingan dengan pemberi pinjaman luar negeri.', 'Gunakan bagian ini untuk data yang sudah masuk komitmen legal, bukan lagi indikasi awal.'],
      ['Proyek terkait', 'Pilih satu atau lebih Proyek Daftar Kegiatan.', 'Satu Perjanjian Pinjaman bisa mencakup proyek dari beberapa surat Daftar Kegiatan.'],
      ['Pemberi pinjaman', 'Pilih pemberi pinjaman Perjanjian Pinjaman.', 'Pemberi pinjaman harus berasal dari rincian pembiayaan semua Proyek Daftar Kegiatan yang dipilih.'],
      ['Identitas pinjaman', 'Kode pinjaman, tanggal perjanjian, tanggal efektif, tanggal penutupan awal, dan tanggal penutupan terbaru.', 'Kode pinjaman harus unik. Tanggal penutupan terbaru tidak boleh lebih awal dari tanggal penutupan awal bila tanggal awal diisi.'],
      ['Nilai dan mata uang', 'Mata uang, nilai pinjaman utama, dan realisasi kumulatif.', 'Realisasi kumulatif diisi manual mengikuti mata uang Perjanjian Pinjaman.'],
      ['Alokasi per proyek', 'Nilai alokasi pinjaman untuk setiap Proyek Daftar Kegiatan.', 'Total alokasi harus sama dengan nilai pinjaman utama.'],
      ['Indikator tampilan', 'Status perpanjangan, jumlah hari perpanjangan, nilai USD tampilan, dan status kinerja.', 'Indikator ini dibaca dari tanggal dan Kurs Tengah BI terbaru; gunakan sebagai alat bantu peninjauan.'],
    ],
  },
]

const revisionHistoryGuides = [
  ['Blue Book', 'Setiap revisi adalah dokumen sendiri. Proyek yang sama pada revisi baru dicatat sebagai salinan baru, tetapi tetap dihubungkan sebagai riwayat proyek yang sama.', 'Saat membuat revisi, gunakan fitur impor proyek dari Blue Book lain agar hubungan riwayat tidak terputus.'],
  ['Green Book', 'Setiap revisi Green Book menyimpan daftar proyek versi dokumen tersebut. Kode proyek yang sama dapat muncul lagi pada revisi baru.', 'Saat proyek Green Book dibuat dari Proyek Blue Book, sistem memakai versi Proyek Blue Book terbaru yang sesuai.'],
  ['Daftar Kegiatan', 'Daftar Kegiatan bersifat final setelah diterbitkan. Proyek di dalamnya tetap menunjuk data Green Book yang dipilih saat dibuat.', 'Jika ada revisi Blue Book atau Green Book setelah Daftar Kegiatan dibuat, data Daftar Kegiatan tidak otomatis berpindah.'],
  ['Perjanjian Pinjaman', 'Perjanjian Pinjaman membaca Proyek Daftar Kegiatan yang dipilih saat perjanjian dibuat.', 'Perubahan tanggal penutupan dapat mengubah informasi perpanjangan dan indikator kinerja tampilan.'],
  ['Perjalanan Proyek', 'Halaman ini dipakai untuk membaca jejak dokumen satu proyek dari awal sampai Perjanjian Pinjaman.', 'Jika ada tanda revisi lebih baru, periksa versi terbaru sebelum menggunakan data lama sebagai dasar keputusan baru.'],
]

const importWorkbookGuides = [
  {
    title: 'Master Data',
    sheets: 'Program Titles, Bappenas Partners, Institutions, Regions, Periods, National Priorities, Lenders, Kurs Tengah.',
    before: ['Rapikan penulisan nama referensi.', 'Pastikan data induk diisi sebelum data anak bila memakai hierarki.', 'Gunakan tipe pemberi pinjaman dan tingkat wilayah yang sesuai daftar pilihan.'],
    preview: ['Status create berarti baris siap ditambahkan.', 'Status skip berarti data sudah ada atau tidak perlu dibuat ulang.', 'Status failed berarti file Excel harus diperbaiki sebelum eksekusi.'],
  },
  {
    title: 'Blue Book',
    sheets: 'Blue Book, Input Data, Relasi - EA, Relasi - IA, Relasi - Locations, Relasi - National Priority, Relasi - Project Cost, Relasi - Lender Indication.',
    before: ['Isi Blue Book Key sebagai kunci sementara untuk menghubungkan semua sheet.', 'Gunakan kode Blue Book yang sama pada sheet relasi untuk mengaitkan data ke proyek.', 'Siapkan Master Instansi, Wilayah, Prioritas Nasional, Program, dan Pemberi Pinjaman terlebih dahulu.'],
    preview: ['Cek baris gagal pada relasi instansi, lokasi, biaya, dan indikasi pemberi pinjaman.', 'Jika kode proyek duplikat di Blue Book yang sama, perbaiki sebelum eksekusi.', 'Impor multi Blue Book dapat membuat beberapa dokumen sekaligus dari satu file Excel.'],
  },
  {
    title: 'Green Book',
    sheets: 'Green Book, Input Data, Relasi - Blue Book Project, Relasi - EA, Relasi - IA, Relasi - Locations, Relasi - Activities, Relasi - Funding Source, Relasi - Disbursement Plan, Relasi - Funding Allocation.',
    before: ['Isi Green Book Key sebagai kunci sementara dokumen.', 'Pastikan Proyek Blue Book yang dirujuk sudah ada dan dipilih dari referensi yang benar.', 'Siapkan mata uang, pemberi pinjaman, instansi, wilayah, dan daftar kegiatan proyek.'],
    preview: ['Relasi Proyek Blue Book wajib valid.', 'Alokasi pendanaan harus selaras dengan daftar kegiatan.', 'Rencana pencairan tidak boleh memiliki tahun yang duplikat untuk proyek yang sama.'],
  },
  {
    title: 'Daftar Kegiatan',
    sheets: 'Daftar Kegiatan, Input Data, Relasi - Green Book Project, Relasi - Locations, Relasi - Financing Detail, Relasi - Loan Allocation, Relasi - Activity Detail.',
    before: ['Isi Daftar Kegiatan Key sebagai kunci sementara surat.', 'Isi nomor surat, perihal, dan tanggal sesuai dokumen surat.', 'Pilih Proyek Green Book yang menjadi dasar proyek Daftar Kegiatan.'],
    preview: ['Nomor surat duplikat akan ditandai agar tidak membuat surat ganda.', 'Rincian pembiayaan boleh memilih pemberi pinjaman dari Master Pemberi Pinjaman.', 'Rincian kegiatan adalah catatan bebas dan tidak harus sama dengan Green Book.'],
  },
  {
    title: 'Perjanjian Pinjaman',
    sheets: 'Loan Agreement dan Relasi - Daftar Kegiatan Project bila satu pinjaman mencakup beberapa Proyek Daftar Kegiatan.',
    before: ['Isi Loan Code sebagai kunci utama impor.', 'Pilih referensi Proyek Daftar Kegiatan dari data master agar tidak salah proyek.', 'Pastikan pemberi pinjaman berasal dari rincian pembiayaan Proyek Daftar Kegiatan yang dipilih.'],
    preview: ['Loan Code harus unik.', 'Total alokasi pada relasi proyek harus sama dengan nilai pinjaman utama.', 'Untuk mata uang selain USD, pastikan Kurs Tengah BI terbaru tersedia.'],
  },
]

const accessManagementGuides = [
  {
    title: 'Membuat pengguna',
    steps: ['Buka menu Pengguna.', 'Klik Tambah Pengguna.', 'Isi nama, username, kata sandi awal, dan peran pengguna.', 'Simpan, lalu sampaikan kredensial awal sesuai prosedur internal.'],
    notes: ['Gunakan ADMIN hanya untuk pengelola aplikasi.', 'Gunakan STAFF untuk pengguna operasional harian.'],
  },
  {
    title: 'Memberi hak akses STAFF',
    steps: ['Buka detail pengguna STAFF.', 'Masuk ke halaman Hak Akses.', 'Centang modul dan aksi yang dibutuhkan.', 'Simpan perubahan dan minta pengguna login ulang bila menu belum berubah.'],
    notes: ['STAFF tidak mendapat akses bila belum diberi izin.', 'Berikan akses sesuai tugas, bukan semua modul sekaligus.'],
  },
  {
    title: 'Memeriksa menu yang tidak muncul',
    steps: ['Pastikan pengguna memakai akun yang benar.', 'Cek peran pengguna.', 'Untuk STAFF, cek hak akses modul yang dibutuhkan.', 'Jika akses baru saja diubah, minta pengguna keluar dan masuk kembali.'],
    notes: ['Menu yang tersembunyi biasanya terkait hak akses.', 'Jika akses sudah benar tetapi menu tetap tidak muncul, lakukan refresh browser.'],
  },
  {
    title: 'Menjaga keamanan akun',
    steps: ['Nonaktifkan akun yang tidak lagi digunakan.', 'Hindari berbagi akun antar pengguna.', 'Gunakan peran ADMIN seperlunya.', 'Audit perubahan besar melalui catatan operasional yang tersedia.'],
    notes: ['Hak akses menentukan data apa yang dapat dilihat atau diubah.', 'Kesalahan hak akses dapat membuat pengguna mengubah data di luar tugasnya.'],
  },
]

const operationalSpecialGuides = [
  {
    title: 'Impor Proyek dari Blue Book Lain',
    steps: [
      'Buka detail Blue Book tujuan, yaitu dokumen yang akan menerima proyek.',
      'Klik Impor Proyek dari Blue Book Lain.',
      'Pilih Blue Book sumber. Untuk kebutuhan revisi, pilih dokumen sumber pada periode yang sama.',
      'Gunakan pencarian bila daftar proyek cukup banyak.',
      'Centang proyek yang ingin dibawa ke Blue Book tujuan.',
      'Periksa proyek yang tidak bisa dipilih. Biasanya proyek tersebut sudah ada di dokumen tujuan atau tidak memenuhi aturan dokumen.',
      'Klik Impor Proyek untuk memasukkan proyek terpilih.',
      'Buka kembali daftar Proyek Blue Book pada dokumen tujuan untuk memastikan proyek sudah masuk.',
    ],
    notes: [
      'Fitur ini dipakai untuk membawa proyek dari dokumen atau revisi sebelumnya tanpa mengetik ulang semua data.',
      'Untuk revisi, proyek yang dibawa tetap terbaca sebagai kelanjutan proyek yang sama pada riwayat.',
      'Jika proyek tidak tersedia untuk dipilih, cek apakah proyek tersebut sudah pernah dibawa ke dokumen tujuan.',
    ],
  },
  {
    title: 'Menggunakan Data Blue Book sebagai Data Green Book',
    steps: [
      'Buka detail Blue Book atau detail Proyek Blue Book yang menjadi dasar proyek Green Book.',
      'Gunakan aksi Gunakan data di Blue Book sebagai data Green Book bila tersedia.',
      'Pilih Green Book tujuan yang akan menerima proyek.',
      'Periksa kembali data yang terisi otomatis, terutama nama proyek, instansi, lokasi, dan nilai awal.',
      'Lengkapi bagian Green Book yang belum ada di Blue Book, seperti kegiatan, sumber pendanaan, rencana pencairan, dan alokasi pendanaan.',
      'Simpan Proyek Green Book setelah semua bagian wajib lengkap.',
    ],
    notes: [
      'Aksi ini membantu mempercepat pengisian Green Book dari proyek yang sudah ada di Blue Book.',
      'Data otomatis tetap perlu diperiksa karena kebutuhan Green Book lebih rinci daripada Blue Book.',
      'Jangan langsung menyimpan sebelum sumber pendanaan dan alokasi sudah sesuai.',
    ],
  },
  {
    title: 'Tambahkan Proyek dari Green Book Lain',
    steps: [
      'Buka detail Green Book tujuan.',
      'Klik Tambahkan Proyek dari Green Book Lain.',
      'Pilih Green Book sumber yang berisi proyek yang ingin dibawa.',
      'Gunakan pencarian untuk menemukan proyek berdasarkan kode atau nama.',
      'Centang proyek yang ingin ditambahkan, atau gunakan pilih semua bila seluruh proyek memang dibutuhkan.',
      'Perhatikan proyek yang bertanda tidak tersedia karena proyek tersebut tidak bisa dibawa ke dokumen tujuan.',
      'Klik impor atau tambahkan proyek terpilih.',
      'Periksa daftar Proyek Green Book pada dokumen tujuan setelah proses selesai.',
    ],
    notes: [
      'Fitur ini berguna saat menyusun dokumen Green Book baru dari dokumen sebelumnya.',
      'Proyek yang dibawa tetap harus diperiksa karena tahun, sumber pendanaan, dan alokasi dapat berubah.',
      'Jika ada proyek yang tidak bisa dipilih, baca keterangan pada dialog sebelum mengubah data sumber.',
    ],
  },
  {
    title: 'Mengisi Kurs Tengah BI secara Massal',
    steps: [
      'Buka menu Kurs Tengah BI pada grup Referensi.',
      'Klik Tambah untuk membuat baris baru.',
      'Pilih mata uang, isi kurs, isi Kurs Tengah BI, dan pilih tanggal cut off.',
      'Ulangi penambahan baris bila perlu memasukkan beberapa mata uang atau beberapa tanggal sekaligus.',
      'Perhatikan label Baru untuk baris yang belum pernah disimpan.',
      'Jika mengubah baris lama, label Diubah akan muncul.',
      'Centang baris bila ingin menghapusnya, lalu gunakan Hapus Terpilih.',
      'Klik Simpan Perubahan untuk menyimpan seluruh baris baru dan baris yang diubah.',
    ],
    notes: [
      'Kurs Tengah BI dipakai untuk indikator tampilan Perjanjian Pinjaman.',
      'Satu mata uang dan satu tanggal cut off tidak boleh dicatat dua kali.',
      'Jika indikator Perjanjian Pinjaman kosong, cek apakah kurs untuk mata uang terkait sudah tersedia.',
    ],
  },
  {
    title: 'Menyimpan Draft Perjanjian Pinjaman',
    steps: [
      'Buka formulir Perjanjian Pinjaman.',
      'Pilih satu atau lebih Proyek Daftar Kegiatan.',
      'Isi pemberi pinjaman, kode pinjaman, tanggal penting, mata uang, nilai pinjaman, realisasi kumulatif, dan alokasi per proyek.',
      'Klik Tambahkan ke Daftar untuk memasukkan isian tersebut ke daftar draft.',
      'Ulangi pengisian bila ingin menyiapkan lebih dari satu Perjanjian Pinjaman dalam satu sesi.',
      'Gunakan edit draft bila ada data yang perlu dikoreksi sebelum disimpan.',
      'Gunakan hapus draft bila baris tersebut tidak jadi disimpan.',
      'Klik Simpan Daftar Perjanjian Pinjaman setelah seluruh draft sudah benar.',
    ],
    notes: [
      'Data belum menjadi Perjanjian Pinjaman final sebelum daftar draft disimpan.',
      'Total alokasi per proyek harus sama dengan nilai pinjaman utama.',
      'Kode pinjaman harus unik dan pemberi pinjaman harus sesuai dengan pembiayaan Proyek Daftar Kegiatan yang dipilih.',
    ],
  },
  {
    title: 'Membaca dan Mengatur Kolom Proyek',
    steps: [
      'Buka menu Proyek.',
      'Gunakan tombol pengaturan kolom di area kanan atas tabel.',
      'Centang kolom tambahan yang ingin ditampilkan, misalnya pemberi pinjaman, lokasi, nilai pinjaman, kode Blue Book, kode Green Book, atau tanggal Daftar Kegiatan.',
      'Lepas centang kolom yang tidak sedang dibutuhkan agar tabel lebih mudah dibaca.',
      'Gunakan pencarian dan filter setelah kolom yang relevan terlihat.',
      'Klik Export Excel bila ingin mengunduh data sesuai filter aktif.',
    ],
    notes: [
      'Kolom yang tampil membantu peninjauan di layar, sedangkan Export Excel mengikuti data yang sudah difilter.',
      'Gunakan kolom tambahan hanya sesuai kebutuhan agar tabel tidak terlalu padat.',
      'Sebelum membagikan Excel, cek kembali label filter aktif.',
    ],
  },
  {
    title: 'Mengelola Data Hierarki',
    steps: [
      'Buka master data yang berbentuk hierarki, seperti Instansi, Wilayah, Judul Program, atau Mitra Kerja Bappenas.',
      'Klik tanda buka pada baris induk untuk melihat data anak di bawahnya.',
      'Gunakan tambah data anak bila ingin menempatkan data baru di bawah induk tertentu.',
      'Pilih induk yang benar sebelum menyimpan.',
      'Edit data anak dari baris yang sama bila ada koreksi nama atau informasi.',
      'Hapus data hanya bila tidak sedang dipakai oleh proyek atau referensi lain.',
    ],
    notes: [
      'Hierarki membantu aplikasi membedakan level kementerian, eselon, provinsi, kabupaten/kota, dan struktur referensi lain.',
      'Kesalahan memilih induk dapat membuat data muncul di tempat yang salah pada formulir dan laporan.',
      'Untuk wilayah, pilihan nasional dapat mewakili cakupan lebih luas daripada provinsi atau kabupaten/kota.',
    ],
  },
  {
    title: 'Menonaktifkan Pengguna dan Mengatur Matriks Hak Akses',
    steps: [
      'Buka menu Pengguna pada grup Akses Admin.',
      'Pilih pengguna yang akan diubah.',
      'Gunakan Nonaktifkan bila akun tidak lagi boleh masuk ke aplikasi.',
      'Untuk pengguna STAFF, buka Atur Hak Akses.',
      'Centang modul yang boleh dibuka dan aksi yang boleh dilakukan.',
      'Simpan perubahan hak akses.',
      'Minta pengguna keluar dan masuk kembali bila menu belum berubah.',
    ],
    notes: [
      'ADMIN otomatis memiliki akses penuh.',
      'STAFF tidak mendapat akses apa pun sebelum hak akses diberikan.',
      'Berikan hak hapus hanya kepada pengguna yang memang bertugas menjaga kualitas data.',
    ],
  },
]

const systemStateGuides = [
  ['Akses ditolak', 'Pengguna tidak memiliki hak akses untuk membuka menu atau menjalankan aksi tertentu.', 'Hubungi ADMIN untuk memeriksa peran dan matriks hak akses. Setelah hak akses diperbarui, keluar dan masuk kembali.'],
  ['Halaman tidak ditemukan', 'Alamat halaman salah, data sudah tidak tersedia, atau halaman yang dibuka belum aktif.', 'Kembali ke sidebar, buka menu resmi, lalu cari data dari daftar. Jika tautan berasal dari bookmark lama, buat ulang bookmark dari halaman terbaru.'],
  ['Data kosong', 'Belum ada data, filter terlalu sempit, atau pencarian tidak cocok dengan data yang tersedia.', 'Klik Reset filter, kosongkan pencarian, lalu coba cari dengan kata kunci yang lebih umum atau kode proyek yang tepat.'],
  ['Formulir tidak bisa disimpan', 'Ada isian wajib yang belum diisi, format tanggal atau nilai tidak sesuai, atau hubungan data belum memenuhi aturan.', 'Baca pesan pada isian terkait, lengkapi data wajib, dan periksa kembali tabel rincian seperti pemberi pinjaman, biaya, alokasi, atau kegiatan.'],
  ['Konfirmasi hapus muncul', 'Sistem meminta kepastian sebelum data dihapus agar pengguna tidak menghapus data secara tidak sengaja.', 'Baca nama data pada dialog konfirmasi. Lanjutkan hanya bila data memang harus dihapus dan belum dipakai tahap berikutnya.'],
  ['Penghapusan ditolak', 'Data masih dipakai oleh dokumen atau tahap lain, misalnya Proyek Blue Book sudah dipakai Green Book atau Proyek Daftar Kegiatan sudah dipakai Perjanjian Pinjaman.', 'Buka detail data untuk melihat hubungan turunannya. Selesaikan hubungan tersebut sesuai arahan pengelola data sebelum menghapus.'],
  ['Preview impor menampilkan create', 'Baris siap dibuat sebagai data baru saat eksekusi impor dijalankan.', 'Periksa jumlah dan contoh barisnya. Jika sudah sesuai, lanjutkan hanya setelah tidak ada baris failed.'],
  ['Preview impor menampilkan skip', 'Baris dikenali sudah ada atau tidak perlu dibuat ulang.', 'Pastikan skip memang sesuai harapan, terutama saat impor revisi atau data master yang sudah pernah dibuat.'],
  ['Preview impor menampilkan failed', 'Baris gagal melewati pemeriksaan template, referensi, nilai wajib, atau aturan bisnis.', 'Baca pesan pada baris tersebut, perbaiki file Excel, unggah ulang, lalu jalankan Preview lagi. Jangan eksekusi sebelum failed menjadi nol.'],
  ['Tombol aksi tidak aktif', 'Pengguna belum memilih data, belum selesai preview, data belum memenuhi syarat, atau hak akses tidak mencukupi.', 'Cek pilihan baris, status preview, kelengkapan formulir, dan hak akses pengguna.'],
]

const troubleshootingGuides = [
  ['Menu tidak muncul', 'Pengguna belum memiliki hak akses atau belum login ulang setelah hak akses diubah.', 'Minta ADMIN memeriksa peran dan hak akses, lalu logout dan login kembali.'],
  ['Filter tidak mengubah angka', 'Filter belum diterapkan, masih ada filter lain yang aktif, atau data memang tidak memenuhi pilihan filter.', 'Klik Terapkan, cek label filter aktif, lalu gunakan Reset bila perlu mulai ulang.'],
  ['Proyek tidak ditemukan', 'Kata kunci terlalu umum, proyek berada pada revisi lama, atau filter menyembunyikan hasil.', 'Cari dengan kode proyek, aktifkan riwayat revisi bila perlu, dan hapus filter yang tidak dibutuhkan.'],
  ['Preview impor gagal', 'File Excel tidak mengikuti template, referensi master belum ada, atau ada nilai wajib yang kosong.', 'Baca baris failed, perbaiki file Excel, unggah ulang, lalu jalankan Preview lagi.'],
  ['Tombol eksekusi impor tidak bisa dipakai', 'Preview belum selesai bersih atau masih ada baris failed.', 'Pastikan total failed = 0 sebelum menjalankan eksekusi impor.'],
  ['Nilai atau indikator Perjanjian Pinjaman kosong', 'Kurs Tengah BI untuk mata uang terkait belum tersedia atau tanggal penting belum lengkap.', 'Lengkapi Kurs Tengah BI dan periksa tanggal efektif serta tanggal penutupan.'],
  ['Peta wilayah terlihat kosong', 'Filter terlalu sempit, lokasi proyek belum diisi, atau data berada di level nasional/provinsi sementara pengguna melihat kota.', 'Reset filter, cek lokasi proyek, lalu lihat tingkat wilayah yang sesuai.'],
  ['Penghapusan data ditolak', 'Data masih dipakai oleh tahap berikutnya.', 'Buka detail data untuk melihat relasi turunan, lalu selesaikan hubungan tersebut sesuai arahan pengelola data.'],
]

const dataManagementSections = [
  {
    id: 'section-blue-book',
    label: '16 - Mengelola Blue Book',
    title: 'Mengelola Blue Book, Proyek Blue Book, Indikasi Pemberi Pinjaman, dan LoI',
    imageKey: 'blue-book',
    intro:
      'Blue Book adalah istilah aplikasi untuk dokumen yang sepadan dengan DRPLN-JM: daftar rencana kegiatan jangka menengah yang dinilai layak dibiayai pinjaman luar negeri. Di dalam PRISM, Blue Book menjadi tempat mencatat dokumen per periode, revisi, usulan proyek, instansi pengusul dan pelaksana, lokasi, prioritas, biaya, indikasi pemberi pinjaman, serta Letter of Intent bila sudah tersedia.',
    items: [
      {
        title: 'Blue Book',
        scope: 'Dokumen utama untuk satu periode perencanaan jangka menengah dan revisinya.',
        create: [
          'Buka menu Blue Book.',
          'Klik tombol Buat Blue Book.',
          'Pilih periode perencanaan yang sesuai, isi tanggal terbit, nomor revisi bila ini dokumen revisi, tahun revisi, dan status dokumen.',
          'Pastikan periode dan revisi menggambarkan dokumen rencana jangka menengah yang benar, karena semua Proyek Blue Book di dalamnya akan membaca konteks tersebut.',
          'Gunakan status Berlaku bila dokumen masih dipakai sebagai acuan kerja, atau Tidak Berlaku bila dokumen hanya menjadi arsip.',
          'Klik Simpan setelah semua informasi utama sudah benar.',
        ],
        read: [
          'Gunakan kolom pencarian atau filter untuk menemukan Blue Book berdasarkan periode, tahun, revisi, atau status.',
          'Klik ikon lihat pada baris dokumen untuk membuka halaman detail.',
          'Di halaman detail, periksa informasi dokumen, daftar Proyek Blue Book, serta riwayat revisi bila tersedia.',
        ],
        update: [
          'Dari halaman detail, klik Edit Blue Book.',
          'Ubah informasi dokumen yang masih perlu dikoreksi, misalnya status, tanggal terbit, atau keterangan revisi.',
          'Klik Simpan dan pastikan perubahan muncul kembali di halaman detail.',
        ],
        delete: [
          'Hapus Blue Book hanya bila dokumen belum memiliki Proyek Blue Book.',
          'Jika dokumen sudah berisi proyek, hapus atau pindahkan dulu proyek yang ada sesuai arahan pengelola data.',
          'Jangan menghapus dokumen yang masih menjadi acuan pekerjaan aktif.',
        ],
      },
      {
        title: 'Proyek Blue Book',
        scope: 'Rincian usulan kegiatan yang berada di dalam satu dokumen Blue Book.',
        create: [
          'Buka detail Blue Book.',
          'Klik Tambah Proyek.',
          'Isi kode proyek, nama proyek, program, instansi pelaksana, lokasi, prioritas nasional, durasi, tujuan, ruang lingkup, keluaran, manfaat, nilai biaya, dan indikasi pemberi pinjaman.',
          'Tuliskan tujuan, ruang lingkup, keluaran, dan manfaat dengan jelas karena bagian ini menjelaskan mengapa kegiatan layak masuk rencana pinjaman luar negeri.',
          'Pastikan kode proyek tidak sama dengan proyek lain di dokumen Blue Book yang sama.',
          'Periksa kembali daftar instansi, lokasi, biaya, dan pemberi pinjaman sebelum menyimpan karena data ini akan menjadi dasar pada tahap berikutnya.',
          'Klik Simpan untuk memasukkan proyek ke dalam Blue Book.',
        ],
        read: [
          'Dari detail Blue Book, lihat tabel Proyek Blue Book.',
          'Gunakan pencarian atau filter bila daftar proyek cukup banyak.',
          'Klik ikon lihat pada proyek untuk membuka rincian proyek, nilai biaya, indikasi pemberi pinjaman, LoI, dan riwayat revisi proyek.',
        ],
        update: [
          'Dari detail proyek, klik Edit.',
          'Perbaiki data proyek yang perlu diperbarui, termasuk pihak terkait, lokasi, prioritas, nilai biaya, atau indikasi pemberi pinjaman.',
          'Klik Simpan, lalu buka kembali detail proyek untuk memastikan koreksi sudah tercatat.',
        ],
        delete: [
          'Hapus proyek dari daftar proyek pada detail Blue Book hanya bila proyek belum dipakai di tahapan Green Book atau tahapan setelahnya.',
          'Jika sistem menolak penghapusan, artinya proyek masih terhubung dengan data lain. Selesaikan dulu hubungan data tersebut sebelum mencoba lagi.',
        ],
      },
      {
        title: 'Indikasi Pemberi Pinjaman',
        scope: 'Calon pemberi pinjaman pada tahap awal, belum menjadi komitmen final dan belum setara dengan Perjanjian Pinjaman.',
        create: [
          'Saat membuat atau mengedit Proyek Blue Book, buka bagian Indikasi Pemberi Pinjaman.',
          'Tambahkan baris baru, pilih pemberi pinjaman dari Master Pemberi Pinjaman, lalu isi jenis pinjaman dan catatan bila diperlukan.',
          'Gunakan catatan untuk menulis informasi singkat yang membantu pembaca memahami status pembicaraan dengan pemberi pinjaman.',
          'Simpan formulir Proyek Blue Book agar indikasi pemberi pinjaman ikut tersimpan.',
        ],
        read: [
          'Buka detail Proyek Blue Book dan lihat bagian Indikasi Pemberi Pinjaman.',
          'Gunakan informasi ini untuk mengetahui pemberi pinjaman mana saja yang sedang atau pernah dipertimbangkan pada tahap Blue Book.',
          'Pahami bahwa indikasi pemberi pinjaman masih bersifat awal dan belum berarti pinjaman sudah disetujui.',
        ],
        update: [
          'Edit Proyek Blue Book, lalu ubah baris Indikasi Pemberi Pinjaman yang perlu diperbaiki.',
          'Sesuaikan pemberi pinjaman, jenis pinjaman, atau catatan bila ada perubahan informasi.',
          'Klik Simpan agar perubahan tersimpan bersama data proyek.',
        ],
        delete: [
          'Edit Proyek Blue Book, hapus baris indikasi yang keliru atau sudah tidak relevan, lalu klik Simpan.',
          'Jangan menghapus indikasi pemberi pinjaman yang masih diperlukan sebagai jejak informasi proyek.',
        ],
      },
      {
        title: 'Letter of Intent (LoI)',
        scope: 'Surat minat dari pemberi pinjaman yang terkait dengan Proyek Blue Book dan membantu memperjelas indikasi pendanaan.',
        create: [
          'Buka detail Proyek Blue Book.',
          'Pada bagian Letter of Intent, klik tombol tambah.',
          'Pilih pemberi pinjaman yang sudah ada di Indikasi Pemberi Pinjaman proyek.',
          'Isi perihal surat, tanggal surat, dan nomor surat bila tersedia.',
          'Klik Simpan untuk mencatat surat minat tersebut.',
        ],
        read: [
          'Daftar Letter of Intent ditampilkan pada detail Proyek Blue Book.',
          'Gunakan daftar ini untuk membedakan proyek yang baru memiliki indikasi pemberi pinjaman dengan proyek yang sudah memiliki minat tertulis.',
        ],
        update: [
          'Jika ada kesalahan pada perihal, tanggal, nomor surat, atau pemberi pinjaman, gunakan aksi edit yang tersedia pada tabel LoI.',
          'Simpan perubahan setelah data surat sudah sesuai dengan dokumen sumber.',
        ],
        delete: [
          'Hapus LoI hanya bila surat tersebut keliru, tidak berlaku, atau memang tidak seharusnya dikaitkan dengan proyek.',
          'Pastikan penghapusan dilakukan oleh pengguna yang memiliki hak akses sesuai.',
        ],
      },
    ],
  },
  {
    id: 'section-green-book',
    label: '17 - Mengelola Green Book',
    title: 'Mengelola Green Book dan Proyek Green Book',
    imageKey: 'green-book',
    intro:
      'Green Book adalah istilah aplikasi untuk dokumen yang sepadan dengan DRPPLN: daftar prioritas tahunan yang berisi rencana kegiatan dengan indikasi pendanaan dan kesiapan lebih lanjut. Di tahap ini, pengguna menghubungkan proyek dengan Proyek Blue Book, mencatat kegiatan, sumber pendanaan, rencana pencairan, dan alokasi pendanaan agar proyek siap dibawa ke Daftar Kegiatan.',
    items: [
      {
        title: 'Green Book',
        scope: 'Dokumen prioritas pendanaan tahunan berdasarkan tahun terbit dan revisi.',
        create: [
          'Buka menu Green Book.',
          'Klik Buat Green Book.',
          'Isi tahun terbit, nomor revisi, dan status dokumen.',
          'Gunakan tahun terbit untuk menunjukkan tahun prioritas pendanaan yang sedang dicatat.',
          'Gunakan status Berlaku bila Green Book menjadi acuan saat ini, atau Tidak Berlaku bila hanya disimpan sebagai arsip.',
          'Klik Simpan untuk membuat dokumen Green Book.',
        ],
        read: [
          'Gunakan daftar Green Book untuk melihat tahun terbit, revisi, status, dan jumlah proyek.',
          'Klik ikon lihat untuk membuka detail dokumen dan daftar Proyek Green Book di dalamnya.',
          'Periksa juga informasi revisi bila dokumen memiliki hubungan dengan Green Book sebelumnya.',
        ],
        update: [
          'Dari detail Green Book klik Edit Green Book.',
          'Ubah informasi dokumen yang perlu diperbaiki, misalnya status atau nomor revisi.',
          'Klik Simpan dan pastikan data terbaru tampil pada detail Green Book.',
        ],
        delete: [
          'Hapus Green Book hanya jika belum memiliki Proyek Green Book dan tidak menjadi sumber revisi Green Book lain.',
          'Jika sudah memiliki proyek, selesaikan dulu data proyek di dalamnya sesuai arahan pengelola data.',
        ],
      },
      {
        title: 'Proyek Green Book',
        scope: 'Rincian proyek prioritas pendanaan yang mengacu ke minimal satu Proyek Blue Book dan memuat kesiapan pendanaan.',
        create: [
          'Buka detail Green Book.',
          'Klik Tambah Proyek.',
          'Pilih minimal satu Proyek Blue Book sebagai dasar proyek.',
          'Jika memilih lebih dari satu Proyek Blue Book, pastikan semuanya berasal dari dokumen Blue Book yang sama.',
          'Isi kode, nama proyek, mata uang, instansi, lokasi, daftar kegiatan, sumber pendanaan, rencana pencairan, dan alokasi pendanaan.',
          'Pastikan sumber pendanaan dan rencana pencairan menggambarkan kesiapan pembiayaan, bukan sekadar indikasi awal.',
          'Periksa tabel kegiatan dan alokasi pendanaan. Alokasi pendanaan mengikuti daftar kegiatan, sehingga perubahan kegiatan dapat memengaruhi alokasi.',
          'Klik Simpan untuk membuat Proyek Green Book.',
        ],
        read: [
          'Buka detail Proyek Green Book untuk melihat dokumen Blue Book asal, kegiatan, sumber pendanaan, rencana pencairan, dan alokasi pendanaan.',
          'Gunakan detail ini untuk memastikan data pendanaan Green Book sudah sesuai sebelum proyek masuk ke Daftar Kegiatan.',
        ],
        update: [
          'Klik Edit dari detail atau daftar proyek.',
          'Ubah data proyek, kegiatan, sumber pendanaan, atau rencana pencairan yang perlu diperbaiki.',
          'Pastikan alokasi pendanaan tetap sesuai setelah kegiatan diubah.',
          'Klik Simpan dan periksa kembali detail proyek.',
        ],
        delete: [
          'Hapus proyek dari daftar Proyek Green Book.',
          'Jika proyek sudah dipakai oleh Daftar Kegiatan, sistem akan menolak penghapusan dan menampilkan hubungan data yang harus diselesaikan terlebih dahulu.',
        ],
      },
    ],
  },
  {
    id: 'section-daftar-kegiatan-perjanjian',
    label: '18 - Mengelola Daftar Kegiatan dan Perjanjian Pinjaman',
    title: 'Mengelola Daftar Kegiatan, Proyek Daftar Kegiatan, dan Perjanjian Pinjaman',
    imageKey: 'daftar-kegiatan',
    intro:
      'Daftar Kegiatan berisi rencana kegiatan yang sudah tercantum dalam Green Book dan siap diusulkan atau dirundingkan dengan calon pemberi pinjaman. Setelah proses perundingan selesai, Perjanjian Pinjaman mencatat komitmen tertulis yang mengikat, termasuk jumlah, peruntukan, hak dan kewajiban, serta ketentuan dan persyaratan pinjaman.',
    items: [
      {
        title: 'Daftar Kegiatan',
        scope: 'Dokumen surat yang memuat daftar kegiatan siap usul atau siap runding.',
        create: [
          'Buka menu Daftar Kegiatan.',
          'Klik Buat Daftar Kegiatan.',
          'Isi perihal surat, tanggal surat, nomor surat, dan informasi pendukung lain yang tersedia.',
          'Pastikan informasi surat sesuai dokumen sumber karena surat ini menjadi wadah Proyek Daftar Kegiatan yang akan dibawa menuju Perjanjian Pinjaman.',
          'Klik Simpan untuk membuat header Daftar Kegiatan.',
        ],
        read: [
          'Cari Daftar Kegiatan dari daftar dokumen surat.',
          'Klik ikon lihat untuk membuka detail surat dan daftar Proyek Daftar Kegiatan di dalamnya.',
          'Periksa status dan isi surat sebelum menambahkan atau mengubah proyek.',
        ],
        update: [
          'Daftar Kegiatan bersifat final setelah diterbitkan.',
          'Sebelum dokumen final, koreksi perihal, tanggal, nomor surat, atau informasi pendukung lain bila diperlukan.',
          'Setelah final, perubahan dibatasi dan biasanya hanya dapat dilakukan oleh pengguna dengan kewenangan khusus.',
        ],
        delete: [
          'Hapus Daftar Kegiatan hanya bila belum memiliki Proyek Daftar Kegiatan.',
          'Jika sudah ada proyek di dalam surat, selesaikan atau hapus proyek tersebut terlebih dahulu sesuai aturan data.',
        ],
      },
      {
        title: 'Proyek Daftar Kegiatan',
        scope: 'Rincian kegiatan di dalam surat Daftar Kegiatan yang berasal dari Proyek Green Book.',
        create: [
          'Buka detail Daftar Kegiatan.',
          'Klik Tambah Proyek ke Surat.',
          'Pilih Proyek Green Book terlebih dahulu agar data dasar seperti nama proyek, instansi, lokasi, dan mitra dapat terisi otomatis.',
          'Periksa kembali data otomatis tersebut karena pengguna masih dapat menyesuaikannya sebelum disimpan.',
          'Lengkapi rincian pembiayaan, alokasi pinjaman, dan rincian kegiatan.',
          'Gunakan rincian pembiayaan untuk mencatat pemberi pinjaman, mata uang, dan nilai yang akan menjadi dasar pemilihan pemberi pinjaman pada Perjanjian Pinjaman.',
          'Rincian kegiatan adalah catatan kegiatan bebas pada tahap Daftar Kegiatan, sehingga tidak harus sama persis dengan daftar kegiatan di Green Book.',
          'Klik Simpan.',
        ],
        read: [
          'Detail Daftar Kegiatan menampilkan setiap proyek dalam panel yang dapat dibuka tutup.',
          'Di setiap panel, periksa dokumen Green Book asal, pembiayaan, alokasi pinjaman, dan rincian kegiatan.',
        ],
        update: [
          'Buka edit Proyek Daftar Kegiatan bila tombol edit tersedia dan hak akses mengizinkan.',
          'Perbarui nama proyek, instansi, lokasi, pembiayaan, alokasi, atau rincian kegiatan sebelum data masuk ke Perjanjian Pinjaman.',
          'Klik Simpan dan periksa kembali detail surat.',
        ],
        delete: [
          'Hapus proyek hanya bila belum dipakai oleh Perjanjian Pinjaman atau data lain setelahnya.',
          'Jika sistem menolak penghapusan, periksa data perjanjian yang masih terhubung dengan proyek tersebut.',
        ],
      },
      {
        title: 'Perjanjian Pinjaman',
        imageKey: 'loan-agreements',
        scope: 'Kesepakatan tertulis hasil perundingan yang menjadi dasar komitmen pinjaman dan dapat mencakup satu atau lebih Proyek Daftar Kegiatan.',
        create: [
          'Buka menu Perjanjian Pinjaman atau gunakan aksi dari detail Proyek Daftar Kegiatan.',
          'Klik Buat Perjanjian Pinjaman. Bila tombol di aplikasi masih memakai istilah Loan Agreement, gunakan tombol tersebut.',
          'Pilih satu atau lebih Proyek Daftar Kegiatan yang sudah memenuhi syarat.',
          'Isi pemberi pinjaman, kode pinjaman, tanggal perjanjian, tanggal efektif, tanggal penutupan awal bila ada, tanggal penutupan terbaru, mata uang, nilai pinjaman, dan realisasi kumulatif.',
          'Gunakan kode pinjaman dan tanggal perjanjian sesuai dokumen perjanjian, bukan berdasarkan indikasi awal di Blue Book.',
          'Isi alokasi pinjaman untuk setiap Proyek Daftar Kegiatan yang dipilih.',
          'Pastikan total alokasi per proyek sama dengan nilai pinjaman utama sebelum menyimpan.',
          'Klik Tambahkan ke Daftar, lalu Simpan Daftar Perjanjian Pinjaman.',
        ],
        read: [
          'Daftar Perjanjian Pinjaman menampilkan kode pinjaman, pemberi pinjaman, proyek yang terkait, tanggal efektif, realisasi pencairan, status kinerja, dan informasi perpanjangan.',
          'Buka detail untuk melihat alokasi pinjaman per Proyek Daftar Kegiatan dan perhitungan tampilan berdasarkan Kurs Tengah BI.',
          'Gunakan detail ini untuk memahami nilai pinjaman secara keseluruhan dan pembagian komitmen ke masing-masing proyek.',
        ],
        update: [
          'Klik Edit pada daftar atau detail Perjanjian Pinjaman.',
          'Perbarui data perjanjian yang masih perlu dikoreksi, seperti tanggal, nilai, realisasi kumulatif, atau alokasi proyek.',
          'Klik Simpan Perubahan dan pastikan ringkasan perjanjian menampilkan nilai terbaru.',
        ],
        delete: [
          'Klik Hapus dari detail Perjanjian Pinjaman hanya bila data perjanjian memang harus dibatalkan.',
          'Pastikan pengguna memiliki hak akses untuk menghapus dan penghapusan sudah sesuai kebijakan pengelolaan data.',
        ],
      },
    ],
  },
]

function escapeHtml(value) {
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function list(items) {
  return `<ul>${items.map((item) => `<li>${escapeHtml(item)}</li>`).join('')}</ul>`
}

function screenshotFigure(key, alt, caption) {
  const relative = `assets/${key}.png`
  return `
    <figure class="screen-figure">
      <img src="${relative}" alt="${escapeHtml(alt)}" />
      <figcaption>${escapeHtml(caption)}</figcaption>
    </figure>
  `
}

function actionBlock(label, items) {
  return `
    <div class="guide-action">
      <strong>${escapeHtml(label)}</strong>
      ${list(items)}
    </div>
  `
}

function tableRows(rows) {
  return rows
    .map((row) => `<tr>${row.map((cell) => `<td>${escapeHtml(cell)}</td>`).join('')}</tr>`)
    .join('')
}

function renderDetailCards(items) {
  return items
    .map(
      (item) => `
        <article class="detail-card">
          <p class="eyebrow">${escapeHtml(item.focus)}</p>
          <h3>${escapeHtml(item.title)}</h3>
          ${list(item.usage)}
        </article>
      `,
    )
    .join('')
}

function renderMasterDataCards(items) {
  return items
    .map(
      (item) => `
        <article class="detail-card">
          <p class="eyebrow">${escapeHtml(item.purpose)}</p>
          <h3>${escapeHtml(item.title)}</h3>
          ${list(item.fill)}
          <p class="note"><strong>Perhatian:</strong> ${escapeHtml(item.caution)}</p>
        </article>
      `,
    )
    .join('')
}

function renderFieldGuides(items) {
  return items
    .map(
      (item) => `
        <article class="module-card">
          <h3>${escapeHtml(item.title)}</h3>
          <table class="compact-table">
            <thead><tr><th>Bagian formulir</th><th>Yang diisi</th><th>Yang perlu diperhatikan</th></tr></thead>
            <tbody>${tableRows(item.rows)}</tbody>
          </table>
        </article>
      `,
    )
    .join('')
}

function renderWorkbookGuides(items) {
  return items
    .map(
      (item) => `
        <article class="module-card">
          <p class="eyebrow">Workbook ${escapeHtml(item.title)}</p>
          <h3>${escapeHtml(item.title)}</h3>
          <p class="mb-small"><strong>Sheet utama:</strong> ${escapeHtml(item.sheets)}</p>
          <div class="two-col compact">
            <div>
              <h4>Sebelum upload</h4>
              ${list(item.before)}
            </div>
            <div>
              <h4>Saat preview</h4>
              ${list(item.preview)}
            </div>
          </div>
        </article>
      `,
    )
    .join('')
}

function renderAccessGuides(items) {
  return items
    .map(
      (item) => `
        <article class="detail-card">
          <h3>${escapeHtml(item.title)}</h3>
          <h4>Langkah</h4>
          ${list(item.steps)}
          <h4 class="mt-small">Catatan</h4>
          ${list(item.notes)}
        </article>
      `,
    )
    .join('')
}

function renderDataManagementSection(section) {
  return `
    <section id="${escapeHtml(section.id)}" class="page data-page">
      <p class="section-label">${escapeHtml(section.label)}</p>
      <h2>${escapeHtml(section.title)}</h2>
      <p class="mb">${escapeHtml(section.intro)}</p>
      <div class="guide-grid">
        ${section.items
          .map(
            (item) => `
              <article class="guide-card">
                <img class="guide-shot" src="assets/${escapeHtml(item.imageKey ?? section.imageKey)}.png" alt="Tampilan ${escapeHtml(item.title)}" />
                <p class="eyebrow">${escapeHtml(item.scope)}</p>
                <h3>${escapeHtml(item.title)}</h3>
                <div class="guide-actions">
                  ${actionBlock('Mencatat data', item.create)}
                  ${actionBlock('Memeriksa data', item.read)}
                  ${actionBlock('Memperbaiki data', item.update)}
                  ${actionBlock('Membatalkan atau menghapus', item.delete)}
                </div>
              </article>
            `,
          )
          .join('')}
      </div>
      <div class="footer"><span>PRISM Manual Book</span><span>${escapeHtml(section.title)}</span></div>
    </section>
  `
}

async function fileExists(filePath) {
  try {
    await fs.access(filePath)
    return true
  } catch {
    return false
  }
}

async function launchBrowser() {
  const candidates = [
    process.env.CHROME_PATH,
    'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files\\Microsoft\\Edge\\Application\\msedge.exe',
    'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe',
  ].filter(Boolean)

  for (const candidate of candidates) {
    if (await fileExists(candidate)) {
      return chromium.launch({ headless: true, executablePath: candidate })
    }
  }

  return chromium.launch({ headless: true })
}

async function captureScreenshots() {
  await fs.mkdir(assetsDir, { recursive: true })

  let session = null
  try {
    const response = await fetch(`${apiUrl}/auth/login`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' }),
    })
    if (response.ok) {
      const payload = await response.json()
      session = payload.data
    } else {
      console.warn(`Login API returned ${response.status}; protected screenshots may show login page.`)
    }
  } catch (error) {
    console.warn(`Could not call login API: ${error.message}`)
  }

  const browser = await launchBrowser()
  const context = await browser.newContext({
    viewport: { width: 1440, height: 980 },
    deviceScaleFactor: 1,
  })
  const page = await context.newPage()
  page.setDefaultTimeout(15000)

  try {
    const loginTarget = screenshotTargets.find((target) => target.key === 'login')
    await page.goto(`${frontendUrl}${loginTarget.route}`, { waitUntil: 'domcontentloaded' })
    await page.waitForTimeout(1200)
    await page.screenshot({ path: path.join(assetsDir, `${loginTarget.key}.png`), fullPage: false })

    if (session?.access_token) {
      await page.evaluate((payload) => {
        window.localStorage.setItem('prism.access_token', payload.access_token)
        window.localStorage.setItem('prism.user', JSON.stringify(payload.user))
        window.localStorage.setItem('prism.permissions', JSON.stringify([]))
      }, session)
    }

    for (const target of screenshotTargets.filter((item) => !item.public)) {
      await page.goto(`${frontendUrl}${target.route}`, { waitUntil: 'domcontentloaded' })
      await page.waitForTimeout(1800)
      await page.screenshot({ path: path.join(assetsDir, `${target.key}.png`), fullPage: false })
    }
  } finally {
    await browser.close()
  }
}

async function buildHtml(tocPageNumbers = {}) {
  const logoPath = path.join(repoRoot, 'prism-frontend', 'public', 'prism-logo.png')
  const logoDataUri = (await fileExists(logoPath))
    ? `data:image/png;base64,${await fs.readFile(logoPath, 'base64')}`
    : ''
  const generatedDate = '10 Mei 2026'

  const moduleCards = modules
    .map(
      (module) => `
        <article class="module-card">
          <p class="eyebrow">${escapeHtml(module.owner)}</p>
          <h3>${escapeHtml(module.title)}</h3>
          <p>${escapeHtml(module.purpose)}</p>
          <div class="two-col">
            <div>
              <h4>Langkah utama</h4>
              ${list(module.steps)}
            </div>
            <div>
              <h4>Catatan penggunaan</h4>
              ${list(module.notes)}
            </div>
          </div>
        </article>
      `,
    )
    .join('')

  const renderFilterGuideCards = (guides) => guides
    .map(
      (guide) => `
        <article class="module-card">
          <p class="eyebrow">${escapeHtml(guide.scope)}</p>
          <h3>${escapeHtml(guide.title)}</h3>
          <h4>Filter dan alat yang tersedia</h4>
          ${list(guide.filters)}
          <div class="two-col compact">
            <div>
              <h4>Cara menggunakan</h4>
              ${list(guide.steps)}
            </div>
            <div>
              <h4>Tips membaca hasil</h4>
              ${list(guide.tips)}
            </div>
          </div>
        </article>
      `,
    )
    .join('')
  const filterGuidePrimaryCards = renderFilterGuideCards(filterGuides.slice(0, 2))
  const filterGuideSpatialCards = renderFilterGuideCards(filterGuides.slice(2))
  const stageFlowCards = stageFlowSteps
    .map(
      (step) => `
        <div class="step ${escapeHtml(step.color)}">
          <img src="assets/${escapeHtml(step.imageKey)}.png" alt="Tampilan ${escapeHtml(step.title)}" />
          <span>${escapeHtml(step.number)}</span>
          <strong>${escapeHtml(step.title)}</strong>
          <p>${escapeHtml(step.description)}</p>
        </div>
      `,
    )
    .join('')

  const planningCards = planningModules
    .map(
      (module, index) => `
        <article class="planning-card">
          <div class="stage-mark">${String(index + 1).padStart(2, '0')}</div>
          <div>
            <h3>${escapeHtml(module.title)}</h3>
            <p class="muted">${escapeHtml(module.subtitle)}</p>
            <p class="context-note">${escapeHtml(module.context)}</p>
            <img class="planning-shot" src="assets/${escapeHtml(module.imageKey)}.png" alt="Tampilan ${escapeHtml(module.title)}" />
            <div class="chips">${module.fields.map((field) => `<span>${escapeHtml(field)}</span>`).join('')}</div>
            <div class="two-col compact">
              <div>
                <h4>Proses</h4>
                ${list(module.process)}
              </div>
              <div>
                <h4>Aturan penting</h4>
                ${list(module.rules)}
              </div>
            </div>
          </div>
        </article>
      `,
    )
    .join('')

  const masterRows = masterData
    .map(([name, note]) => `<tr><td>${escapeHtml(name)}</td><td>${escapeHtml(note)}</td></tr>`)
    .join('')

  const importRows = importKinds
    .map(([name, note]) => `<tr><td>${escapeHtml(name)}</td><td>${escapeHtml(note)}</td></tr>`)
    .join('')
  const legalRows = tableRows(legalReferenceRows)
  const dataManagementPages = dataManagementSections.map(renderDataManagementSection).join('')
  const dashboardDetailCards = renderDetailCards(dashboardDetailPanels)
  const projectMasterRows = tableRows(projectMasterDetails)
  const masterDataGuideCards = renderMasterDataCards(masterDataGuides)
  const fieldGuideCards = renderFieldGuides(formFieldGuides)
  const revisionHistoryRows = tableRows(revisionHistoryGuides)
  const importWorkbookCards = renderWorkbookGuides(importWorkbookGuides)
  const accessGuideCards = renderAccessGuides(accessManagementGuides)
  const operationalGuideCards = renderAccessGuides(operationalSpecialGuides)
  const systemStateRows = tableRows(systemStateGuides)
  const troubleshootingRows = tableRows(troubleshootingGuides)

  const html = `<!doctype html>
<html lang="id">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Manual Book PRISM</title>
  <style>
    @page {
      size: A4;
      margin: 14mm 13mm 16mm;
    }

    :root {
      --ink: #10202f;
      --muted: #5d6b7a;
      --line: #d9e2e9;
      --paper: #fbfcfb;
      --teal: #0b6f73;
      --teal-2: #12918f;
      --green: #1fa06f;
      --gold: #f2b63f;
      --orange: #d97706;
      --violet: #7252aa;
      --slate: #2f4051;
      --soft: #eef6f4;
      --soft-gold: #fff5db;
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      color: var(--ink);
      background: white;
      font-family: Inter, "Segoe UI", Arial, sans-serif;
      font-size: 10.4pt;
      line-height: 1.52;
      letter-spacing: 0;
    }

    h1, h2, h3, h4, p { margin: 0; }

    h1 {
      font-size: 40pt;
      line-height: 1;
      letter-spacing: 0;
    }

    h2 {
      font-size: 20pt;
      line-height: 1.15;
      margin-bottom: 10mm;
    }

    h3 {
      font-size: 13.5pt;
      line-height: 1.2;
      margin-bottom: 3mm;
    }

    h4 {
      font-size: 9.5pt;
      text-transform: uppercase;
      letter-spacing: .08em;
      color: var(--teal);
      margin-bottom: 2mm;
    }

    p { color: var(--muted); }
    strong { color: var(--ink); }
    ul { margin: 0; padding-left: 4.8mm; }
    li { margin: 0 0 1.5mm; }

    .page {
      min-height: 267mm;
      padding-bottom: 14mm;
      page-break-after: auto;
      break-after: auto;
      position: relative;
    }

    .page + .page {
      page-break-before: always;
      break-before: page;
    }

    .data-page {
      min-height: auto;
      padding-bottom: 8mm;
      page-break-after: auto;
      break-after: auto;
    }

    .cover {
      margin: -14mm -13mm -16mm;
      min-height: 297mm;
      padding: 22mm 19mm;
      color: white;
      background:
        linear-gradient(135deg, rgba(11,111,115,.98), rgba(47,64,81,.98)),
        repeating-linear-gradient(0deg, rgba(255,255,255,.06) 0 1px, transparent 1px 18px),
        repeating-linear-gradient(90deg, rgba(255,255,255,.05) 0 1px, transparent 1px 18px);
      display: flex;
      flex-direction: column;
      justify-content: space-between;
    }

    .cover .brand {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 700;
      color: white;
    }

    .cover .brand img {
      width: 34px;
      height: 34px;
      border-radius: 8px;
      background: white;
      padding: 3px;
    }

    .cover .kicker {
      color: var(--gold);
      text-transform: uppercase;
      font-size: 10pt;
      letter-spacing: .14em;
      font-weight: 700;
      margin-bottom: 8mm;
    }

    .cover p {
      color: rgba(255,255,255,.82);
      max-width: 145mm;
      font-size: 13pt;
      margin-top: 7mm;
    }

    .cover .meta {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 7mm;
      margin-top: 18mm;
    }

    .cover .meta div {
      border-top: 1px solid rgba(255,255,255,.26);
      padding-top: 4mm;
      color: rgba(255,255,255,.8);
      font-size: 9.5pt;
    }

    .cover .meta span {
      display: block;
      color: white;
      font-weight: 700;
      margin-bottom: 1mm;
    }

    .section-label,
    .eyebrow {
      color: var(--teal);
      font-weight: 800;
      letter-spacing: .09em;
      text-transform: uppercase;
      font-size: 8.5pt;
      margin-bottom: 3mm;
    }

    .intro-grid,
    .toc-grid,
    .role-grid,
    .workflow-grid,
    .screenshot-grid {
      display: grid;
      gap: 5mm;
    }

    .intro-grid { grid-template-columns: 1.05fr .95fr; }
    .toc-grid { grid-template-columns: repeat(2, 1fr); }
    .role-grid { grid-template-columns: repeat(2, 1fr); }
    .workflow-grid { grid-template-columns: repeat(2, 1fr); }
    .screenshot-grid { grid-template-columns: 1fr 1fr; }

    .intro-copy {
      border-left: 4px solid var(--teal);
      padding: 1mm 0 1mm 5mm;
      margin-bottom: 6mm;
      max-width: 174mm;
    }

    .intro-copy p {
      margin-bottom: 3mm;
      color: #405163;
      font-size: 10.8pt;
    }

    .intro-copy p:last-child { margin-bottom: 0; }

    .panel,
    .module-card,
    .planning-card {
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #fff;
      padding: 5mm;
      break-inside: avoid;
    }

    .panel.teal {
      background: var(--soft);
      border-color: #bcd9d3;
    }

    .panel.gold {
      background: var(--soft-gold);
      border-color: #f4d58d;
    }

    .toc-item {
      border-bottom: 1px solid var(--line);
      padding: 0;
    }

    .toc-link {
      display: grid;
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 5mm;
      align-items: baseline;
      padding: 3.2mm 0;
      color: inherit;
      text-decoration: none;
    }

    .toc-main {
      display: grid;
      gap: 1mm;
    }

    .toc-main strong { font-size: 10.8pt; }
    .toc-main span { color: var(--muted); font-size: 9pt; }

    .toc-page {
      min-width: 10mm;
      text-align: right;
      color: var(--teal);
      font-weight: 800;
      font-size: 10pt;
    }

    .flow {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 4mm;
      margin: 7mm 0;
    }

    .flow .step {
      border-radius: 8px;
      padding: 3mm;
      color: white;
      min-height: 61mm;
      overflow: hidden;
    }

    .step.blue { background: var(--teal); }
    .step.green { background: var(--green); }
    .step.orange { background: var(--orange); }
    .step.violet { background: var(--violet); }
    .step img {
      display: block;
      width: 100%;
      height: 22mm;
      object-fit: cover;
      object-position: top left;
      border-radius: 6px;
      margin-bottom: 3mm;
      border: 1px solid rgba(255,255,255,.26);
      background: rgba(255,255,255,.18);
    }
    .step span { display: block; opacity: .72; font-size: 8.5pt; margin-bottom: 2mm; }
    .step p { color: rgba(255,255,255,.84); font-size: 9pt; margin-top: 2mm; }

    .two-col {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 6mm;
      margin-top: 5mm;
    }

    .two-col.compact { gap: 5mm; }

    .role-card h3 {
      border-left: 4px solid var(--gold);
      padding-left: 3mm;
    }

    .module-card {
      margin-bottom: 5mm;
    }

    .detail-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 4mm;
      margin-bottom: 6mm;
    }

    .detail-card {
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #fff;
      padding: 4.2mm;
      break-inside: avoid;
    }

    .detail-card h3 {
      margin-bottom: 2.4mm;
    }

    .detail-card .note {
      margin-top: 3mm;
      padding-top: 2.4mm;
      border-top: 1px solid var(--line);
      font-size: 9.2pt;
    }

    .guide-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 4mm;
      padding-bottom: 12mm;
    }

    .data-page .guide-grid {
      display: block;
      padding-bottom: 4mm;
    }

    .guide-card {
      border: 1px solid var(--line);
      border-radius: 8px;
      background: white;
      padding: 4.2mm;
      break-inside: avoid;
    }

    .data-page .guide-card {
      margin-bottom: 5mm;
    }

    .guide-card h3 {
      margin-bottom: 2mm;
    }

    .guide-shot {
      display: block;
      width: 100%;
      height: 28mm;
      object-fit: cover;
      object-position: top left;
      border: 1px solid var(--line);
      border-radius: 7px;
      margin-bottom: 3mm;
      background: #f7faf9;
    }

    .guide-actions {
      display: grid;
      gap: 2.4mm;
      margin-top: 3mm;
    }

    .guide-action {
      border-top: 1px solid var(--line);
      padding-top: 2.4mm;
      font-size: 9.2pt;
    }

    .guide-action:first-child {
      border-top: none;
      padding-top: 0;
    }

    .guide-action strong {
      display: inline-block;
      color: var(--teal);
      font-size: 8.6pt;
      text-transform: uppercase;
      letter-spacing: .08em;
      margin-bottom: 1.2mm;
    }

    .guide-action li {
      margin-bottom: 1mm;
    }

    .planning-card {
      display: grid;
      grid-template-columns: 15mm 1fr;
      gap: 5mm;
      margin-bottom: 5mm;
    }

    .context-note {
      margin-top: 2.5mm;
      color: #405163;
    }

    .stage-mark {
      height: 15mm;
      width: 15mm;
      border-radius: 8px;
      background: var(--teal);
      color: white;
      display: grid;
      place-items: center;
      font-weight: 800;
    }

    .planning-shot {
      display: block;
      width: 100%;
      height: 35mm;
      object-fit: cover;
      object-position: top left;
      border: 1px solid var(--line);
      border-radius: 8px;
      margin: 4mm 0;
      background: #f7faf9;
    }

    .workflow-step-shot {
      display: block;
      width: 100%;
      height: 20mm;
      object-fit: cover;
      object-position: top left;
      border: 1px solid var(--line);
      border-radius: 6px;
      margin-bottom: 3mm;
      background: #f7faf9;
    }

    .chips {
      display: flex;
      flex-wrap: wrap;
      gap: 2mm;
      margin: 4mm 0 1mm;
    }

    .chips span {
      border: 1px solid #cbd8dd;
      background: #f7faf9;
      border-radius: 999px;
      padding: 1.2mm 2.4mm;
      color: var(--slate);
      font-size: 8.6pt;
      white-space: nowrap;
    }

    table {
      width: 100%;
      border-collapse: collapse;
      break-inside: avoid;
      background: white;
      border: 1px solid var(--line);
      border-radius: 8px;
      overflow: hidden;
    }

    thead { display: table-header-group; }
    tr { break-inside: avoid; }

    th, td {
      text-align: left;
      padding: 3mm;
      border-bottom: 1px solid var(--line);
      vertical-align: top;
    }

    th {
      background: var(--soft);
      color: var(--teal);
      font-size: 8.8pt;
      text-transform: uppercase;
      letter-spacing: .08em;
    }

    tr:last-child td { border-bottom: none; }
    td:first-child { width: 35%; font-weight: 700; color: var(--ink); }

    .compact-table td:first-child { width: 22%; }
    .compact-table td:nth-child(2) { width: 38%; }
    .compact-table th,
    .compact-table td {
      padding: 2.4mm;
      font-size: 9.1pt;
    }

    .legal-table td:first-child { width: 24%; }
    .legal-table td:nth-child(2) { width: 38%; }

    .screen-figure {
      margin: 0 0 6mm;
      break-inside: avoid;
      border: 1px solid var(--line);
      border-radius: 9px;
      background: #fff;
      overflow: hidden;
      box-shadow: 0 8px 22px rgba(16,32,47,.08);
    }

    .screen-figure img {
      display: block;
      width: 100%;
      height: auto;
    }

    .screen-figure figcaption {
      padding: 3mm 4mm;
      color: var(--muted);
      font-size: 9pt;
      border-top: 1px solid var(--line);
      background: #fbfcfb;
    }

    .callout {
      border-left: 4px solid var(--gold);
      background: var(--soft-gold);
      padding: 4mm 5mm;
      border-radius: 6px;
      margin: 6mm 0;
      break-inside: avoid;
    }

    .checklist {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 4mm;
    }

    .check {
      border: 1px solid var(--line);
      border-radius: 8px;
      padding: 4mm;
      background: white;
      min-height: 26mm;
    }

    .check strong {
      display: block;
      margin-bottom: 1.6mm;
    }

    .footer {
      position: absolute;
      left: 0;
      right: 0;
      bottom: 0;
      display: flex;
      justify-content: space-between;
      color: #82909c;
      font-size: 8.2pt;
      border-top: 1px solid var(--line);
      padding-top: 3mm;
    }

    .data-page .footer {
      display: none;
    }

    .muted { color: var(--muted); }
    .mb { margin-bottom: 6mm; }
    .mt { margin-top: 6mm; }
    .mb-small { margin-bottom: 3mm; }
    .mt-small { margin-top: 3mm; }
    .no-break { break-inside: avoid; }
  </style>
</head>
<body>
  <section class="cover page">
    <div>
      <div class="brand">
        ${logoDataUri ? `<img src="${logoDataUri}" alt="" />` : ''}
        <span>PRISM</span>
      </div>
      <div style="margin-top:38mm">
        <p class="kicker">Manual Book Pengguna</p>
        <h1>Project Loan Integrated Monitoring System</h1>
        <p>Dokumen panduan operasional untuk membaca dashboard, mengelola data referensi, mengisi alur Blue Book sampai Perjanjian Pinjaman, menjalankan impor file Excel, dan mengatur hak akses pengguna.</p>
      </div>
      <div class="meta">
        <div><span>Versi</span>Manual pengguna</div>
        <div><span>Tanggal</span>${generatedDate}</div>
        <div><span>Lingkup</span>Aplikasi PRISM berbasis web</div>
      </div>
    </div>
    <div>
      <p style="font-size:9.5pt;color:rgba(255,255,255,.7)">Disusun dari struktur menu, dokumen aturan bisnis, dan tampilan aplikasi PRISM pada workspace lokal.</p>
    </div>
  </section>

  <section id="section-orientasi" class="page">
    <p class="section-label">01 - Orientasi</p>
    <h2>Pengantar</h2>
    <div class="intro-copy">
      <p>PRISM adalah aplikasi untuk membantu pengelolaan informasi pinjaman luar negeri secara lebih tertib, mulai dari tahap usulan awal sampai menjadi perjanjian pinjaman. Di dalam satu aplikasi, pengguna dapat melihat hubungan antara Blue Book, Green Book, Daftar Kegiatan, dan Perjanjian Pinjaman sehingga riwayat sebuah proyek lebih mudah ditelusuri.</p>
      <p>Manual ini disusun sebagai panduan kerja sehari-hari bagi pengguna aplikasi. Isi manual tidak ditulis sebagai dokumen teknis, melainkan sebagai petunjuk praktis: menu apa yang perlu dibuka, data apa yang perlu diperiksa, langkah apa yang perlu dilakukan, dan hal apa yang harus diperhatikan sebelum menyimpan atau memakai data untuk peninjauan.</p>
      <p>Gunakan manual ini dari bagian awal bila baru pertama kali memakai PRISM. Untuk pekerjaan harian, pengguna dapat langsung membuka bagian sesuai kebutuhan, misalnya filter Dashboard, halaman Proyek, pengisian Blue Book, impor file Excel, pengaturan hak akses, atau penanganan pesan sistem.</p>
    </div>
    <div class="intro-grid">
      <div class="panel teal">
        <h3>Tujuan aplikasi</h3>
        <p>PRISM mencatat alur perencanaan pinjaman luar negeri dari usulan awal, pematangan prioritas, penetapan kegiatan, sampai perjanjian pinjaman. Sistem membantu menjaga jejak perubahan data, menyaring portofolio, dan mengatur hak akses per modul.</p>
      </div>
      <div class="panel gold">
        <h3>Urutan kerja yang disarankan</h3>
        <p>Pastikan master data lengkap, isi dokumen perencanaan secara berurutan, gunakan halaman Proyek untuk pemeriksaan lintas modul, lalu pakai Dashboard, Perjalanan Proyek, dan Sebaran Wilayah untuk peninjauan portofolio.</p>
      </div>
    </div>
    <div class="flow">
      ${stageFlowCards}
    </div>
    <div class="role-grid">
      <div class="panel role-card">
        <h3>ADMIN</h3>
        <p>Memiliki akses penuh untuk mengelola pengguna, hak akses, master data, impor data, serta koreksi data lintas modul bila diperlukan.</p>
      </div>
      <div class="panel role-card">
        <h3>STAFF</h3>
        <p>Mengakses modul sesuai hak akses yang diberikan ADMIN. Jika hak akses belum diberikan, menu terkait tidak akan muncul atau tidak dapat digunakan.</p>
      </div>
    </div>
    <div class="callout"><strong>Prinsip penting:</strong> Blue Book dan Green Book menyimpan data per dokumen dan per revisi. Daftar Kegiatan dan Perjanjian Pinjaman tetap memakai data proyek yang dipilih saat dibuat, sehingga tidak otomatis berubah ketika ada revisi baru.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Orientasi</span></div>
  </section>

  <section id="section-daftar-isi" class="page">
    <p class="section-label">02 - Daftar Isi</p>
    <h2>Struktur Manual</h2>
    <div class="toc-grid">
      ${tocEntries
        .map(
          ([id, title, note]) => {
            const pageNumber = tocPageNumbers[id] ? String(tocPageNumbers[id]) : ''
            return `<div class="toc-item"><a class="toc-link" href="#${escapeHtml(id)}"><span class="toc-main"><strong>${escapeHtml(title)}</strong><span>${escapeHtml(note)}</span></span><span class="toc-page">${escapeHtml(pageNumber)}</span></a></div>`
          },
        )
        .join('')}
    </div>
    <div class="footer"><span>PRISM Manual Book</span><span>Daftar Isi</span></div>
  </section>

  <section id="section-akses" class="page">
    <p class="section-label">03 - Akses</p>
    <h2>Login dan Navigasi</h2>
    <div class="two-col">
      <div>
        <h3>Masuk ke aplikasi</h3>
        ${list([
          'Buka alamat aplikasi PRISM pada browser.',
          'Masukkan username dan kata sandi.',
          'Klik Masuk.',
          'Jika berhasil, sistem mengarahkan pengguna ke halaman default sesuai hak akses.',
        ])}
        <div class="callout">Jika login gagal, periksa kembali username, kata sandi, dan status akun. Untuk STAFF, hubungi ADMIN bila menu yang dibutuhkan tidak muncul.</div>
      </div>
      <div>
        ${screenshotFigure('login', 'Halaman login PRISM', 'Tampilan login PRISM dengan isian username dan kata sandi.')}
      </div>
    </div>
    <div class="two-col mt">
      <div class="panel">
        <h3>Sidebar</h3>
        <p>Sidebar memuat menu utama: Dashboard, Project, Perjalanan Proyek, Sebaran Wilayah, Dokumen Perencanaan, Referensi, dan Akses Admin. Gunakan kotak Cari menu untuk menemukan modul lebih cepat.</p>
      </div>
      <div class="panel">
        <h3>Hak akses</h3>
        <p>Menu muncul sesuai peran dan hak akses. ADMIN melihat seluruh menu admin. STAFF hanya melihat modul yang diberi hak oleh ADMIN.</p>
      </div>
    </div>
    <div class="footer"><span>PRISM Manual Book</span><span>Login dan Navigasi</span></div>
  </section>

  <section id="section-dashboard" class="page">
    <p class="section-label">04 - Dashboard dan Analitik</p>
    <h2>Memantau Portofolio</h2>
    ${screenshotFigure('dashboard', 'Dashboard PRISM', 'Dashboard menampilkan funnel tahapan dan panel distribusi portofolio.')}
    <div class="two-col">
      <div class="panel">
        <h3>Membaca funnel</h3>
        ${list([
          'Blue Book menunjukkan basis usulan proyek.',
          'Green Book menunjukkan proyek yang sudah masuk prioritas pendanaan.',
          'Daftar Kegiatan menunjukkan proyek dalam surat kegiatan.',
          'Perjanjian Pinjaman menunjukkan proyek yang sudah memiliki kontrak pinjaman.',
        ])}
      </div>
      <div class="panel">
        <h3>Aksi dari dashboard</h3>
        ${list([
          'Gunakan filter periode untuk membatasi angka portofolio.',
          'Klik panel untuk membuka halaman Proyek dengan filter relevan.',
          'Bandingkan gap antar tahap untuk menentukan tindak lanjut.',
        ])}
      </div>
    </div>
    <div class="callout"><strong>Filter periode:</strong> pilihan periode di Dashboard memengaruhi funnel, panel distribusi, dan tautan lanjutan ke halaman Proyek atau Sebaran Wilayah. Jika semua periode dipilih, Dashboard membaca seluruh data yang tersedia.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Dashboard</span></div>
  </section>

  <section id="section-detail-dashboard" class="page">
    <p class="section-label">05 - Detail Dashboard</p>
    <h2>Cara Membaca Panel Dashboard</h2>
    <p class="mb">Dashboard dipakai sebagai titik awal membaca portofolio. Baca angka sebagai ringkasan, lalu buka halaman Proyek, Perjalanan Proyek, atau Sebaran Wilayah untuk memeriksa daftar proyek pembentuk angka tersebut.</p>
    <div class="detail-grid">
      ${dashboardDetailCards}
    </div>
    <div class="callout"><strong>Prinsip membaca:</strong> angka dashboard mengikuti data yang sudah tersimpan dan filter periode yang sedang aktif. Jika hasil terlihat tidak sesuai, cek filter periode, kelengkapan data proyek, dan tahap dokumen yang sudah dicatat.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Detail Dashboard</span></div>
  </section>

  <section id="section-modul-peninjauan" class="page">
    <p class="section-label">06 - Modul Peninjauan</p>
    <h2>Proyek, Perjalanan, dan Sebaran Wilayah</h2>
    ${moduleCards}
    <div class="footer"><span>PRISM Manual Book</span><span>Modul Peninjauan</span></div>
  </section>

  <section id="section-proyek-ekspor" class="page">
    <p class="section-label">07 - Proyek dan Ekspor</p>
    <h2>Menggunakan Tabel Proyek</h2>
    <p class="mb">Halaman Proyek adalah daftar gabungan yang membantu pengguna mencari proyek lintas dokumen. Gunakan halaman ini ketika ingin memeriksa data proyek, menelusuri tahapnya, atau menyiapkan file Excel sesuai filter aktif.</p>
    <table>
      <thead><tr><th>Area</th><th>Fungsi</th><th>Cara pakai</th></tr></thead>
      <tbody>${projectMasterRows}</tbody>
    </table>
    <div class="callout"><strong>Export Excel:</strong> hasil unduhan mengikuti pencarian, filter aktif, dan cakupan data yang sedang dipakai. Selalu cek label filter aktif sebelum membagikan file hasil unduhan.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Proyek</span></div>
  </section>

  <section id="section-penggunaan-filter" class="page">
    <p class="section-label">08 - Penggunaan Filter</p>
    <h2>Filter Dashboard dan Proyek</h2>
    <p class="mb">Filter membantu pengguna mempersempit data agar analisis lebih tepat. Gunakan pencarian untuk menemukan data cepat, lalu gunakan Filter lanjutan bila perlu membatasi data berdasarkan tahap, status, pemberi pinjaman, wilayah, atau periode.</p>
    ${filterGuidePrimaryCards}
    <div class="footer"><span>PRISM Manual Book</span><span>Penggunaan Filter</span></div>
  </section>

  <section id="section-filter-lanjutan" class="page">
    <p class="section-label">09 - Filter Lanjutan Peninjauan</p>
    <h2>Filter Perjalanan Proyek dan Sebaran Wilayah</h2>
    <p class="mb">Perjalanan Proyek memakai pencarian proyek sebagai filter utama. Sebaran Wilayah memakai filter peta dan daftar proyek agar pengguna dapat membaca konsentrasi proyek per wilayah.</p>
    ${filterGuideSpatialCards}
    <div class="footer"><span>PRISM Manual Book</span><span>Filter Lanjutan Peninjauan</span></div>
  </section>

  <section id="section-tampilan-peninjauan" class="page">
    <p class="section-label">10 - Tampilan Peninjauan</p>
    <h2>Contoh Layar Peninjauan</h2>
    <div class="screenshot-grid">
      ${screenshotFigure('project-master', 'Halaman Proyek PRISM', 'Halaman Proyek dengan pencarian, penyaring data, dan unduh Excel.')}
      ${screenshotFigure('spatial', 'Sebaran Wilayah PRISM', 'Sebaran Wilayah dengan peta dan daftar proyek fokus.')}
    </div>
    <div class="callout"><strong>Tips peninjauan:</strong> Mulai dari Dashboard untuk gambaran umum, turun ke halaman Proyek untuk daftar proyek, lalu gunakan Perjalanan Proyek atau Sebaran Wilayah untuk konteks proyek dan wilayah.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Tampilan Peninjauan</span></div>
  </section>

  <section id="section-master-data" class="page">
    <p class="section-label">11 - Master Data</p>
    <h2>Referensi Sebelum Pengisian Dokumen</h2>
    <p class="mb">Master data menjadi sumber pilihan pada formulir dokumen. Lengkapi referensi sebelum impor atau pengisian manual agar pemeriksaan data berjalan lancar.</p>
    <table>
      <thead><tr><th>Master</th><th>Fungsi</th></tr></thead>
      <tbody>${masterRows}</tbody>
    </table>
    <div class="callout">Perubahan master data dapat memengaruhi pilihan pada formulir, pemeriksaan file impor, perhitungan kurs, dan konsistensi laporan. Lakukan perubahan master data dengan rujukan dokumen yang jelas.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Master Data</span></div>
  </section>

  <section id="section-panduan-referensi" class="page">
    <p class="section-label">12 - Panduan Referensi</p>
    <h2>Cara Menggunakan Master Data</h2>
    <p class="mb">Master data adalah daftar acuan yang dipakai berulang pada formulir dan impor. Isi dengan rapi terlebih dahulu agar pengguna tidak perlu mengetik bebas dan agar data mudah disaring.</p>
    <div class="detail-grid">
      ${masterDataGuideCards}
    </div>
    <div class="footer"><span>PRISM Manual Book</span><span>Panduan Referensi</span></div>
  </section>

  <section id="section-dokumen-perencanaan" class="page">
    <p class="section-label">13 - Dokumen Perencanaan</p>
    <h2>Dasar Istilah dan Alur Pengisian Dokumen</h2>
    <div class="intro-copy">
      <p>Manual ini memakai nama menu yang ada di PRISM, lalu memadankannya dengan istilah perencanaan pinjaman luar negeri pada PP Nomor 10 Tahun 2011 dan Permen PPN/Bappenas Nomor 4 Tahun 2011.</p>
      <p>Padanan ini membantu pengguna membaca aplikasi secara lebih utuh: Blue Book berada pada tahap rencana jangka menengah, Green Book berada pada tahap prioritas tahunan, Daftar Kegiatan berada pada tahap siap usul atau siap runding, dan Perjanjian Pinjaman berada pada tahap pengikatan tertulis.</p>
    </div>
    <table class="compact-table legal-table mb">
      <thead><tr><th>Istilah</th><th>Makna dalam dasar hukum</th><th>Penggunaan di PRISM</th></tr></thead>
      <tbody>${legalRows}</tbody>
    </table>
    <div class="callout"><strong>Catatan pemakaian:</strong> PRISM tidak mengganti dokumen hukum. Aplikasi membantu mencatat data operasional agar alur dari rencana, prioritas, kesiapan, sampai perjanjian pinjaman dapat ditelusuri dengan bahasa yang konsisten.</div>
    <h3>Urutan Kerja di Aplikasi</h3>
    ${planningCards}
    <div class="footer"><span>PRISM Manual Book</span><span>Dokumen Perencanaan</span></div>
  </section>

  <section id="section-panduan-formulir" class="page">
    <p class="section-label">14 - Panduan Isi Formulir</p>
    <h2>Isian Penting pada Formulir Dokumen</h2>
    <p class="mb">Bagian ini menjelaskan isian formulir dengan bahasa operasional. Gunakan sebagai daftar periksa sebelum menyimpan data atau sebelum memperbaiki data yang berasal dari impor.</p>
    ${fieldGuideCards}
    <div class="footer"><span>PRISM Manual Book</span><span>Panduan Isi Formulir</span></div>
  </section>

  <section id="section-contoh-layar-dokumen" class="page">
    <p class="section-label">15 - Contoh Layar Dokumen</p>
    <h2>Daftar Dokumen Perencanaan</h2>
    <div class="screenshot-grid">
      ${screenshotFigure('blue-book', 'Daftar Blue Book', 'Daftar Blue Book sebagai pintu masuk proyek indikatif.')}
      ${screenshotFigure('green-book', 'Daftar Green Book', 'Daftar Green Book untuk proyek prioritas pendanaan.')}
      ${screenshotFigure('daftar-kegiatan', 'Daftar Kegiatan', 'Daftar Kegiatan untuk surat dan proyek pembiayaan.')}
      ${screenshotFigure('loan-agreements', 'Perjanjian Pinjaman', 'Perjanjian Pinjaman untuk kontrak dan kinerja pinjaman.')}
    </div>
    <div class="footer"><span>PRISM Manual Book</span><span>Layar Dokumen</span></div>
  </section>

  ${dataManagementPages}

  <section id="section-alur-operasional" class="page">
    <p class="section-label">19 - Alur Operasional Khusus</p>
    <h2>Detail Penggunaan Fitur yang Sering Terlewat</h2>
    <p class="mb">Bagian ini menjelaskan alur layar yang tidak selalu terlihat dari daftar menu, tetapi penting dalam pekerjaan harian. Gunakan sebagai panduan saat membawa proyek antar dokumen, mengisi kurs, menyiapkan draft perjanjian, mengatur kolom, dan mengelola data hierarki.</p>
    <div class="detail-grid">
      ${operationalGuideCards}
    </div>
    <div class="callout"><strong>Prinsip kerja:</strong> bila fitur membawa data dari dokumen lain, gunakan hasilnya sebagai awal pengisian. Tetap periksa ulang isi dokumen tujuan karena kebutuhan setiap tahap tidak sama.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Alur Operasional Khusus</span></div>
  </section>

  <section id="section-pesan-sistem" class="page">
    <p class="section-label">20 - Pesan Sistem dan Kondisi Layar</p>
    <h2>Membaca Status, Validasi, dan Hasil Preview</h2>
    <p class="mb">PRISM menampilkan pesan dan status untuk membantu pengguna memahami apa yang sedang terjadi. Jangan langsung mengulang aksi berkali-kali; baca dulu pesan yang muncul, cek filter atau hak akses, lalu lakukan perbaikan sesuai konteks.</p>
    <table>
      <thead><tr><th>Kondisi</th><th>Arti sederhananya</th><th>Yang perlu dilakukan</th></tr></thead>
      <tbody>${systemStateRows}</tbody>
    </table>
    <div class="callout"><strong>Khusus impor:</strong> create berarti siap dibuat, skip berarti dilewati karena sudah ada atau tidak perlu dibuat ulang, dan failed berarti file Excel harus diperbaiki sebelum eksekusi.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Pesan Sistem</span></div>
  </section>

  <section id="section-riwayat-revisi" class="page">
    <p class="section-label">21 - Riwayat Revisi</p>
    <h2>Membedakan Data Lama dan Data Terbaru</h2>
    <p class="mb">PRISM menyimpan jejak dokumen per revisi. Dengan cara ini, pengguna dapat melihat data yang dipakai pada saat dokumen dibuat tanpa kehilangan riwayat ketika ada versi baru.</p>
    <table>
      <thead><tr><th>Area</th><th>Arti sederhananya</th><th>Cara memakai</th></tr></thead>
      <tbody>${revisionHistoryRows}</tbody>
    </table>
    <div class="callout"><strong>Bahasa sederhananya:</strong> revisi baru tidak menghapus cerita lama. Data lama tetap dapat dibaca sebagai arsip, sedangkan versi terbaru dipakai untuk keputusan berikutnya.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Riwayat Revisi</span></div>
  </section>

  <section id="section-impor-data" class="page">
    <p class="section-label">22 - Impor Data</p>
    <h2>Impor File Excel</h2>
    <div class="two-col">
      <div>
        <p class="mb">Impor Data dipakai untuk memasukkan banyak data sekaligus melalui file Excel. Setiap jenis impor memiliki template resmi dari sistem.</p>
        <div class="workflow-grid">
          <div class="panel"><img class="workflow-step-shot" src="assets/import-data.png" alt="Tampilan tombol template impor" /><h3>1</h3><p>Unduh template dari tombol Template.</p></div>
          <div class="panel"><img class="workflow-step-shot" src="assets/import-data.png" alt="Tampilan file impor" /><h3>2</h3><p>Isi file Excel sesuai sheet dan aturan kolom.</p></div>
          <div class="panel"><img class="workflow-step-shot" src="assets/import-data.png" alt="Tampilan preview impor" /><h3>3</h3><p>Pilih file dan jalankan Preview untuk memeriksa isinya.</p></div>
          <div class="panel"><img class="workflow-step-shot" src="assets/import-data.png" alt="Tampilan eksekusi impor" /><h3>4</h3><p>Jalankan eksekusi hanya setelah Preview tidak menampilkan baris gagal.</p></div>
        </div>
      </div>
      <div>
        ${screenshotFigure('import-data', 'Impor Data PRISM', 'Impor Data menyediakan template, pemeriksaan awal, dan eksekusi file Excel.')}
      </div>
    </div>
    <table class="mt">
      <thead><tr><th>Jenis impor</th><th>Cakupan</th></tr></thead>
      <tbody>${importRows}</tbody>
    </table>
    <div class="callout"><strong>Aturan kerja:</strong> Jangan eksekusi file yang masih memiliki baris gagal. Perbaiki data di file Excel, unggah ulang, lalu ulangi Preview.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Impor Data</span></div>
  </section>

  <section id="section-detail-file-excel" class="page">
    <p class="section-label">23 - Detail File Excel Impor</p>
    <h2>Sheet dan Pemeriksaan File Excel</h2>
    <p class="mb">Setiap file Excel memiliki sheet Panduan dan sheet data. Jangan mengubah nama sheet atau header kolom pada template karena Preview membaca struktur tersebut untuk memeriksa isi file.</p>
    ${importWorkbookCards}
    <div class="callout"><strong>Preview dulu:</strong> Preview tidak menyimpan data. Gunakan hasil Preview untuk memperbaiki file sampai tidak ada baris failed, baru jalankan eksekusi impor.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Detail File Excel Impor</span></div>
  </section>

  <section id="section-admin" class="page">
    <p class="section-label">24 - Admin</p>
    <h2>Manajemen Pengguna dan Hak Akses</h2>
    <div class="role-grid">
      <div class="panel">
        <h3>Mengelola pengguna</h3>
        ${list([
          'Buka menu Pengguna pada grup Akses Admin.',
          'Tambah pengguna baru atau edit pengguna yang sudah ada.',
          'Tetapkan peran ADMIN atau STAFF sesuai kebutuhan organisasi.',
          'Nonaktifkan atau hapus akses sesuai kebijakan operasional.',
        ])}
      </div>
      <div class="panel">
        <h3>Mengelola hak akses STAFF</h3>
        ${list([
          'Buka halaman hak akses pengguna.',
          'Aktifkan hak akses per modul dan aksi.',
          'Simpan perubahan hak akses.',
          'Minta pengguna login ulang bila menu belum berubah.',
        ])}
      </div>
    </div>
    <div class="callout">ADMIN otomatis memiliki akses penuh. STAFF hanya dapat membuka menu dan menjalankan aksi yang sudah diberi hak oleh ADMIN.</div>
    <h3 class="mt">Checklist sebelum go-live data</h3>
    <div class="checklist">
      ${[
        ['Master data', 'Negara, pemberi pinjaman, instansi, wilayah, periode, mata uang, kurs, program, dan prioritas sudah lengkap.'],
        ['Dokumen', 'Blue Book, Green Book, Daftar Kegiatan, dan Perjanjian Pinjaman mengikuti urutan bisnis.'],
        ['Impor', 'Semua file Excel sudah melewati Preview tanpa baris gagal sebelum eksekusi.'],
        ['Hak akses', 'Setiap STAFF hanya memiliki modul yang sesuai tugasnya.'],
        ['Peninjauan', 'Dashboard, halaman Proyek, Perjalanan Proyek, dan Sebaran Wilayah sudah dicek setelah pengisian data besar.'],
        ['Revisi', 'Data versi lama dan versi terbaru dibedakan sebelum membuat keputusan lanjutan.'],
      ].map(([title, note]) => `<div class="check"><strong>${escapeHtml(title)}</strong><p>${escapeHtml(note)}</p></div>`).join('')}
    </div>
    <div class="footer"><span>PRISM Manual Book</span><span>Admin dan Checklist</span></div>
  </section>

  <section id="section-hak-akses-detail" class="page">
    <p class="section-label">25 - Hak Akses Detail</p>
    <h2>Alur Pengguna dan Hak Akses</h2>
    <p class="mb">Hak akses menentukan menu dan aksi yang dapat digunakan. ADMIN memiliki akses penuh, sedangkan STAFF hanya mendapat akses sesuai modul yang diberikan.</p>
    <div class="detail-grid">
      ${accessGuideCards}
    </div>
    <div class="callout"><strong>Prinsip akses:</strong> berikan hak sesuai tugas pengguna. Jika pengguna hanya perlu melihat data, jangan berikan hak ubah atau hapus tanpa kebutuhan jelas.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Hak Akses Detail</span></div>
  </section>

  <section id="section-troubleshooting" class="page">
    <p class="section-label">26 - Troubleshooting</p>
    <h2>Masalah Umum dan Cara Menanganinya</h2>
    <p class="mb">Gunakan tabel ini sebagai panduan awal sebelum meminta bantuan teknis. Sebagian besar masalah harian dapat ditelusuri dari hak akses, filter aktif, kelengkapan master data, atau hasil Preview impor.</p>
    <table>
      <thead><tr><th>Gejala</th><th>Kemungkinan penyebab</th><th>Langkah yang disarankan</th></tr></thead>
      <tbody>${troubleshootingRows}</tbody>
    </table>
    <div class="callout"><strong>Jika masih gagal:</strong> catat menu yang dibuka, filter yang aktif, nama file impor bila ada, dan pesan error yang muncul. Informasi ini membantu ADMIN menelusuri masalah lebih cepat.</div>
    <div class="footer"><span>PRISM Manual Book</span><span>Troubleshooting</span></div>
  </section>
</body>
</html>`

  await fs.writeFile(htmlPath, html, 'utf8')
}

async function exportPdf() {
  const browser = await launchBrowser()
  const page = await browser.newPage({ viewport: { width: 1440, height: 1200 } })
  try {
    await page.goto(pathToFileURL(htmlPath).href, { waitUntil: 'load' })
    await page.pdf({
      path: pdfPath,
      format: 'A4',
      printBackground: true,
      preferCSSPageSize: true,
      displayHeaderFooter: false,
      margin: { top: '0', right: '0', bottom: '0', left: '0' },
    })
  } finally {
    await browser.close()
  }
}

async function extractSectionPageNumbers() {
  const script = String.raw`
import json
import re
import sys
from pypdf import PdfReader

pdf_path = sys.argv[1]
markers = json.loads(sys.argv[2])

def normalize(value):
    return re.sub(r"\s+", " ", value or "").strip().upper()

reader = PdfReader(pdf_path)
page_texts = [normalize(page.extract_text() or "") for page in reader.pages]
result = {}

for marker in markers:
    needle = normalize(marker["marker"])
    for index, text in enumerate(page_texts, start=1):
        if needle in text:
            result[marker["id"]] = index
            break

print(json.dumps(result))
`

  const { stdout } = await execFileAsync(
    'python',
    ['-c', script, pdfPath, JSON.stringify(sectionPageMarkers)],
    { maxBuffer: 10 * 1024 * 1024 },
  )
  return JSON.parse(stdout)
}

function samePageNumbers(a, b) {
  return tocEntries.every(([id]) => String(a[id] ?? '') === String(b[id] ?? ''))
}

async function addPdfPageNumbers() {
  const bytes = await fs.readFile(pdfPath)
  const pdf = await PDFDocument.load(bytes)
  const font = await pdf.embedFont(StandardFonts.Helvetica)
  const pages = pdf.getPages()
  const fontSize = 8.5
  const color = rgb(0.32, 0.39, 0.45)

  pages.forEach((page, index) => {
    if (index === 0) return
    const pageNumber = index + 1
    const text = `Halaman ${pageNumber}`
    const { width } = page.getSize()
    const textWidth = font.widthOfTextAtSize(text, fontSize)
    page.drawText(text, {
      x: width - 36 - textWidth,
      y: 18,
      size: fontSize,
      font,
      color,
    })
  })

  await fs.writeFile(pdfPath, await pdf.save())
}

async function getPdfInfo() {
  const bytes = await fs.readFile(pdfPath)
  const pdf = await PDFDocument.load(bytes)
  return {
    pages: pdf.getPageCount(),
    bytes: bytes.length,
  }
}

async function main() {
  await fs.mkdir(assetsDir, { recursive: true })
  await captureScreenshots()
  await buildHtml()
  if (htmlOnly) {
    console.log(JSON.stringify({ htmlPath, mode: 'html-only' }, null, 2))
    return
  }
  await exportPdf()
  let tocPageNumbers = await extractSectionPageNumbers()
  await buildHtml(tocPageNumbers)
  await exportPdf()
  const finalTocPageNumbers = await extractSectionPageNumbers()
  if (!samePageNumbers(tocPageNumbers, finalTocPageNumbers)) {
    tocPageNumbers = finalTocPageNumbers
    await buildHtml(tocPageNumbers)
    await exportPdf()
  }
  await addPdfPageNumbers()
  const info = await getPdfInfo()
  console.log(JSON.stringify({ htmlPath, pdfPath, ...info }, null, 2))
}

main().catch((error) => {
  console.error(error)
  process.exitCode = 1
})
