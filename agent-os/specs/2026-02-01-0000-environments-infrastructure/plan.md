# Environments Infrastructure Plan

## Översikt
Skapa en komplett environments-struktur för begbot med stöd för både nuvarande Supabase-konfiguration och framtida Kubernetes-deployments.

## Mappstruktur
```
agent-os/specs/2026-02-01-0000-environments-infrastructure/
├── plan.md                    # Detta dokument
├── shape.md                   # Shaping-beslut och kontext
├── standards.md              # Relevanta standarder
├── references.md             # Referenser till liknande kod
├── visuals/                  # Mockups/screenshots
└── environments/
    ├── kubernetes/
    │   ├── base/             # Base manifests (deployment, service, etc.)
    │   ├── overlays/
    │   │   ├── dev/          # Development environment
    │   │   └── prod/         # Production environment
    │   └── services/
    │       ├── api/          # Backend API service
    │       ├── workers/      # Background workers
    │       └── frontend/     # Frontend (om tillämpligt)
    └── supabase/
        ├── migrations/       # SQL-migreringar
        └── config/           # Supabase-konfiguration
```

## Tasks

### Task 1: Save spec documentation
Spara spec-dokumentation i `agent-os/specs/2026-02-01-0000-environments-infrastructure/`

### Task 2: Skapa Kubernetes base manifests
- ~~`base/deployment.yaml`~~ ✅ Generisk deployment för begbot-tjänster
- ~~`base/service.yaml`~~ ✅ ClusterIP service
- ~~`base/kustomization.yaml`~~ ✅ Kustomize config
- ~~`base/configmap.yaml`~~ ✅ Applikationskonfiguration
- ~~`base/secret.yaml`~~ ✅ Hemligheter (template)

### Task 3: Skapa Kubernetes service-specifika overlays
- ~~`services/api/kustomization.yaml`~~ ✅ API service konfiguration
- ~~`services/workers/kustomization.yaml`~~ ✅ Workers konfiguration
- ~~`services/frontend/kustomization.yaml`~~ ✅ Frontend konfiguration

### Task 4: Skapa dev overlay
- ~~`overlays/dev/kustomization.yaml`~~ ✅
- ~~`overlays/dev/replicas.yaml`~~ ✅
- ~~`overlays/dev/env.yaml`~~ ✅
- ~~`overlays/dev/namespace.yaml`~~ ✅

### Task 5: Skapa prod overlay
- ~~`overlays/prod/kustomization.yaml`~~ ✅
- ~~`overlays/prod/replicas.yaml`~~ ✅
- ~~`overlays/prod/hpa.yaml`~~ ✅
- ~~`overlays/prod/pdb.yaml`~~ ✅
- ~~`overlays/prod/env.yaml`~~ ✅
- ~~`overlays/prod/namespace.yaml`~~ ✅

### Task 6: Skapa PostgreSQL manifest för framtida Kubernetes
- ~~`services/db/postgres.yaml`~~ ✅ PostgreSQL statefulset
- ~~`services/db/pvc.yaml`~~ ✅ Persistent volume claim
- ~~`services/db/secrets.yaml`~~ ✅ Database credentials

### Task 7: Inkludera Supabase schema
- ~~Kopiera `schema_improved.sql`~~ ✅
- ~~Skapa `supabase/config/supabase-config.toml`~~ ✅

### Task 8: Dokumentera databas-migreringsstrategi
- ~~`supabase/migrations/README.md`~~ ✅ Migreringsprocess
- ~~`supabase/config/environment-variables.md`~~ ✅ Miljövariabler för databas

## Leverabler
1. Komplett environments-mappstruktur
2. Kubernetes-manifests för alla tjänster
3. Supabase-konfiguration bevarad
4. Dokumentation för databas-strategi
