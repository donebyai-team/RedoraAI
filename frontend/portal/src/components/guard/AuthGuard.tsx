'use client'

import { ReactNode, ReactElement, useEffect, useState, useRef } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@doota/ui-core/hooks/useAuth'
import { browserTokenStore } from '@doota/ui-core/provider/BrowserStores'
import { routes } from '@doota/ui-core/routing'
import { useAppDispatch } from '../../../store/hooks'
import {
  goToStep,
  setIsOnboardingDone,
  setProject,
} from '../../../store/Onboarding/OnboardingSlice'
import { Project } from '@doota/pb/doota/core/v1/core_pb'
import { isActivePath } from '../../utils/url'

interface AuthGuardProps {
  children: ReactNode
  fallback: ReactElement | null
}

const AuthGuard = ({ children, fallback }: AuthGuardProps) => {
  const { user, loading: authLoading } = useAuth()
  const router = useRouter()
  const path = usePathname()
  const dispatch = useAppDispatch()
  const [isReady, setIsReady] = useState(false)

  const isEditProduct = isActivePath(routes.new.edit_product, path);

  function calculateNextStep(data: Project | null): number {
    if (!data) return 1;
    const { id, website, name, description, targetPersona, keywords, sources } = data
    const hasBasicInfo = Boolean(id && website && name && description && targetPersona)
    const hasKeywords = Array.isArray(keywords) && keywords.length > 0
    const hasSources = Array.isArray(sources) && sources.length > 0
    if (hasBasicInfo && hasKeywords && hasSources) return 3
    if (hasBasicInfo && hasKeywords) return 3
    if (hasBasicInfo) return 2
    return 1
  }

  const hasCheckedInitialLoad = useRef(false);

  useEffect(() => {
    const checkAuthAndSetup = async () => {
      if (authLoading) return

      if (!user) {
        const token = await browserTokenStore.Get()
        if (!token) {
          router.replace(routes.app.auth.login)
          return
        }
      }

      if (user) {
        const data = user.projects?.[0]
        const isOnboardingDone = user.isOnboardingDone

        const nextStep = calculateNextStep(data)
        dispatch(setIsOnboardingDone(isOnboardingDone))
        if (!hasCheckedInitialLoad.current) {
          dispatch(setProject(data));
          hasCheckedInitialLoad.current = true;
        }
        if (!isEditProduct) {
          dispatch(goToStep(nextStep))
        }

        if (!isOnboardingDone && !path.startsWith(routes.app.auth.onboarding)) {
          router.replace(routes.app.auth.onboarding)
          return
        } else if (isOnboardingDone && path.startsWith(routes.app.auth.onboarding)) {
          router.replace(routes.new.dashboard)
          return
        }
      }

      // Mark as ready ONLY after everything else
      setIsReady(true)
    }

    checkAuthAndSetup()
  }, [authLoading, user, path, dispatch, router, isEditProduct])

  // Strict block on rendering until ready
  if (!isReady) {
    return fallback
  }

  return <>{children}</>
}

export default AuthGuard