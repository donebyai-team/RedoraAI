'use client'

import { ReactNode, FC } from 'react'
import { BaseExecutionRuntimeProvider } from '@doota/ui-core/context/ExecutionRuntimeContext'
import toast from 'react-hot-toast'

export const PortalExecutionRuntimeProvider: FC<{ children: ReactNode }> = ({ children }) => {
  return (
    <BaseExecutionRuntimeProvider
      store={{
        runtime: 'portal',

        // All images here are relative to '<root>/frontend/portal/public' folder!
        imageSourceRegistry: {
          logo: '/images/logo.svg',
          logo_new: '/images/logo_new.png',
          dat_logo: '/images/dat_logo.png',
          dat_fuel_gauge_logo: '/images/dat_fuel_gauge_logo.png'
        },
        action: {
          composeNewEmail: (body: string) => {
            navigator.clipboard
              .writeText(body)
              .then(() => toast.success('Email draft copied to clipboard!'))
              .catch(err => console.warn('Failed to copy email draft to clipboard: ', err))
          },
          replyToEmail: (_emailId: string, body: string) => {
            navigator.clipboard
              .writeText(body)
              .then(() => toast.success('Email draft copied to clipboard!'))
              .catch(err => console.warn('Failed to copy email draft to clipboard: ', err))
          }
        }
      }}
    >
      {children}
    </BaseExecutionRuntimeProvider>
  )
}
