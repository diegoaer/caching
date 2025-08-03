package lru

import (
	"fmt"
	"time"
)

// ObservableCacheItem represents an item in the observable cache state.
// This is intended to be used for demo purposes, where we want to return the state of the cache as a JSON object.
type ObservableCacheItem struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"` // Value is stored as a string for JSON serialization
	ExpiresAt time.Time `json:"expires_at"`
	Prev      string    `json:"prev"`
	Next      string    `json:"next"`
}

type ObservableCache struct {
	Cache *SafeLRUCache // The underlying SafeLRUCache
}

type ObservableCacheState struct {
	Capacity int                   `json:"capacity"`
	Items    []ObservableCacheItem `json:"items"`
}

func NewObservableCache(capacity int) *ObservableCache {
	cache := NewSafeLRUCache(capacity)
	return &ObservableCache{
		Cache: cache,
	}
}

func (observable *ObservableCache) State() ObservableCacheState {
	observable.Cache.mutex.Lock()
	defer observable.Cache.mutex.Unlock()

	lru, ok := observable.Cache.cache.(*LRUCache)
	if !ok {
		return ObservableCacheState{}
	}

	// This is not performant, but it is a simple way to get the state of the cache.
	// In a real application, observability in cache is often done with metrics,
	// but here we want to return the state as a JSON object.
	// If you must use this in production, consider implementing a more efficient way to get the state.
	items := make([]ObservableCacheItem, 0, len(lru.items))
	prev := ""
	for e := lru.usageOrder.Front(); e != nil; e = e.Next() {
		ent := e.Value.(*entry)
		next := ""
		if e.Next() != nil {
			next = e.Next().Value.(*entry).key
		}
		items = append(items, ObservableCacheItem{
			Key:       ent.key,
			Value:     fmt.Sprintf("%v", ent.value), // Convert value to string for JSON serialization
			ExpiresAt: ent.expiresAt,
			Prev:      prev,
			Next:      next,
		})
		prev = ent.key
	}

	return ObservableCacheState{
		Capacity: lru.capacity,
		Items:    items,
	}
}
