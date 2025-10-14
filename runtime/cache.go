package runtime

import (
	"sync"
	"time"
)

// CacheEntry represents a cached template with metadata
type CacheEntry struct {
	Template    *Template
	LoadedAt    time.Time
	ExpiresAt   time.Time
	Dependencies map[string]time.Time
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false // No expiration set
	}
	return time.Now().After(e.ExpiresAt)
}

// IsValid checks if the cache entry and its dependencies are still valid
func (e *CacheEntry) IsValid(loader Loader) bool {
	if e.IsExpired() {
		return false
	}

	// Check if any dependencies have been modified
	for depPath, modTime := range e.Dependencies {
		currentModTime, err := getModTime(loader, depPath)
		if err != nil || currentModTime.After(modTime) {
			return false
		}
	}

	return true
}

// TemplateCache provides thread-safe template caching with TTL support
type TemplateCache struct {
	entries map[string]*CacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
 maxSize int
}

// NewTemplateCache creates a new template cache
func NewTemplateCache(ttl time.Duration, maxSize int) *TemplateCache {
	return &TemplateCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get retrieves a template from the cache
func (c *TemplateCache) Get(name string) (*Template, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, ok := c.entries[name]
	if !ok {
		return nil, false
	}

	// Check if entry is still valid
	if !entry.IsValid(nil) { // loader not needed for basic validation
		delete(c.entries, name)
		return nil, false
	}

	return entry.Template, true
}

// Set stores a template in the cache
func (c *TemplateCache) Set(name string, template *Template, dependencies map[string]time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Evict entries if cache is full
	if len(c.entries) >= c.maxSize {
		c.evictOldest()
	}

	now := time.Now()
	var expiresAt time.Time
	if c.ttl > 0 {
		expiresAt = now.Add(c.ttl)
	}

	entry := &CacheEntry{
		Template:     template,
		LoadedAt:     now,
		ExpiresAt:    expiresAt,
		Dependencies: dependencies,
	}

	c.entries[name] = entry
}

// Delete removes a template from the cache
func (c *TemplateCache) Delete(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, name)
}

// Clear removes all entries from the cache
func (c *TemplateCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// Invalidate removes templates that depend on the given file
func (c *TemplateCache) Invalidate(filePath string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Find and remove entries that depend on the given file
	for name, entry := range c.entries {
		if _, depends := entry.Dependencies[filePath]; depends {
			delete(c.entries, name)
		}
	}
}

// Size returns the current number of cached entries
func (c *TemplateCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.entries)
}

// Clean removes expired entries from the cache
func (c *TemplateCache) Clean(loader Loader) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for name, entry := range c.entries {
		if !entry.IsValid(loader) {
			delete(c.entries, name)
		}
	}
}

// evictOldest removes the oldest entry from the cache
func (c *TemplateCache) evictOldest() {
	var oldestName string
	var oldestTime time.Time

	for name, entry := range c.entries {
		if oldestName == "" || entry.LoadedAt.Before(oldestTime) {
			oldestName = name
			oldestTime = entry.LoadedAt
		}
	}

	if oldestName != "" {
		delete(c.entries, oldestName)
	}
}

// getModTime gets the modification time of a template file
// This is a placeholder - in a real implementation, this would check file modification times
func getModTime(loader Loader, path string) (time.Time, error) {
	// For now, just return the current time
	// In a real implementation, this would check the actual file modification time
	return time.Now(), nil
}