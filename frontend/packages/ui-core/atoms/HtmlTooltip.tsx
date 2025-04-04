import React, { ReactElement, ReactNode } from 'react'
import Tooltip, { TooltipProps } from '@mui/material/Tooltip'

interface Props extends Omit<TooltipProps, 'children'> {
  title: ReactNode
  children: ReactElement
}

export const HtmlTooltip: React.FC<Props> = ({ title, children, ...props }) => (
  <Tooltip title={title} placement='top-end' arrow {...props} disableInteractive>
    {children}
  </Tooltip>
)
