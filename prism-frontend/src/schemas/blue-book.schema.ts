import { z } from 'zod'

const optionalNumber = z.preprocess(
  (value) => (value === '' || value === null ? undefined : value),
  z.number().int().optional(),
)

const optionalText = z
  .string()
  .optional()
  .transform((value) => value?.trim() || undefined)

export const blueBookSchema = z.object({
  period_id: z.string().uuid('Period wajib dipilih'),
  publish_date: z.string().min(1, 'Tanggal terbit wajib diisi'),
  revision_number: z.number().int().min(0, 'Revisi minimal 0'),
  revision_year: optionalNumber,
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
  program_title_id: z.string().uuid('Program Title wajib dipilih'),
  bappenas_partner_id: z.string().uuid('Bappenas Partner wajib dipilih'),
  bb_code: z.string().min(1, 'BB Code wajib diisi'),
  project_name: z.string().min(1, 'Nama proyek wajib diisi'),
  duration: optionalText,
  objective: optionalText,
  scope_of_work: optionalText,
  outputs: optionalText,
  outcomes: optionalText,
  executing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Executing Agency'),
  implementing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Implementing Agency'),
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

