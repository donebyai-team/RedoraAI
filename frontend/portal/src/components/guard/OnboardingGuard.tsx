'use client'

import { ReactNode, ReactElement, useEffect } from 'react'
import { useRouter, usePathname } from 'next/navigation'
import { routes } from '@doota/ui-core/routing'
import { useOnboardingStatus } from '../../hooks/useOnboardingStatus'

interface OnboardingGuardProps {
    children: ReactNode
    fallback: ReactElement | null
}

const OnboardingGuard = ({ children, fallback }: OnboardingGuardProps) => {
    const router = useRouter()
    const path = usePathname()
    const { loading, isOnboardingDone } = useOnboardingStatus()

    useEffect(() => {
        if (loading || isOnboardingDone === null) return

        if (!isOnboardingDone && !path.startsWith(routes.app.auth.onboarding)) {
            router.replace(routes.app.auth.onboarding)
        } else if (isOnboardingDone && path.startsWith(routes.app.auth.onboarding)) {
            router.replace(routes.app.home)
        }
    }, [isOnboardingDone, loading, path, router])

    if (loading || isOnboardingDone === null) {
        return fallback
    }

    return <>{children}</>
}

export default OnboardingGuard
