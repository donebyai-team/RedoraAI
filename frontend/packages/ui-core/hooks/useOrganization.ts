import { Organization } from '@doota/pb/doota/portal/v1/portal_pb'
import { AuthContext } from '../context/AuthContext'
import { useContext } from 'react'

export const useOrganization = (): [Organization | undefined, (organization: Organization) => Promise<void>] => {
  const { user, organization, setOrganization } = useContext(AuthContext)
  if (organization != null) {
    return [organization, setOrganization]
  }

  if (user != null) {
    return [user.organizations[0], setOrganization]
  }

  return [undefined, setOrganization]
}
