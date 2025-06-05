'use client'

import { DashboardFooter } from '@/components/dashboard/DashboardFooter';
import { DashboardHeader } from '@/components/dashboard/DashboardHeader';
import Onboarding from '@/components/pages/Onboarding';

export default function Page() {

    return (<>
        <div className="min-h-screen flex flex-col bg-gradient-to-b from-background to-secondary/20">
            <DashboardHeader />
            <Onboarding />
            <DashboardFooter />
        </div>
    </>);
}
