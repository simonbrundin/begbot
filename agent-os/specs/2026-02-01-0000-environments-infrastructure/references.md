# References

## Intern referens
- **Plan-projektet**: `/home/simon/repos/plan/environments/kubernetes/`
  - Kustomize-struktur med base + overlays
  - PostgreSQL + Hasura konfiguration
  - Development och production overlays
  - Kargo/ArgoCD integrationsexempel

## Extern dokumentation
- [Kustomize Documentation](https://kubectl.docs.kubernetes.io/guides/config_management/kustomize/)
- [Kubernetes Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [Kubernetes StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
- [Supabase Migrations](https://supabase.com/docs/guides/cli/local-development#database-migrations)

## Filer i begbot-projektet
- `schema_improved.sql` - Befintligt databas-schema
- `config.yaml` - Applikationskonfiguration
- `cmd/main.go` - Huvudenträde för applikationen

## Mönster att följa
1. Kustomize overlay-pattern från plan-projektet
2. Separata namespace per miljö
3. Environment-specifika konfigurationer via patches
4. Centraliserad kustomization.yaml som aggregat
