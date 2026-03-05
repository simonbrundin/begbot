References in this repo:

- Tradera SOAP client: `internal/marketplaces/tradera.go`
- HTML parsing / scraping: `internal/services/marketplace.go` (ParseTraderaDoc & fetchTraderaAds)
- Email policy enforcement: `internal/services/bot.go:shouldSendTradingRuleEmail`
- Existing specs that touched scraping/ads: `agent-os/specs/2026-02-04-1900-scraper-log-streaming` and `agent-os/specs/2026-02-03-0900-manual-ad-fetch`

External references to consult if needed:

- Tradera API docs (internal/archival) — SOAP Search/GetItem semantics
