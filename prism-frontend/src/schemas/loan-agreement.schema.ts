import { z } from 'zod'

export const loanAgreementSchema = z
  .object({
    dk_project_id: z.string().uuid('DK Project wajib dipilih'),
    lender_id: z.string().uuid('Lender wajib dipilih'),
    loan_code: z.string().min(1, 'Kode Loan wajib diisi'),
    agreement_date: z.string().min(1, 'Tanggal agreement wajib diisi'),
    effective_date: z.string().min(1, 'Tanggal efektif wajib diisi'),
    original_closing_date: z.string().min(1, 'Original Closing Date wajib diisi'),
    closing_date: z.string().min(1, 'Closing Date wajib diisi'),
    currency: z
      .string()
      .min(3, 'Kode mata uang minimal 3 karakter (ISO 4217)')
      .max(3, 'Kode mata uang maksimal 3 karakter (ISO 4217)'),
    amount_original: z.number().positive('Amount harus lebih dari 0'),
    amount_usd: z.number().positive('Amount USD harus lebih dari 0'),
  })
  .refine((data) => new Date(data.closing_date) >= new Date(data.original_closing_date), {
    message: 'Closing Date tidak boleh lebih awal dari Original Closing Date',
    path: ['closing_date'],
  })
