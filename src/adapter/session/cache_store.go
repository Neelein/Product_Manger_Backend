package session

import (
	"context"
	"sync"
	"time"

	domain "backend/src/domain/model"

	"github.com/google/uuid"
)

type SessionCache struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
	ttl      time.Duration
	stopCh   chan struct{}
}

func NewSessionCache(ttl time.Duration) *SessionCache {
	c := &SessionCache{
		sessions: make(map[string]*domain.Session),
		ttl:      ttl,
		stopCh:   make(chan struct{}),
	}
	go c.cleanupLoop()
	return c
}

func (c *SessionCache) Stop() {
	close(c.stopCh)
}

func (c *SessionCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCh:
			return
		}
	}
}

func (c *SessionCache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, s := range c.sessions {
		if s.ExpiresAt.Before(now) {
			delete(c.sessions, key)
		}
	}
}

func (c *SessionCache) Create(_ context.Context, session *domain.Session) error {
	session.ID = uuid.New().String()
	session.SessionKey = uuid.New().String()
	session.CreatedAt = time.Now()
	session.ExpiresAt = time.Now().Add(c.ttl)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessions[session.SessionKey] = session
	return nil
}

func (c *SessionCache) GetByKey(_ context.Context, sessionKey string) (*domain.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	s, ok := c.sessions[sessionKey]
	if !ok {
		return nil, nil
	}
	if s.ExpiresAt.Before(time.Now()) {
		delete(c.sessions, sessionKey)
		return nil, nil
	}
	s.ExpiresAt = time.Now().Add(c.ttl)
	cp := *s
	return &cp, nil
}

func (c *SessionCache) Delete(_ context.Context, sessionKey string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.sessions, sessionKey)
	return nil
}

func (c *SessionCache) DeleteByMemberID(_ context.Context, memberID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, s := range c.sessions {
		if s.MemberID == memberID {
			delete(c.sessions, key)
		}
	}
	return nil
}
