# DootaAI

This repository contains the full code for running all the components of LoadLogic. The main important code folders are:

- `backend/` the Go backend, which includes the "portal-api".
- `frontend` the shared frontend logic/UI across all frontend apps (portal, Chrome extension, etc.)
- `frontend/portal` the frontend application, which connects to the Go backend with Connect-Web.

## Dev environment

### Requirements

- Docker (recent enough)
- Golang 1.21+
- Node.js 20+ & PNPM

### Configuration
*_
_*Start up local PostgresSQL database and Redis locally:

```bash
# This is essentially a wrapper script around `docker compose up`
./devel/up.sh
```

We use environment variables to configure our developer environment to avoid having to specify each flag manually and to kept secrets out of committed Git revisions. We suggest to setup [direnv](https://direnv.net/) to auto-load the `.envrc` containing environment variables when you from your terminal you `cd` in the project.

Copy `.envrc.example` to `.envrc` and edit it:

```bash
cp .envrc.example .envrc
```

All the variables to be filled in are denoted as _`<name>`_ you need to replace them with the proper value(s). Ask your actual manager at LoadLogic to provide you those values.

You will also need to have a Google account to access some infrastructure parts required for development (signing JWT tokens). Again, ask your manager at LoadLogic to take the necessary steps to grant you access.

#### Golang

Ensure that your version of Golang is correct using `go version`. You should be able to run the test suite also to prove that everything is in order:

```bash
cd backend
go test ./...
go build -o doota && ./doota start
```

##### PostgreSQL

You need to bootstrap your database initially, to do so, use the [migrate.sh](./backend/script/migrate.sh), the script assumes environment variables are properly loaded, ensure you performed `direnv allow` or loaded them manually. To migrate to latest schema version do:

```bash
./backend/script/migrate.sh up
```

#### Frontend

The `frontend` contains a mono-repo of the various frontend related packages and applications we have:

- LoadLogic User Portal at [portal](./frontend/portal/).
- LoadLogic User Chrome Extension at [extension/chrome](./frontend/extension/chrome).

As well as shared packages:

- [packages/pb](./frontend/packages/pb) for Protobuf bindings sharing
- [packages/ui-core](./frontend/packages/ui-core) for shared simple and more complex UI components between browser and the various web-based extensions (Chrome, Outlook to come, etc.)
- [packages/mui-config](./frontend/packages/mui-config) for Material UI theme sharing
- [packages/tailwind-config](./frontend/packages/tailwind-config) for Tailwind CSS theme sharing

```bash
cd frontend
pnpm install
```

This installs everything related to frontend packages.

### Running

You will need to run 3 components

1. Docker to setup your postgres, redis, gcp pubsub emulator
2. The backend server with `reflex`
3. The frontend

To start docker simply run the following from the root of the project:

```bash
./devel/up.sh
```

Once docker is up and running you should be able to open `http://localhost:8081/` to see a running instance `pgweb` a Cross-platform client for PostgreSQL databases.

For development purposes, we suggest installing [reflex](https://github.com/cespare/reflex) which will check for file modifications and re-launch the necessary servers/generators. We provide a [`.reflex` config file](./.reflex) that is pre-configured.

```
reflex -c .reflex
```

This starts everything you need to have a fully working backend environment. In another terminal, starts the frontend development server:

```bash
cd frontend
pnpm dev:portal
```

Then you should be open `http://localhost:3000`, register using one of the listed provider and then launch an investigation by drag-and-dropping the `./devel/mock_paul_bailey.csv` to the upload page on the website.

#### Portal Setup

The database needs to be populated with some information if you want to use the `compose email` feature of the portal. Connects to http://localhost:8081 and run the following SQL query, adjust variables to fit your required values:

> [!NOTE]
> You will need to insert organization, then user, then message sources because you need organization and user ids which will be available only after first `INSERT`

```SQL
INSERT INTO organizations (name) VALUES ('DootaAI');
INSERT INTO users (auth0_id, email, email_verified, organization_id, role, state) VALUES ('', 'shank@dootaai.com', true, 'YOUR ORG ID', 'PLATFORM_ADMIN', 'ACTIVE');
INSERT INTO projects (
    name, 
    organization_id, 
    description, 
    customer_persona, 
    goals, 
    website
)
VALUES (
    'MiraAI',
    'YOUR ORG ID',
    'MiraAI is an intelligent assistant platform for analyzing and summarizing pitch decks.', 
    'Venture capital analysts and associates at VC firms screening startup deals.', 
    'Streamline deal screening by summarizing PDFs and extracting key insights automatically.', 
    'https://miraai.com'
)
RETURNING id;

INSERT INTO sources (
    name,
    description,
    project_id,
    source_type
)
VALUES (
    'all',
    'across subreddits',
    'XXX',
    'SUBREDDIT'
)

INSERT INTO keyword_trackers (
    keyword_id,
    source_id
)
VALUES (
    '12757f86-d517-4009-a3fc-4a09bbbec9ff',
    'de452b3e-59a3-401f-8d0f-47832cfe6e4b'
)

```

### Tracing
We have tracing setup in the codebase using OpenTelemetry. To enable tracing, you need to set the `SF_TRACING` environment variable.
The value of the variable should be a valid collector URL.

We have tested 2 collector `Zipkin` and `Jaeger`. : We have experienced none-deterministic issues with jaeger

Running `Zipkin` locally:
```shell
docker run -p 9411:9411 openzipkin/zipkin
```
Running `Jaeger` locally:
```shell
docker run --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:1.35
```

Once you have chosen a collector, you can set the `SF_TRACING` environment variable to the collector URL
* Zipkin: `SF_TRACING=zipkin://localhost:9411?scheme=http`
* Jaeger: `SF_TRACING=jaeger://localhost:4317?scheme=http`

Wait a few seconds after the test is complete to see the trace in your browser:
* Zipkin: http://localhost:9411/
* Jaeger: http://localhost:16686/


## Release
## Tests

```bash
cd backend
go test ./...
```

## Deployment

With a proper `gcloud` env authenticated and having access to the `doota` project:

```bash
gcloud auth configure-docker us-east1-docker.pkg.dev
```

```bash
gcloud container clusters get-credentials saas-us-east1 --region us-east1 --project doota
```

## Tools

There is a `redis-tools` pod that allows you to access the redis
_NOTE_ Make sure you use the correct redis ip based on the ENV you are looking at

```bash
redis-cli -h 10.109.86.67 --scan --pattern '*'
```

## Troubleshooting

**If you encounter an issue related to OSError: [Errno 24] Too many open files**

```bash
ulimit -a # check the number of file descriptors open that you are allow to have with
ulimit -Sn 1000000 # to change it to 1000000
```

**SQL adding organization feature flag**
```sql
UPDATE organizations SET feature_flags = jsonb_set(feature_flags, '{enable_load_diff_email}', 'true') WHERE id='xx'
UPDATE organizations SET feature_flags = jsonb_set(feature_flags, '{enable_auto_comment}', 'true') WHERE id='e250ced8-7441-4805-b9dd-2686d9492c4f'
UPDATE organizations SET feature_flags = jsonb_set(feature_flags, '{relevancy_llm_model}', '"redora-dev-claude-thinking"') WHERE id='e250ced8-7441-4805-b9dd-2686d9492c4f'
update projects set is_active=false where id='d1732e25-386a-48dc-9851-a8fea2156bf2';
UPDATE organizations SET feature_flags = jsonb_set(feature_flags,'{subscription,metadata,relevant_posts,per_day}','5',false) WHERE id='8c0ca052-82f2-439f-ac74-a360b4624599';

```

## Seed DB

You can find a sample seed file here [`./devel/seed.sql`](`./devel/seed.sql`)

## Create a new migration
```bash
./backend/script/migrate.sh new customer
```

## Spooler 
1. Call between given hours
2. Per day max no of calls limits.
3. Total max no of calls limits
4. Fanout limits to max no of calls per org at a time
5. Polling interval

```
DAT:
freight  tools integrations dat create <org-id> {\"auth_host\":\"identity.api.staging.dat.com\",\"api_host\":\"analytics.api.staging.dat.com\",\"org_user\":\"dat@loadlogic.ai\",\"org_password\":\"CHANGE_PASSWORD\",\"user_account\":\"mdm@streamingfast.io\"}
doota tools integrations slack_webhook create 9ce764d3-8663-476f-829d-3181109df3e1 {\"channel\":\"redora-daily-leads-alert\",\"webhook\":\"https://hooks.slack.com/services/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\"}
```

## Deployment
Redora is deployed on railway app. It uses google cloud store for openai debug and KMS for jwt
Postgress and Redis is inside the railway itself. 
Frontend is deployed via it's dockerfile and backend direct