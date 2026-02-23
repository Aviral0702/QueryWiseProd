# QueryWise

A lightweight, secure agent-based system designed to analyze and optimize cloud database costs.

## Overview

QueryWise helps developers and DevOps teams identify:
- **Expensive queries** — by cost and I/O
- **Unused indexes** — via usage stats
- **Inefficient patterns** — without exposing sensitive data

## Architecture

| Component | Tech | Role |
|-----------|------|------|
| **Agent** | Go | Runs near the DB; collects from `pg_stat_statements`, hashes query text, sends metrics to backend |
| **Backend** | Node.js + Express | Ingest API, cost engine, reports API |
| **Frontend** | Next.js + Tailwind + shadcn | Dashboard: top queries, cost, response time trend |

## Quick start (Docker)

```bash
# Start backend + frontend + Postgres
docker compose up -d db backend frontend

# Run migrations (backend runs them on start). Create an instance and get API key:
docker compose exec backend node scripts/seed-instance.js my-db
# Copy the printed api_key into agent config.

# Open dashboard
open http://localhost:3001
```

Backend: http://localhost:3000  
Frontend: http://localhost:3001  

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
go build -o querywise-agent ./cmd/main.go
./querywise-agent -config config/agent.yml
```

Postgres must have `pg_stat_statements` enabled and the agent DB user must have `SELECT` on it (see [agent/README.md](./agent/README.md)).

## Login (Google SSO)

The dashboard is protected by **Google OAuth 2**. To enable:

1. In [Google Cloud Console](https://console.cloud.google.com/apis/credentials), create an OAuth 2.0 Client ID (Web application).
2. Add authorized redirect URI: `http://localhost:3000/api/v1/auth/google/callback` (or your backend URL + `/api/v1/auth/google/callback`).
3. Set in backend `.env`: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and `FRONTEND_URL` (e.g. `http://localhost:3001`).

Without these, the app redirects to `/login`; with them, users sign in with Google and get a JWT. Instances and reports APIs require `Authorization: Bearer <token>`.

**Bypass auth for testing:** set `ENABLE_GOOGLE_AUTH=false` in the backend and `NEXT_PUBLIC_REQUIRE_AUTH=false` in the frontend (e.g. in `.env.local`). Then instances and reports are accessible without login. The nav shows a “Testing — auth disabled” badge.

## API

- `GET /api/v1/auth/google` — Redirect to Google sign-in.
- `GET /api/v1/auth/google/callback` — OAuth callback (used by Google).
- `GET /api/v1/auth/me` — Current user (header: `Authorization: Bearer <token>`).
- `POST /api/v1/ingest` — Agent sends metrics (header: `X-API-Key`).
- `GET /api/v1/reports/:db_id` — Report for dashboard (requires auth).
- `GET /api/v1/instances` — List instances. `POST /api/v1/instances` — Create (body: `{ "name": "my-db" }`; requires auth). You can also create instances from the dashboard (Home).

## Directory structure

```
QueryWise/
├── agent/          # Go agent (config, collector, client)
├── backend/        # Node.js API, migrations, cost engine
├── frontend/       # Next.js dashboard
├── config/         # agent.example.yml
├── docker-compose.yml
└── README.md
```

## Documentation

- [Architecture](./ARCHITECTURE.md)
- [Agent](./agent/README.md)
- [Backend](./backend/README.md)
