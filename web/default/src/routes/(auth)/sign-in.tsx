import { z } from 'zod'
import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { SignIn } from '@/features/auth/sign-in'
import { saveAffiliateCode } from '@/features/auth/lib/storage'

const searchSchema = z.object({
  redirect: z.string().optional(),
  aff: z.string().optional(),
})

export const Route = createFileRoute('/(auth)/sign-in')({
  component: SignIn,
  validateSearch: searchSchema,
  beforeLoad: async ({ search }) => {
    const { auth } = useAuthStore.getState()

    // Save affiliate code if present in URL
    if (search.aff) {
      saveAffiliateCode(search.aff)
    }

    // 如果已经有用户信息，说明已登录
    if (auth.user) {
      // 优先使用 redirect 参数（用户之前想去的地方）
      // 否则跳转到 dashboard
      throw redirect({ to: search?.redirect || '/dashboard' })
    }
  },
})
