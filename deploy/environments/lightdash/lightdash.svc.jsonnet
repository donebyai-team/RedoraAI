{
  "apiVersion": "v1",
  "kind": "Service",
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
    "type": "NodePort",
    "ports": [
      {
        "port": 8080,
        "targetPort": "http",
        "protocol": "TCP",
        "name": "http"
      }
    ],
    "selector": {
      "app.kubernetes.io/name": "lightdash",
      "app.kubernetes.io/instance": "lightdash"
    }
  }
}