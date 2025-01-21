// replacement for k.libsonnet, opiniated for streamingfast (tweaks over grafana's already opiniated tweaks)
//
// Reference for ksonnet k8s https://jsonnet-libs.github.io/k8s-libsonnet/ (pick the right version based on which cluster
// version you want to deploy to).
(import 'ksonnet-util/kausal.libsonnet') +
{
  MiB: (1024 * 1024),
  GiB: (1024 * 1024 * 1024),

  apps+: {
    v1+: {
      deployment+: {
        new(name, replicas, containers, labels, initContainers=[])::
          super.new(name, replicas, containers, labels) +
          //super.mixin.spec.withMinReadySeconds(10) + //default from grafana.libsonnet
          //super.mixin.spec.withRevisionHistoryLimit(10) + //default from grafana.libsonnet
          super.mixin.metadata.withLabels(labels) +
          super.mixin.spec.withProgressDeadlineSeconds(600) +
          super.mixin.spec.strategy.withType('RollingUpdate') +
          super.mixin.spec.strategy.rollingUpdate.withMaxSurge(1) +
          super.mixin.spec.strategy.rollingUpdate.withMaxUnavailable(1) +
          super.spec.template.spec.withInitContainers(initContainers) +
          super.mixin.spec.template.spec.withTerminationGracePeriodSeconds(30),
      },

      statefulSet+: {
        new(name, replicas, containers, labels, serviceName, volumeClaims=[], initContainers=[])::
          super.new(name, replicas, containers, volumeClaims=volumeClaims, podLabels=labels) +
          //super.mixin.spec.updateStrategy.withType('RollingUpdate') + // default from grafana.libsonnet
          super.mixin.metadata.withLabels(labels) +
          super.spec.withPodManagementPolicy('Parallel') +
          super.spec.template.spec.withInitContainers(initContainers) +
          super.mixin.spec.withServiceName(serviceName),
      },
    },
  },

  core+: {
    v1+: {
      container+: {
        withCommand(args):
          super.withCommand(std.prune(args)),

        withHealthzReadiness(port, ssl=false):
          self.withHttpReadiness(port, path='/healthz', ssl=ssl),

        withHttpReadiness(port, path='/', ssl=false, periodSeconds=6, failureThreshold=3, timeoutSeconds=1, initialDelaySeconds=5):
          super.mixin.readinessProbe.httpGet.withPath(path) +
          super.mixin.readinessProbe.httpGet.withPort(port) +
          super.mixin.readinessProbe.httpGet.withScheme(if ssl then 'HTTPS' else 'HTTP') +
          super.mixin.readinessProbe.withInitialDelaySeconds(initialDelaySeconds) +
          super.mixin.readinessProbe.withTimeoutSeconds(timeoutSeconds) +
          super.mixin.readinessProbe.withPeriodSeconds(periodSeconds) +
          super.mixin.readinessProbe.withFailureThreshold(failureThreshold),

        withStartupProbe(port, path='/', ssl=false):
          super.mixin.startupProbe.httpGet.withPath(path) +
          super.mixin.startupProbe.withFailureThreshold(60) +
          super.mixin.startupProbe.httpGet.withPort(port) +
          super.mixin.startupProbe.httpGet.withScheme(if ssl then 'HTTPS' else 'HTTP') +
          super.mixin.startupProbe.withTimeoutSeconds(10) +
          super.mixin.startupProbe.withPeriodSeconds(1),

        withLivenessProbe(port, path='/', ssl=false):
          super.mixin.livenessProbe.httpGet.withPath(path) +
          super.mixin.livenessProbe.withFailureThreshold(3) +
          super.mixin.livenessProbe.httpGet.withPort(port) +
          super.mixin.livenessProbe.httpGet.withScheme(if ssl then 'HTTPS' else 'HTTP') +
          // The probe will fail if it does not respond in 10 seconds
          super.mixin.livenessProbe.withTimeoutSeconds(10) +
          // The probe will be checked every 10 seconds.
          super.mixin.livenessProbe.withPeriodSeconds(10),

        withGRPCReadiness(port):
          super.readinessProbe.exec.withCommandMixin(['/app/grpc_health_probe', '-addr=:' + port]) +
          super.readinessProbe.withInitialDelaySeconds(5) +
          super.readinessProbe.withTimeoutSeconds(1) +
          super.readinessProbe.withPeriodSeconds(6) +
          super.readinessProbe.withFailureThreshold(3) +
          super.lifecycle.postStart.exec.withCommand([
            '/bin/sh',
            '-c',
            'test -e /app/grpc_health_probe || (curl -L https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.9/grpc_health_probe-linux-amd64 -o /app/grpc_health_probe && chmod +x /app/grpc_health_probe)',  // may not be needed but you never know
          ]),

        withTCPReadiness(port):
          super.mixin.readinessProbe.tcpSocket.withPort(port),

      },
      backendConfig+: {
        new(service, healthCheck=null, labels={}, mixin={}):: {
          apiVersion: 'cloud.google.com/v1',
          kind: 'BackendConfig',
          metadata: {
            labels: labels,
            name: service,
          },
          spec: {
            [if healthCheck != null then 'healthCheck']: healthCheck,
          },
        } + mixin,

        healthCheckHttp(port, requestPath='/', checkIntervalSec=6, timeoutSec=3)::
          self.healthCheck('HTTP', port, requestPath, checkIntervalSec, timeoutSec),

        healthCheckHttps(port, requestPath='/', checkIntervalSec=6, timeoutSec=3)::
          self.healthCheck('HTTPS', port, requestPath, checkIntervalSec, timeoutSec),

        healthCheck(type, port, requestPath='/', checkIntervalSec=6, timeoutSec=3):: {
          type: type,
          port: port,
          requestPath: requestPath,
          checkIntervalSec: checkIntervalSec,
          timeoutSec: timeoutSec,
        },

        mixin: {
          spec: {
            withTimeoutSec(sec):: {
              spec+: {
                timeoutSec: sec,
              },
            },
            withConnectionDraining(sec):: {
              spec+: {
                connectionDraining: {
                  drainingTimeoutSec: sec,
                },
              },
            },
          },
        },
      },
    },
  },

  networking+: {
    v1+: {
      ingress+: {
        path(path, service, port):: {
          backend: {
            service: {
              name: service,
              port: {
                [if std.isNumber(port) then 'number']: port,
                [if std.isString(port) then 'name']: port,
              },
            },
          },
          [if path != '' then 'path']: path,
          pathType: 'ImplementationSpecific',
        },
      },

    },
  },

  gke: {
    ingress: {
      //  {
      //    '<host>': {
      //        '<path>': <serviceDefinition>,
      //    }
      //  }
      //
      new(name, config)::
        local ingress = $.networking.v1.ingress;
        local managedCertificateName(name) = std.strReplace(name, '.', '-');
        local managedCertificates = { [managedCertificateName(domain)]: domain for domain in std.objectFields(config) };
        local managedCertificateNames = std.objectFields(managedCertificates);
        local serviceDefPath(path, serviceDef) =
          ingress.path(path=path, service=serviceDef.metadata.name, port=serviceDef.spec.ports[0].port);

        // Gives [ { host: <host1>, [ <pathDef>, ...] }, { host: <host2>, [ <pathDef>, ... ]}
        local rules = [
          {
            host: host,
            http: {
              paths: [serviceDefPath(path, config[host][path]) for path in std.objectFields(config[host])],
            },
          }
          for host in std.objectFields(config)
        ];

        {
          mcrts:
            { [name]: $.gke.managedCertificate(name, managedCertificates[name]) for name in managedCertificateNames },

          root:
            ingress.new(name=name) +
            ingress.metadata.withAnnotations({
              'kubernetes.io/ingress.class': 'gce',
              [if std.length(managedCertificateNames) > 0 then 'networking.gke.io/managed-certificates']: std.join(', ', managedCertificateNames),
            }) +
            ingress.spec.withRules(rules),
        },
    },

    managedCertificate(name, domains):: {
      apiVersion: 'networking.gke.io/v1',
      kind: 'ManagedCertificate',
      metadata: {
        name: name,
      },
      spec: {
        domains: if std.isArray(v=domains) then domains else [domains],
      },
    },
  },

  util+:: {
    local util = self,
    monitoringRoles(namespace):: {
      roleBinding:
        $.rbac.v1.roleBinding.new('prometheus-k8s') +
        $.rbac.v1.roleBinding.bindRole(self.role) +
        $.rbac.v1.roleBinding.withSubjects([
          {
            kind: 'ServiceAccount',
            name: 'prometheus-k8s',
            namespace: namespace,
          },
        ]),

      role:
        $.rbac.v1.role.new('prometheus-k8s') +
        $.rbac.v1.role.withRules([
          $.rbac.v1.policyRule.withApiGroups(['']) +
          $.rbac.v1.policyRule.withResources(['services', 'endpoints', 'pods']) +
          $.rbac.v1.policyRule.withVerbs(['get', 'list', 'watch']),
        ]),

      serviceMonitor: {
        apiVersion: 'monitoring.coreos.com/v1',
        kind: 'ServiceMonitor',
        metadata: {
          name: 'prometheus',
        },
        spec: {
          endpoints: [
            {
              port: 'prometheus',
            },
          ],
          selector: {
            matchLabels: {
              monitor: 'prometheus',
            },
          },
        },
      },
    },

    withVolumeMountsFromDiskIfSet(name, path, disk=null):: (
      local container = $.core.v1.container;
      local volumeMount = $.core.v1.volumeMount;

      if disk != null && std.get(disk, 'size') != null then
        container.withVolumeMountsMixin([volumeMount.new(name, path)])
      else
        {}
    ),

    attachPVCTemplateFromDisk(name, disk):: (
      assert name != null && name != '' : 'Received a null "name" parameter, this is invalid';
      assert disk != null : 'Received a null "disk" parameter, this is invalid';

      local max_disk = std.get(disk, 'max_size', null);

      util.attachPVCTemplate(
        'datadir',
        disk.size,
        disk.storage_class,
        resize=max_disk != null && max_disk != '',
        resizeLimit=max_disk,
      )
    ),

    attachPVCTemplateFromDiskIfSet(name, disk=null):: (
      if disk != null && std.get(disk, 'size') != null then
        util.attachPVCTemplateFromDisk(name, disk)
      else
        {}
    ),

    attachPVCTemplate(name, size, storageClass, resize=true, resizeLimit='3Ti', resizeStep='10%')::
      local pvc = $.core.v1.persistentVolumeClaim;
      local sts = $.apps.v1.statefulSet;
      sts.spec.withVolumeClaimTemplatesMixin(
        pvc.new('datadir') +
        (if resize then pvc.mixin.metadata.withAnnotations({
           'resize.topolvm.io/increase': resizeStep,
           'resize.topolvm.io/storage_limit': resizeLimit,
           'resize.topolvm.io/threshold': resizeStep,
         }) else {}) +
        pvc.mixin.spec.resources.withRequests({ storage: size }) +
        pvc.mixin.spec.withAccessModes(['ReadWriteOnce']) +
        pvc.mixin.spec.withStorageClassName(storageClass) +
        pvc.mixin.spec.withVolumeMode('Filesystem')
      ),

    internalServiceFor(targetResource, publishNotReadyAddresses=false, headless=false, exposedPort=null, withPublicIP=false, backendConfig=null, ignoredPorts=[9102])::
      local service = $.core.v1.service;
      self.serviceFor(targetResource, ignored_ports=ignoredPorts) +
      (if exposedPort == null then service.mixin.metadata.withAnnotationsMixin({
         'cloud.google.com/neg': '{"ingress": false}',
       }) else service.mixin.metadata.withAnnotationsMixin({
         'cloud.google.com/neg': '{"ingress": false, "exposed_ports":{"%d":{}}}' % exposedPort,
       })) +
      (
        if backendConfig == null
        then {}
        else service.mixin.metadata.withAnnotationsMixin({ 'beta.cloud.google.com/backend-config': '{"default": "%s"}' % backendConfig })
      ) +
      (if publishNotReadyAddresses then service.spec.withPublishNotReadyAddresses(true) else {}) +
      (if withPublicIP then (service.spec.withAllocateLoadBalancerNodePorts(true) + service.spec.withType('LoadBalancer')) else {}) +
      (if headless then service.spec.withClusterIP('None') else {})
    ,

    publicServiceFor(targetResource, name=null, grpc_portnames=[], backendConfig=null, ignoredPorts=[9102])::
      local service = $.core.v1.service;
      self.serviceFor(targetResource, ignored_ports=ignoredPorts) +
      (
        if name == null then {} else
          service.mixin.metadata.withName(name) +
          service.mixin.metadata.withLabels({ name: name })
      ) +
      service.mixin.metadata.withAnnotationsMixin({
        'cloud.google.com/neg': '{"ingress": true}',
        [if grpc_portnames != [] then 'cloud.google.com/app-protocols']:
          '{%s}' % std.join(',', ['"%s":"HTTP2"' % port for port in grpc_portnames]),
      }) +
      (
        if backendConfig == null
        then {}
        else service.mixin.metadata.withAnnotationsMixin({ 'beta.cloud.google.com/backend-config': '{"default": "%s"}' % backendConfig })
      ) +
      service.mixin.spec.withType('NodePort'),

    unsetReplicas(obj):: obj { spec: util.objectPop(obj.spec, 'replicas') },
    objectPop(obj, key): {
      [k]: obj[k]
      for k in std.objectFieldsAll(obj)
      if k != key
    },

    hpaFor(parent, max_replicas, min_replicas=1, avg_cpu=50, avg_mem=0, scaleUpSeconds=70, scaleDownSeconds=60)::
      local as = $.autoscaling.v2;
      local hpa = as.horizontalPodAutoscaler;
      local hpa_spec = as.horizontalPodAutoscalerSpec;

      hpa.new(parent.metadata.name) +
      {
        spec: hpa_spec.behavior.scaleUp.withPolicies(
                {
                  periodSeconds: scaleUpSeconds,
                  type: 'Pods',
                  value: 4,
                },
              ) +
              hpa_spec.behavior.scaleUp.withStabilizationWindowSeconds(0) +
              hpa_spec.behavior.scaleDown.withPolicies(
                {
                  periodSeconds: scaleDownSeconds,
                  type: 'Pods',
                  value: 1,
                },
              ) +
              hpa_spec.behavior.scaleDown.withStabilizationWindowSeconds(30) +
              hpa_spec.withMinReplicas(min_replicas) +
              hpa_spec.withMaxReplicas(max_replicas) +
              hpa_spec.scaleTargetRef.withKind(parent.kind) +
              hpa_spec.scaleTargetRef.withApiVersion(parent.apiVersion) +
              hpa_spec.scaleTargetRef.withName(parent.metadata.name) +
              hpa_spec.withMetrics(
                std.prune([
                  (if avg_mem != 0 then
                     as.metricSpec.resource.target.withType('Utilization') +
                     as.metricSpec.resource.target.withAverageUtilization(avg_mem) +
                     as.metricSpec.resource.withName('memory') +
                     as.metricSpec.withType('Resource')
                   else null),
                  (if avg_cpu != 0 then
                     as.metricSpec.resource.target.withType('Utilization') +
                     as.metricSpec.resource.target.withAverageUtilization(avg_cpu) +
                     as.metricSpec.resource.withName('cpu') +
                     as.metricSpec.withType('Resource')
                   else null),
                ])
              ),
      },

    serviceFor(deployment, ignored_labels=[], nameFormat='%s', ignored_ports=[])::
      local container = $.core.v1.container;
      local service = $.core.v1.service;
      local servicePort = $.core.v1.servicePort;
      local ports = [
        servicePort.newNamed(
          name=('%(container)s-%(port)s' % { container: c.name, port: port.name }),
          port=port.containerPort,
          targetPort=port.containerPort
        ) +
        if std.objectHas(port, 'protocol')
        then servicePort.withProtocol(port.protocol)
        else {}
        for c in deployment.spec.template.spec.containers
        for port in (c + container.withPortsMixin([])).ports
        if std.count(ignored_ports, port.containerPort) == 0
      ];
      local labels = {
        [x]: deployment.spec.template.metadata.labels[x]
        for x in std.objectFields(deployment.spec.template.metadata.labels)
        if std.count(ignored_labels, x) == 0
      };

      service.new(
        deployment.metadata.name,  // name
        labels,  // selector
        ports,
      ) +
      service.mixin.metadata.withLabels({ name: deployment.metadata.name }),


    monitorServiceFor(deployment, ignored_labels=[])::
      local service = $.core.v1.service;
      local servicePort = $.core.v1.servicePort;
      local ports = [servicePort.newNamed(
        name='prometheus',
        port=9102,
        targetPort=9102,
      )];
      local labels = {
        [x]: deployment.spec.template.metadata.labels[x]
        for x in std.objectFields(deployment.spec.template.metadata.labels)
        if std.count(ignored_labels, x) == 0
      };

      service.new(
        '%s-monitor' % deployment.metadata.name,  // name
        labels,  // selector
        ports,
      ) +
      service.mixin.metadata.withLabels({
        name: deployment.metadata.name,
        monitor: 'prometheus',
      }) +
      service.spec.withPublishNotReadyAddresses(true),

    renameSTS(newName):: {
      metadata+: {
        name: newName,
        labels+: {
          name: newName,
        },
      },
      spec+: {
        template+: {
          metadata+: {
            labels+: {
              name: newName,
            },
          },
        },
        selector+: {
          matchLabels+: {
            name: newName,
          },
        },
      },
    },

    // Sets `k.util.templateToleration` and `k.util.templateAffinity` if `nodePool != null && nodePool != ''`.

    runOnNodePoolAndZoneOnlyIfSet(zones, nodePool):: (
      if zones != null && zones != '' && nodePool != null && nodePool != '' then (
        util.templateToleration(util.tolerationNodeInPool(nodePool)) +
        util.templateAffinity(util.affinityBoth(zones, nodePool))
      ) else self.runOnZoneOnlyIfSet(zones) + self.runOnNodePoolOnlyIfSet(nodePool)
    ),

    runOnZoneOnlyIfSet(zones):: (
      if zones != null && zones != '' then (
        util.templateAffinity(util.affinityNodeInZones(zones))
      ) else {}
    ),

    runOnNodePoolOnlyIfSet(nodePool):: (
      if nodePool != null && nodePool != '' then (
        util.templateToleration(util.tolerationNodeInPool(nodePool)) +
        util.templateAffinity(util.affinityNodeInPools(nodePool))
      ) else
        {}
    ),

    templateToleration(tol):: {
      spec+: {
        template+: {
          spec+: {
            tolerations: if std.isArray(v=tol) then tol else [tol],
          },
        },
      },
    },

    templateTolerationMixin(tol):: {
      spec+: {
        template+: {
          spec+: {
            tolerations+: if std.isArray(v=tol) then tol else [tol],
          },
        },
      },
    },

    templateAffinity(aff):: {
      spec+: {
        template+: {
          spec+: {
            affinity: aff,
          },
        },
      },
    },
    templateAffinityMixin(aff):: {
      spec+: {
        template+: {
          spec+: {
            affinity+: aff,
          },
        },
      },
    },

    affinityBoth(zones, nodePools, zone_key='topology.kubernetes.io/zone', pool_key='pool-id'):: {
      nodeAffinity+: {
        requiredDuringSchedulingIgnoredDuringExecution+: {
          nodeSelectorTerms+: [
            {
              matchExpressions: [
                {
                  key: zone_key,
                  operator: 'In',
                  values: if std.isArray(v=zones) then zones else [zones],
                },
                {
                  key: pool_key,
                  operator: 'In',
                  values: if std.isArray(v=nodePools) then nodePools else [nodePools],
                },
              ],
            },
          ],
        },
      },
    },

    affinityNodeInZones(zones, key='topology.kubernetes.io/zone'):: {
      nodeAffinity+: {
        requiredDuringSchedulingIgnoredDuringExecution+: {
          nodeSelectorTerms+: [
            {
              matchExpressions: [
                {
                  key: key,
                  operator: 'In',
                  values: if std.isArray(v=zones) then zones else [zones],
                },
              ],
            },
          ],
        },
      },
    },

    affinityNodeInPools(nodePools, key='pool-id'):: {
      nodeAffinity+: {
        requiredDuringSchedulingIgnoredDuringExecution+: {
          nodeSelectorTerms+: [
            {
              matchExpressions: [
                {
                  key: key,
                  operator: 'In',
                  values: if std.isArray(v=nodePools) then nodePools else [nodePools],
                },
              ],
            },
          ],
        },
      } + if std.isArray(v=nodePools) && std.length(nodePools) > 1 then (
        {
          preferredDuringSchedulingIgnoredDuringExecution: [
            {
              weight: 1,
              preference: {
                matchExpressions: [
                  {
                    key: 'pool-id',
                    operator: 'In',
                    values: [
                      if std.isArray(v=nodePools) then nodePools[0] else nodePools,
                    ],
                  },
                ],
              },
            },
          ],
        }
      ) else {},
    },

    tolerationNodeInPool(nodePool, key='compute-class'):: {
      effect: 'NoSchedule',
      key: key,
      operator: 'Equal',
      value: nodePool,
    },

    //  affinityAntiMatchApp(app, weight=10):: {
    //    podAntiAffinity: {
    //      preferredDuringSchedulingIgnoredDuringExecution: [
    //        {
    //          podAffinityTerm: {
    //            labelSelector: {
    //              matchExpressions: [
    //                {
    //                  key: 'app',
    //                  operator: 'In',
    //                  values: [
    //                    app,
    //                  ],
    //                },
    //              ],
    //            },
    //            topologyKey: 'kubernetes.io/hostname',
    //          },
    //          weight: weight,
    //        },
    //      ],
    //    },
    //  },
    //

    //  affinityAntiSelf(podTemplate, weight=10):: {
    //    podAntiAffinity: {
    //      preferredDuringSchedulingIgnoredDuringExecution: [
    //        {
    //          podAffinityTerm: {
    //            labelSelector: {
    //              matchExpressions: [
    //                {
    //                  key: 'app',
    //                  operator: 'In',
    //                  values: [
    //                    podTemplate.metadata.labels.app,
    //                  ],
    //                },
    //              ],
    //            },
    //            topologyKey: 'kubernetes.io/hostname',
    //          },
    //          weight: weight,
    //        },
    //      ],
    //    },
    //  },


    stsServiceAccount(acct)::
      local sts = $.apps.v1.statefulSet;
      (if acct == '' then {} else sts.spec.template.spec.withServiceAccountName(acct)),

    deployServiceAccount(acct)::
      local deploy = $.apps.v1.deployment;
      (if acct == '' then {} else deploy.spec.template.spec.withServiceAccountName(acct)),

    replaceSTSImage(newImg, containerIndex=0, alwaysPull=false)::
      local container = $.core.v1.container;
      util.mixinSTSContainer(container.withImage(newImg) + (if alwaysPull then container.withImagePullPolicy('Always') else {}), containerIndex),

    mixinSTSContainer(mixin, containerIndex=0):: {
      local container = $.core.v1.container,
      spec+: {
        template+: {
          spec+: {
            containers:
              if containerIndex > std.length(super.containers) - 1 then super.containers + [mixin] else (
                super.containers[0:containerIndex] +
                [
                  super.containers[containerIndex] + mixin,
                ] + super.containers[containerIndex + 1:]
              ),
          },
        },
      },
    },

    // Extracts the pod's name from a either a app definition (`apps.v1.StatefulSet`) or from a
    // "Firehose" component (e.g. one that is defined such that it have a child key named `StatefulSet`
    // and it's value is an object of type `apps.v1.StatefulSet`).
    podNameFromSts(input, index=0):: (
      local statefulSet = std.get(input, 'statefulSet', input);
      local apiVersion = std.get(statefulSet, 'apiVersion');
      local kind = std.get(statefulSet, 'kind');

      assert apiVersion == 'apps/v1' && kind == 'StatefulSet' : 'Received a non-StatefulSet component %s/%s' % [apiVersion, kind];

      '%s-%d' % [statefulSet.metadata.name, index]
    ),

    // Extracts the service's name from a Service definition (`apps.v1.Service`)
    serviceName(service):: (
      local apiVersion = std.get(service, 'apiVersion');
      local kind = std.get(service, 'kind');

      assert apiVersion == 'v1' && kind == 'Service' : 'Received a non-Service component %s/%s' % [apiVersion, kind];

      service.metadata.name
    ),

    // Extracts the service's port value for given named port from a Service definition (`apps.v1.Service`)
    servicePort(service, named_port):: (
      local apiVersion = std.get(service, 'apiVersion');
      local kind = std.get(service, 'kind');

      assert apiVersion == 'v1' && kind == 'Service' : 'Received a non-Service component %s/%s' % [apiVersion, kind];

      local portNames = [x.name for x in service.spec.ports];
      local portByName = { [x.name]: x for x in service.spec.ports };

      assert std.objectHas(portByName, named_port) : 'No port named %s found in service %s (valid names are %s)' % [named_port, service.metadata.name, std.join(', ', portNames)];

      portByName[named_port].port
    ),

    // Extracts the service's DNS resolvable hostname from a Service definition (`apps.v1.Service`) and a specific named port
    serviceDnsHostname(service, named_port):: (
      $.util.serviceGrpcAddr('dns', service, named_port)
    ),

    serviceGrpcAddr(scheme, service, named_port):: (
      local apiVersion = std.get(service, 'apiVersion');
      local kind = std.get(service, 'kind');

      assert apiVersion == 'v1' && kind == 'Service' : 'Received a non-Service component %s/%s' % [apiVersion, kind];

      '%s:///%s:%d' % [scheme, service.metadata.name, $.util.servicePort(service, named_port)]
    ),

    // Extracts the service's <name>:<port> address from a "Firehose" component (e.g. one that is defined
    // such that it have a child key named `internalService` and it's value is an object of type
    // `core.v1.Service`).
    //
    // To be used like `k.util.internalServiceAddr($.substreams_tier1, 'grpc')` which would yield
    // something like `substreams-tier1:9000`.
    internalServiceAddr(input, named_port):: (
      local name = $.util.serviceName(std.get(input, 'internalService', input));
      local port = $.util.servicePort(std.get(input, 'internalService', input), named_port);

      '%s:%s' % [name, port]
    ),

    // Extracts the service's name from a either a Service definition (`apps.v1.Service`) or from a
    // "Firehose" component (e.g. one that is defined such that it have a child key named `internalService`
    // and it's value is an object of type `core.v1.Service`).
    //
    // To be used like `k.util.internalServiceAddr($.substreams_tier1)` which would yield
    // something like `substreams-tier1` (e.g. the service's name).
    internalServiceName(input):: (
      $.util.serviceName(std.get(input, 'internalService', input))
    ),

    // Extracts the service's hostname from a either a Service definition (`apps.v1.Service`) or from a
    // "Firehose" component (e.g. one that is defined such that it have a child key named `internalService`
    // and it's value is an object of type `core.v1.Service` and generates a gRPC address using the `dns:///`
    // resolver.
    //
    // To be used like `k.util.internalServiceDnsHostname($.substreams_tier1, 'grpc')` which would yield
    // something like `dns:///substreams-tier1:9000`.
    internalServiceDnsHostname(input, named_port):: (
      $.util.serviceGrpcAddr('dns', std.get(input, 'internalService', input), named_port)
    ),

    // Extracts the service's hostname from a either a Service definition (`apps.v1.Service`) or from a
    // "Firehose" component (e.g. one that is defined such that it have a child key named `internalService`
    // and it's value is an object of type `core.v1.Service`) and generates a gRPC address using the `kubernetes:///`
    // resolver.
    //
    // To be used like `k.util.internalServiceKubernetesGrpcAddr($.substreams_tier1, 'grpc')` which would yield
    // something like `kubernetes:///substreams-tier1:9000`.
    internalServiceKubernetesGrpcAddr(input, named_port):: (
      $.util.serviceGrpcAddr('kubernetes', std.get(input, 'internalService', input), named_port)
    ),

    // Extracts the service's name from a either from a "Firehose" component (e.g. one that is defined
    // such that it have a child key named `publicService` and it's value is an object of type
    // `core.v1.Service`).
    //
    // To be used like `k.util.publicServiceName($.substreams_tier1)` which would yield
    // something like `substreams-tier1`.
    publicServiceName(input):: (
      $.util.serviceName(std.get(input, 'publicService', input))
    ),

    // Extracts the service's <name>:<port> address from a "Firehose" component (e.g. one that is defined
    // such that it have a child key named `publicService` and it's value is an object of type
    // `core.v1.Service`).
    //
    // To be used like `k.util.publicServiceAddr($.substreams_tier1, 'grpc')` which would yield
    // something like `substreams-tier1:9000`.
    publicServiceAddr(input, named_port):: (
      local name = $.util.serviceName(std.get(input, 'publicService', input));
      local port = $.util.servicePort(std.get(input, 'publicService', input), named_port);

      '%s:%s' % [name, port]
    ),

    // Extracts the service's hostname from "Firehose" component (e.g. one that is defined
    // such that it have a child key named `publicService` and it's value is an object of type
    // `core.v1.Service`)  and generates a gRPC address using the `dns:///`
    // resolver.
    //
    // To be used like `k.util.publicServiceDnsHostname($.substreams_tier1, 'grpc')` which would yield
    // something like `dns:///substreams-tier1:9000`.
    publicServiceDnsHostname(input, named_port):: (
      $.util.serviceGrpcAddr('dns', std.get(input, 'publicService', input), named_port)
    ),

    // Extracts the service's hostname from "Firehose" component (e.g. one that is defined
    // such that it have a child key named `publicService` and it's value is an object of type
    // `core.v1.Service`)  and generates a gRPC address using the `kubernetes:///`
    // resolver.
    //
    // To be used like `k.util.publicServiceKubernetesGrpcAddr($.substreams_tier1, 'grpc')` which would yield
    // something like `kubernetes:///substreams-tier1:9000`.
    publicServiceKubernetesGrpcAddr(input, named_port):: (
      $.util.serviceGrpcAddr('kubernetes', std.get(input, 'publicService', input), named_port)
    ),

    // can replace by 'null' to prune
    replaceCommand(cmds, prefix, new)::
      std.prune([if std.startsWith(x, prefix) then new else x for x in cmds]),

    setResources(obj)::
      $.util.resourcesRequests(obj.requests[0], obj.requests[1]) +
      $.util.resourcesLimits(obj.limits[0], obj.limits[1]),

    gcpServiceAccount(k8sAccount, gcpAccount)::
      $.core.v1.serviceAccount.new(k8sAccount) +
      $.core.v1.serviceAccount.metadata.withAnnotations({
        'iam.gke.io/gcp-service-account': gcpAccount,
      }),


    newGKEPublicInterface(name, managed_certs, extra_annotations, rules):: {
      ingress:
        $.networking.v1.ingress.new(name=name) +
        $.networking.v1.ingress.metadata.withAnnotations({
          'kubernetes.io/ingress.class': 'gce',
          'networking.gke.io/managed-certificates': std.join(', ', std.objectFields(managed_certs)),
        } + extra_annotations) +
        $.networking.v1.ingress.spec.withRules(std.map(function(rule) {
          host: rule.host,
          http: {
            paths: [$.networking.v1.ingress.path(path=path.path, service=path.service, port=path.port) for path in rule.paths],
          },
        }, rules)),

      _certs_array::
        std.map(function(key) {
          key: key,
          value: $.gke.managedCertificate(key, managed_certs[key]),
        }, std.objectFields(managed_certs)),

      managed_certs: std.foldl(function(out, cert) out { [cert.key]: cert.value }, self._certs_array, {}),
    },

    limitRange(type, requests, limits, name='default-limit-range')::
      $.core.v1.limitRange.new(name) +
      $.core.v1.limitRange.spec.withLimits([
        $.core.v1.limitRangeItem.withDefault({ cpu: limits[0], memory: limits[1] }) +
        $.core.v1.limitRangeItem.withDefaultRequest({ cpu: requests[0], memory: requests[1] }) +
        $.core.v1.limitRangeItem.withType(type),
      ]),

    withRollingUpdate(maxUnavailable, maxSurge)::
      $.apps.v1.deployment.spec.strategy.withType('RollingUpdate') +
      $.apps.v1.deployment.spec.strategy.rollingUpdate.withMaxUnavailable(maxUnavailable) +
      $.apps.v1.deployment.spec.strategy.rollingUpdate.withMaxSurge(maxSurge),

    // Deprecated: Use `k.gke` directly instead
    gke: $.gke,
  },
}
