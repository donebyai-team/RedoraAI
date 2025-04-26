'use client'

import React, { FC, ReactNode } from 'react'
import { Grid } from '@mui/material'
import NavBar from './NavBar'
// import { Sidebar } from './Sidebar'
import { usePathname } from 'next/navigation'
import { isActivePath } from '../utils/url'

export const DashboardLayout: FC<{ children: ReactNode }> = ({ children }) => {
  // const [hoverActive, setHoverActive] = useState(false)
  const pathname = usePathname()
  const isSetting = isActivePath('/dashboard/settings', pathname)

  // const handleMouseEnter = () => {
  //   setHoverActive(true)
  // }

  // const handleMouseLeave = () => {
  //   setHoverActive(false)
  // }

  return (
    <>
      <Grid container sx={{ display: { xs: 'inline-block', sm: 'flex' } }}>
        {/* <Grid
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
        </Grid> */}

        <NavBar />

        <div className={isSetting ? 'flex flex-col flex-1' : 'contents'}>
          {/* <Sidebar/> */}

          {/* <Grid item className={!isSetting ? 'sm:h-dvh flex-1 w-0' : 'flex-1 bg-neutral-50 h-full p-2'}>             */}
            <Grid item className={'flex-1 bg-neutral-50 h-full p-2'}>            
            {children}
          </Grid>            

        </div>

      </Grid>
    </>
  )
}
