# QueryWise Architecture

## System Overview

QueryWise is a two-tier system designed to securely analyze database costs without exposing sensitive customer data.

```
┌─────────────────────────────────────────────────────────────┐
│                  Customer Infrastructure                     │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   PostgreSQL Database                  │ │
│  │  - Query Performance Metadata (pg_stat_statements)     │ │
│  │  - Index Usage Statistics                              │ │
│  │  - Cache Hit Ratios                                    │ │
│  └────────────────────────────────────────────────────────┘ │
│                          ▲                                   │
│                          │ Queries                           │
│                          │ (Read-only)                       │
│  ┌────────────────────────────────────────────────────────┐ │
│  │        QueryWise Agent (Go-based)                      │ │
│  │  - Collects performance metadata                       │ │
│  │  - Filters sensitive data                              │ │
│  │  - Sends anonymized metrics                            │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ Secure HTTPS
                          │ Anonymized Metrics Only
                          ▼
┌─────────────────────────────────────────────────────────────┐
│              Central Analytics Backend                       │
│  ┌────────────────────────────────────────────────────────┐ │
│  │      QueryWise Backend (Node.js)                       │ │
│  │  - Receives metrics                                    │ │
│  │  - Analyzes query performance                          │ │
│  │  - Identifies optimization opportunities              │ │
│  │  - Generates cost reports                              │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │          Dashboard & Visualization                     │ │
│  │  - Cost breakdowns                                     │ │
│  │  - Query recommendations                               │ │
│  │  - Index optimization suggestions                      │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Components

### Agent (Go)
- **Location**: Near customer database (on-premise or same VPC)
- **Responsibility**: Collect metrics, sanitize data, send to backend
- **Data Collected**: Query execution times, index usage, cache statistics
- **Security**: No sensitive data leaves customer infrastructure

### Backend (Node.js)
- **Responsibility**: Aggregate metrics, analyze patterns, provide insights
- **Features**: Cost analysis, optimization suggestions, historical tracking
- **API**: RESTful endpoints for dashboard and programmatic access

## Data Flow

1. Agent queries PostgreSQL system views (read-only)
2. Agent aggregates and anonymizes metrics
3. Agent sends summary statistics to backend
4. Backend stores and analyzes data
5. Dashboard visualizes insights

## Security Model

- **Data Isolation**: Each customer's metrics are isolated
- **No Raw Data**: Only aggregated statistics are transmitted
- **Minimal Permissions**: Agent uses read-only database user
- **Encrypted Transport**: All communication is HTTPS
