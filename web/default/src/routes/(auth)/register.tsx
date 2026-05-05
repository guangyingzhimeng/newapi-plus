import { createFileRoute } from '@tanstack/react-router'
import { SignUp } from '@/features/auth'
import { saveAffiliateCode } from '@/features/auth/lib/storage'
import { z } from 'zod'

export const Route = createFileRoute('/(auth)/register')({
  validateSearch: z.object({
    aff: z.string().optional(),
  }),
  beforeLoad: ({ search }) => {
    if (search.aff) {
      saveAffiliateCode(search.aff)
    }
  },
  component: SignUp,
})
