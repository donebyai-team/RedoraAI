import { ReactNode, ReactElement, useEffect, useState, FC } from 'react'
import { useClientsContext } from '../../context/ClientContext'
import { init as initFullStory, isInitialized as isFullStoryInitialized } from '@fullstory/browser'

interface Props {
  children: ReactNode
  fallback: ReactElement | null
}

export const ConfigGuard: FC<Props> = ({ children, fallback }) => {
  const [loading, setLoading] = useState(false)
  const { portalClient } = useClientsContext()

  useEffect(() => {
    if (isFullStoryInitialized()) {
      console.info('fullstory already initialized')
      return
    }

    setLoading(true)

    portalClient
      .getConfig({})
      .then(config => {
        initFullStory({ orgId: config.fullStoryOrgId, debug: true }, ({ sessionUrl }) =>
          console.info('fullstory started session %s', sessionUrl)
        )
      })
      .catch(error => {
        console.error('failed to get config %s', error)
      })
      .finally(() => {
        setLoading(false)
      })
  }, [setLoading, portalClient])

  if (loading) {
    return fallback
  }

  return children
}
