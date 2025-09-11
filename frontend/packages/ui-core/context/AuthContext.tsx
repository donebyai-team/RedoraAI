import { FC, ReactNode, createContext, useEffect, useState } from 'react'
import { JWT, Organization, User } from '@doota/pb/doota/portal/v1/portal_pb'
import toast from 'react-hot-toast'

import { errorToMessage } from '@doota/pb/utils/errors'
import { FullStory, isInitialized as isFullStoryInitialized } from '@fullstory/browser'
import { useClientsContext } from './ClientContext'
import { TokenStore, OrganizationStore } from '@doota/store'
import { Subscription, SubscriptionPlanID, SubscriptionStatus } from '@doota/pb/doota/core/v1/core_pb'
import { logDailyVisit } from '../amplitude'
import { isPlatformAdmin } from '../helper/role'
import { Timestamp } from '@bufbuild/protobuf/dist/cjs/wkt/gen/google/protobuf/timestamp_pb'


const DEFAULT_PLAN: Subscription = {
  $typeName: "doota.core.v1.Subscription",
  planId: SubscriptionPlanID.SUBSCRIPTION_PLAN_FREE,
  status: SubscriptionStatus.ACTIVE,
  maxKeywords: 0,
  maxSources: 0,
  comments: undefined,
  dm: undefined,
  createdAt: undefined,
  expiresAt: undefined,
};

export type AuthValuesType = {
  user: User | null
  organization: Organization | null
  loading: boolean

  logout: () => void
  login: (jwt: JWT) => Promise<void>
  refreshSession: () => Promise<void>

  setOrganization: (org: Organization) => Promise<void>
  setUser: React.Dispatch<React.SetStateAction<User | null>>
  getOrganization: () => Organization | null
  getPlanDetails: () => Subscription,
  planDetails: Subscription,
  currentOrganization: Organization | null,
}

// ** Defaults
const defaultProvider: AuthValuesType = {
  user: null,
  organization: null,
  loading: true,
  refreshSession: () => Promise.resolve(),
  login: () => Promise.resolve(),
  logout: () => Promise.resolve(),

  setOrganization: () => Promise.resolve(),
  setUser: () => { },
  getOrganization: () => null,
  getPlanDetails: () => DEFAULT_PLAN,
  planDetails: DEFAULT_PLAN,
  currentOrganization: null
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

      if (!isPlatformAdmin(user) && user.organizations.length > 0) {
        const orgID = user.organizations[0].id
        let name = user.organizations[0].name
        if (user.projects.length > 0) {
          name = name + "-" + user.projects[0].name
        }
        const plan = user.organizations[0].featureFlags?.subscription?.planId
        logDailyVisit(orgID, name, {
          plan: plan?.toString(),
          page: window.location.pathname,
          createdAt: user.createdAt ? timestampToISOString(user.createdAt) : undefined,
        })
      }

    } catch (err) {
      if (onRefreshSessionError) {
        onRefreshSessionError(err)
      }
    }
    setLoading(false)
  }

  const getOrganization = () => {
    if (organization) {
      return organization;
    }
    return user?.organizations?.[0] ?? null;
  }

  const getPlanDetails = (): Subscription => {
    const currentOrganization = getOrganization();
    if (currentOrganization?.featureFlags?.subscription) {
      return currentOrganization?.featureFlags?.subscription
    }
    return DEFAULT_PLAN;
  };

  const planDetails = getPlanDetails();
  const currentOrganization = getOrganization();

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
    },
    setUser,
    getOrganization,
    getPlanDetails,
    planDetails,
    currentOrganization
  }

  return <AuthContext.Provider value={values}>{children}</AuthContext.Provider>
}

function timestampToISOString(ts?: Timestamp): string | undefined {
  if (!ts) return undefined;
  const millis = Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000);
  return new Date(millis).toISOString();
}
