local k = import 'sf/k8s.libsonnet';  // some constructor overrides over grafana's "kausal"
local ingress = k.networking.v1.ingress;


// In staging environment images are deployed via continuous deployment (github actions)
// This causes images in tanka to be "out of sync" with the cluster images. Kubernetes requires
// us to push deployments with images. To ensure that we do not need to always manually update the
// images in the tanka, we have a small tk proxy script (tk.sh) that fetches the current images via
// K8s and creates the images.json file. Use tk.sh like you use `tk`
//        ./tk.sh apply environments/staging
local k8sImages = (import 'images.json');


(import '../production/main.jsonnet') +
{
  local doota = super.doota,
  _images+:: {
    frontend: k8sImages.frontend,
    portalApi: k8sImages['portal-api'],
    migrator: k8sImages.migrator,
    frontend_env: 'staging',
  },

  svcAccount: doota.newServiceAccount([
    'portal-api-app-stag',
    'github-actions-to-gcr',
  ]),

  gaRbac: doota.newContinuousDeployment('github-actions-to-gcr', [
    'frontend',
    'portal-api',
    'migrator',
  ]),

  _config+:: {
    local top = self,

    apiDomain: 'api.staging.dootaai.com',
    appDomain: 'app.staging.dootaai.com',

    default_redis_addr: '10.182.57.35:6379',
    default_openai_debug_store: 'gs://doota-ai-debug-stag',
    default_tracing: '',

    default_database+:: {
      project: 'doota',
      region: 'us-east1',
      instance: 'doota',
      username: 'doota_stag_user',
      name: 'doota-stag',
      host: '127.0.0.1',
      secret: 'sql-database-pword',
    },

    migrator+:: {
      // since enable auth migration is enabled, the desired version is ignored
      desiredVersion: 0,
      enableAutoMigration: true,
    },

    frontend+:: {
      replicas: 1,
    },

    portalApi+:: {
      replicas: 1,
      jwt_kms: 'projects/doota/locations/global/keyRings/api-auth/cryptoKeys/jwt-signing-dev/cryptoKeyVersions/1',
      auth0+: {
        domain: 'doota.us.auth0.com',
      },
      fullstory_org_id: '',
      database+: {
        username: 'doota_stag_user',
        name: 'doota-stag',
      },
      service_account: 'portal-api-app-stag',
    },

    public_interface+:: {
      extra_annotations: { 'kubernetes.io/ingress.global-static-ip-name': 'ingress-staging-us-east1' },
    },

  },
}
