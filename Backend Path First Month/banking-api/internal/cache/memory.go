package cache

import (
	"context"
	"sync"
	"time"

	"banking-api/internal/domain"
)

type memoryEntry struct {
	value     string
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]memoryEntry
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{items: make(map[string]memoryEntry)}
	go c.cleanupLoop()
	return c
}

func (c *MemoryCache) Get(_ context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return "", domain.ErrCacheMiss
	}
	return entry.value, nil
}

func (c *MemoryCache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = memoryEntry{value: value, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for key, entry := range c.items {
			if now.After(entry.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
