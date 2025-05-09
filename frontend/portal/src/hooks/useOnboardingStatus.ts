'use client';

import { Project } from "@doota/pb/doota/core/v1/core_pb";
import { useClientsContext } from "@doota/ui-core/context/ClientContext"
import { useEffect, useState } from "react"


interface OnboardingStatus {
    loading: boolean;
    data: Project | null | undefined;
    isOnboardingDone: boolean | null;
    error: Error | null;
}

export const useOnboardingStatus = (): OnboardingStatus => {
    const { portalClient } = useClientsContext()
    const [loading, setLoading] = useState(true)
    const [isOnboardingDone, setIsOnboardingDone] = useState<boolean | null>(null)
    const [data, seData] = useState<Project | null | undefined>(undefined)
    const [error, setError] = useState<Error | null>(null)

    useEffect(() => {
        const fetchStatus = async () => {
            setLoading(true)
            try {
                const data = await portalClient.getProjects({})
                setIsOnboardingDone(data?.isOnboardingDone ?? false)
                seData(data.projects?.[0] ?? null)
                setError(null)
            } catch (err: any) {
                console.error('Onboarding check failed:', err)
                setError(err)
                seData(null)
                setIsOnboardingDone(null)
            } finally {
                setLoading(false)
            }
        }

        fetchStatus()
    }, [portalClient])

    return { loading, data, isOnboardingDone, error }
}