'use client'

import React, { FC, ReactNode, useState } from 'react'
import { Grid } from '@mui/material'
import NavBar from './NavBar'

export const DashboardLayout: FC<{ children: ReactNode }> = ({ children }) => {
  const [hoverActive, setHoverActive] = useState(false)

  const handleMouseEnter = () => {
    setHoverActive(true)
  }

  const handleMouseLeave = () => {
    setHoverActive(false)
  }

  return (
    <>
      <Grid container sx={{ display: { xs: 'inline-block', sm: 'flex' } }}>
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
          <NavBar hoverActive={hoverActive} />
        </Grid>
      </Grid>
    </>
  )
}
