# Shape: Secrets Management

## Problem
`config.yaml` contains hardcoded sensitive data (database passwords, API keys). This prevents safe commits to public repository.

## Scope
Remove all secrets from `config.yaml`, load from environment variables instead.

## Decisions
1. **Config file remains YAML** - `config.yaml` stays for structure/non-secrets
2. **Environment variable fallback** - Code reads env vars, with defaults for non-secrets
3. **Local development** - `.env` file loaded by app (gitignored)
4. **Production** - Kubernetes Secrets via External Secrets Operator

## Changes Required
- `config.yaml`: Replace values with placeholders or env var references
- Go code: Add env var loading
- `.gitignore`: Add `.env`
- Kubernetes: Already has external secrets, just update references

## Constraints
- Must not break existing functionality
- Must work locally and in Kubernetes
