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
    home: '/dashboard/leads',
    settings: {
      account: '/dashboard/settings/integrations',
      security: '/dashboard/settings/security',
      notification: '/dashboard/settings/notification',
      edit_product: '/dashboard/settings/product'
    },
    auth: {
      login: '/auth/login',
      callback: '/auth/callback',
      onboarding: '/onboarding'
    },
    serverError: '/500'
  },

  new: {
    dashboard: '/dashboard',
    keywords: '/keywords',
    leads: '/leads',
    interactions: '/interactions',
    insights: '/insights',
    edit_product: '/settings/product',
    integrations: '/settings/integrations',
    automation: '/settings/automation',
    billing: '/settings/billing',
    onboarding: '/onboarding',
  }
}

export const isPublicRoute = (path: string): boolean => {
  const publicRoutes = [routes.app.auth.login, routes.app.auth.callback]
  return publicRoutes.some(route => {
    return path.startsWith(route)
  })
}
