package tenant

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const maxCacheEntries = 10_000

type CachedResolver struct {
	next        Resolver
	ttl         time.Duration
	negativeTTL time.Duration
	now         func() time.Time
	mu          sync.RWMutex
	entries     map[string]cacheEntry
}

type cacheEntry struct {
	tenant    Tenant
	valid     bool
	expiresAt time.Time
}

func NewCachedResolver(next Resolver, ttl, negativeTTL time.Duration) *CachedResolver {
	return &CachedResolver{
		next:        next,
		ttl:         ttl,
		negativeTTL: negativeTTL,
		now:         time.Now,
		entries:     make(map[string]cacheEntry),
	}
}

func (r *CachedResolver) Resolve(ctx context.Context, slug, apiKey string) (Tenant, error) {
	if slug == "" || apiKey == "" {
		return Tenant{}, ErrUnauthenticated
	}

	key := cacheKey(slug, apiKey)
	if entry, ok := r.get(key); ok {
		if !entry.valid {
			return Tenant{}, ErrUnauthenticated
		}
		return entry.tenant, nil
	}

	resolved, err := r.next.Resolve(ctx, slug, apiKey)
	if err == nil {
		if r.ttl > 0 {
			r.put(key, cacheEntry{tenant: resolved, valid: true, expiresAt: r.now().Add(r.ttl)})
		}
		return resolved, nil
	}
	if errors.Is(err, ErrUnauthenticated) && r.negativeTTL > 0 {
		r.put(key, cacheEntry{expiresAt: r.now().Add(r.negativeTTL)})
	}
	return Tenant{}, err
}

func (r *CachedResolver) get(key string) (cacheEntry, bool) {
	now := r.now()
	r.mu.RLock()
	entry, ok := r.entries[key]
	r.mu.RUnlock()
	if !ok || !now.Before(entry.expiresAt) {
		if ok {
			r.mu.Lock()
			if current, found := r.entries[key]; found && !now.Before(current.expiresAt) {
				delete(r.entries, key)
			}
			r.mu.Unlock()
		}
		return cacheEntry{}, false
	}
	return entry, true
}

func (r *CachedResolver) put(key string, entry cacheEntry) {
	r.mu.Lock()
	if _, exists := r.entries[key]; !exists && len(r.entries) >= maxCacheEntries {
		now := r.now()
		for cachedKey, cachedEntry := range r.entries {
			if !now.Before(cachedEntry.expiresAt) {
				delete(r.entries, cachedKey)
			}
		}
		if len(r.entries) >= maxCacheEntries {
			for cachedKey := range r.entries {
				delete(r.entries, cachedKey)
				break
			}
		}
	}
	r.entries[key] = entry
	r.mu.Unlock()
}

func cacheKey(slug, apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return slug + ":" + hex.EncodeToString(sum[:])
}
