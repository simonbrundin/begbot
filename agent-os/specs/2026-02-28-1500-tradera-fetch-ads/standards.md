Relevant standards to follow:

- backend/api — follow API client patterns (Tradera client already follows SOAP request building conventions)
- backend/services — Dependency injection and adapters (marketplace client is injected into MarketplaceService)
- backend/queries — respect rate-limits and avoid unbounded API calls
- testing/test-writing — add unit and integration tests with table-driven style
- global/error-handling — return wrapped errors with context
