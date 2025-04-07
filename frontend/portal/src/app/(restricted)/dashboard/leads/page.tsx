import { Box, Typography } from '@mui/material'

export default function Page() {
  return (
      <Box
        sx={{
          paddingTop: '200px'
        }}
      >
        <Typography
          sx={{
            fontSize: '14px',
            fontWeight: '600',
            lineHeight: '19px',
            letterSpacing: '0em',
            textAlign: 'center'
          }}
        >
          Choose a conversation
        </Typography>
      </Box>
  )
}
