'use client'
import { ReactNode, FC } from 'react'
import { BaseElementRegistryProvider } from '@doota/ui-core/context/ElementRegistryContext'
import Link from 'next/link'
import Image from 'next/image'

export const NextElementRegistryProvider: FC<{ children: ReactNode }> = ({ children }) => {
  return (
    <BaseElementRegistryProvider Link={Link} Image={Image}>
      {children}
    </BaseElementRegistryProvider>
  )
}
