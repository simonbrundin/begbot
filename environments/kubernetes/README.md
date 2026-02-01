# Kubernetes Services

## API Service

Frontend för Slack-integrationer och API-anrop.

```bash
# Deploy
kubectl apply -k environments/kubernetes/services/api/overlays/dev

# Skala
kubectl scale deployment begbot-api --replicas=3 -n begbot-dev
```

## Workers Service

Bakgrundsjobb för schemalagda uppgifter.

```bash
# Deploy
kubectl apply -k environments/kubernetes/services/workers/overlays/dev

# Logs
kubectl logs -l component=workers -n begbot-dev -f
```

## Frontend Service

Webb-interface (om tillämpligt).

```bash
# Deploy
kubectl apply -k environments/kubernetes/services/frontend/overlays/dev
```

## Database Service

PostgreSQL för framtida self-hosted databas.

```bash
# Initialisera
kubectl apply -k environments/kubernetes/services/db/

# Backup
kubectl exec begbot-postgres-0 -n begbot-prod -- pg_dump -U postgres begbot > backup.sql
```
