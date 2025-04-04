import { Config, ConfigSchema } from '@doota/pb/doota/portal/v1/portal_pb'
import { portalClient } from './grpc'
import { log } from './logger'
import { create } from '@bufbuild/protobuf'

// this is present on build (i.e. http://api.freightstream.ai)
export const CONFIG_API_URI = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8787'

// this is present on build (i.e. http://app.freightstream.ai)
export const CONFIG_PORTAL_URI = process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'

export class ConfigProvider {
  config: Config

  constructor() {
    this.config = create(ConfigSchema, {
      auth0Domain: 'domain.auth0.com',
      auth0ClientId: 'xxxxxxxxxxxxxxxx',
      auth0Scope: 'openid email',
      msoftAuth0CallbackUrl: 'http://msoftcallback',
      googleAuth0CallbackUrl: 'http://googlecallback'
    })
  }

  async bootstrap(): Promise<Config> {
    this.config = await this.buildConfig()

    return this.config
  }

  async fetchFromBackend(): Promise<Config> {
    return portalClient.getConfig({})
  }

  async buildConfig(): Promise<Config> {
    const backendConfig = await this.fetchFromBackend()

    if (backendConfig === null) {
      throw new Error('No backend configuration found')
    }

    log.info('retrieve config', { config: backendConfig })

    return backendConfig
  }
}

export const configProvider = new ConfigProvider()
