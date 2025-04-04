export type ErrCallbackType = (err: { [key: string]: string }) => void

export type ConfigType = {
  auth0Domain: string
  auth0ClientId: string
  auth0Scope: string
  msoftAuth0CallbackUrl: string
  googleAuth0CallbackUrl: string
}
