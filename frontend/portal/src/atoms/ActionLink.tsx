import { LinkProps } from '@mui/material'
import { FC } from 'react'
import MuiLink from '@mui/material/Link'

type Props = LinkProps<'button', { active?: boolean }>

export const ActionLink: FC<Props> = ({ children, active, sx, ...props }) => {
  const sxMerged: Props['sx'] = {
    fontWeight: active ? '600' : undefined,
    ...sx
  }

  return (
    <MuiLink component='button' underline={active ? 'always' : 'none'} sx={sxMerged} {...props}>
      {children}
    </MuiLink>
  )
}
