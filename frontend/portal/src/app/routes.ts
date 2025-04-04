import { CONFIG_PORTAL_URI } from '../services/config'

export const buildAppUrl = (path: string): string => {
  return CONFIG_PORTAL_URI + path
}
