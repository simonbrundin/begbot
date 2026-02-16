# Plan: Secrets Management

## Goal
Make it safe to commit to public repo by removing hardcoded secrets from `config.yaml`.

## Context
Code already has `applyEnvOverrides()` in `config.go:90` that reads from environment variables. `.env` is already in `.gitignore`. Only `config.yaml` needs fixing.

---

## Spec Status: ✅ DONE

Completed: 2026-02-02

---

## Tasks

### Task 1: Save spec documentation
Save this plan and related spec files.

### Task 2: Replace secrets in config.yaml with placeholders ✅ COMPLETED
Update `config.yaml`:
- Database password → `env:DATABASE_PASSWORD` or empty placeholder
- LLM API key → `env:LLM_API_KEY` or empty placeholder
- SMTP credentials → `env:SMTP_USERNAME`, `env:SMTP_PASSWORD`
**Status:** config.yaml already had empty placeholders with env var comments

### Task 3: Create local .env file with actual secrets ✅ COMPLETED
Create `.env` file (gitignored) with:
```
DATABASE_PASSWORD=vaf8PNB@gqj5wux4cje
LLM_API_KEY=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
SMTP_USERNAME=simonbrundin@gmail.com
SMTP_PASSWORD=ephy cbhm lvtk gpdj
SMTP_FROM=simonbrundin@gmail.com
```
**Status:** .env file created and verified in .gitignore

### Task 4: Update .env.example ✅ COMPLETED
Add empty value placeholders for all required secrets.
**Status:** .env.example already complete with TODO comments

### Task 5: Verify Kubernetes External Secrets ✅ COMPLETED
Ensure Kubernetes manifests correctly reference secrets from External Secrets Operator.
**Status:** base/externalsecret.yaml created to pull secrets from Vault
**Status:** ClusterSecretStore `vault-backend` exists in infrastructure repo
**Status:** Deployment uses envFrom: secretRef to inject secrets as environment variables

### How it works:
1. Secrets stored in Vault at path `prod/begbot`
2. ExternalSecret pulls secrets and creates Kubernetes Secret `begbot-secrets`
3. Deployment injects via envFrom: secretRef

### Task 6: Document workflow ✅ COMPLETED
Add README section explaining:
- `config.yaml` is safe to commit
- `.env` file required for local development
- Production uses Vault via External Secrets Operator
**Status:** README.md updated with Secrets Management section

## Decision
- App already supports env var overrides
- Just need to clean up config.yaml and ensure .env file is used locally
- Production uses Vault via External Secrets Operator

---

## Spec Status: ✅ DONE

Completed: 2026-02-02
