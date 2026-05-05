import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { CopyButton } from '@/components/copy-button'
import { ApiKeysDialogs } from './components/api-keys-dialogs'
import { ApiKeysProvider } from './components/api-keys-provider'
import { ApiKeysTable } from './components/api-keys-table'

const API_ENDPOINT = 'https://guangyingzhimeng.dpdns.org/new-api/v1'

export function ApiKeys() {
  const { t } = useTranslation()
  return (
    <ApiKeysProvider>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('API Keys')}</SectionPageLayout.Title>
        <SectionPageLayout.Description>
          {t('Manage your API keys for accessing the service')}
        </SectionPageLayout.Description>
        <SectionPageLayout.Content>
          <div className='space-y-3'>
            <div className='bg-muted/30 flex flex-col gap-2 rounded-lg border px-3 py-2.5 sm:flex-row sm:items-center sm:justify-between'>
              <div className='min-w-0'>
                <div className='text-xs font-medium'>
                  {t('API Endpoint')}
                </div>
                <div className='text-muted-foreground truncate font-mono text-xs sm:text-sm'>
                  {API_ENDPOINT}
                </div>
              </div>
              <CopyButton
                value={API_ENDPOINT}
                variant='outline'
                size='sm'
                tooltip={t('Copy API endpoint')}
                successTooltip={t('Copied!')}
                aria-label={t('Copy API endpoint')}
              />
            </div>
            <ApiKeysTable />
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <ApiKeysDialogs />
    </ApiKeysProvider>
  )
}
