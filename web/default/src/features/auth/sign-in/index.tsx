import { useSearch } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { AuthLayout } from '../auth-layout'
import { TermsFooter } from '../components/terms-footer'
import { UserAuthForm } from './components/user-auth-form'

export function SignIn() {
  const { t } = useTranslation()
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })

  return (
    <AuthLayout>
      <div className='w-full space-y-8'>
        <div className='space-y-2'>
          <h2 className='text-center text-2xl font-semibold tracking-tight sm:text-left'>
            {t('WeChat QR code sign in')}
          </h2>
          <p className='text-muted-foreground text-center text-sm sm:text-left sm:text-base'>
            {t('Please use WeChat to scan the QR code and confirm in DreamAuth.')}
          </p>
        </div>

        <UserAuthForm redirectTo={redirect} />

        <TermsFooter variant='sign-in' className='text-center' />
      </div>
    </AuthLayout>
  )
}
