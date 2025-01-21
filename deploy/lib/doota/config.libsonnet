{
  _images+:: {
    local top = self,
    img(name, tag, env=''): 'us-east1-docker.pkg.dev/doota/docker/%s:%s%s' % [name, tag, if env != '' then '.' + env else ''],

    latest_tag: error 'needs to set "latest_tag"',
    frontend_env: '',
    frontend: self.img('frontend', self.latest_tag, top.frontend_env),
    portalApi: self.img('backend', self.latest_tag),
    migrator: self.img('backend', self.latest_tag),
    sqlBastion: 'gcr.io/cloudsql-docker/gce-proxy:1.22.0',
  },

  _config+:: {
    local top = self,

    default_gcp_project: error 'needs to set default_gcp_project',
    default_frontend_port: error 'needs to set default_frontend_port',
    default_portal_api_port: error 'needs to set default_portal_api_port',
    default_gpt_model: error 'needs to set default_gpt_model',
    default_openai_debug_store: top.default_openai_debug_store,
    default_tracing: '',
    svcAccounts: [],


    default_database: {
      project: 'doota',
      region: 'us-east1',
      instance: 'doota',
      username: error 'needs to set default_database.username',
      name: error 'needs to set default_database.name',
      host: '127.0.0.1',
      secret: 'sql-database-pword',
      encryption_key: 'sql-database-encryption-key',
    },

    bastion: {
      name: 'sql-bastion',
      service_account: 'sql-bastion-app',

      project: 'doota',
      region: 'us-east1',
      resource: {
        requests: ['50m', '10Mi'],
        limits: ['100m', '1Gi'],
      },

      databases: [],
    },

    migrator: {
      name: 'migrator',
      sql_proxy_img: 'gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.8.2',
      database: top.default_database,
      desiredVersion: error 'needs to set desiredVersion',
      enableAutoMigration: error 'needs to set enableAutoMigration',
      resources: {
        requests: ['200m', '50Mi'],
        limits: ['1', '200Mi'],
      },
    },

    portalApi: {
      replicas: 1,
      name: 'portal-api',
      cors_url_regex_allow: error 'need to set cors_url_regex_allow',
      jwt_kms: error 'need to set jwt_kms',
      auth0: {
        domain: error 'need to set auth.domain',
        api_redirect: error 'need to set auth.api_redirect',
      },
      database: top.default_database,
      gcp_project: top.default_gcp_project,
      http_listen_port: top.default_portal_api_port,
      service_account: error 'need to set portalApi.service_account',
      resources: {
        requests: ['200m', '50Mi'],
        limits: ['1', '200Mi'],
      },
    },

    frontend: {
      replicas: 2,
      name: 'frontend',
      default_http_port: 3000,
      resources: {
        requests: ['200m', '50Mi'],
        limits: ['1', '200Mi'],
      },
    },
  },
}
