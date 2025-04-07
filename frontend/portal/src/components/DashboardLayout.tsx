'use client'

import React, { FC, ReactNode, useState } from 'react'
import { Grid } from '@mui/material'
import NavBar from './NavBar'
import { Sidebar } from './Sidebar'
import { usePathname } from 'next/navigation'
import { isActivePath } from '../utils/url'

export const DashboardLayout: FC<{ children: ReactNode }> = ({ children }) => {
  const [hoverActive, setHoverActive] = useState(false)
  const pathname = usePathname()
  const isSetting = isActivePath('/dashboard/settings', pathname)

  const handleMouseEnter = () => {
    setHoverActive(true)
  }

  const handleMouseLeave = () => {
    setHoverActive(false)
  }

  return (
    <>
      <Grid container sx={{display: {xs: 'inline-block', sm: 'flex'}}}>
        <Grid
            item
            xs={12}
            sm='auto'
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
            className={`group bg-neutral-950 h-[60px] sm:h-dvh transition-all ease-in-out duration-300 !w-full ${
                hoverActive ? 'sm:!w-[300px]' : 'sm:!w-[65px]'
            }`}
        >
          <NavBar hoverActive={hoverActive}/>
        </Grid>

        <div className={isSetting ? 'flex flex-col flex-1' : 'contents'}>
          <Sidebar/>
        </div>

      </Grid>
    </>
  )
}
