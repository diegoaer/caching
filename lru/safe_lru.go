package lru

import (
	"sync"
	"time"
)

type SafeLRUCache struct {
	cache Cache      // The underlying LRU cache
	mutex sync.Mutex // Mutex to ensure thread safety
}

var _ Cache = (*SafeLRUCache)(nil) // Ensure SafeLRUCache implements the Cache interface

func NewSafeLRUCache(capacity int) *SafeLRUCache {
	cache := NewLRUCache(capacity)
	cache.name = metricCacheTypeSafeLRU // Set a different name for the safe cache
	return &SafeLRUCache{
		cache: cache,
	}
}

// NewSafeLRUCacheFrom creates a SafeLRUCache from an existing LRUCache.
// This is useful for wrapping an existing cache without losing its state.
// The new SafeLRUCache will be thread-safe.
// It does not copy the items from the original cache, so it should be used with caution.
// Used in tests
func NewSafeLRUCacheFrom(cache Cache) *SafeLRUCache {
	return &SafeLRUCache{
		cache: cache,
	}
}

// Get retrieves an item from the cache by its key.
// It returns the value and a boolean indicating whether the item was found.
// If the ttl has expired, the item will be removed and not found.
// It is thread-safe.
func (safeCache *SafeLRUCache) Get(key string) (value any, found bool) {
	safeCache.mutex.Lock()
	defer safeCache.mutex.Unlock()

	return safeCache.cache.Get(key)
}

// Set adds or updates an item in the cache with no expiration.
// The item will not expire unless explicitly removed.
// If the key already exists, both its value and expiration will be overridden.
// It is thread-safe.
func (safeCache *SafeLRUCache) Set(key string, value any) (status string) {
	safeCache.mutex.Lock()
	defer safeCache.mutex.Unlock()

	return safeCache.cache.Set(key, value)
}

// SetWithTTL adds or updates an item in the cache with a specified expiration time. (TTL: time to live).
// It calls the internal set method with the expiration time.
// It is thread-safe.
func (safeCache *SafeLRUCache) SetWithTTL(key string, value any, ttl time.Duration) (status string) {
	safeCache.mutex.Lock()
	defer safeCache.mutex.Unlock()

	return safeCache.cache.SetWithTTL(key, value, ttl)
}

// Remove deletes an item from the cache by key.
// If the item does not exist, it does nothing.
// It is thread-safe.
func (safeCache *SafeLRUCache) Remove(key string) {
	safeCache.mutex.Lock()
	defer safeCache.mutex.Unlock()

	safeCache.cache.Remove(key)
}

// Capacity returns the maximum number of items that can be stored in the cache.
// This value is fixed at initialization and does not require locking.
func (safeCache *SafeLRUCache) Capacity() int {
	return safeCache.cache.Capacity()
}

// Len returns the number of items currently in the cache.
// It is thread-safe.
func (safeCache *SafeLRUCache) Len() int {
	safeCache.mutex.Lock()
	defer safeCache.mutex.Unlock()

	return safeCache.cache.Len()
}

// UnsafePeek retrieves the value for a key without updates to its usage order nor expiration.
// This method is not thread-safe and may return expired items.
// It assumes the underlying cache is an LRUCache, if not, it will panic.
// It returns the value and a boolean indicating whether the item was found.
func (safeCache *SafeLRUCache) UnsafePeek(key string) (value any, found bool) {
	if lru, ok := safeCache.cache.(*LRUCache); ok {
		if elem, found := lru.items[key]; found {
			return elem.Value.(*entry).value, true
		}
	} else {
		panic("UnsafePeek can only be used with LRUCache")
	}
	return nil, false // Item not found, or not an LRUCache
}

// UnsafeLen returns the number of items in the cache without locking.
// This is not thread-safe and may return an inaccurate length.
// It is intended for use in scenarios where the returned length doesn't need to be 100% accurate.
func (safeCache *SafeLRUCache) UnsafeLen() int {
	return safeCache.cache.Len()
}
