local k = import 'sf/k8s.libsonnet';  // some constructor overrides over grafana's "kausal"
local ingress = k.networking.v1.ingress;

(import 'doota/doota.libsonnet') +
{
  local doota = super.doota,

  _config+:: {
    local top = self,

    lightdashApiFQDN: 'lightdash.doota.ao',

    public_interface+: {
      name: 'default-ingress',
      managed_certs: {
        [std.strReplace(top.lightdashApiFQDN, '.', '-')]: top.lightdashApiFQDN,
      },
      extra_annotations: { 'kubernetes.io/ingress.global-static-ip-name': 'ingress-lightdash-us-east1' },
      rules: [
        { host: top.lightdashApiFQDN, paths: [
          { path: '/*', service: 'lightdash', port: 8080 },
        ] },
      ],
    },
  },

  newGKEPublicInterface(config):: {
    ingress:
      ingress.new(name=config.name) +
      ingress.metadata.withAnnotations({
        'kubernetes.io/ingress.class': 'gce',
        'networking.gke.io/managed-certificates': std.join(', ', std.objectFields(config.managed_certs)),
      } + config.extra_annotations) +
      ingress.spec.withRules(std.map(function(rule) {
        host: rule.host,
        http: {
          paths: [ingress.path(path=path.path, service=path.service, port=path.port) for path in rule.paths],
        },
      }, config.rules)),

    certs_array::
      std.map(function(key) {
        key: key,
        value: k.gke.managedCertificate(key, config.managed_certs[key]),
      }, std.objectFields(config.managed_certs)),

    managed_certs: std.foldl(function(out, cert) out { [cert.key]: cert.value }, self.certs_array, {}),
  },

  public_interface: self.newGKEPublicInterface($._config.public_interface),
  lightdashBrowserlessChromeServiceAccount: (import 'lightdash-browserless-chrome.serviceaccounts.jsonnet'),
  lightdashBrowserlessChromeService: (import 'lightdash-browserless-chrome.svc.jsonnet'),
  lightdashBrowserlessChromeDeployment: (import 'lightdash-browserless-chrome.deployment.jsonnet'),
  lightdashBrowserlessChromeTest: (import 'lightdash-browserless-chrome-test-connection.pod.jsonnet'),
  ligthdashServiceAccount: (import 'lightdash.serviceaccounts.jsonnet'),
  ligthdashService: (import 'lightdash.svc.jsonnet'),
  ligthdashConfigMap: (import 'lightdash.configmap.jsonnet'),
  ligthdashDeployment: (import 'lightdash.deployment.jsonnet'),
  ligthdashTest: (import 'lightdash-test-connection.pod.jsonnet'),

}
