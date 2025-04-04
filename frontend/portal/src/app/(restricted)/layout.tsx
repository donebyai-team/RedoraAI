import Box from '@mui/material/Box'
import { PortalAuthProvider } from '../../provider/PortalAuthProvider'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <PortalAuthProvider>
      <Box sx={{ display: 'flex', minHeight: '100vh' }}>{children}</Box>
    </PortalAuthProvider>
  )
}
