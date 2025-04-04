import { useContext } from 'react'
import { ConfigContext } from '../context/ConfigContext'
import { Config } from '@doota/pb/doota/portal/v1/portal_pb'

export const useConfig = (): Config => {
  const config = useContext(ConfigContext)
  return config
}
