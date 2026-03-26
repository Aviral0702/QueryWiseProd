# QueryWise Architecture

## System Overview

QueryWise collects **anonymized PostgreSQL query statistics** from `pg_stat_statements` (hashed query text, timing, block counters) and stores time-series snapshots in a central backend. A web dashboard lists **per-user database instances** and shows **cost-ranked queries** and simple **response-time trends**.

```
┌─────────────────────────────────────────────────────────────┐
│                  Customer Infrastructure                     │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   PostgreSQL Database                  │ │
│  │  - pg_stat_statements (read-only)                      │ │
│  └────────────────────────────────────────────────────────┘ │
│                          ▲                                   │
│                          │ Read-only queries                 │
│  ┌────────────────────────────────────────────────────────┐ │
│  │        QueryWise Agent (Go)                            │ │
│  │  - Hashes query text (SHA-256)                         │ │
│  │  - Sends metrics with instance API key                 │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          │ HTTPS (recommended in production)
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Central Backend (Node.js)                       │
│  - Ingest + rate limiting                                    │
│  - Latest snapshot per query hash for rankings               │
│  - Optional Google SSO; instances scoped by owner_user_id    │
└─────────────────────────────────────────────────────────────┘
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Dashboard (Next.js)                             │
│  - Instances, API keys, reports                              │
└─────────────────────────────────────────────────────────────┘
```

## Components

### Agent (Go)

- **Location:** Near the customer database (on-premise or same VPC).
- **Responsibility:** Read `pg_stat_statements`, hash query text, POST JSON to `/api/v1/ingest` with `X-API-Key`.
- **Security:** Raw SQL is not sent to the backend; only hashes and numeric stats.

### Backend (Node.js)

- **Responsibility:** Store metrics, deduplicate snapshots per query hash (latest row wins for rankings), estimate relative cost, expose REST APIs.
- **Auth:** Optional JWT after Google OAuth; `db_instances.owner_user_id` ties instances to `users.id`.

### Frontend (Next.js)

- **Responsibility:** Sign-in (when enabled), create/list instances, open reports.

## Data flow

1. Agent queries PostgreSQL `pg_stat_statements` (read-only).
2. Agent sends batches to the backend with the instance API key.
3. Backend inserts rows tagged with `db_instance_id` and `recorded_at`.
4. Reports use **the latest snapshot per `query_hash`** for rankings (cumulative stats are not summed across time).
5. Dashboard calls authenticated APIs to list instances and load reports.

## Security model

- **Tenant isolation:** When Google auth is enabled, instance list/report access is restricted to rows where `owner_user_id` matches the JWT user. Ingest remains keyed only by API secret (`X-API-Key`).
- **No raw query text** in the central store—only hashes and aggregates.
- **Transport:** Use HTTPS in production between agent, backend, and browsers.

## Not in scope (today)

- Unused-index analysis, automatic rewrite recommendations, or non-Postgres databases.
- Row-level encryption at rest (operate like any service on your Postgres deployment).
