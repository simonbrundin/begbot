package services

import (
	"context"
	"sync"
	"time"

	"begbot/internal/config"
)

type CacheService struct {
	cfg   *config.Config
	local map[string]time.Time
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewCacheService(cfg *config.Config) *CacheService {
	ttl := 24 * time.Hour
	if cfg != nil && cfg.App.CacheTTL > 0 {
		ttl = cfg.App.CacheTTL
	}
	return &CacheService{
		cfg:   cfg,
		local: make(map[string]time.Time),
		ttl:   ttl,
	}
}

func (s *CacheService) IsCached(link string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expiredAt, exists := s.local[link]
	if !exists {
		return false
	}
	return time.Now().Before(expiredAt)
}

func (s *CacheService) Add(link string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.local[link] = time.Now().Add(s.ttl)
}

func (s *CacheService) Filter(ctx context.Context, links []string) ([]string, []string) {
	var newLinks []string
	var cachedLinks []string

	for _, link := range links {
		if s.IsCached(link) {
			cachedLinks = append(cachedLinks, link)
		} else {
			newLinks = append(newLinks, link)
			s.Add(link)
		}
	}

	return newLinks, cachedLinks
}

func (s *CacheService) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for link, expiredAt := range s.local {
		if now.After(expiredAt) {
			delete(s.local, link)
		}
	}
}
