## Deployment

`
`Deployment is managed with [Tanka](https://tanka.dev/) which makes it possible to easily create the Kubernetes manifests.

### Getting Started

First install the [Tanka](https://tanka.dev/install) CLI tool and also ensure that `kubectl` is properly configure to contact
your cluster.

You can easily see all the manifests that are going to be deploying to the cluster with (all commands assumes you are
at the project root):

```
tk show deploy/environments/production
```

> You can filter for specific kind adding `-t deployment/.*` or specific name `-t .*/.*name.*`.

> If you hit an error saying `Error: evaluating jsonnet: RUNTIME ERROR: couldn't open import "secrets.jsonnet"`, you forgot to setup the `secrets.jsonnet` file correctly.

### Cluster Setup Procedure

1. Create the k8s namespaces. We create the namespace via `kubcetl` directly

```bash
kubectl create namespace production
kubectl create namespace staging
```

2. We will need to manually create the secrets in ALL namespaces

```bash
kubectl create secret generic airbrake --from-literal=projectId=<PROJECT_ID> --from-literal=projectKey=<PROJECT_KEY> -n <production|staging>
kubectl create secret generic auth0 --from-literal=clientId=<CLIENT_ID> --from-literal=clientSecret=67oCdQp-<CLIENT_SECRET> -n <production|staging>
kubectl create secret generic sql-database-pword --from-literal="password=<PASSWORD>" -n <production|staging>
kubectl create secret generic sql-bastion --from-literal="pgdsn=postgres://<USER>:<PASSWORD>@<IP>:5432/<DBNAME>?sslmode=disable" -n <production|staging>
kubectl create secret generic openai --from-literal="api_key=<API-KEY>" --from-literal="org=<OR-ID>" -n production
kubectl create secret generic unidoc --from-literal="api_key=<API-KEY>" -n <production|staging>
kubectl create secret generic langsmith --from-literal=api_key=<API-KEY> --from-literal=project=<PROJECT>> -n <production|staging>
kubectl create secret generic microsoft --from-literal="client_id=<CLIENT-ID>" --from-literal="client_secret=<CLIENT-SECRET>" -n <production|staging>
kubectl create secret generic shipwell-tms --from-literal="username=<USERNAME>" --from-literal="password=<PASSWORD>"  -n <production|staging>
kubectl create secret generic agl-tms --from-literal="token=<TOKEN>" -n <production|staging>
kubectl create secret generic turvo-tms \
  --from-literal="username=<USERNAME>" \
  --from-literal="password=<PASSWORD>" \
  --from-literal="client_id=<CLIENT_ID>" \
  --from-literal="client_secret=<CLIENT_SECRET>" \
  --from-literal="api_key=<API_KEY>" \
  --from-literal="api_host=<API_HOST>" \
  --from-literal="ui_host=<UI_HOST>" \
  -n <production|staging>
kubectl create secret generic google --from-literal="client_id=<CLIENT-ID>" --from-literal="client_secret=<CLIENT-SECRET>" -n  <production|staging>
kubectl create secret generic google-geocoding-api --from-literal="api_key=<GOOGLE-GEOCODING-API-KEY" -n  <production|staging>
```



3. Setting up the service account for the SQL Proxy. We need to setup the cloud SQL proxy service account in the k8s cluster.
   You can follow this guide for more information https://cloud.google.com/sql/docs/mysql/connect-kubernetes-engine
   The service account has already been created in terraform, we simply need to import the service account key as a secret

```bash
# Create a credential file for your service account key:
gcloud iam service-accounts keys create cloud-sql-sa.json --iam-account=cloud-sql-proxy@doota.iam.gserviceaccount.com
# Turn your service account key into a k8s Secret:
kubectl create secret generic sql-cloud-proxy-sa --from-file=service_account.json=cloud-sql-sa.json -n <production|staging>
```

4. Github Config
   If you which to enable Github Config, you must create the secret

```bash
kubectl create secret generic github-config --from-file=github_config.yml -n production
```

Note we have a template file [here github_config.template.yml](./github_config.template.yml). You must replace the information from the github app.

5. Deploying

Deploying is simply a matter of checking the differences with the actual cluster state and then apply it.

```
tk diff deploy/environments/production
```

And to apply:

```
tk apply deploy/environments/production
```

note: the commands above are for the production environment, you can replace `production` with `staging` to deploy to the staging environment.
