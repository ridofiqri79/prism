import { z } from 'zod'

export const createUserSchema = z.object({
  username: z.string().min(3, 'Username minimal 3 karakter'),
  email: z.string().email('Format email tidak valid'),
  password: z.string().min(8, 'Password minimal 8 karakter'),
  role: z.enum(['ADMIN', 'STAFF']),
})

export const updateUserSchema = createUserSchema.omit({ password: true }).extend({
  is_active: z.boolean(),
})

export type CreateUserFormValues = z.infer<typeof createUserSchema>
export type UpdateUserFormValues = z.infer<typeof updateUserSchema>
