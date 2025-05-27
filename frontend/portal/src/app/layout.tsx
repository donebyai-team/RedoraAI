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
import Script from 'next/script'

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
      {/* ✅ HubSpot Chat Script */}
      <Script
          id="hubspot-chat"
          strategy="afterInteractive"
          src="//js-na2.hs-scripts.com/242526027.js"
      />
        <StoreProvider>
          <NextElementRegistryProvider>
            <ConfigGuard fallback={<FallbackSpinner />}>
              <PortalClientProvider>
                <ConfigProvider>
                  <PortalExecutionRuntimeProvider>
                    <ThemeRegistry>
                      {children}
                      <StyledReactHotToast>
                        <Toaster position='top-right' toastOptions={{ className: 'react-hot-toast mt-[47px]' }} />
                      </StyledReactHotToast>
                    </ThemeRegistry>
                  </PortalExecutionRuntimeProvider>
                </ConfigProvider>
              </PortalClientProvider>
            </ConfigGuard>
          </NextElementRegistryProvider>
        </StoreProvider>
      </body>
    </html>
  )
}
