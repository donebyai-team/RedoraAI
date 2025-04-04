import { createContext, FC, ReactNode } from 'react'

export type ExecutionRuntime = ExecutionRuntimeStore['runtime']

export type ExecutionRuntimeStore = {
  runtime: 'portal' | 'chrome_extension' | 'outlook_addin'

  imageSourceRegistry: {
    logo: string
    logo_new: string
    dat_logo: string
    dat_fuel_gauge_logo: string
  }

  action: {
    replyToEmail: (emailId: string, body: string) => void
    composeNewEmail: (body: string, to?: string, subject?: string) => void
  }
}

// Assume portal runtime if there is no provider bound that sets the value, this avoid the need for
// use to add a specific portal provider.
export const ExecutionRuntimeContext = createContext<ExecutionRuntimeStore>({
  runtime: 'portal',

  // All images here are relative to '<root>/frontend/portal/public' folder!
  imageSourceRegistry: {
    logo: '/images/logo.svg',
    // In Browser for now, logo_new is the same as logo
    logo_new: '/images/logo.svg',

    dat_logo: '/images/dat_logo.png',
    dat_fuel_gauge_logo: '/images/dat_fuel_gauge_logo.png'
  },
  action: {
    composeNewEmail: () => {},
    replyToEmail: () => {}
  }
})

type Props = {
  children: ReactNode
  store: ExecutionRuntimeStore
}

export const BaseExecutionRuntimeProvider: FC<Props> = ({ store, children }) => {
  return <ExecutionRuntimeContext.Provider value={store}>{children}</ExecutionRuntimeContext.Provider>
}
