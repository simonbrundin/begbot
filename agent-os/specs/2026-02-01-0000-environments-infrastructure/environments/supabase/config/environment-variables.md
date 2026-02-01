# Database Environment Variables

## Overview

This document describes the environment variables required for database connections across different environments.

## Supabase Connection Variables

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `DATABASE_URL` | Full connection string | Yes | `postgresql://user:pass@host:5432/db` |
| `DATABASE_HOST` | Database host | Yes | `aws-1-eu-west-1.pooler.supabase.com` |
| `DATABASE_PORT` | Database port | Yes | `5432` |
| `DATABASE_USER` | Database user | Yes | `postgres.fxhknzpqhrkpqothjvrx` |
| `DATABASE_PASSWORD` | Database password | Yes | `your-password` |
| `DATABASE_NAME` | Database name | Yes | `postgres` |
| `DATABASE_SSLMODE` | SSL mode | No | `require` or `disable` |

## Environment-Specific Configurations

### Development

```yaml
DATABASE_HOST: localhost
DATABASE_PORT: 5432
DATABASE_NAME: begbot_dev
DATABASE_SSLMODE: disable
```

### Production (Supabase)

```yaml
DATABASE_HOST: aws-1-eu-west-1.pooler.supabase.com
DATABASE_PORT: 5432
DATABASE_NAME: postgres
DATABASE_USER: postgres.fxhknzpqhrkpqothjvrx
DATABASE_PASSWORD: <from secrets>
DATABASE_SSLMODE: require
```

### Kubernetes (Future)

```yaml
DATABASE_HOST: postgres.begbot.svc.cluster.local
DATABASE_PORT: 5432
DATABASE_NAME: begbot
DATABASE_SSLMODE: disable
```

## Secrets Management

### Development

- Store in `.env` file (not committed to git)
- Use `.env.example` as template

### Production (Supabase)

- Use Supabase secrets management
- Or Kubernetes Secrets for deployments

### Kubernetes

- Store in Kubernetes Secrets manifests
- Reference via `secretKeyRef` in deployments

## Driver Configuration

The application uses `jackc/pgx/v5` driver (imported as `pgx/stdlib`).

```go
import (
    "github.com/jackc/pgx/v5/stdlib"
)
```

Connection pool settings should be configured in the application config:

```yaml
database:
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
```
