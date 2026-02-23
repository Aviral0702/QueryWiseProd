# QueryWise Agent

The Go-based database agent that runs near the customer's PostgreSQL database.

## Overview

The agent collects query performance metadata from PostgreSQL and sends it securely to the QueryWise backend for analysis.

## Features

- Reads PostgreSQL performance metrics (via `pg_stat_statements`)
- Filters and anonymizes sensitive data
- Securely sends metrics to central backend
- Lightweight and low-overhead
- Configurable collection intervals

## Requirements

- Go 1.21 or later
- PostgreSQL with `pg_stat_statements` extension enabled
- Network connectivity to QueryWise backend

## Setup

```bash
cd agent
go build -o querywise-agent ./cmd/main.go
```

## Configuration

See `config/agent.example.yml` for configuration options.

## Running

```bash
./querywise-agent -config config/agent.yml
```

## Database Permissions

The agent requires a read-only PostgreSQL user with access to:
- `pg_stat_statements`
- `pg_stat_user_tables`
- `pg_stat_user_indexes`
- `pg_index`

Example:
```sql
CREATE ROLE querywise_reader WITH LOGIN PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE your_db TO querywise_reader;
GRANT USAGE ON SCHEMA pg_catalog TO querywise_reader;
GRANT SELECT ON pg_stat_statements TO querywise_reader;
GRANT SELECT ON pg_stat_user_tables TO querywise_reader;
GRANT SELECT ON pg_stat_user_indexes TO querywise_reader;
GRANT SELECT ON pg_index TO querywise_reader;
```
