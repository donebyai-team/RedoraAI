'use client'

import React from 'react'
import { routes } from '@doota/ui-core/routing'
import toast from 'react-hot-toast'
import { useAuth } from '@doota/ui-core/hooks/useAuth'
import { JWT } from '@doota/pb/doota/portal/v1/portal_pb'
import { LoginPanel } from '@doota/ui-core/components/LoginPanel'
import { Box } from '@mui/material'

export default function Page() {
  const { login } = useAuth()

  return (
    <Box sx={{ width: ['100%', '100%', '450px'] }}>
      <LoginPanel
        onPasswordlessStarted={toast.success}
        onPasswordlessStartError={(errorMessage: string) => {
          toast.error(errorMessage)
        }}
        onPasswordlessVerified={async (jwt: JWT) => {
          return login(jwt).then(() => {
            window.location.href = routes.app.home
          })
        }}
        onPasswordlessVerifyError={(errorMessage: string) => {
          toast.error(errorMessage)
        }}
      />
    </Box>
  )
}
