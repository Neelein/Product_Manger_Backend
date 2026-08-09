package session

import (
	"time"
)

type Cache = SessionCache

func NewCache(ttl time.Duration) *Cache { return NewSessionCache(ttl) }
