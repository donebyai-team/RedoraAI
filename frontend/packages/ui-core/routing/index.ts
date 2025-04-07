// We keep routes definition in `ui-core` so they can be shared between the portal and the extension
// when the extension needs to construct a URL to open a portal tab.
export const routes = {
  api: {
    config: '/config',
    auth: {
      state: '/auth/state',
      callback: '/auth/callback',
      issue: '/auth/issue'
    }
  },

  app: {
    home: '/dashboard',
    settings: {
      account: '/dashboard/settings/account',
      security: '/dashboard/settings/security',
      notification: '/dashboard/settings/notification'
    },
    auth: {
      login: '/auth/login',
      callback: '/auth/callback'
    },
    serverError: '/500'
  }
}
