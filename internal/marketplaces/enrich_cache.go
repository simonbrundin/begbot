package marketplaces

import (
	"sync"
	"time"
)

// simple in-memory TTL cache for GetItem responses keyed by item ID.
// Not intended to be distributed — small, conservative implementation.
type enrichCacheEntry struct {
	details   *AdDetails
	expiresAt time.Time
}

type EnrichCache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[int]*enrichCacheEntry
}

func NewEnrichCache(ttl time.Duration) *EnrichCache {
	return &EnrichCache{
		ttl:   ttl,
		items: make(map[int]*enrichCacheEntry),
	}
}

func (c *EnrichCache) Get(id int) (*AdDetails, bool) {
	c.mu.RLock()
	e, ok := c.items[id]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		c.mu.Lock()
		// double-check and delete stale
		if cur, ok2 := c.items[id]; ok2 && cur == e {
			delete(c.items, id)
		}
		c.mu.Unlock()
		return nil, false
	}
	return e.details, true
}

func (c *EnrichCache) Set(id int, details *AdDetails) {
	if details == nil {
		return
	}
	c.mu.Lock()
	c.items[id] = &enrichCacheEntry{
		details:   details,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}
