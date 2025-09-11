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

// const amplitudeApiKey = process.env.NEXT_PUBLIC_AMPLITUDE_API_KEY || '';

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

        <script
          src="https://browser.sentry-cdn.com/8.47.0/bundle.tracing.replay.min.js"
          integrity="sha384-VaqNrma84jlgEWxBCMOnatKAHLSjaKGmo8Biuj3NQEg1MrmeukY8s6pnaTgRVjKM"
          crossOrigin="anonymous"
        ></script>

        <script
          dangerouslySetInnerHTML={{
            __html: `
              Sentry.init({
                dsn: "https://9ddcdfafdb7a4f7e8ffa1b4a653678da@o5948691128404174.ingest.us.getdecipher.com/5382214313016608",
                integrations: [
                  Sentry.replayIntegration({
                    maskAllText: false,
                    blockAllMedia: false,
                    maskAllInputs: true,
                    networkDetailAllowUrls: [/^.*$/],
                  }),
                ],
                replaysOnErrorSampleRate: 1.0,
                replaysSessionSampleRate: 1.0,
              });
            `,
          }}
        />

      </head>

      <body>
        {/* ✅ HubSpot Chat Script */}
        <Script
          id="hubspot-chat"
          strategy="afterInteractive"
          src="//js-na2.hs-scripts.com/242526027.js"
        />

        {/* ✅ Clarity Script */}
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

        {/* ✅ Amplitude Scripts */}
        {/* <Script src="https://cdn.amplitude.com/libs/analytics-browser-2.11.1-min.js.gz" />
        <Script src="https://cdn.amplitude.com/libs/plugin-session-replay-browser-1.8.0-min.js.gz" />
        <Script
          id="amplitude-scripts"
          type="text/javascript"
          dangerouslySetInnerHTML={{
            __html: `
              window.amplitude.add(window.sessionReplay.plugin({ sampleRate: 1 }));
              window.amplitude.init('${amplitudeApiKey}', {
                autocapture: { elementInteractions: true }
              });
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
