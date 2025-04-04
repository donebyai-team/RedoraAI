import { LoadingButtonProps, LoadingButton as MuiLoadingButton } from '@mui/lab'
import { ButtonProps, Button as MuiButton, styled } from '@mui/material'
import { FC } from 'react'

const Btn = styled(MuiButton)(() => ({
  // boxShadow: '0px 4px 4px 0px #0000000D',
  // padding: '10px 9px 10px 9px',
  borderRadius: '10px',
  gap: '10px',
  fontSize: '12px',
  fontWeight: '700',
  textTransform: 'none'
}))

const LoadingBtn = styled(MuiLoadingButton)(() => ({
  boxShadow: '0px 4px 4px 0px #0000000D',
  padding: '6px 17px',
  borderRadius: '10px',
  gap: '10px',
  fontSize: '14px',
  fontWeight: '700',
  height: '44px',
  textTransform: 'none'
}))

export const Button: FC<ButtonProps> = props => {
  return <Btn {...props} />
}

export const LoadingButton: FC<LoadingButtonProps> = props => {
  return <LoadingBtn {...props} />
}
