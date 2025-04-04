import { LinkProps } from '@mui/material'
import { FC } from 'react'
import NextLink from 'next/link'
import MuiLink from '@mui/material/Link'

type Props = LinkProps<typeof NextLink, { active?: boolean }>

export const Link: FC<Props> = ({ children, active, sx, ...props }) => {
  const sxMerged: Props['sx'] = {
    fontWeight: active ? '600' : undefined,
    ...sx
  }

  return (
    <MuiLink component={NextLink} underline={active ? 'always' : 'none'} sx={sxMerged} {...props}>
      {children}
    </MuiLink>
  )
}
