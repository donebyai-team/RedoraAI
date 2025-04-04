'use client'

import { FallbackSpinner } from '../atoms/FallbackSpinner'
import { ReactNode, createContext, useEffect, useState } from 'react'
import { Config, ConfigSchema } from '@doota/pb/doota/portal/v1/portal_pb'
import { portalClient } from '../services/grpc'
import { create } from '@bufbuild/protobuf'

// ** Defaults
const defaultProvider = create(ConfigSchema, {
  auth0Domain: 'domain.auth0.com',
  auth0ClientId: 'xxxxxxxxxxxxxxxx',
  auth0Scope: 'openid email',
  msoftAuth0CallbackUrl: 'http://msoftcallback'
})

const ConfigContext = createContext<Config>(defaultProvider)

type Props = {
  children: ReactNode
}

const ConfigProvider = ({ children }: Props) => {
  const [config, setConfig] = useState<Config>(defaultProvider)
  const [loading, setLoading] = useState<boolean>(true)

  useEffect(() => {
    portalClient
      .getConfig({})
      .then(config => {
        setConfig(config)
      })
      .finally(() => {
        setLoading(false)
      })
  }, [])

  if (loading) {
    return <FallbackSpinner />
  }

  return <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
}

export { ConfigContext, ConfigProvider }
