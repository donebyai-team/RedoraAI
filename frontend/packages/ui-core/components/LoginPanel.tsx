import { useState } from 'react'
import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'
import { Button, TextField } from '@mui/material'
import { useClientsContext } from '../context/ClientContext'
import { errorToMessage } from '@doota/pb/utils/errors'
import { JWT } from '@doota/pb/doota/portal/v1/portal_pb'
import { FC } from 'react'
import { Link } from '../atoms/Link'
import { useIsExecutionRuntimeInPortal } from '../hooks/useExecutionRuntime'
import { Logo } from './Logo'

const validateEmail = (email: string) => {
  return String(email)
    .toLowerCase()
    .match(
      /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|.(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
    )
}

const onFormSubmit = (e: { preventDefault: () => void }) => {
  e.preventDefault()
}

type Props = {
  onPasswordlessStarted: (message: string) => void
  onPasswordlessVerified: (jwt: JWT) => Promise<void>
  onPasswordlessStartError: (message: string, error: unknown) => void
  onPasswordlessVerifyError: (message: string, error: unknown) => void
}

export const LoginPanel: FC<Props> = ({
  onPasswordlessStarted,
  onPasswordlessStartError,
  onPasswordlessVerified,
  onPasswordlessVerifyError
}) => {
  const [optState, setOPTState] = useState<'start' | 'verify'>('start')
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const { portalClient } = useClientsContext()
  const isInPortal = useIsExecutionRuntimeInPortal()

  const onOTPLogin = (resend?: boolean) => {
    if (!validateEmail(email)) {
      return
    }

    portalClient
      .passwordlessStart({ email: email })
      .then(() => {
        setOPTState('verify')
        let verb = 'sent'
        if (resend) {
          verb = 're-sent'
        }

        onPasswordlessStarted(`We've ${verb} your verification code to your email`)
      })
      .catch((err: unknown) => {
        onPasswordlessStartError(errorToMessage(err), err)
      })
  }

  const onOTPVerify = (email: string, code: string) => {
    portalClient
      .passwordlessVerify({ email: email, code: code })
      .then((jwt: JWT) => {
        console.log(`Passwordless received JWT expiring at ${new Date(Number(jwt.expiresAt))}`)

        onPasswordlessVerified(jwt)
          .then(() => {
            console.log('Logged in')
          })
          .catch(err => {
            onPasswordlessVerifyError(errorToMessage(err), err)
          })
      })
      .catch((err: unknown) => {
        onPasswordlessVerifyError(errorToMessage(err), err)
      })
  }

  function renderStartOTP() {
    return (
      <form onSubmit={onFormSubmit} className='w-full'>
        <Typography>Enter your email to sign in</Typography>
        <TextField value={email} onChange={e => setEmail(e.target.value)} fullWidth size={'small'} />
        <Button
          variant='contained'
          onClick={() => onOTPLogin()}
          disabled={!validateEmail(email)}
          type='submit'
          fullWidth
          sx={{ marginTop: '10px', marginBottom: '10px' }}
        >
          Submit
        </Button>
      </form>
    )
  }

  function renderVerifyOTP() {
    return (
      <form onSubmit={onFormSubmit}>
        <Box
          sx={{
            mb: 4,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'flex-end',
            textAlign: 'center'
          }}
        >
          <Typography variant='h6'>{`We've sent a verification code to ${email}`}</Typography>
          <Typography variant='caption'>
            {`Didn't get a code?`}{' '}
            <Link href='#' onClick={() => onOTPLogin(true)}>
              Click to resend
            </Link>
          </Typography>
          <TextField value={code} onChange={e => setCode(e.target.value)} fullWidth size={'small'} />
          <Button
            variant='contained'
            type='submit'
            fullWidth
            sx={{ marginTop: '10px' }}
            onClick={() => {
              onOTPVerify(email, code)
            }}
          >
            Verify
          </Button>
          <Typography variant='caption'>Code expires in 5 minutes.</Typography>
          <Typography variant='caption'>
            <Link href='#' onClick={() => setOPTState('start')}>
              Try another email.
            </Link>
          </Typography>
        </Box>
      </form>
    )
  }

  let topBoxSx = {}
  if (isInPortal) {
    topBoxSx = { borderWidth: 1, borderStyle: 'solid', borderColor: 'divider' }
  }

  return (
    <Box sx={topBoxSx}>
      <Box
        sx={{
          p: 6,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          backgroundColor: 'background.paper'
        }}
      >
        <Box sx={{ width: '100%' }}>
          <Box
            sx={{
              display: 'flex',
              justifyContent: 'center',
              mb: 6
            }}
          >
            <div className='flex w-full justify-center'>
              <Logo />
            </div>
          </Box>
          {optState == 'start' && renderStartOTP()}
          {optState == 'verify' && renderVerifyOTP()}
        </Box>
      </Box>
    </Box>
  )
}
