import '../styles/global.css'

import { ConfigProvider } from '../context/ConfigContext'
import ThemeRegistry from '../theme/ThemeRegistry'
import ConfigGuard from '../components/guard/ConfigGuard'
import { FallbackSpinner } from '../atoms/FallbackSpinner'
import { StyledReactHotToast } from '@doota/ui-core/components/StyledReactHotToast'
import { Toaster } from 'react-hot-toast'
import { NextElementRegistryProvider } from '../context/NextElementRegistryProvider'
import { PortalClientProvider } from '../provider/PortalClientProvider'
import { PortalExecutionRuntimeProvider } from '../provider/PortalExecutionRuntimeProvider'
import StoreProvider from '../../store/StoreProvider'

export const metadata = {
  title: 'RedoraAI',
  description: ''
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang='en'>
      <head>
        <meta name='robots' content='noindex,follow' />
        <link rel="icon" href="/images/favicon.ico" />
        <title>RedoraAI</title>
      </head>

      <body>
        <NextElementRegistryProvider>
          <ConfigGuard fallback={<FallbackSpinner />}>
            <PortalClientProvider>
              <ConfigProvider>
                <StoreProvider>
                  <PortalExecutionRuntimeProvider>
                    <ThemeRegistry>
                      {children}
                      <StyledReactHotToast>
                        <Toaster position='top-right' toastOptions={{ className: 'react-hot-toast mt-[47px]' }} />
                      </StyledReactHotToast>
                    </ThemeRegistry>
                  </PortalExecutionRuntimeProvider>
                </StoreProvider>
              </ConfigProvider>
            </PortalClientProvider>
          </ConfigGuard>
        </NextElementRegistryProvider>
      </body>
    </html>
  )
}
