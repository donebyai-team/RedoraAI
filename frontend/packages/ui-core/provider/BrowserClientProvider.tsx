import { ReactNode, FC } from 'react'
import { BaseClientProvider } from '../context/ClientContext'
import { createClients } from '@doota/client'

type Clients = ReturnType<typeof createClients>

export const BrowserClientProvider: FC<{ children: ReactNode; clients: Clients & { apiEndpointUrl: string } }> = ({
  clients,
  children
}) => {
  return <BaseClientProvider clientStore={clients}>{children}</BaseClientProvider>
}
