package lru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstructLRUCache(t *testing.T) {
	cache := NewLRUCache(5)
	assert.NotNil(t, cache)
	assert.Equal(t, 5, cache.Capacity())
	assert.Equal(t, 0, cache.Len())
}

func TestSet(t *testing.T) {
	cache := NewLRUCache(5)
	status := cache.Set("key1", "value1")
	assert.Equal(t, "added", status)
	assert.Equal(t, 1, cache.Len())
}

func TestSetOverride(t *testing.T) {
	cache := NewLRUCache(5)
	cache.Set("key1", "value1")
	status := cache.Set("key1", "value1_updated")
	assert.Equal(t, "updated", status)
	assert.Equal(t, 1, cache.Len())

	// Check if the value was updated
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1_updated", value)
}

func TestSetWithTTL(t *testing.T) {
	cache := NewLRUCache(5)

	status := cache.SetWithTTL("key2", "value1", 100*time.Millisecond) // With expiration
	assert.Equal(t, "added", status)
	assert.Equal(t, 1, cache.Len())

	// Check if the item with expiration is retrievable
	value, found := cache.Get("key2")
	assert.True(t, found)
	assert.Equal(t, "value1", value)
}

func TestSetWithTTLIfExpired(t *testing.T) {
	cache := NewLRUCache(5)

	status := cache.Set("key1", "value1")
	assert.Equal(t, "added", status)
	cache.SetWithTTL("key3", "value1", 100*time.Millisecond) // With expiration
	assert.Equal(t, 2, cache.Len())
	status = cache.SetWithTTL("key3", "value2", 0) // Update with current time, should make it expire
	assert.Equal(t, "expired", status)
	assert.Equal(t, 1, cache.Len())
}

func TestSetAfterSetWithTTL(t *testing.T) {
	cache := NewLRUCache(5)

	cache.SetWithTTL("key1", "value1", 10*time.Millisecond)
	cache.SetWithTTL("key2", "value2", 10*time.Millisecond)
	assert.Equal(t, 2, cache.Len())
	time.Sleep(11 * time.Millisecond)   // Wait for the items to expire
	cache.Set("key1", "value1_updated") // Set without expiration
	assert.Equal(t, 2, cache.Len())     // We haven't accessed key2, so it should not be removed yet
	value, found := cache.Get("key2")   // Should not be found, as it has expired
	assert.False(t, found)
	assert.Nil(t, value)
	assert.Equal(t, 1, cache.Len()) // Only key1 should be present now
}

func TestLengthNotSurpassCapacity(t *testing.T) {
	cases := []struct {
		key   string
		value string
		want  int
	}{
		{"key1", "value1", 1},
		{"key2", "value2", 2},
		{"key3", "value3", 3},
		{"key4", "value4", 4},
		{"key5", "value5", 5},
		{"key6", "value6", 5}, // This should not increase the length beyond capacity
	}

	cache := NewLRUCache(5)
	assert.Equal(t, 5, cache.Capacity())
	assert.Equal(t, 0, cache.Len())

	for i, c := range cases {
		cache.Set(c.key, c.value)
		assert.Equal(t, c.want, cache.Len(), "case %d: expected length %d, got %d", i+1, c.want, cache.Len())
	}

	cache2 := NewLRUCache(1)
	assert.NotNil(t, cache2)

	cache2.Set("key1", "value1")
	assert.Equal(t, 1, cache2.Len())

	cache2.Set("key2", "value2")
	assert.Equal(t, 1, cache2.Len())
}

func TestBasicGet(t *testing.T) {
	cache := NewLRUCache(5)
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)
}

func TestOverrideGet(t *testing.T) {
	cache := NewLRUCache(5)
	cache.Set("key1", "value1")
	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	cache.Set("key1", "value1_updated")
	value, found = cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1_updated", value)
}

func TestGetNonExistentKey(t *testing.T) {
	cache := NewLRUCache(5)
	cache.Set("key1", "value1")
	value, found := cache.Get("non_existent_key")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestGetAfterEjectingOldest(t *testing.T) {
	cache := NewLRUCache(2)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	value, found := cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, value)

	value, found = cache.Get("key2")
	assert.True(t, found)
	assert.Equal(t, "value2", value)

	value, found = cache.Get("key3")
	assert.True(t, found)
	assert.Equal(t, "value3", value)
}

func TestGetAfterExpiration(t *testing.T) {
	cache := NewLRUCache(5)
	cache.Set("key1", "value1")
	cache.SetWithTTL("key2", "value2", 10*time.Millisecond)
	cache.SetWithTTL("key3", "value3", 15*time.Millisecond)

	time.Sleep(11 * time.Millisecond) // Wait for the items to expire

	value, found := cache.Get("key1")
	assert.True(t, found)
	assert.Equal(t, "value1", value)

	value, found = cache.Get("key2")
	assert.False(t, found)
	assert.Nil(t, value)

	value, found = cache.Get("key3")
	assert.True(t, found)
	assert.Equal(t, "value3", value)
}

func TestDeleteOldestAfterGet(t *testing.T) {
	cache := NewLRUCache(2)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Get("key1") // Access key1 to make it most recently used
	cache.Set("key3", "value3")

	value, found := cache.Get("key2") // key2 should be the least recently used and thus removed
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestDeleteOldestAfterSet(t *testing.T) {
	cache := NewLRUCache(2)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key1", "value1_updated") // Access key1 to make it most recently used
	cache.Set("key3", "value3")

	value, found := cache.Get("key2") // key2 should be the least recently used and thus removed
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestRemove(t *testing.T) {
	cache := NewLRUCache(5)
	cache.Set("key1", "value1")
	cache.Remove("key1")
	value, found := cache.Get("key1")
	assert.False(t, found)
	assert.Nil(t, value)
}
