import { createContext, FC, ReactNode, useContext } from 'react'
import { createClients } from '@doota/client'

const apiEndpointUrl = getNextPublicApiUrl()

const defaultClients = createClients(
  getNextPublicApiUrl(),
  {
    Get: () => Promise.resolve(undefined),
    Set: async () => {},
    Del: async () => {}
  },
  {
    Get: () => Promise.resolve(undefined),
    Set: async () => {},
    Del: async () => {}
  }
)

export type ClientStore = typeof defaultClients & {
  /**
   * The URL of the API endpoint used to communicate with the server.
   */
  apiEndpointUrl: string
}

export const ClientContext = createContext<ClientStore>({ ...defaultClients, apiEndpointUrl: apiEndpointUrl ?? '' })

export const useClientsContext = (): ClientStore => useContext(ClientContext)

function getNextPublicApiUrl(): string | undefined {
  // @ts-ignore Replaced by Vite, didn't adjusted tsconfig to reflect this, will fix later
  return process.env.NEXT_PUBLIC_API_URL
}

type Props = {
  children: ReactNode
  clientStore: ClientStore
}

export const BaseClientProvider: FC<Props> = ({ clientStore, children }) => {
  return <ClientContext.Provider value={clientStore}>{children}</ClientContext.Provider>
}
