'use client'

import OnboadingForm from '@/components/Onboarding/MainForm';
import { DashboardFooter } from '@/components/dashboard/DashboardFooter';
import { DashboardHeader } from '@/components/dashboard/DashboardHeader';

export default function Page() {

    return (<>
        <div className="min-h-screen flex flex-col bg-gradient-to-b from-background to-secondary/20">
            <DashboardHeader />

            <div className="flex-1 overflow-auto">
                <main className="container mx-auto px-4 py-6 md:px-6">
                    <OnboadingForm />
                </main>
            </div>

            <DashboardFooter />
        </div>


    </>);
}
