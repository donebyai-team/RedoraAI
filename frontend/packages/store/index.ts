import { JWT, Organization } from '@doota/pb/doota/portal/v1/portal_pb'

export interface TokenStore {
  Set(token: JWT): Promise<void>
  Get(): Promise<JWT | undefined>
  Del(): Promise<void>
}

export interface OrganizationStore {
  Set(org: Organization): Promise<void>
  Get(): Promise<Organization | undefined>
  Del(): Promise<void>
}
