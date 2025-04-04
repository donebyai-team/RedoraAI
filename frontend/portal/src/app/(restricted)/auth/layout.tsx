import { Box } from '@mui/system'

export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <Box
      sx={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        overflowX: 'hidden',
        position: 'relative',
        width: '100%'
      }}
    >
      {children}
    </Box>
  )
}
