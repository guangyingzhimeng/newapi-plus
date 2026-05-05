import { useEffect, useRef, useState } from 'react'
import { CheckCircle2, Loader2, QrCode, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useStatus } from '@/hooks/use-status'
import { Button } from '@/components/ui/button'
import { LegalConsent } from '@/features/auth/components/legal-consent'
import {
  completeDreamAuthLogin,
  createDreamAuthSession,
  getDreamAuthStatus,
} from '@/features/auth/api'
import type {
  AuthFormProps,
  DreamAuthSession,
  DreamAuthStatus,
} from '@/features/auth/types'
import { useAuthRedirect } from '@/features/auth/hooks/use-auth-redirect'

const POLL_INTERVAL_MS = 2000
const QR_REFRESH_INTERVAL_MS = 60_000

export function UserAuthForm({ redirectTo }: AuthFormProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const { handleLoginSuccess } = useAuthRedirect()
  const [session, setSession] = useState<DreamAuthSession | null>(null)
  const [scanStatus, setScanStatus] = useState<DreamAuthStatus | null>(null)
  const [loading, setLoading] = useState(false)
  const [completing, setCompleting] = useState(false)
  const [agreedToLegal, setAgreedToLegal] = useState(false)
  const completingRef = useRef(false)

  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)
  const requiresLegalConsent = hasUserAgreement || hasPrivacyPolicy
  const canStart = !requiresLegalConsent || agreedToLegal

  useEffect(() => {
    setAgreedToLegal(!requiresLegalConsent)
  }, [requiresLegalConsent])

  const startSession = async () => {
    if (!canStart) {
      toast.error(t('Please agree to the legal terms first'))
      return
    }

    completingRef.current = false
    setCompleting(false)
    setLoading(true)
    setScanStatus(null)
    try {
      const res = await createDreamAuthSession('user')
      if (!res.success || !res.data) {
        throw new Error(res.message || t('Failed to get QR code'))
      }
      setSession(res.data)
    } catch (error) {
      toast.error(
        error instanceof Error ? error.message : t('Failed to get QR code')
      )
      setSession(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (canStart) {
      void startSession()
    }
    // startSession intentionally reads changing state; canStart is the trigger.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [canStart])

  useEffect(() => {
    if (!session?.sessionNo || completing) return
    const timer = window.setTimeout(() => {
      void startSession()
    }, QR_REFRESH_INTERVAL_MS)
    return () => window.clearTimeout(timer)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [completing, session?.sessionNo])

  useEffect(() => {
    if (!session?.sessionNo) return
    let cancelled = false
    let timer: number | null = null

    const poll = async () => {
      try {
        const next = await getDreamAuthStatus(session.sessionNo)
        if (cancelled) return
        if (!next.success || !next.data) {
          throw new Error(next.message || t('Failed to get login status'))
        }
        setScanStatus(next.data)

        if (next.data.loginReady && !completingRef.current) {
          completingRef.current = true
          setCompleting(true)
          const result = await completeDreamAuthLogin(session.sessionNo)
          if (!result.success) {
            throw new Error(result.message || t('Login failed'))
          }
          await handleLoginSuccess(result.data ?? null, redirectTo)
          toast.success(t('Welcome back!'))
          return
        }

        if (next.data.expired || next.data.status === 4 || next.data.status === 6) {
          return
        }
      } catch (error) {
        if (!cancelled) {
          completingRef.current = false
          setCompleting(false)
          setScanStatus((prev) => ({
            sessionNo: session.sessionNo,
            status: prev?.status ?? 0,
            statusText:
              error instanceof Error
                ? error.message
                : t('Failed to get login status'),
            loginReady: false,
            expired: false,
          }))
        }
      } finally {
        if (!cancelled && !completingRef.current) {
          timer = window.setTimeout(poll, POLL_INTERVAL_MS)
        }
      }
    }

    timer = window.setTimeout(poll, 400)
    return () => {
      cancelled = true
      if (timer) window.clearTimeout(timer)
    }
  }, [handleLoginSuccess, redirectTo, session?.sessionNo, t])

  const showSuccess = completing || scanStatus?.loginReady
  const statusText = completing
    ? t('Authorization successful, signing in...')
    : scanStatus?.statusText || t('Please scan with WeChat')
  const needsManualRefresh =
    scanStatus?.expired || scanStatus?.status === 4 || scanStatus?.status === 6

  return (
    <div className='grid gap-5'>
      <div className='relative mx-auto'>
        <div className='bg-primary/10 absolute -inset-3 rounded-lg blur-sm' />
        <div className='bg-background relative flex aspect-square w-[260px] items-center justify-center rounded-lg border p-4 shadow-sm'>
          {loading ? (
            <Loader2 className='text-muted-foreground h-8 w-8 animate-spin' />
          ) : session?.qrcode ? (
            <img
              src={session.qrcode}
              alt={t('WeChat login QR code')}
              className='h-full w-full object-contain'
            />
          ) : (
            <QrCode className='text-muted-foreground h-12 w-12' />
          )}
        </div>
      </div>

      <div className='text-muted-foreground flex min-h-6 items-center justify-center gap-2 text-sm'>
        {completing ? (
          <Loader2 className='text-primary h-4 w-4 animate-spin' />
        ) : showSuccess ? (
          <CheckCircle2 className='h-4 w-4 text-emerald-500' />
        ) : (
          <span className='bg-primary h-2 w-2 rounded-full' />
        )}
        <span>{statusText}</span>
      </div>

      <LegalConsent
        status={status}
        checked={agreedToLegal}
        onCheckedChange={setAgreedToLegal}
      />

      <Button
        type='button'
        variant={needsManualRefresh || !session ? 'default' : 'outline'}
        className='h-11 w-full justify-center gap-2 rounded-lg'
        disabled={loading || completing || !canStart}
        onClick={() => void startSession()}
      >
        {loading ? (
          <Loader2 className='h-4 w-4 animate-spin' />
        ) : (
          <RefreshCw className='h-4 w-4' />
        )}
        {session ? t('Refresh QR code') : t('Get QR code')}
      </Button>
    </div>
  )
}
