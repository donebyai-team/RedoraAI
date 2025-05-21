import { AuthLoading } from '@/app/(restricted)/dashboard/layout'
import AuthGuard from '@/components/guard/AuthGuard'
import DashboardLayout from '@/components/layout/dashboard'

export default function Layout({ children }: { children: React.ReactNode }) {
    return (
        <AuthGuard fallback={<AuthLoading />}>
            <DashboardLayout>{children}</DashboardLayout>
        </AuthGuard>
    )
}
