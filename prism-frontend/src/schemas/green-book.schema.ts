import { z } from 'zod'

const optionalText = z
  .string()
  .optional()
  .transform((value) => value?.trim() || undefined)

const optionalDurationMonths = z
  .number()
  .int('Durasi harus berupa bulan bulat')
  .positive('Durasi harus lebih dari 0 bulan')
  .optional()
  .nullable()

export const greenBookSchema = z.object({
  publish_year: z.number().int().min(1900, 'Tahun terbit wajib diisi'),
  revision_number: z.number().int().min(0, 'Revisi minimal 0'),
  status: z.enum(['active', 'superseded']),
})

export const importGBProjectsFromGreenBookSchema = z.object({
  source_green_book_id: z.string().uuid('Green Book sumber wajib dipilih'),
  project_ids: z.array(z.string().uuid('Project Green Book tidak valid')).min(1, 'Minimal satu Project Green Book dipilih'),
})

export const gbProjectSchema = z.object({
  program_title_id: z.string().uuid('Judul program wajib dipilih'),
  gb_code: z.string().min(1, 'Kode Green Book wajib diisi'),
  project_name: z.string().min(1, 'Nama proyek wajib diisi'),
  duration: optionalDurationMonths,
  objective: optionalText,
  scope_of_project: optionalText,
  bb_project_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Proyek Blue Book'),
  bappenas_partner_ids: z.array(z.string().uuid('Mitra Kerja Bappenas tidak valid')),
  executing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 executing agency'),
  implementing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 implementing agency'),
  location_ids: z.array(z.string().uuid()).min(1, 'Lokasi wajib dipilih'),
})

export type GBProjectBaseFormValues = z.infer<typeof gbProjectSchema>
