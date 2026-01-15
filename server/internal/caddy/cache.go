package caddy

import (
	"sync"
	"time"

	"github.com/sitepod/sitepod/internal/storage"
)

// refCache caches ref data for projects
type refCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
	ttl   time.Duration
}

type cacheItem struct {
	data      *storage.RefData
	expiresAt time.Time
}

func newRefCache(ttl time.Duration) *refCache {
	return &refCache{
		items: make(map[string]*cacheItem),
		ttl:   ttl,
	}
}

func (c *refCache) Get(key string) (*storage.RefData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok || time.Now().After(item.expiresAt) {
		return nil, false
	}
	return item.data, true
}

func (c *refCache) Set(key string, data *storage.RefData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *refCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// routingCache caches domain routing index
type routingCache struct {
	mu    sync.RWMutex
	index *RoutingIndex
	ttl   time.Duration
	timer time.Time
}

// RoutingIndex maps domains to projects with slug prefixes
type RoutingIndex struct {
	Entries   []RoutingEntry `json:"entries"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// RoutingEntry represents a domain+slug to project mapping
type RoutingEntry struct {
	Domain    string `json:"domain"`
	Slug      string `json:"slug"`
	ProjectID string `json:"project_id"`
	Project   string `json:"project"`
	Env       string `json:"env,omitempty"`
}

func newRoutingCache(ttl time.Duration) *routingCache {
	return &routingCache{ttl: ttl}
}

func (c *routingCache) Get() (*RoutingIndex, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.index == nil || time.Now().After(c.timer) {
		return nil, false
	}
	return c.index, true
}

func (c *routingCache) Set(index *RoutingIndex) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.index = index
	c.timer = time.Now().Add(c.ttl)
}
