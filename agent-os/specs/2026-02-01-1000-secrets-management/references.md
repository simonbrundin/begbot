# References: Secrets Management

## Existing Code
- `/home/simon/repos/begbot/config.yaml` - Current config with secrets
- `/home/simon/repos/begbot/.env.example` - Template for env vars
- `/home/simon/repos/begbot/internal/db/postgres.go:21` - Database connection
- `/home/simon/repos/begbot/environments/kubernetes/base/secret.yaml` - K8s secrets
- `/home/simon/repos/begbot/environments/kubernetes/base/external-secret.yaml` - External Secrets Operator

## Go Libraries for Env Vars
- `github.com/kelseyhightower/envconfig` - Common for 12-factor apps
- `github.com/caarlos0/env/v9` - Simpler alternative
- Standard `os.Getenv` - Minimal dependency

## External Resources
- [12-factor App: Config](https://12factor.net/config)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
