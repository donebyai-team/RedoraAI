import { FC, ReactNode, createContext, useEffect, useState } from 'react'
import { JWT, Organization, User } from '@doota/pb/doota/portal/v1/portal_pb'
import toast from 'react-hot-toast'
import { errorToMessage } from '@doota/pb/utils/errors'
import { FullStory, isInitialized as isFullStoryInitialized } from '@fullstory/browser'
import { useClientsContext } from './ClientContext'
import { TokenStore, OrganizationStore } from '@doota/store'

export type AuthValuesType = {
  user: User | null
  organization: Organization | null
  loading: boolean

  logout: () => void
  login: (jwt: JWT) => Promise<void>
  refreshSession: () => Promise<void>

  setOrganization: (org: Organization) => Promise<void>
}

// ** Defaults
const defaultProvider: AuthValuesType = {
  user: null,
  organization: null,
  loading: true,
  refreshSession: () => Promise.resolve(),
  login: () => Promise.resolve(),
  logout: () => Promise.resolve(),

  setOrganization: () => Promise.resolve()
}

export const AuthContext = createContext<AuthValuesType>(defaultProvider)

type Props = {
  tokenStore: TokenStore
  organizationStore: OrganizationStore
  onLogoutCompleted?: () => void
  onRefreshSessionError?: (error: unknown) => void
  children: ReactNode
}

export const BaseAuthProvider: FC<Props> = ({
  tokenStore,
  organizationStore,
  onLogoutCompleted,
  onRefreshSessionError,
  children
}) => {
  const [user, setUser] = useState<User | null>(defaultProvider.user)
  const [organization, setOrganization] = useState<Organization | null>(defaultProvider.organization)
  const [loading, setLoading] = useState<boolean>(defaultProvider.loading)
  const { portalClient } = useClientsContext()

  useEffect(() => {
    const initAuth = async (): Promise<void> => {
      const jwt = await tokenStore.Get()
      if (jwt === undefined) {
        setLoading(false)
        return
      }

      await refreshSession()
    }

    initAuth()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tokenStore])

  useEffect(() => {
    organizationStore
      .Get()
      .then(org => {
        setOrganization(org ?? null)
      })
      .catch(err => {
        console.error('failed to get organization', { err: err })
      })
  }, [organizationStore])

  const getUser = (): Promise<User> => {
    console.debug('getting self')

    return portalClient.self({})
  }

  const setFullStoreIdentity = (user: User) => {
    if (isFullStoryInitialized()) {
      FullStory('setIdentity', {
        uid: user.id,
        properties: {
          email: user.email,
          role: user.role,
          Organization: user.organizations[0].name
        }
      })
    }
  }

  const handleLogin = async (jwt: JWT) => {
    await tokenStore.Set(jwt)
    getUser()
      .then(user => {
        setUser(user)
        setFullStoreIdentity(user)
        if (isFullStoryInitialized()) {
          FullStory('trackEvent', {
            name: 'user_logged_in',
            properties: {}
          })
        }
      })
      .catch(err => {
        toast.error(errorToMessage(err))
      })

    return
  }

  const handleLogout = async () => {
    setUser(null)
    await tokenStore.Del()
    await organizationStore.Del()

    if (onLogoutCompleted) {
      onLogoutCompleted()
    }
  }

  const refreshSession = async () => {
    setLoading(true)
    try {
      const user = await getUser()
      setUser(user)
      setFullStoreIdentity(user)
    } catch (err) {
      if (onRefreshSessionError) {
        onRefreshSessionError(err)
      }
    }
    setLoading(false)
  }

  const values = {
    user,
    organization,
    loading,
    setLoading,
    refreshSession,
    login: handleLogin,
    logout: handleLogout,
    setOrganization: async (org: Organization) => {
      await organizationStore.Set(org)
      setOrganization(org)
    }
  }

  return <AuthContext.Provider value={values}>{children}</AuthContext.Provider>
}
