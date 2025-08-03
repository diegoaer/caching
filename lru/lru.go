package lru

import (
	"container/list"
	"time"
)

const (
	setStatusAdded   = "added"
	setStatusUpdated = "updated"
	setStatusExpired = "expired"
)

type entry struct {
	key       string    // The key for the cached item
	value     any       // The value for the cached item
	expiresAt time.Time // Optional expiration time for the cached item
}

// hasExpired checks if the entry has expired based on its expiration time.
func (e *entry) hasExpired() bool {
	return hasExpired(e.expiresAt)
}

// hasExpired checks if an expiration date has expired.
// If the expiration date is zero, it means the item does not expire.
// If the expiration date is in the past, it means the item has expired.
func hasExpired(expiration time.Time) bool {
	return !expiration.IsZero() && expiration.Before(time.Now())
}

type LRUCache struct {
	capacity   int                      // The capacity of this cache, when full, the least recently used item will be removed
	items      map[string]*list.Element // Provides easy access to the cached elements
	usageOrder *list.List               // Holds the cached elements in order
	name       string                   // Name of the cache, used for metrics
}

var _ Cache = (*LRUCache)(nil) // Ensure LRUCache implements the Cache interface

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity:   capacity,
		items:      make(map[string]*list.Element),
		usageOrder: list.New(),
		name:       metricCacheTypeLRU, // Default name for the cache
	}
}

// Get retrieves an item from the cache by its key.
// It returns the value and a boolean indicating whether the item was found.
// If the ttl has expired, the item will be removed and not found.
func (cache *LRUCache) Get(key string) (value any, found bool) {
	if elem, found := cache.items[key]; found {
		if elem.Value.(*entry).hasExpired() {
			cache.remove(key, metricReasonExpired) // Remove the item if it has expired
			return nil, false                      // Item expired and removed
		}

		// Move the accessed item to the front of the usage order list
		cache.usageOrder.MoveToFront(elem)

		cacheHits.WithLabelValues(cache.name, metricOpGet).Inc() // Increment cache hit metric
		return elem.Value.(*entry).value, true
	}
	cacheMisses.WithLabelValues(cache.name, metricOpGet).Inc() // Increment cache miss metric
	return nil, false                                          // Item not found
}

// update updates the value and expiration time of an existing item in the cache.
// It moves the item to the front of the usage order list to mark it as recently used.
func (cache *LRUCache) update(element *list.Element, value any, expiration time.Time) {
	// Update the value and move it to the front of the usage order list
	element.Value.(*entry).value = value
	element.Value.(*entry).expiresAt = expiration
	cache.usageOrder.MoveToFront(element)

	cacheHits.WithLabelValues(cache.name, metricOpSet).Inc() // Increment cache hit metric
}

// checkCapacity checks if the cache has reached its capacity.
// If it has, it removes the least recently used item.
// This method is called before adding a new item to ensure the cache does not exceed its capacity.
func (cache *LRUCache) checkCapacity() {
	if cache.usageOrder.Len() >= cache.capacity {
		// Remove the least recently used item
		leastRecentlyUsed := cache.usageOrder.Back()
		if leastRecentlyUsed != nil {
			cache.remove(leastRecentlyUsed.Value.(*entry).key, metricReasonEvicted)
		}
	}
}

// Set adds or updates an item in the cache.
// If the cache is full, it removes the least recently used item.
// If the item already exists, it updates the value and expiration time.
// If the expiration time is in the past, the item will be removed immediately.
// If the expiration time is zero, the item will not expire.
func (cache *LRUCache) set(key string, value any, expiration time.Time) (status string) {
	if elem, found := cache.items[key]; found {
		cache.update(elem, value, expiration) // Update existing item
		return setStatusUpdated
	} else {
		cache.checkCapacity() // Check capacity before adding a new item
		// Create a new entry and add it to the cache
		newEntry := &entry{key: key, value: value, expiresAt: expiration}
		newElem := cache.usageOrder.PushFront(newEntry)
		cache.items[key] = newElem

		cacheMisses.WithLabelValues(cache.name, metricOpSet).Inc()                               // Increment cache miss metric
		totalItems.WithLabelValues(cache.name, metricOpSet).Set(float64(cache.usageOrder.Len())) // Update total items metric
		return setStatusAdded
	}
}

// Set adds or updates an item in the cache with no expiration.
// The item will not expire unless explicitly removed.
// If the key already exists, both its value and expiration will be overridden.
func (cache *LRUCache) Set(key string, value any) (status string) {
	return cache.set(key, value, time.Time{}) // No expiration
}

// SetWithTTL adds or updates an item in the cache with a specified expiration time.
// It calls the internal set method with the expiration time.
func (cache *LRUCache) SetWithTTL(key string, value any, ttl time.Duration) (status string) {
	expiration := time.Now().Add(ttl)

	if !hasExpired(expiration) {
		status = cache.set(key, value, expiration)
	} else {
		cache.remove(key, metricReasonExpired) // Remove the item if it has expired
		status = setStatusExpired
	}

	expirationHistogram.WithLabelValues(cache.name).Observe(ttl.Seconds()) // Record the expiration duration in the histogram
	return status
}

// Remove deletes an item from the cache by key.
// If the item does not exist, it does nothing.
// It also updates the metrics for eviction and total items.
// The reason parameter is used to specify why the item is being removed (e.g., "manual", "expired", "evicted").
func (cache *LRUCache) remove(key string, reason string) {
	if elem, found := cache.items[key]; found {
		// Remove the item from the cache
		cache.usageOrder.Remove(elem)
		delete(cache.items, key)

		evictionCount.WithLabelValues(cache.name, metricOpRemove, reason).Inc()                     // Increment eviction metric
		totalItems.WithLabelValues(cache.name, metricOpRemove).Set(float64(cache.usageOrder.Len())) // Update total items metric
	}
}

// Remove deletes an item from the cache by key.
func (cache *LRUCache) Remove(key string) {
	cache.remove(key, metricReasonManual) // Default reason is "manual"
}

// Capacity returns the maximum number of items that can be stored in the cache.
func (cache *LRUCache) Capacity() int {
	return cache.capacity
}

// Len returns the number of items currently in the cache.
func (cache *LRUCache) Len() int {
	return cache.usageOrder.Len()
}
