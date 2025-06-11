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
import NotificationProvider from '@/components/layout/notification'

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
        <Script
          id="clarity-script"
          type="text/javascript"
          dangerouslySetInnerHTML={{
            __html: `(function(c,l,a,r,i,t,y){
            c[a]=c[a]||function(){(c[a].q=c[a].q||[]).push(arguments)};
            t=l.createElement(r);t.async=1;t.src="https://www.clarity.ms/tag/"+i;
            y=l.getElementsByTagName(r)[0];y.parentNode.insertBefore(t,y);
            })(window, document, "clarity", "script", "rt2azu4mlk");`
          }}
        />

        {/* <Script src="https://cdn.amplitude.com/libs/analytics-browser-2.11.1-min.js.gz" />
        <Script src="https://cdn.amplitude.com/libs/plugin-session-replay-browser-1.8.0-min.js.gz" />
        <Script
          id="amplitude-scripts"
          type="text/javascript"
          dangerouslySetInnerHTML={{
            __html: `
              window.amplitude.add(window.sessionReplay.plugin({sampleRate: 1}));window.amplitude.init('e678ed1fa2f36bf86c1527a2f8fb9862', {"autocapture":{"elementInteractions":true}});
            `,
          }}
        /> */}

        <StoreProvider>
          <NextElementRegistryProvider>
            <ConfigGuard fallback={<FallbackSpinner />}>
              <PortalClientProvider>
                <ConfigProvider>
                  <PortalExecutionRuntimeProvider>
                    <ThemeRegistry>
                      <NotificationProvider>
                        {children}
                      </NotificationProvider>
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
