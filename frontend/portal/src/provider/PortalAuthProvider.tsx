'use client'
import { BaseAuthProvider } from '@doota/ui-core/context/AuthContext'
import { useRouter } from 'next/navigation'
import { browserOrganizationStore, browserTokenStore } from '@doota/ui-core/provider/BrowserStores'
import { routes } from '@doota/ui-core/routing'
import { FC, ReactNode } from 'react'

export const PortalAuthProvider: FC<{ children: ReactNode }> = ({ children }) => {
  const router = useRouter()

  return (
    <BaseAuthProvider
      tokenStore={browserTokenStore}
      organizationStore={browserOrganizationStore}
      onLogoutCompleted={() => {
        router.replace(routes.app.auth.login)
      }}
      onRefreshSessionError={() => {
        // This might not be the most appropriate way to do this, for
        // now let's do like it was in AuthContext before this was moved here
        router.replace(routes.app.auth.login)
      }}
    >
      {children}
    </BaseAuthProvider>
  )
}
