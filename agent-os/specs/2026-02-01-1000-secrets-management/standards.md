# Standards: Secrets Management

## Applied Standards

### configuration-structure
Config uses nested structs with yaml tags. Environment variable loading should follow same pattern.

## Additional Guidelines

### Secret Handling
- Never log secrets
- Use environment variables for runtime secrets
- `.env` files must be gitignored
- Kubernetes Secrets for production
