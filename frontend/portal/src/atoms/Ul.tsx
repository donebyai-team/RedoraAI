import { OverrideProps } from '@mui/material/OverridableComponent'
import { Box, BoxTypeMap } from '@mui/system'
import { FC } from 'react'

type UlProps = OverrideProps<BoxTypeMap, 'ul'>

export const Ul: FC<UlProps> = ({ children, sx, ...props }) => {
  const sxMerged: UlProps['sx'] = {
    display: 'flex',
    listStyleType: 'none',
    flexDirection: 'column',
    gap: 4,
    paddingInlineStart: 4,
    ...sx
  }

  return (
    <Box component='ul' sx={sxMerged} {...props}>
      {children}
    </Box>
  )
}
