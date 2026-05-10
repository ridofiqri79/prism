import { z } from 'zod'

export const loanAgreementSchema = z
  .object({
    dk_project_id: z.string().optional().default(''),
    dk_project_allocations: z
      .array(
        z.object({
          dk_project_id: z.string().uuid('Proyek Daftar Kegiatan wajib dipilih'),
          allocation_original: z.number().positive('Alokasi harus lebih dari 0'),
        }),
      )
      .min(1, 'Minimal satu Proyek Daftar Kegiatan wajib dipilih'),
    lender_id: z.string().uuid('Lender wajib dipilih'),
    loan_code: z.string().min(1, 'Kode loan wajib diisi'),
    agreement_date: z.string().min(1, 'Tanggal agreement wajib diisi'),
    effective_date: z.string().min(1, 'Tanggal efektif wajib diisi'),
    original_closing_date: z.string().trim().optional().default(''),
    closing_date: z.string().min(1, 'Tanggal closing wajib diisi'),
    currency: z
      .string()
      .min(3, 'Kode mata uang minimal 3 karakter (ISO 4217)')
      .max(3, 'Kode mata uang maksimal 3 karakter (ISO 4217)'),
    amount_original: z.number().positive('Nilai pinjaman harus lebih dari 0'),
    amount_usd: z.number().nonnegative('Nilai pinjaman USD tidak boleh negatif').default(0),
    cumulative_disbursement: z.number().nonnegative('Cumulative disbursement tidak boleh negatif'),
  })
  .refine(
    (data) => {
      if (!data.original_closing_date) return true
      return new Date(data.closing_date) >= new Date(data.original_closing_date)
    },
    {
      message: 'Tanggal closing tidak boleh lebih awal dari tanggal closing awal',
      path: ['closing_date'],
    },
  )
  .refine(
    (data) => {
      const total = data.dk_project_allocations.reduce(
        (sum, item) => sum + item.allocation_original,
        0,
      )
      return (
        Math.abs(Math.round(total * 100) / 100 - Math.round(data.amount_original * 100) / 100) <
        0.005
      )
    },
    {
      message: 'Total alokasi harus sama dengan nilai Loan Agreement',
      path: ['dk_project_allocations'],
    },
  )
