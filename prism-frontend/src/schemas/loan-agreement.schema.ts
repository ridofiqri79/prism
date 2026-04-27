import { z } from 'zod'

export const loanAgreementSchema = z
  .object({
    dk_project_id: z.string().uuid('DK Project wajib dipilih'),
    lender_id: z.string().uuid('Lender wajib dipilih'),
    loan_code: z.string().min(1, 'Kode loan wajib diisi'),
    agreement_date: z.string().min(1, 'Tanggal agreement wajib diisi'),
    effective_date: z.string().min(1, 'Tanggal efektif wajib diisi'),
    original_closing_date: z.string().min(1, 'Tanggal closing awal wajib diisi'),
    closing_date: z.string().min(1, 'Tanggal closing wajib diisi'),
    currency: z
      .string()
      .min(3, 'Kode mata uang minimal 3 karakter (ISO 4217)')
      .max(3, 'Kode mata uang maksimal 3 karakter (ISO 4217)'),
    amount_original: z.number().positive('Nilai pinjaman harus lebih dari 0'),
    amount_usd: z.number().positive('Nilai pinjaman USD harus lebih dari 0'),
  })
  .refine((data) => new Date(data.closing_date) >= new Date(data.original_closing_date), {
    message: 'Tanggal closing tidak boleh lebih awal dari tanggal closing awal',
    path: ['closing_date'],
  })
