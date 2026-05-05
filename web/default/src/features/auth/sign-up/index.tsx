import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { AuthLayout } from '../auth-layout'
import { TermsFooter } from '../components/terms-footer'
import { UserAuthForm } from '../sign-in/components/user-auth-form'

export function SignUp() {
  const { t } = useTranslation()

  return (
    <AuthLayout>
      <div className='w-full space-y-8'>
        <div className='space-y-2'>
          <h2 className='text-center text-2xl font-semibold tracking-tight sm:text-left'>
            {t('WeChat QR code sign up/in')}
          </h2>
          <p className='text-muted-foreground text-center text-sm sm:text-left sm:text-base'>
            {t('Please use WeChat to scan the QR code to register or sign in.')}
          </p>
        </div>

        <UserAuthForm />

        <div className='text-center text-sm'>
          <p className='text-muted-foreground'>
            {t('Already have an account?')}{' '}
            <Link
              to='/sign-in'
              className='hover:text-primary font-medium underline underline-offset-4'
            >
              {t('Sign in')}
            </Link>
          </p>
        </div>

        <TermsFooter variant='sign-up' className='text-center' />
      </div>
    </AuthLayout>
  )
}
