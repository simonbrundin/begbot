# Standards

## Tillämpade standarder

### Kubernetes Best Practices
- **Resurs-request/limits**: Alla containers måste ha resource requests och limits
- **Liveness/readiness probes**: För alla tjänster
- **PodDisruptionBudget**: För produktionsservices
- **HorizontalPodAutoscaler**: För automatisk skalning i prod

### Security
- **Secrets hantering**: Inga secrets i plain text i git
- **Least privilege**: Service accounts med minimala rättigheter
- **Network policies**: Namespace-isolering

### Database
- **Idempotenta migreringar**: Alla SQL-migreringar måste vara idempotenta
- **Migrations med version**: Numererade migreringar i kronologisk ordning
- **Backup-strategi**: Dokumenterad backup-procedur

### CI/CD
- **Infrastructure as Code**: Alla konfigurationer versionerade
- **Immutable artifacts**: Docker images med specifika tags

## Externa standarder
- [12-Factor App](https://12factor.net/) - Konfigurationshantering
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [SQL Style Guide](https://www.sqlstyle.guide/) - SQL-kodstandard
