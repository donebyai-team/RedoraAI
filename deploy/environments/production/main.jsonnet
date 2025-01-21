local k = import 'sf/k8s.libsonnet';  // some constructor overrides over grafana's "kausal"
local ingress = k.networking.v1.ingress;


(import 'doota/doota.libsonnet') +
{
  local doota = super.doota,

  _images+:: {
    latest_tag: 'v1.0.11',
  },

  public_interface: doota.newGKEPublicInterface($._config.public_interface),
  frontend: doota.newFrontend($._images.frontend, $._config.frontend),
  portalApi: doota.newPortalApi($._images.portalApi, $._config.portalApi),
  migrator: doota.newMigrator($._images.migrator, $._config.migrator),
  sql_bastion: doota.newSqlBastion($._images.sqlBastion, $._config.bastion),

  redisTool: doota.newRedisTool(),
  svcAccount: doota.newServiceAccount([
    'portal-api-app-prod',
    'sql-bastion-app',
  ]),

  _config+:: {
    local c = self,

    apiDomain: 'api.dootaai.com',
    appDomain: 'app.dootaai.com',

    default_redis_addr: '10.19.182.227:6379',
    default_gcp_project: 'doota',
    default_frontend_port: 3000,
    default_portal_api_port: 9000,
    default_gpt_model: 'gpt-4o-2024-08-06',
    default_openai_debug_store: 'gs://doota-ai-debug-prod',
//    default_tracing: 'cloudtrace://?project_id=doota&ratio=1',

    default_database+:: {
      project: 'doota',
      region: 'us-east1',
      instance: 'doota',
      username: 'doota',
      name: 'doota-prod',
      host: '127.0.0.1',
      secret: 'sql-database-pword',
    },

    migrator+:: {
      desiredVersion: 55,
      enableAutoMigration: false,
    },

    portalApi+:: {
      replicas: 2,
      resources: {
        requests: ['1', '50Mi'],
        limits: ['2', '500Mi'],
      },
      service_account: 'portal-api-app-prod',

      jwt_kms: 'projects/doota/locations/global/keyRings/api-auth/cryptoKeys/jwt-signing/cryptoKeyVersions/1',
      // The `chrome://` is there to allow CORS being performed by the Chrome Extension for which origin is `chrome://<id>`.
      // It's possible that published extension has a fixed <id> in which case we could protect the CORS to be performed
      // only from our extension (update 'staging' if you restrict it to a single extension as staging picks up production
      // value for this config).
      cors_url_regex_allow: '^(https://%s|chrome-extension://|http://localhost:300[0-9]|https://localhost:400[0-9])' % c.appDomain,

      auth0: {
        domain: 'doota-prod.us.auth0.com',
        api_redirect: 'https://%s/auth/callback' % c.apiDomain,
      },
      fullstory_org_id: 'o-1XZ1WK-na1',
    },

    public_interface+: {
      local c = $._config,

      name: 'default-ingress',
      managed_certs: {
        [std.strReplace(c.apiDomain, '.', '-')]: c.apiDomain,
        [std.strReplace(c.appDomain, '.', '-')]: c.appDomain,
        [std.strReplace(c.webhookMiscrosoftDomain, '.', '-')]: c.webhookMiscrosoftDomain,
      },
      extra_annotations: { 'kubernetes.io/ingress.global-static-ip-name': 'ingress-production-us-east1' },
      rules: [
        {
          host: c.apiDomain,
          paths: [
            { path: '/', service: 'portal-api-public', port: c.default_portal_api_port },
            { path: '/*', service: 'portal-api-public', port: c.default_portal_api_port },
          ],
        },
        {
          host: c.appDomain,
          paths: [
            { path: '/outlook/*', service: 'outlook-addin-server-public', port: 8000 },
            { path: '/', service: 'frontend-public', port: c.default_frontend_port },
            { path: '/*', service: 'frontend-public', port: c.default_frontend_port },
          ],
        },
        {
          host: c.webhookMiscrosoftDomain,
          paths: [
            { path: '/', service: 'msoft-service-public', port: c.default_webhook_port },
            { path: '/*', service: 'msoft-service-public', port: c.default_webhook_port },
          ],
        },

      ],
    },
  },
}
