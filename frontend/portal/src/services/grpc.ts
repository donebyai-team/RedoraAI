import { browserTokenStore, browserOrganizationStore } from '@doota/ui-core/provider/BrowserStores'
import { createClients } from '@doota/client'

const apiEndpointUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8787'

export const frontendClients = {
  apiEndpointUrl,
  ...createClients(apiEndpointUrl, browserTokenStore, browserOrganizationStore)
}

export const portalClient = frontendClients.portalClient