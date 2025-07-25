package pokecache

import "time"

func NewCache(interval time.Duration) *Cache {
	c := new(Cache)

	c.entries = make(map[string]cacheEntry)
	c.interval = interval
	go c.reapLoop()

	return c
}

func (c *Cache) Add(key string, val []byte) {
	newCache := new(cacheEntry)

	newCache.createdAt = time.Now()
	newCache.val = val

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = *newCache
}

func (c *Cache) Get(key string) ([]byte, bool) {

	c.mu.Lock()

	entry, err := c.entries[key]

	c.mu.Unlock()

	if !err {
		return nil, false
	}

	return entry.val, true
}

func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()

			for key := range c.entries {
				now := time.Since(c.entries[key].createdAt)

				if now > c.interval {
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
		}
	}
}
