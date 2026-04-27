import { z } from 'zod'

const optionalUuid = z
  .string()
  .uuid()
  .optional()
  .or(z.literal('').transform(() => undefined))

export const countrySchema = z.object({
  name: z.string().min(1, 'Nama wajib diisi'),
  code: z
    .string()
    .length(3, 'Kode harus 3 karakter')
    .transform((value) => value.toUpperCase()),
})

export const lenderSchema = z
  .object({
    name: z.string().min(1, 'Nama wajib diisi'),
    short_name: z.string().optional(),
    type: z.enum(['Bilateral', 'Multilateral', 'KSA']),
    country_id: optionalUuid,
  })
  .refine(
    (data) => {
      if (data.type !== 'Multilateral') return Boolean(data.country_id)
      return true
    },
    { message: 'Negara wajib diisi', path: ['country_id'] },
  )

export const institutionSchema = z.object({
  name: z.string().min(1, 'Nama wajib diisi'),
  short_name: z.string().optional(),
  level: z.enum([
    'Kementerian/Badan/Lembaga',
    'Eselon I',
    'BUMN',
    'Pemerintah Daerah',
    'BUMD',
    'Lainnya',
  ]),
  parent_id: optionalUuid,
})

export const regionSchema = z.object({
  code: z.string().min(1, 'Kode wajib diisi').max(10, 'Kode maksimal 10 karakter'),
  name: z.string().min(1, 'Nama wajib diisi'),
  type: z.enum(['COUNTRY', 'PROVINCE', 'CITY']),
  parent_code: z.string().optional(),
})

export const periodSchema = z
  .object({
    name: z.string().min(1, 'Nama wajib diisi'),
    year_start: z.number().int().min(2000, 'Tahun awal minimal 2000'),
    year_end: z.number().int().min(2000, 'Tahun akhir minimal 2000'),
  })
  .refine((data) => data.year_end > data.year_start, {
    message: 'Tahun akhir harus lebih besar dari tahun awal',
    path: ['year_end'],
  })

export const nationalPrioritySchema = z.object({
  period_id: z.string().uuid('Period wajib dipilih'),
  title: z.string().min(1, 'Judul wajib diisi'),
})

export const programTitleSchema = z.object({
  title: z.string().min(1, 'Judul wajib diisi'),
  parent_id: optionalUuid,
})

export const bappenasPartnerSchema = z
  .object({
    name: z.string().min(1, 'Nama wajib diisi'),
    level: z.enum(['Eselon I', 'Eselon II']),
    parent_id: optionalUuid,
  })
  .refine(
    (data) => {
      if (data.level === 'Eselon II') return Boolean(data.parent_id)
      return true
    },
    { message: 'Eselon I parent wajib diisi untuk Eselon II', path: ['parent_id'] },
  )
