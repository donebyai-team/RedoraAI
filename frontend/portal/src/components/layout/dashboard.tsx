'use client'

import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { AppSidebar } from '@/components/dashboard/AppSidebar'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <SidebarProvider>
        <div className='min-h-screen flex w-full'>
          <AppSidebar />
          <SidebarInset className='bg-gradient-to-b from-background to-secondary/20'>{children}</SidebarInset>
        </div>
      </SidebarProvider>
    </>
  )
}
