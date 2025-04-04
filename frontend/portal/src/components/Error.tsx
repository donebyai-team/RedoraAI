import { FC } from 'react'
import { Typography } from '@mui/material'
import CrisisAlert from '@mui/icons-material/CrisisAlert'
import { Box, SxProps } from '@mui/system'
import { errorToMessage } from '@doota/pb/utils/errors'

export const FocusPanel: FC<{ children: React.ReactNode; sx?: SxProps }> = ({ children, sx }) => (
  <Box sx={{ borderRadius: 2, bgcolor: 'customColors.focusPanelBg', p: 4, mb: 6, ...sx }}>{children}</Box>
)

export const Error: FC<{ error: Error; title?: string }> = ({ error, title }) => (
  <FocusPanel>
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
      <CrisisAlert />
      <Typography variant='h5'>{title || 'Error'}</Typography>
    </Box>
    <Typography variant='caption'>{errorToMessage(error)}</Typography>
  </FocusPanel>
)
