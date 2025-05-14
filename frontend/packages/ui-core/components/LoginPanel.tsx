// import { useState } from 'react'
import Box from '@mui/material/Box'
import Typography from '@mui/material/Typography'
import {
  Button,
  styled,
  // TextField 
} from '@mui/material'
import { useClientsContext } from '../context/ClientContext'
import { errorToMessage } from '@doota/pb/utils/errors'
import { IntegrationType, JWT } from '@doota/pb/doota/portal/v1/portal_pb'
import { FC } from 'react'
// import { Link } from '../atoms/Link'
import { useIsExecutionRuntimeInPortal } from '../hooks/useExecutionRuntime'
import { Logo } from './Logo'
import { buildAppUrl } from '../../../portal/src/app/routes'
import { routes } from '../routing'

const GoogleIcon = () => (
  <svg width="18" height="18" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48">
    <path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z" />
    <path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z" />
    <path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z" />
    <path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z" />
    <path fill="none" d="M0 0h48v48H0z" />
  </svg>
);

const SocialButton = styled(Button)({
  width: 'auto',
  justifyContent: 'flex-start',
  textAlign: 'left',
  padding: '12px 24px',
  border: '1px solid #E0E0E0',
  borderRadius: 8,
  color: '#333',
  backgroundColor: '#fff',
  transition: 'all 0.2s ease-in-out',
  '&:hover': {
    backgroundColor: '#f8f8f8',
    borderColor: '#d0d0d0',
    transform: 'translateY(-1px)',
  },
  '& .MuiButton-startIcon': {
    marginRight: 16,
  },
  boxShadow: '0 1px 2px rgba(0, 0, 0, 0.05)',
  fontSize: '1rem',
  fontWeight: 500,
});

// const validateEmail = (email: string) => {
//   return String(email)
//     .toLowerCase()
//     .match(
//       /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|.(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
//     )
// }

// const onFormSubmit = (e: { preventDefault: () => void }) => {
//   e.preventDefault()
// }

type Props = {
  onPasswordlessStarted: (message: string) => void
  onPasswordlessVerified: (jwt: JWT) => Promise<void>
  onPasswordlessStartError: (message: string, error: unknown) => void
  onPasswordlessVerifyError: (message: string, error: unknown) => void
}

export const LoginPanel: FC<Props> = ({
  // onPasswordlessStarted,
  onPasswordlessStartError,
  // onPasswordlessVerified,
  // onPasswordlessVerifyError
}) => {
  // const [optState, setOPTState] = useState<'start' | 'verify'>('start')
  // const [email, setEmail] = useState('')
  // const [code, setCode] = useState('')
  const { portalClient } = useClientsContext()
  const isInPortal = useIsExecutionRuntimeInPortal()

  // const onOTPLogin = (resend?: boolean) => {
  //   if (!validateEmail(email)) {
  //     return
  //   }

  //   portalClient
  //     .passwordlessStart({ email: email })
  //     .then(() => {
  //       setOPTState('verify')
  //       let verb = 'sent'
  //       if (resend) {
  //         verb = 're-sent'
  //       }

  //       onPasswordlessStarted(`We've ${verb} your verification code to your email`)
  //     })
  //     .catch((err: unknown) => {
  //       onPasswordlessStartError(errorToMessage(err), err)
  //     })
  // }

  // const onOTPVerify = (email: string, code: string) => {
  //   portalClient
  //     .passwordlessVerify({ email: email, code: code })
  //     .then((jwt: JWT) => {
  //       console.log(`Passwordless received JWT expiring at ${new Date(Number(jwt.expiresAt))}`)

  //       onPasswordlessVerified(jwt)
  //         .then(() => {
  //           console.log('Logged in')
  //         })
  //         .catch(err => {
  //           onPasswordlessVerifyError(errorToMessage(err), err)
  //         })
  //     })
  //     .catch((err: unknown) => {
  //       onPasswordlessVerifyError(errorToMessage(err), err)
  //     })
  // }

  // function renderStartOTP() {
  //   return (
  //     <form onSubmit={onFormSubmit} className='w-full'>
  //       <Typography>Enter your email to sign in</Typography>
  //       <TextField value={email} onChange={e => setEmail(e.target.value)} fullWidth size={'small'} />
  //       <Button
  //         variant='contained'
  //         onClick={() => onOTPLogin()}
  //         disabled={!validateEmail(email)}
  //         type='submit'
  //         fullWidth
  //         sx={{ marginTop: '10px', marginBottom: '10px' }}
  //       >
  //         Submit
  //       </Button>
  //     </form>
  //   )
  // }

  // function renderVerifyOTP() {
  //   return (
  //     <form onSubmit={onFormSubmit}>
  //       <Box
  //         sx={{
  //           mb: 4,
  //           display: 'flex',
  //           flexDirection: 'column',
  //           alignItems: 'center',
  //           justifyContent: 'flex-end',
  //           textAlign: 'center'
  //         }}
  //       >
  //         <Typography variant='h6'>{`We've sent a verification code to ${email}`}</Typography>
  //         <Typography variant='caption'>
  //           {`Didn't get a code?`}{' '}
  //           <Link href='#' onClick={() => onOTPLogin(true)}>
  //             Click to resend
  //           </Link>
  //         </Typography>
  //         <TextField value={code} onChange={e => setCode(e.target.value)} fullWidth size={'small'} />
  //         <Button
  //           variant='contained'
  //           type='submit'
  //           fullWidth
  //           sx={{ marginTop: '10px' }}
  //           onClick={() => {
  //             onOTPVerify(email, code)
  //           }}
  //         >
  //           Verify
  //         </Button>
  //         <Typography variant='caption'>Code expires in 5 minutes.</Typography>
  //         <Typography variant='caption'>
  //           <Link href='#' onClick={() => setOPTState('start')}>
  //             Try another email.
  //           </Link>
  //         </Typography>
  //       </Box>
  //     </form>
  //   )
  // }

  const handleLoginWithGoogleButton = () => {

    portalClient
      .oauthAuthorize({
        integrationType: IntegrationType.GOOGLE,
        redirectUrl: buildAppUrl(routes.app.auth.callback)
      })
      .then(oAuthAuthorizeResp => {
        window.open(oAuthAuthorizeResp.authorizeUrl, '_self')
      })
      .catch((err: unknown) => {
        onPasswordlessStartError(errorToMessage(err), err)
      })
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
          backgroundColor: 'background.paper',
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

          <Box display={"flex"} justifyContent={"center"} alignItems={"center"} width={"100%"}>
            <Typography color="text.secondary" mb={5}>
              Redora AI
            </Typography>
          </Box>
          <Box display={"flex"} justifyContent={"center"} alignItems={"center"} width={"100%"} mb={5}>
            <SocialButton
              variant="outlined"
              startIcon={<GoogleIcon />}
              onClick={handleLoginWithGoogleButton}
            >
              Continue with Google
            </SocialButton>
          </Box>

          {/* {optState == 'start' && renderStartOTP()} */}
          {/* {optState == 'verify' && renderVerifyOTP()} */}
        </Box>
      </Box>
    </Box>
  )
}