import { FC, ReactNode } from 'react'
import { BaseAuthProvider } from '../context/AuthContext'
import { browserTokenStore, browserOrganizationStore } from '../provider/BrowserStores'

export const BrowserAuthProvider: FC<{ children: ReactNode }> = ({ children }) => {
  return (
    <BaseAuthProvider
      tokenStore={browserTokenStore}
      organizationStore={browserOrganizationStore}
      onLogoutCompleted={() => {
        // Nothing to perform in the Outlook extension, the component will be re-rendered
        // and will display the login form.
      }}
    >
      {children}
    </BaseAuthProvider>
  )
}
