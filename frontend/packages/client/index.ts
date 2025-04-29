import { Interceptor, createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { OrganizationStore, TokenStore } from '@doota/store'
import { PortalService } from '@doota/pb/doota/portal/v1/portal_pb'

const createAuthInterceptor: (tokenStore: TokenStore) => Interceptor = tokenStore => {
  return next => async req => {
    const jwt = await tokenStore.Get()
    if (jwt !== undefined) {
      req.header.set('Authorization', `Bearer ${jwt.token}`)
    }

    return await next(req)
  }
}

const createOrgInterceptor: (organizationStore: OrganizationStore) => Interceptor = organizationStore => {
  return next => async req => {
    const organization = await organizationStore.Get()
    if (organization) {
      req.header.set('X-ORGANIZATION-ID', organization.id)
    }

    return await next(req)
  }
}

export const createClients = (
  baseUrl: string | undefined,
  tokenStore: TokenStore,
  organizationStore: OrganizationStore
) => {
  const auth = createAuthInterceptor(tokenStore)
  const org = createOrgInterceptor(organizationStore)

  // The transport defines what type of endpoint we're hitting.
  // In our example we'll be communicating with a Connect endpoint.
  const transport = createConnectTransport({
    baseUrl: baseUrl || 'http://localhost:8787',
    useBinaryFormat: false,
    interceptors: [auth, org],
  })

  return {
    portalClient: createClient(PortalService, transport),
  }
}
