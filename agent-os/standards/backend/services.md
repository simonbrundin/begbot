## Service Dependency Pattern

### Constructor Injection

Use constructor injection for all service dependencies. This makes dependencies explicit and enables mocking in tests.

```go
type BotService struct {
    cfg                *config.Config
    marketplaceService *MarketplaceService
    cacheService       *CacheService
    llmService         *LLMService
    database           *db.Postgres
}

func NewBotService(cfg *config.Config, marketplaceService *MarketplaceService, /* ... */) *BotService {
    return &BotService{
        cfg:                cfg,
        marketplaceService: marketplaceService,
        // ...
    }
}
```

### Job-Aware Services

Use constructor variants only when a service needs optional job context:

```go
func NewBotService(cfg *config.Config, /* ... */) *BotService {
    // Basic constructor for standalone use
}

func NewBotServiceWithJob(cfg *config.Config, /* ..., */ jobService *JobService, jobID string) *BotService {
    // Variant for job context - enables job logging
    return &BotService{
        // ...
        jobService: jobService,
        jobID:      jobID,
    }
}
```

### Logging Pattern

Services should have a `log` method that handles both stdout and job logging:

```go
func (s *BotService) log(level LogLevel, format string, args ...interface{}) {
    message := fmt.Sprintf(format, args...)
    log.Printf("[%s] %s", level, message)

    if s.jobService != nil && s.jobID != "" {
        s.jobService.AddLog(s.jobID, level, message)
    }
}
```
