'use client'

import { ReactNode, ReactElement, useEffect, useState } from 'react'
import { usePathname, useRouter } from 'next/navigation'
import { useAuth } from '@doota/ui-core/hooks/useAuth'
import { browserTokenStore } from '@doota/ui-core/provider/BrowserStores'
import { routes } from '@doota/ui-core/routing'
import { useAppDispatch } from '../../../store/hooks'
import {
  ProjectTypes,
  setIsOnboardingDone,
  setProjects,
  setStep,
} from '../../../store/Onboarding/OnboardingSlice'

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

  function calculateNextStep(data: ProjectTypes): number {
    const { id, website, name, description, targetPersona, keywords, sources } = data
    const hasBasicInfo = Boolean(id && website && name && description && targetPersona)
    const hasKeywords = Array.isArray(keywords) && keywords.length > 0
    const hasSources = Array.isArray(sources) && sources.length > 0
    if (hasBasicInfo && hasKeywords && hasSources) return 2
    if (hasBasicInfo && hasKeywords) return 2
    if (hasBasicInfo) return 1
    return 0
  }

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

        const newData = {
          id: data?.id ?? "",
          name: data?.name ?? "",
          description: data?.description ?? "",
          website: data?.website ?? "",
          targetPersona: data?.targetPersona ?? "",
          keywords: data?.keywords?.map(keyword => keyword.name) ?? [],
          sources: data?.sources?.map(source => ({ id: source.id, name: source.name })) ?? [],
          suggestedKeywords: data?.suggestedKeywords ?? [],
          suggestedSources: data?.suggestedSources ?? [],
        }

        const nextStep = calculateNextStep(newData)
        dispatch(setIsOnboardingDone(isOnboardingDone))
        dispatch(setProjects(newData))
        dispatch(setStep(nextStep))

        if (!isOnboardingDone && !path.startsWith(routes.app.auth.onboarding)) {
          router.replace(routes.app.auth.onboarding)
          return
        } else if (isOnboardingDone && path.startsWith(routes.app.auth.onboarding)) {
          router.replace(routes.app.home)
          return
        }
      }

      // Mark as ready ONLY after everything else
      setIsReady(true)
    }

    checkAuthAndSetup()
  }, [authLoading, user, path, dispatch, router])

  // Strict block on rendering until ready
  if (!isReady) {
    return fallback
  }

  return <>{children}</>
}

export default AuthGuard