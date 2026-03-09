package cache

import "sync"

type RedirectCache struct {
	mu    sync.RWMutex
	items map[string]string
}

func New() *RedirectCache {
	return &RedirectCache{items: make(map[string]string)}
}

func (c *RedirectCache) Get(hash string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.items[hash]
	return v, ok
}

func (c *RedirectCache) Set(hash, url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[hash] = url
}

func (c *RedirectCache) Delete(hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, hash)
}

func (c *RedirectCache) BulkSet(entries map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range entries {
		c.items[k] = v
	}
}
