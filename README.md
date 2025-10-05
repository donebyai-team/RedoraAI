# 🧠 RedoraAI

**RedoraAI** is an open-source, AI-powered lead generation platform for **Reddit**.\
It automates lead discovery, monitoring, and engagement — helping you find and connect with high-intent users across Reddit while staying fully compliant with community rules.

> 🚀 Automate your Reddit lead generation with AI agents that discover relevant subreddits, monitor discussions, and craft authentic, rule-safe engagement messages.

---

## 📋 Table of Contents

- [Features](#-features)
- [Architecture Overview](#-architecture-overview)
- [Tech Stack](#-tech-stack)
- [Getting Started](#-getting-started)
    - [Configuration](#configuration)
    - [Backend Setup](#backend-setup)
    - [Frontend Setup](#frontend-setup)
    - [Running the Project](#running-the-project)
- [Integrations](#-integrations)
- [Deployment](#-deployment)
- [Account Safety Strategies](#-account-safety-strategies)
- [Gotchas & TODOs](#-gotchas--todos)
- [Contributing](#-contributing)
- [License](#-license)

---

## ✨ Features

### 🕵️ Lead Generation

- Extract relevant posts for given **keywords**, **subreddits**, and **product details**
- AI-assisted **keyword** and **subreddit** suggestions
- Intelligent post scoring using LLMs

### ⚙️ Automation

- AI-generated **DMs** and **comments** tailored to community guidelines
- **Scheduled auto-replies** on relevant posts
- **Auto DMs** with configurable timing and frequency

### 🔔 Notifications

- Email notifications via [Resend](https://resend.com)
- Slack alerts for daily or weekly summaries

### 👥 Multi-Account Management

- Connect multiple Reddit accounts
- Auto-rotation for comments and DMs
- Rotation strategies:
    - `Random` — pick a random account
    - `Specific` — use a chosen account
    - `Most Qualified` — based on karma, age, etc.

### 📊 Reporting

- Daily and weekly engagement summaries

### 💳 Subscription Management

- Simple in-app subscription logic with plan limits

### 💬 Interactions

- Manage all AI-generated **comments** and **DMs**

### 🗓️ Posting

- Generate and schedule Reddit posts
- Posts follow subreddit rules and guidelines

---

## 🏗️ Architecture Overview

This repository contains all components required to run RedoraAI.

```
.
├── backend/             # Go backend services
│   ├── portal-api/      # Public API layer for frontend (gRPC + Connect)
│   └── spooler/         # Core tracking engine for subreddits and posts
├── frontend/            # Frontend mono-repo
│   ├── portal/          # Web app (Next.js + PNPM)
│   └── packages/        # Shared UI, config, and protobuf packages
└── devel/               # Local development setup scripts
```

### Backend Services

- **Portal** — API layer for frontend over gRPC/Connect-Web
- **Spooler** — Tracks relevant posts based on subreddit/keyword pairs every 24h
    - Configurable fanout limits, polling intervals, and daily quotas

### LLM Layer

- Powered by **LiteLLM** (deployed separately)
- Supports **OpenAI** and **Gemini** APIs interchangeably
- Handles scoring, comment/DM generation, and keyword/subreddit suggestions

---

## 🧰 Tech Stack

**Backend**

- Go `1.23+`
- PostgreSQL
- Redis
- Docker
- LiteLLM
- Playwright

**Frontend**

- Node.js `20+`
- PNPM
- Next.js / React
- Tailwind CSS / Material UI

**Auth & APIs**

- Auth0 (passwordless login)
- Reddit OAuth APIs
- Resend (emails)
- Browserless / Steel.dev (CDP automation)
- DODO Payments (subscriptions)

---

## ⚙️ Getting Started

### Prerequisites

Ensure you have installed:

- Docker
- Go `1.23+`
- Node.js `20+`
- PNPM
- [direnv](https://direnv.net/) for environment variables

---

### Configuration

Start local PostgreSQL and Redis:

```bash
./devel/up.sh
```

Copy the environment file and configure:

```bash
cp .envrc.example .envrc
direnv allow
```

Replace placeholders (`<value>`) with your actual secrets and keys.

---

### Backend Setup

Run tests and start the backend:

```bash
cd backend
go test ./...
go build -o redora && ./redora start
```

Initialize the database:

```bash
./backend/script/migrate.sh up
```

Create a new migration:

```bash
./backend/script/migrate.sh new <migration_name>
```

---

### Frontend Setup

Install dependencies:

```bash
cd frontend
pnpm install
```

Start the development server:

```bash
pnpm dev:portal
```

Visit: [http://localhost:3000](http://localhost:3000)

---

### Running the Project

You’ll need three components running:

1. **Docker** — for Postgres, Redis, Pub/Sub emulator
   ```bash
   ./devel/up.sh
   ```
2. **Backend** — use [reflex](https://github.com/cespare/reflex) for live reload
   ```bash
   reflex -c .reflex
   ```
3. **Frontend** — Next.js app
   ```bash
   cd frontend && pnpm dev:portal
   ```

Visit:

- `http://localhost:8081` → pgweb (Postgres UI)
- `http://localhost:3000` → Redora Portal

---

## 🔌 Integrations

Integrations store external service credentials and configuration.

| Type               | Description                                 |
| ------------------ | ------------------------------------------- |
| **Reddit Cookies** | User-provided cookies for Reddit automation |
| **Slack Webhook**  | Notifications and alerts                    |
| **OAuth Tokens**   | Reddit access/refresh tokens                |

**Manually insert an integration using tools. Example (CLI):**

```bash
doota tools integrations slack_webhook create <org-id> '{"channel":"redora-alerts","webhook":"<slack-url>"}'
```
---

## ☁️ Deployment

- Hosted on **Railway.app**
- GCP used for LLM and Playwright storage
- Secrets managed via **GCP KMS**
- PostgreSQL and Redis hosted on Railway

---

## 🛡️ Account Safety Strategies

To minimize Reddit account bans:

1. **Age-based limits** — automate only with accounts > 2 weeks old
2. **Gradual scaling** — increase activity slowly
3. **DM-first approach** — DMs are safer than comments
4. **Rule adherence** — generate context-aware replies
5. **Consistent engagement** — reply and post regularly
6. **Occasional posting** — maintain activity score

---

## ⚠️ Gotchas & TODOs

- We currently use OpenAI for keyword/subreddit suggestions during onboarding and LiteLLM for other AI related tasks. Ideally, we should use a single AI provider for all tasks.
- When scoring a post, if the score is >90, we double check it with an advance model. This could be improved by selecting a better default model.
- Comment and DM generation should be moved into separate LLM calls. Right now, scoring, comment and DM generation are all done in a single LLM call.
- We should add the ability to regenerate comments or DMs.
- To avoid getting banned, we only use Reddit accounts that are > 2 weeks old for AI generation. This is a temporary solution and we should come up with a better way to handle account warmup.

---

## 🤝 Contributing

We welcome contributions!

1. Fork the repo
2. Create a new branch:
   ```bash
   git checkout -b feature/your-feature
   ```
3. Commit your changes
4. Open a Pull Request

Please follow Go and JS linting rules before submitting.

---

## 📜 License

Released under the [MIT License](LICENSE).\
© 2025 DoneByAI — building AI tools that work *for* you.

---

### 💡 Maintainers

- [DoneByAI Team](https://donebyai.team)
- [Shashank Agarwal](https://github.com/shank318)