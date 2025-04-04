'use client'

// ** MUI Imports
import Box, { BoxProps } from '@mui/material/Box'
import CircularProgress from '@mui/material/CircularProgress'
import Image from 'next/image'
import { useEffect, useState } from 'react'

export const FallbackSpinner = ({ sx }: { sx?: BoxProps['sx'] }) => {
  const [inInitializationGracePeriod, setInInitializationGracePeriod] = useState(true)

  useEffect(() => {
    const timeout = setTimeout(() => {
      setInInitializationGracePeriod(false)
    }, 200)

    return () => clearTimeout(timeout)
  }, [])

  return (
    <Box
      sx={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        flexDirection: 'column',
        justifyContent: 'center',
        ...sx
      }}
    >
      {inInitializationGracePeriod ? null : (
        <>
          <Image width={150} height={50} alt='freightstream' priority src={'/images/logo.svg'} />
          <CircularProgress disableShrink sx={{ mt: 6, color: 'grey.600' }} />
        </>
      )}
    </Box>
  )
}
