{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "lightdash",
    "namespace": "lightdash",
    "labels": {
      "helm.sh/chart": "lightdash-0.8.9",
      "app.kubernetes.io/name": "lightdash",
      "app.kubernetes.io/instance": "lightdash",
      "app.kubernetes.io/version": "0.778.1",
      "app.kubernetes.io/managed-by": "Helm"
    }
  },
  "spec": {
    "replicas": 1,
    "selector": {
      "matchLabels": {
        "app.kubernetes.io/name": "lightdash",
        "app.kubernetes.io/instance": "lightdash"
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "app.kubernetes.io/name": "lightdash",
          "app.kubernetes.io/instance": "lightdash"
        }
      },
      "spec": {
        "securityContext": {},
        "serviceAccountName": "lightdash",
        "containers": [
          {
            "name": "lightdash",
            "securityContext": {},
            "image": "lightdash/lightdash:0.778.1",
            "imagePullPolicy": "IfNotPresent",
            "command": null,
            "args": null,
            "env": [
              {
                "name": "PGPASSWORD",
                "valueFrom": {
                  "secretKeyRef": {
                    "name": "lightdash-externaldb",
                    "key": "postgresql-password"
                  }
                }
              },
              {
                "name": "EMAIL_SMTP_PASSWORD",
                "valueFrom": {
                  "secretKeyRef": {
                    "name": "lightdash-smtp",
                    "key": "password"
                  }
                }
              }
            ],
            "envFrom": [
              {
                "configMapRef": {
                  "name": "lightdash"
                }
              },
              {
                "secretRef": {
                  "name": "lightdash"
                }
              }
            ],
            "ports": [
              {
                "name": "http",
                "containerPort": 8080,
                "protocol": "TCP"
              }
            ],
            "livenessProbe": {
              "initialDelaySeconds": 30,
              "timeoutSeconds": 60,
              "periodSeconds": 30,
              "httpGet": {
                "path": "/api/v1/livez",
                "port": 8080
              }
            },
            "readinessProbe": {
              "initialDelaySeconds": 30,
              "periodSeconds": 60,
              "timeoutSeconds": 30,
              "httpGet": {
                "path": "/api/v1/health",
                "port": 8080
              }
            },
            "resources": {}
          }
        ]
      }
    }
  }
}