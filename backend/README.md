# QueryWise Backend

The Node.js-based central analytics backend for QueryWise.

## Overview

The backend receives metrics from deployed agents, analyzes database cost patterns, and provides optimization recommendations via a REST API and dashboard.

## Features

- Aggregates metrics from multiple agents
- Analyzes query performance and cost patterns
- Identifies optimization opportunities (indexes, query rewrites)
- Provides cost breakdowns and trends
- RESTful API for programmatic access
- Web dashboard for visualization

## Requirements

- Node.js 18+ and npm
- PostgreSQL for metrics storage
- Redis for caching (optional)

## Setup

```bash
cd backend
npm install
```

## Configuration

Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
# Edit .env with your database and service configuration
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

## API Endpoints

- `GET /api/health` - Health check
- `POST /api/metrics` - Receive metrics from agents
- `GET /api/analysis/:agentId` - Get analysis for an agent
- `GET /api/dashboard` - Dashboard data

See API documentation in `docs/API.md`

## Database

The backend stores metrics and analysis results in PostgreSQL. Run migrations:

```bash
npm run migrate
```
