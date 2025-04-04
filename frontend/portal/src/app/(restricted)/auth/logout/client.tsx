'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@doota/ui-core/hooks/useAuth'
import { FallbackSpinner } from '../../../../atoms/FallbackSpinner'
import { routes } from '@doota/ui-core/routing'

const LogoutPage = () => {
  const { logout } = useAuth()
  const router = useRouter()

  useEffect(() => {
    logout()
    router.push(routes.app.auth.login)
  })

  return <FallbackSpinner />
}

export default LogoutPage
