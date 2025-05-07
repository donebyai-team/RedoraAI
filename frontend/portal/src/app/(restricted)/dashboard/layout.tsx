import Box from '@mui/material/Box'
import { FallbackSpinner } from '../../../atoms/FallbackSpinner'
import AuthGuard from '../../../components/guard/AuthGuard'
import { DashboardLayout } from "../../../components/DashboardLayout";
import OnboardingGuard from '../../../components/guard/OnboardingGuard';

export const metadata = {
  title: 'RedoraAI',
  description: ''
}

export const AuthLoading = () => (
  <Box
    sx={{
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '100vh',
      overflowX: 'hidden',
      position: 'relative',
      width: '100%'
    }}
  >
    <FallbackSpinner />
  </Box>
)

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <OnboardingGuard fallback={<AuthLoading />}>
      <AuthGuard fallback={<AuthLoading />}>
        <DashboardLayout>{children}</DashboardLayout>
      </AuthGuard>
    </OnboardingGuard>
  )
}
