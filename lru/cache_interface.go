package lru

import (
	"time"
)

type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any) (status string)
	SetWithTTL(key string, value any, ttl time.Duration) (status string)
	Remove(key string)
	Len() int
	Capacity() int
}
