# Plan: Secrets Management

## Goal
Make it safe to commit to public repo by removing hardcoded secrets from `config.yaml`.

## Context
Code already has `applyEnvOverrides()` in `config.go:90` that reads from environment variables. `.env` is already in `.gitignore`. Only `config.yaml` needs fixing.

## Tasks

### Task 1: Save spec documentation
Save this plan and related spec files.

### Task 2: Replace secrets in config.yaml with placeholders
Update `config.yaml`:
- Database password → `env:DATABASE_PASSWORD` or empty placeholder
- LLM API key → `env:LLM_API_KEY` or empty placeholder  
- SMTP credentials → `env:SMTP_USERNAME`, `env:SMTP_PASSWORD`

### Task 3: Create local .env file with actual secrets
Create `.env` file (gitignored) with:
```
DATABASE_PASSWORD=actual_password
LLM_API_KEY=actual_key
SMTP_PASSWORD=actual_password
```

### Task 4: Update .env.example
Add empty value placeholders for all required secrets.

### Task 5: Verify Kubernetes External Secrets
Ensure Kubernetes manifests correctly reference secrets from External Secrets Operator.

### Task 6: Document workflow
Add README section explaining:
- `config.yaml` is safe to commit
- `.env` file required for local development
- Production uses Kubernetes Secrets

## Decision
- App already supports env var overrides
- Just need to clean up config.yaml and ensure .env file is used locally
- Production uses Kubernetes Secrets via External Secrets Operator
