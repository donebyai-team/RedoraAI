import React, { FC, useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { isActivePath } from '../utils/url'

export const Sidebar: FC<{ drawerWidth?: number }> = ({ drawerWidth }) => {
  const pathname = usePathname()
  const isSetting = isActivePath('/dashboard/settings', pathname)

  useEffect(() => {
  }, [])

  return (
    <>
      {isSetting && (
        <div className='w-full border-b border-neutral-200 h-[56px] px-2 flex items-center'>
          <div className='text-xl font-semibold'>Settings</div>
        </div>
      )}
    </>
  )
}
