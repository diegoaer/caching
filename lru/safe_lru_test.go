package lru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeLRUCache struct {
	getCalled        bool
	setCalled        bool
	setWithTTLCalled bool
	removeCalled     bool
	lenCalled        bool
	capacityCalled   bool
}

func (f *fakeLRUCache) Get(key string) (any, bool) {
	f.getCalled = true
	return nil, false
}

func (f *fakeLRUCache) Set(key string, value any) (status string) {
	f.setCalled = true
	return "added"
}

func (f *fakeLRUCache) SetWithTTL(key string, value any, ttl time.Duration) (status string) {
	f.setWithTTLCalled = true
	return "added"
}

func (f *fakeLRUCache) Remove(key string) {
	f.removeCalled = true
}

func (f *fakeLRUCache) Len() int {
	f.lenCalled = true
	return 0
}

func (f *fakeLRUCache) Capacity() int {
	f.capacityCalled = true
	return 0
}

func TestConstructSafeLRUCache(t *testing.T) {
	cache := NewSafeLRUCache(5)
	assert.NotNil(t, cache)
	assert.Equal(t, 5, cache.Capacity())
	assert.Equal(t, 0, cache.Len())
}

func TestConstructSafeLRUCacheFrom(t *testing.T) {
	original := NewLRUCache(5)
	original.Set("key1", "value1")
	original.Set("key2", "value2")

	cache := NewSafeLRUCacheFrom(original)
	assert.NotNil(t, cache)
	assert.Equal(t, 5, cache.Capacity())
	assert.Equal(t, 2, cache.Len())
}

func TestCacheGet(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	value, found := safeCache.Get("testKey")
	assert.False(t, found)
	assert.Nil(t, value)
	assert.True(t, fake.getCalled, "Get should call the underlying cache's Get method")
}

func TestCacheSet(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	safeCache.Set("testKey", "testValue")
	assert.True(t, fake.setCalled, "Set should call the underlying cache's Set method")
}

func TestCacheSetWithTTL(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	safeCache.SetWithTTL("testKey", "testValue", time.Minute)
	assert.True(t, fake.setWithTTLCalled, "SetWithTTL should call the underlying cache's SetWithTTL method")
}

func TestCacheRemove(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	safeCache.Remove("testKey")
	assert.True(t, fake.removeCalled, "Remove should call the underlying cache's Remove method")
}

func TestCacheLen(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	length := safeCache.Len()
	assert.Equal(t, 0, length)
	assert.True(t, fake.lenCalled, "Len should call the underlying cache's Len method")
}

func TestCacheCapacity(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	capacity := safeCache.Capacity()
	assert.Equal(t, 0, capacity)
	assert.True(t, fake.capacityCalled, "Capacity should call the underlying cache's Capacity method")
}

func TestCacheUnsafePeek(t *testing.T) {
	safeCache := NewSafeLRUCache(5)
	safeCache.Set("testKey", "testValue")

	value, found := safeCache.UnsafePeek("testKey")
	assert.True(t, found)
	assert.Equal(t, "testValue", value)
}

func TestCacheUnsafePeekNonExistent(t *testing.T) {
	safeCache := NewSafeLRUCache(5)
	value, found := safeCache.UnsafePeek("nonExistentKey")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestCacheUnsafePeekExpired(t *testing.T) {
	safeCache := NewSafeLRUCache(5)
	safeCache.SetWithTTL("testKey", "testValue", 50*time.Millisecond) // Set an expired item
	time.Sleep(60 * time.Millisecond)                                 // Wait for the item to expire
	value, found := safeCache.UnsafePeek("testKey")

	assert.True(t, found, "UnsafePeek should return the value even if it is expired")
	assert.Equal(t, "testValue", value)
}

func TestCacheUnsafePeekNotMoveItems(t *testing.T) {
	safeCache := NewSafeLRUCache(2)
	safeCache.Set("testKey", "testValue")
	safeCache.Set("testKey2", "testValue2")

	safeCache.UnsafePeek("testKey")         // This should not update the usage order
	safeCache.Set("testKey3", "testValue3") // This should evict "testKey" since it was not accessed
	value, found := safeCache.Get("testKey")
	assert.False(t, found)
	assert.Nil(t, value)

	// Check that Get still works as expected (updates usage order)
	safeCache.Get("testKey2")               // Access "testKey2" to keep it in the cache
	safeCache.Set("testKey4", "testValue4") // This should evict "testKey3"
	value, found = safeCache.Get("testKey3")
	assert.False(t, found)
	assert.Nil(t, value)

	value, found = safeCache.Get("testKey2")
	assert.True(t, found)
	assert.Equal(t, "testValue2", value)
}

func TestUnsafePeekPanicOnNonLRUCache(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	assert.Panics(t, func() {
		safeCache.UnsafePeek("testKey")
	}, "UnsafePeek should panic if the underlying cache is not an LRUCache")
}

func TestCacheUnsafeLen(t *testing.T) {
	fake := &fakeLRUCache{}
	safeCache := NewSafeLRUCacheFrom(fake)

	length := safeCache.UnsafeLen()
	assert.Equal(t, 0, length)
	assert.True(t, fake.lenCalled, "UnsafeLen should call the underlying cache's Len method")
}
