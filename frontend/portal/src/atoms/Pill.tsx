import { SxProps, lighten } from '@mui/system'
import { Box, Theme, Typography, useTheme } from '@mui/material'
import { FC } from 'react'

export type PillProps = {
  text?: string
  variant?: 'pill' | 'circle'
  color: string | ((theme: Theme) => string)
  sx?: SxProps<Theme>
  refTarget?: any
  className?: string
}

export const Pill: FC<PillProps> = ({ text, color, variant, sx, refTarget, className, ...props }) => {
  const theme = useTheme()

  let computedColor: string
  if (typeof color === 'function') {
    computedColor = color(theme)
  } else if (typeof color === 'string') {
    computedColor = color
  } else {
    throw new Error('Invalid color type')
  }

  if (variant === 'circle') {
    return (
      <Box
        ref={refTarget}
        sx={{
          display: 'inline-block',
          width: '0.75rem',
          height: '0.75rem',
          color: computedColor,
          bgcolor: lighten(computedColor, 0.2),
          borderRadius: '50%',
          ...sx
        }}
        {...props}
      />
    )
  }
  return (
    <Typography
      component='span'
      className={className}
      ref={refTarget}
      sx={{
        color: computedColor === undefined ? 'black' : computedColor,
        bgcolor: computedColor !== undefined ? lighten(computedColor, 0.8) : 'black',
        px: 3,
        borderRadius: 2,
        textTransform: 'capitalize',
        fontSize: '0.875rem',
        ...sx
      }}
      {...props}
    >
      {text}
    </Typography>
  )
}
