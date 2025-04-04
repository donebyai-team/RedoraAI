import { fromJsonString, toJsonString } from '@bufbuild/protobuf'
import { JWT, JWTSchema, Organization, OrganizationSchema } from '@doota/pb/doota/portal/v1/portal_pb'

export const TOKEN_KEY = 'token'
export const ORG_KEY = 'og'

export class BrowserTokenStore {
  async Set(token: JWT): Promise<void> {
    window.localStorage.setItem(TOKEN_KEY, toJsonString(JWTSchema, token))
  }

  async Get(): Promise<JWT | undefined> {
    const storedToken = window.localStorage.getItem(TOKEN_KEY)
    console.log("Rgsegsegse", storedToken)
    if (!storedToken) {
      return undefined
    }

    return fromJsonString(JWTSchema, storedToken)
  }

  async Del(): Promise<void> {
    window.localStorage.removeItem(TOKEN_KEY)
  }
}

export class BrowserOrganizationStore {
  async Set(organization: Organization): Promise<void> {
    window.localStorage.setItem(ORG_KEY, toJsonString(OrganizationSchema, organization))
  }

  async Get(): Promise<Organization | undefined> {
    const storedOrganization = window.localStorage.getItem(ORG_KEY)
    if (!storedOrganization) {
      return undefined
    }

    return fromJsonString(OrganizationSchema, storedOrganization)
  }
}

export const browserTokenStore = new BrowserTokenStore()
export const browserOrganizationStore = new BrowserOrganizationStore()
