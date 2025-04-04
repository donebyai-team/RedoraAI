import { Authentication, WebAuth } from 'auth0-js'
import { configProvider } from '../config'

export const auth0 = () => {
  return new WebAuth({
    clientID: configProvider.config.auth0ClientId,
    domain: configProvider.config.auth0Domain,
    responseType: 'token id_token',
    scope: configProvider.config.auth0Scope
  })
}

/**
 * Used to construct a url for auth0 social login
 */
export const authentication = () => {
  return new Authentication({
    clientID: configProvider.config.auth0ClientId,
    domain: configProvider.config.auth0Domain,
    responseType: 'token id_token',
    scope: configProvider.config.auth0Scope
  })
}
