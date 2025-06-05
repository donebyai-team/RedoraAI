'use client'

import React from 'react'
import { TooltipProvider } from '@/components/ui/tooltip'
import { Toaster } from '@/components/ui/toaster'
import { Toaster as Sonner } from '@/components/ui/sonner'

function NotificationProvider({ children }: { children: React.ReactNode }) {
  return (
    <>
      <TooltipProvider>
        <Sonner />
        <Toaster />
        {children}
      </TooltipProvider>
    </>
  )
}

export default NotificationProvider
