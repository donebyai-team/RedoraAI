import Box from '@mui/material/Box'
import { FallbackSpinner } from '../../../atoms/FallbackSpinner'
import AuthGuard from '../../../components/guard/AuthGuard'
import { DashboardLayout } from "../../../components/DashboardLayout";
import OnboardingGuard from '../../../components/guard/OnboardingGuard';

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
    <AuthGuard fallback={<AuthLoading />}>
      <OnboardingGuard fallback={<AuthLoading />}>
        <DashboardLayout>{children}</DashboardLayout>
      </OnboardingGuard>
    </AuthGuard>
  )
}
