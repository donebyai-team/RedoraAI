local k = import 'sf/k8s.libsonnet';

// Reference documentation to find the right thing to call https://jsonnet-libs.github.io/k8s-libsonnet/1.20
local deployment = k.apps.v1.deployment;
local sts = k.apps.v1.statefulSet;
local container = k.core.v1.container;
local configMap = k.core.v1.configMap;
local port = k.core.v1.containerPort;
local service = k.core.v1.service;
local job = k.batch.v1.job;
local servicePort = k.core.v1.servicePort;
local ingress = k.networking.v1.ingress;
local configMap = k.core.v1.configMap;
local sts = k.apps.v1.statefulSet;
local envVar = k.core.v1.envVar;
local backendConfig = k.core.v1.backendConfig;
local volumeMount = k.core.v1.volumeMount;
local rbac = k.rbac.v1;
local rbacRole = rbac.role;
local rbacRoleBinding = rbac.roleBinding;
local nameAppend(name, tag) = (if tag == '' then name else '%s-%s' % [name, tag]);
local iamServiceAccount(serviceAccount) = (serviceAccount + '@doota.iam.gserviceaccount.com');


local initCloudSqlProxy(database) = (
  container.new('cloud-sql-proxy', 'gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.8.2') +
  container.withImagePullPolicy('Always') +
  container.withEnvMap({
    CSQL_PROXY_HEALTH_CHECK: 'true',
    CSQL_PROXY_HTTP_PORT: '9801',
    CSQL_PROXY_HTTP_ADDRESS: '0.0.0.0',
  }) +
  // https://github.com/GoogleCloudPlatform/cloud-sql-proxy/blob/main/examples/k8s-health-check/README.md
  container.withRestartPolicy('Always') +
  container.withStartupProbe(9801, path='/startup', ssl=false) +
  container.withLivenessProbe(9801, path='/liveness', ssl=false) +
  // container.withHttpReadiness(9801, path='/readiness', ssl=false, failureThreshold=6, periodSeconds=30, timeoutSeconds=10, initialDelaySeconds=30) +
  container.withArgs([
    '--private-ip',
    '--structured-logs',
    '--credentials-file=/secrets/service_account.json',
    database.project + ':' + database.region + ':' + database.instance,
  ]) +
  k.util.setResources({ requests: ['200m', '1Gi'], limits: ['1', '2Gi'] })
);

(import 'config.libsonnet') + {
  doota: {
    local c = self,

    local newServiceAccount(serviceAccount) =
      k.util.gcpServiceAccount(serviceAccount, iamServiceAccount(serviceAccount)),

    newServiceAccount(serviceAccounts):: {
      [x]: newServiceAccount(x)
      for x in serviceAccounts
    },

    dbDSN(database):: 'postgresql://%s:${PGPASS}@%s/%s?enable_incremental_sort=off&sslmode=disable&encryptionKey=${PGENCRYPTIONKEY}' % [database.username, database.host, database.name],


    newFrontend(image, config):: {
      local this = self,

      deployment:
        deployment.new(
          name=config.name,
          replicas=config.replicas,
          labels={ app: config.name },
          containers=[
            container.new('frontend', image) +
            container.withPorts([
              port.new('http', config.default_http_port),
            ]) +
            container.withCommand([
              'npm',
              'run',
              'start',
            ]) +
            k.util.setResources(config.resources) +
            container.withTCPReadiness(config.default_http_port),
          ]
        ),

      publicService:
        k.util.publicServiceFor(self.deployment, name=config.name + '-public'),
    },

    newPortalApi(image, config):: {
      local this = self,
      local publicServiceName = config.name + '-public',

      deployment:
        deployment.new(
          name=config.name,
          replicas=config.replicas,
          labels={ app: config.name },
          initContainers=[
            initCloudSqlProxy(config.database),
          ],
          containers=[
            container.new('portal-api', image) +
            container.withPorts([
              port.new('grpc', config.http_listen_port),
            ]) +
            container.withCommand([
              '/app/doota',
              'start',
              'portal-api',
              '--log-format=json',
              '--pg-dsn=' + c.dbDSN(config.database),
              '--jwt-kms-keypath=' + config.jwt_kms,
              '--portal-cors-url-regex-allow=' + config.cors_url_regex_allow,
              '--common-pubsub-project=' + config.gcp_project,
              '--portal-http-listen-addr=:%s' % config.http_listen_port,
            ]) +
            container.withEnvMap({
              DLOG: 'doota.*=info',
            }) +
            container.withEnvMixin([
              envVar.fromSecretRef('PGPASS', config.database.secret, 'password'),
              envVar.fromSecretRef('PGENCRYPTIONKEY', config.database.encryption_key, 'encryption_key'),
            ]) +
            k.util.setResources(config.resources) +
            container.withHealthzReadiness(config.http_listen_port, ssl=false),
          ]
        ) +
        k.util.stsServiceAccount(config.service_account) +
        deployment.secretVolumeMount('sql-cloud-proxy-sa', '/secrets/', 420, {}, {}, ['cloud-sql-proxy'], includeInitContainers=true),

      backendConfig:
        backendConfig.new(
          service=publicServiceName,
          healthCheck=backendConfig.healthCheckHttp(port=config.http_listen_port, requestPath='/healthz'),
          mixin=backendConfig.mixin.spec.withTimeoutSec(86400),
        ),

      internalService:
        k.util.internalServiceFor(self.deployment),

      publicService:
        k.util.publicServiceFor(
          self.deployment,
          name=publicServiceName,
          grpc_portnames=['grpc'],
          backendConfig=publicServiceName,
        ),
    },

    newMigrator(image, config):: {
      sts:
        sts.new(
          name=config.name,
          replicas=1,
          serviceName=config.name,
          labels={ app: config.name },
          initContainers=[
            initCloudSqlProxy(config.database),
          ],
          containers=[
            container.new('migrator', image) +
            container.withCommand([
              '/app/doota',
              'migrator',
              '--pg-dsn=' + c.dbDSN(config.database),
              '--desired-version=' + config.desiredVersion,
              (if config.enableAutoMigration then '--enable-auto-migration' else null),
            ]) + container.withEnvMap({
              DLOG: 'doota.*=info',
            }) +
            container.withEnvMixin([
              envVar.fromSecretRef('PGPASS', 'sql-database-pword', 'password'),
            ]),
          ],
        ) +
        sts.secretVolumeMount('sql-cloud-proxy-sa', '/secrets/', 420, {}, {}, ['cloud-sql-proxy'], includeInitContainers=true),
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

    newSqlBastion(image, config):: {
      local newProxyContainer(database) =
        container.new(database.instance, image) +
        container.withPorts([
          port.new('sql', database.to_port),
        ]) +
        container.withCommand([
          '/cloud_sql_proxy',
          '-ip_address_types=PRIVATE',
          '-instances=%(project)s:%(region)s:%(instance)s=tcp:%(to_port)s' % {
            project: config.project,
            region: config.region,
            instance: database.instance,
            to_port: database.to_port,
          },
        ]) +
        container.securityContext.withRunAsNonRoot(true) +
        k.util.setResources(config.resource),

      local newViewContainer(database) =
        container.new(nameAppend(database.instance, 'view'), 'sosedoff/pgweb:0.11.12') +
        container.withPorts([
          port.new('web', database.web_port),
        ]) +
        container.withCommand([
          'pgweb',
          '--bind=0.0.0.0',
          '--listen=%s' % database.web_port,
          '--binary-codec=hex',
        ]) +

        container.withEnvMixin([
          envVar.fromSecretRef('DATABASE_URL', database.dsnSecret.name, database.dsnSecret.key),
        ]) +
        k.util.setResources(config.resource),

      statefulSet:
        if std.length(config.databases) == 0 then null else
          sts.new(
            name=config.name,
            serviceName=config.name,
            replicas=1,
            labels=std.get(config, 'labels', {}),
            containers=
            [newProxyContainer(x) for x in config.databases] +
            [newViewContainer(x) for x in config.databases],
          ) +
          k.util.stsServiceAccount(config.service_account),
    },

    newContinuousDeployment(serviceAccount, apps):: {
      githubActionRbac: {
        role: rbacRole.new('image-setter') +
              rbacRole.withRules([
                rbac.policyRule.withApiGroups(['apps']) +  // "" indicates the core API group,
                rbac.policyRule.withResources(['deployments', 'statefulsets']) +
                rbac.policyRule.withResourceNames(apps) +
                rbac.policyRule.withVerbs(['get', 'patch']),
              ]),
        roleBinding: rbacRoleBinding.new('image-setter') +
                     rbacRoleBinding.bindRole(self.role) +
                     rbacRoleBinding.withSubjects([{
                       kind: 'User',
                       name: iamServiceAccount(serviceAccount),
                     }]),
      },
    },
  },
}
