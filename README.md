# QueryWise

A lightweight, secure agent-based system designed to analyze and optimize cloud database costs using **PostgreSQL `pg_stat_statements`** metrics (hashed query text only—no raw SQL stored centrally).

## Overview

QueryWise helps developers and DevOps teams identify:

- **Expensive queries** — ranked by a simple cost model (CPU time + block I/O + per-call overhead)
- **Response-time trends** — aggregated from ingested snapshots

The agent does **not** collect index-usage reports or automated rewrite recommendations today; those would be future work.

## Architecture

| Component | Tech | Role |
|-----------|------|------|
| **Agent** | Go | Runs near the DB; collects from `pg_stat_statements`, hashes query text, sends metrics to backend |
| **Backend** | Node.js + Express | Ingest API, cost estimates, reports API |
| **Frontend** | Next.js | Dashboard: instances, API keys, top queries, summary stats |

## Quick start (Docker)

```bash
# Start backend + frontend + Postgres
docker compose up -d db backend frontend

# Migrations run on backend start. Create an instance and get API key:
docker compose exec backend node scripts/seed-instance.js my-db
# Copy the printed api_key into agent config.

# Open dashboard
open http://localhost:3001
```

Backend: http://localhost:3000  
Frontend: http://localhost:3001  

With **Google auth enabled** (`ENABLE_GOOGLE_AUTH` not `false`), instances created via the CLI seed have **no owner** until you either:

- Create instances from the dashboard while signed in (recommended), or  
- Set ownership after sign-in, e.g. `UPDATE db_instances SET owner_user_id = (SELECT id FROM users WHERE email = 'you@example.com' LIMIT 1) WHERE name = 'my-db';`, or  
- Run seed with `OWNER_USER_EMAIL=you@example.com docker compose exec -e OWNER_USER_EMAIL=... backend node scripts/seed-instance.js my-db` (user must have signed in once so the row exists).

## Quick start (local)

**1. Backend**

```bash
cd backend
cp .env.example .env   # set DB_* and API_SECRET
npm install
npm run migrate
npm run seed my-db    # creates instance, prints API key
npm run dev
```

**2. Frontend**

```bash
cd frontend
cp .env.local.example .env.local   # NEXT_PUBLIC_API_URL=http://localhost:3000
npm install
npm run dev
```

Open http://localhost:3001

**3. Agent** (on a host that can reach your Postgres and the backend)

```bash
cd agent
cp ../config/agent.example.yml config/agent.yml
# Edit config/agent.yml: database.*, backend.url, backend.api_key (from step 1)
# Set agent.name to the same string as the QueryWise instance *name* so ingest labels match the instance.
go build -o querywise-agent ./cmd/main.go
./querywise-agent -config config/agent.yml
```

Postgres must have `pg_stat_statements` enabled and the agent DB user must have `SELECT` on it (see [agent/README.md](./agent/README.md)).

## Login (Google SSO)

The dashboard can be protected by **Google OAuth 2**. To enable:

1. In [Google Cloud Console](https://console.cloud.google.com/apis/credentials), create an OAuth 2.0 Client ID (Web application).
2. Add authorized redirect URI: `http://localhost:3000/api/v1/auth/google/callback` (or your backend URL + `/api/v1/auth/google/callback`).
3. Set in backend `.env`: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `FRONTEND_URL` (e.g. `http://localhost:3001`).

If SSO is not configured, opening `/api/v1/auth/google` in a **browser** redirects to `/login?error=sso_not_configured` instead of returning raw JSON.

**Bypass auth for testing:** set `ENABLE_GOOGLE_AUTH=false` in the backend and `NEXT_PUBLIC_REQUIRE_AUTH=false` in the frontend (e.g. in `.env.local`). Then instances and reports are accessible without login. The nav shows a “Testing — auth disabled” badge.

## API

- `GET /api/v1/auth/google` — Redirect to Google sign-in.
- `GET /api/v1/auth/google/callback` — OAuth callback (used by Google).
- `GET /api/v1/auth/me` — Current user (header: `Authorization: Bearer <token>`).
- `POST /api/v1/ingest` — Agent sends metrics (header: `X-API-Key`). The payload may include `db_id`; it should match the instance name or id (otherwise the backend logs a warning).
- `GET /api/v1/reports/:db_id` — Report for dashboard (requires auth when enabled). Each user only sees instances they own.
- `GET /api/v1/instances` — List instances. `POST /api/v1/instances` — Create (body: `{ "name": "my-db" }`; requires auth when enabled). Owner is set from the JWT.

## Directory structure

```
QueryWise/
├── agent/          # Go agent (config, collector, client)
├── backend/        # Node.js API, migrations, cost engine
├── frontend/       # Next.js dashboard
├── config/         # agent.example.yml (if present)
├── docker-compose.yml
└── README.md
```

## Documentation

- [Architecture](./ARCHITECTURE.md)
- [Agent](./agent/README.md)
- [Backend](./backend/README.md)
