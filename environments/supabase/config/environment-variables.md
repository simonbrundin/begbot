# Environment Variables

## Application Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_HOST` | PostgreSQL host | - | Yes |
| `DATABASE_PORT` | PostgreSQL port | 5432 | Yes |
| `DATABASE_USER` | Database user | - | Yes |
| `DATABASE_PASSWORD` | Database password | - | Yes |
| `DATABASE_NAME` | Database name | postgres | Yes |
| `DATABASE_SSLMODE` | SSL mode | require | No |

## LLM Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `LLM_PROVIDER` | LLM provider | openrouter |
| `LLM_API_KEY` | API key | - |
| `LLM_DEFAULT_MODEL` | Default model | deepseek/deepseek-v3.2 |

## Scraping Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `SCRAPING_TRADERA_ENABLED` | Enable Tradera scraping | true |
| `SCRAPING_TRADERA_TIMEOUT` | Timeout for Tradera | 30s |
| `SCRAPING_BLOKET_ENABLED` | Enable Blocket scraping | true |
| `SCRAPING_BLOKET_TIMEOUT` | Timeout for Blocket | 30s |

## Valuation Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `VALUATION_TARGET_SELL_DAYS` | Target days to sell | 14 |
| `VALUATION_MIN_PROFIT_MARGIN` | Minimum profit margin | 0.15 |
| `VALUATION_SAFETY_MARGIN` | Safety margin | 0.2 |

## Email Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP host | smtp.gmail.com |
| `SMTP_PORT` | SMTP port | 587 |
| `SMTP_USERNAME` | SMTP username | - |
| `SMTP_PASSWORD` | SMTP password | - |
| `EMAIL_FROM` | From address | - |

## App Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Log level | info |
| `ENVIRONMENT` | Environment | development |
| `CACHE_TTL` | Cache TTL | 24h |
