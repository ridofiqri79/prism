import { z } from 'zod'

export const quarterSchema = z.enum(['TW1', 'TW2', 'TW3', 'TW4'])

export const monitoringKomponenSchema = z.object({
  id: z.string().uuid().optional(),
  component_name: z.string().min(1, 'Nama komponen wajib diisi'),
  planned_la: z.number().min(0, 'Rencana LA tidak boleh negatif'),
  planned_usd: z.number().min(0, 'Rencana USD tidak boleh negatif'),
  planned_idr: z.number().min(0, 'Rencana IDR tidak boleh negatif'),
  realized_la: z.number().min(0, 'Realisasi LA tidak boleh negatif'),
  realized_usd: z.number().min(0, 'Realisasi USD tidak boleh negatif'),
  realized_idr: z.number().min(0, 'Realisasi IDR tidak boleh negatif'),
})

export const monitoringSchema = z.object({
  budget_year: z.number().int('Tahun harus bilangan bulat').min(2000, 'Tahun minimal 2000'),
  quarter: quarterSchema,
  exchange_rate_usd_idr: z.number().positive('Kurs USD/IDR harus lebih dari 0'),
  exchange_rate_la_idr: z.number().positive('Kurs LA/IDR harus lebih dari 0'),
  planned_la: z.number().min(0, 'Rencana LA tidak boleh negatif'),
  planned_usd: z.number().min(0, 'Rencana USD tidak boleh negatif'),
  planned_idr: z.number().min(0, 'Rencana IDR tidak boleh negatif'),
  realized_la: z.number().min(0, 'Realisasi LA tidak boleh negatif'),
  realized_usd: z.number().min(0, 'Realisasi USD tidak boleh negatif'),
  realized_idr: z.number().min(0, 'Realisasi IDR tidak boleh negatif'),
  komponen: z.array(monitoringKomponenSchema).optional(),
})
