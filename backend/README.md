# QueryWise Backend

Node.js API for ingesting agent metrics and serving cost-style reports to the dashboard.

## Overview

- Accepts **POST `/api/v1/ingest`** with `X-API-Key` (maps to a `db_instance`).
- Serves **GET `/api/v1/reports/:db_id`** (instance id or name) with optional JWT when `ENABLE_GOOGLE_AUTH` is not `false`.
- Serves **GET/POST `/api/v1/instances`** (and regenerate-key, delete) with the same auth rules; new instances get `owner_user_id` from the JWT when auth is on.

## Requirements

- Node.js 18+ and npm
- PostgreSQL for metrics storage

## Setup

```bash
cd backend
npm install
cp .env.example .env
# Set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, API_SECRET / JWT_SECRET
npm run migrate
```

## Running

Development:

```bash
npm run dev
```

Production:

```bash
npm start
```

## API (summary)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | — | Health |
| GET | `/api/v1/health` | — | Health |
| GET | `/api/v1/auth/google` | — | OAuth redirect |
| GET | `/api/v1/auth/google/callback` | — | OAuth callback |
| GET | `/api/v1/auth/me` | Bearer | Current user |
| POST | `/api/v1/ingest` | `X-API-Key` | Agent metrics |
| GET/POST | `/api/v1/instances` | Bearer* | List / create instances |
| GET | `/api/v1/instances/:id` | Bearer* | Instance detail; `?reveal=true` for API key |
| POST | `/api/v1/instances/:id/regenerate-key` | Bearer* | New API key |
| DELETE | `/api/v1/instances/:id` | Bearer* | Remove instance |
| GET | `/api/v1/reports/:db_id` | Bearer* | Ranked queries + trend |

\* Bearer not required when `ENABLE_GOOGLE_AUTH=false`.

## Database

Run migrations:

```bash
npm run migrate
```

Create an instance from the shell (optional owner via `OWNER_USER_EMAIL`):

```bash
node scripts/seed-instance.js my-db
```
