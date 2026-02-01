# Database Migration Strategy

## Current State
- **Production**: Supabase (managed PostgreSQL)
- **Future**: Self-hosted PostgreSQL on Kubernetes

## Migration Plan

### Phase 1: Schema as Code (Complete)
All schemas are versioned in `supabase/migrations/`

### Phase 2: Dual-Write Period
```
┌─────────────┐     ┌─────────────┐
│   App       │────▶│   Supabase  │
│             │     │   (primary) │
└─────────────┘     └─────────────┘
                          │
                     ┌────▼────┐
                     │  K8s DB │
                     │ (sync)  │
                     └─────────┘
```

### Phase 3: Switch to Kubernetes
1. Update `secret.yaml` to point to K8s PostgreSQL
2. Run final migration
3. Decommission Supabase

## Migrering Commands

### Apply migrations to Supabase
```bash
supabase db push
```

### Apply migrations to K8s PostgreSQL
```bash
kubectl exec -it begbot-postgres-0 -- psql -U postgres -d begbot -f /migrations/001_initial_schema.sql
```

## Backup Strategy

### Supabase
- Automatic daily backups via Supabase dashboard
- Point-in-time recovery available

### Kubernetes PostgreSQL
```yaml
# Use Velero for backups
apiVersion: velero.io/v1
kind: Backup
metadata:
  name: begbot-backup
spec:
  includedNamespaces:
    - begbot-prod
```

## Rollback Procedure
1. Keep Supabase active for 7 days post-migration
2. Document last migration timestamp
3. Test rollback procedure before cutover
