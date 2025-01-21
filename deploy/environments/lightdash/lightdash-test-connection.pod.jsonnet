{
  "apiVersion": "v1",
  "kind": "Pod",
  "metadata": {
    "name": "lightdash-test-connection",
    "namespace": "lightdash",
    "labels": {
      "helm.sh/chart": "lightdash-0.8.9",
      "app.kubernetes.io/name": "lightdash",
      "app.kubernetes.io/instance": "lightdash",
      "app.kubernetes.io/version": "0.778.1",
      "app.kubernetes.io/managed-by": "Helm"
    },
    "annotations": {
      "helm.sh/hook": "test"
    }
  },
  "spec": {
    "containers": [
      {
        "name": "wget",
        "image": "busybox",
        "command": [
          "wget"
        ],
        "args": [
          "lightdash:8080"
        ]
      }
    ],
    "restartPolicy": "Never"
  }
}