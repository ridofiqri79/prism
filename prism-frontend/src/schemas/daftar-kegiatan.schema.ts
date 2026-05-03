import { z } from 'zod'

const optionalUuid = z.string().uuid().or(z.literal('')).optional().nullable()
const money = z.number().min(0, 'Nilai tidak boleh negatif')
const optionalDurationMonths = z
  .number()
  .int('Durasi harus berupa bulan bulat')
  .positive('Durasi harus lebih dari 0 bulan')
  .optional()
  .nullable()

export const daftarKegiatanSchema = z.object({
  subject: z.string().min(1, 'Perihal wajib diisi'),
  date: z.string().min(1, 'Tanggal wajib diisi'),
  letter_number: z.string().optional().nullable(),
})

export const dkFinancingDetailSchema = z.object({
  lender_id: z.string().uuid('Lender wajib dipilih'),
  currency: z.string().min(3, 'Mata uang wajib diisi').max(3, 'Gunakan kode ISO 4217'),
  amount_original: money,
  grant_original: money,
  counterpart_original: money,
  amount_usd: money,
  grant_usd: money,
  counterpart_usd: money,
  remarks: z.string().optional().nullable(),
})

export const dkLoanAllocationSchema = z.object({
  institution_id: z.string().uuid('Instansi wajib dipilih'),
  currency: z.string().min(3, 'Mata uang wajib diisi').max(3, 'Gunakan kode ISO 4217'),
  amount_original: money,
  grant_original: money,
  counterpart_original: money,
  amount_usd: money,
  grant_usd: money,
  counterpart_usd: money,
  remarks: z.string().optional().nullable(),
})

export const dkActivityDetailSchema = z.object({
  activity_number: z.number().int().positive(),
  activity_name: z.string().min(1, 'Nama aktivitas wajib diisi'),
})

export const dkProjectSchema = z.object({
  program_title_id: optionalUuid,
  institution_id: z.string().uuid('Executing agency wajib dipilih'),
  project_name: z.string().min(1, 'Nama proyek wajib diisi'),
  duration: optionalDurationMonths,
  objectives: z.string().optional().nullable(),
  gb_project_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Proyek Green Book'),
  bappenas_partner_ids: z.array(z.string().uuid('Mitra Kerja Bappenas tidak valid')),
  location_ids: z.array(z.string().uuid()).min(1, 'Lokasi wajib dipilih'),
  financing_details: z.array(dkFinancingDetailSchema).min(1, 'Minimal 1 rincian pembiayaan'),
  loan_allocations: z.array(dkLoanAllocationSchema).min(1, 'Minimal 1 alokasi pinjaman'),
  activity_details: z.array(dkActivityDetailSchema).min(1, 'Minimal 1 rincian kegiatan'),
})
