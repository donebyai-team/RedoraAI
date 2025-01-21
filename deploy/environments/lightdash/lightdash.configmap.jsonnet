{
  "apiVersion": "v1",
  "kind": "ConfigMap",
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
  "data": {
    "PGUSER": "lightdash",
    "PGHOST": "34.148.82.41",
    "PGPORT": "5432",
    "PGDATABASE": "lightdash",
    "HEADLESS_BROWSER_HOST": "lightdash-browserless-chrome",
    "HEADLESS_BROWSER_PORT": "80",
    "SCHEDULER_ENABLED": "true",
    "DBT_PROJECT_DIR": "",
    "PORT": "8080",
    "SECURE_COOKIES": "true",
    "SITE_URL": "https://lightdash.cort3x.xyz",
    "TRUST_PROXY": "true"
  }
}