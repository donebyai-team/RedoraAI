'use client'

import { ReactNode, ReactElement, useEffect, useState } from 'react'
import { configProvider } from '../../services/config'
import { init as initFullStory, isInitialized as isFullStoryInitialized } from '@fullstory/browser'
import { log } from '../../services/logger'
import { initAmplitude } from '@doota/ui-core/amplitude'

interface ConfigGuardProps {
  children: ReactNode
  fallback: ReactElement | null
}

const ConfigGuard = (props: ConfigGuardProps) => {
  const { children, fallback } = props
  // const [loading, setLoading] = useState(false)

  const apiKey = process.env.NEXT_PUBLIC_AMPLITUDE_API_KEY;
  initAmplitude(apiKey)

  // useEffect(() => {
  //   setLoading(true)
  //   configProvider.bootstrap().finally(() => {
  //     setLoading(false)

  //     if (configProvider.config.fullStoryOrgId === '') {
  //       log.info('skipping fullstory setup')
  //       return
  //     }

  //     if (isFullStoryInitialized()) {
  //       log.info('fullstory already initialized')
  //       return
  //     }

  //     initFullStory({ orgId: configProvider.config.fullStoryOrgId, debug: true }, ({ sessionUrl }) =>
  //       log.info('fullstory started session %s', sessionUrl)
  //     )
  //   })
  // }, [])

  // if (loading) {
  //   return fallback
  // }

  return <>{children}</>
}

export default ConfigGuard
