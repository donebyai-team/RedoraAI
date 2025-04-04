const AUTH_REDRIECT_PREFIX = 'aR'
export const storeAuthRedirectURI = (redirectURI: string): void => {
  localStorage.setItem(AUTH_REDRIECT_PREFIX, redirectURI)
}

export const getAuthRedirectURI = (): string => {
  return String(localStorage.getItem(AUTH_REDRIECT_PREFIX))
}
