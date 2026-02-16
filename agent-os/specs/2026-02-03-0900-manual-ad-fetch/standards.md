# Standards: Manual Ad Fetch

## Applicable Standards

From `agent-os/standards/index.yml`:

### 1. configuration-structure
**Applies:** Yes - New config for job tracking if needed

**Application:**
- Use nested config structs for FetchConfig
- Use time.Duration for timeouts

### 2. currency-storage
**Applies:** No - This feature doesn't handle monetary values

### 3. email-sending
**Applies:** No - No email functionality in this feature

### 4. llm-service-functions
**Applies:** No - LLM service not used for ad fetching

### 5. tech-stack
**Applies:** Yes - Follow Go and Nuxt conventions

**Application:**
- Go: Standard library net/http, context.Context everywhere
- Nuxt: useFetch, reactive state, TypeScript types

## Additional Conventions

### API Design
- RESTful endpoints with proper HTTP methods
- JSON responses
- Error responses with proper status codes
- CORS headers (already configured in cmd/api/main.go)

### Frontend Patterns
- Reuse existing `config.public.apiBase` pattern
- Follow existing scraping.vue component structure
- Use TypeScript types from `~/types/database`
- Tailwind CSS for styling (existing project uses it)

### Go Patterns
- context.Context for all operations
- Structured logging with log package
- Error wrapping with %w
