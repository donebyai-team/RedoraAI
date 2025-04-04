import { FallbackSpinner } from '../atoms/FallbackSpinner'
import { Box } from '@mui/system'

export default function Loading() {
  return (
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
}
