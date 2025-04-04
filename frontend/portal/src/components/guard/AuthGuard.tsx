'use client'

import { ReactNode, ReactElement, useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@doota/ui-core/hooks/useAuth'
import { routes } from '@doota/ui-core/routing'
import { browserTokenStore } from '@doota/ui-core/provider/BrowserStores'

interface AuthGuardProps {
  children: ReactNode
  fallback: ReactElement | null
}

const AuthGuard = (props: AuthGuardProps) => {
  const { children, fallback } = props

  const { user, loading } = useAuth()
  const router = useRouter()
  const path = usePathname()

  useEffect(
    () => {
      if (user === null) {

        // It's possible that user is still loading, in the browser, we redirect only if
        // the token is not present in the browser store also
        browserTokenStore.Get().then(token => {
          if (token === undefined) {
            router.replace(routes.app.auth.login)
          }
        })
        return
      }
      // if (auth.user !== null && !isAdmin(auth.user)) {
      //   if (auth.user.organizationId === undefined) {
      //     router.replace(routes.app.onboarding)
      //   }
      // }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [user, path]
  )

  if (loading || user === null) {
    return fallback
  }

  return <>{children}</>
}

export default AuthGuard
