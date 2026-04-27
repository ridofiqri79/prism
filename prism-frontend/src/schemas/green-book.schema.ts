import { z } from 'zod'

const optionalText = z
  .string()
  .optional()
  .transform((value) => value?.trim() || undefined)

export const greenBookSchema = z.object({
  publish_year: z.number().int().min(1900, 'Tahun terbit wajib diisi'),
  revision_number: z.number().int().min(0, 'Revisi minimal 0'),
})

export const gbProjectSchema = z.object({
  program_title_id: z.string().uuid('Program Title wajib dipilih'),
  gb_code: z.string().min(1, 'GB Code wajib diisi'),
  project_name: z.string().min(1, 'Nama proyek wajib diisi'),
  duration: optionalText,
  objective: optionalText,
  scope_of_project: optionalText,
  bb_project_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 BB Project'),
  executing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Executing Agency'),
  implementing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Implementing Agency'),
  location_ids: z.array(z.string().uuid()).min(1, 'Lokasi wajib dipilih'),
})

export type GBProjectBaseFormValues = z.infer<typeof gbProjectSchema>

