# Shaping Decisions

## Scope
Skapa en environments-mapp som stöder både nuvarande och framtida infrastrukturbehov för begbot.

## Beslut

### Beslut 1: Hybrid-strategi för databas
- **Val**: Bibehåll Supabase för nuvarande produktion, förbered Kubernetes PostgreSQL för framtid
- **Motivering**: Supabase fungerar väl nu, men Kubernetes ger mer kontroll vid skalning
- **Konsekvens**: Två databas-konfigurationer måste同步iseras vid migrering

### Beslut 2: Kustomize för Kubernetes
- **Val**: Använd Kustomize istället för Helm
- **Motivering**: Enklare för mindre projekt, bättre Git-integration, färre dependencies
- **Konsekvens**: Standard Kubernetes-manifests med overlay-pattern

### Beslut 3: Separata services för API och Workers
- **Val**: dela upp i `api` och `workers` tjänster
- **Motivering**: Olika skalningsbehov, olika resurskrav
- **Konsekvens**: Två separate deployments med egen konfiguration

### Beslut 4: Schema som kod
- **Val**: Alla databas-schema i migrations-filer
- **Motivering**: Versionshantering, reproducerbar miljö, CI/CD-integration
- **Konsekvens**: Migreringar måste vara bakåtkompatibla

## Konfiguration
- **Applikationsconfig**: config.yaml -> Kubernetes ConfigMap
- **Hemligheter**: Kubernetes Secrets (inte i git)
- **Databas-credentials**: Supabase credentials -> Kubernetes Secrets

## Miljöer
- **dev**: Lokal/utvecklingsmiljö med simpla resources
- **prod**: Produktionsmiljö med HPA, PDB, och full redundans

## Kända begränsningar
1. Inga ArgoCD/Kargo-configurations inkluderade (kan läggas till senare)
2. Ingen ingress-konfiguration för närvarande
3. Inga monitoring/observability configs
