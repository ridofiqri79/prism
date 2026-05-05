import { z } from 'zod'

const optionalNumber = z.preprocess(
  (value) => (value === '' || value === null ? undefined : value),
  z.number().int().optional(),
)

const optionalText = z
  .string()
  .optional()
  .transform((value) => value?.trim() || undefined)

const optionalUUID = (message: string) =>
  z.preprocess(
    (value) => (value === '' || value === null ? undefined : value),
    z.string().uuid(message).optional(),
  )

const optionalDurationMonths = z
  .number()
  .int('Durasi harus berupa bulan bulat')
  .positive('Durasi harus lebih dari 0 bulan')
  .optional()
  .nullable()

export const blueBookSchema = z.object({
  period_id: z.string().uuid('Periode wajib dipilih'),
  replaces_blue_book_id: optionalUUID('Blue Book sumber tidak valid'),
  publish_date: z.string().min(1, 'Tanggal terbit wajib diisi'),
  revision_number: z.number().int().min(0, 'Revisi minimal 0'),
  revision_year: optionalNumber,
  status: z.enum(['active', 'superseded']),
})

export const importBBProjectsFromBlueBookSchema = z.object({
  source_blue_book_id: z.string().uuid('Blue Book sumber wajib dipilih'),
  project_ids: z.array(z.string().uuid('Project Blue Book tidak valid')).min(1, 'Minimal satu Project Blue Book dipilih'),
})

export const projectCostSchema = z.object({
  funding_type: z.enum(['Foreign', 'Counterpart']),
  funding_category: z.string().min(1, 'Kategori wajib dipilih'),
  amount_usd: z.number().min(0, 'Nilai tidak boleh negatif'),
})

export const lenderIndicationSchema = z.object({
  lender_id: z.string().uuid('Lender wajib dipilih'),
  remarks: optionalText,
})

export const bbProjectSchema = z.object({
  project_identity_id: optionalUUID('Project identity tidak valid'),
  program_title_id: z.string().uuid('Judul program wajib dipilih'),
  bappenas_partner_ids: z.array(z.string().uuid('Mitra Kerja Bappenas tidak valid')),
  bb_code: z.string().min(1, 'Kode Blue Book wajib diisi'),
  project_name: z.string().min(1, 'Nama proyek wajib diisi'),
  duration: optionalDurationMonths,
  objective: optionalText,
  scope_of_work: optionalText,
  outputs: optionalText,
  outcomes: optionalText,
  executing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 executing agency'),
  implementing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 implementing agency'),
  location_ids: z.array(z.string().uuid()).min(1, 'Lokasi wajib dipilih'),
  national_priority_ids: z.array(z.string().uuid()),
})

export const loiSchema = z.object({
  lender_id: z.string().uuid('Lender wajib dipilih'),
  subject: z.string().min(1, 'Perihal wajib diisi'),
  date: z.string().min(1, 'Tanggal wajib diisi'),
  letter_number: optionalText,
})

export type BlueBookFormValues = z.infer<typeof blueBookSchema>
export type BBProjectBaseFormValues = z.infer<typeof bbProjectSchema>
export type LoIFormValues = z.infer<typeof loiSchema>
