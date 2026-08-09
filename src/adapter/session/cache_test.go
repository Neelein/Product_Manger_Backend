package session

import (
	"context"
	"testing"
	"time"

	"backend/src/domain/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCache() *SessionCache {
	return NewSessionCache(time.Hour)
}

func createTestSession(t *testing.T, cache *SessionCache) *model.Session {
	t.Helper()
	session := &model.Session{MemberID: "member-1"}
	require.NoError(t, cache.Create(context.Background(), session))
	return session
}

func TestSessionCache_Create_And_Get(t *testing.T) {
	cache := newTestCache()
	defer cache.Stop()

	session := createTestSession(t, cache)

	got, err := cache.GetByKey(context.Background(), session.SessionKey)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "member-1", got.MemberID)
}

func TestSessionCache_GetByKey_RefreshesExpiry(t *testing.T) {
	cache := newTestCache()
	defer cache.Stop()

	session := createTestSession(t, cache)

	got, err := cache.GetByKey(context.Background(), session.SessionKey)
	require.NoError(t, err)
	require.NotNil(t, got)

	got.ExpiresAt = time.Now().Add(-30 * time.Minute)

	refreshed, err := cache.GetByKey(context.Background(), session.SessionKey)
	require.NoError(t, err)
	require.NotNil(t, refreshed)
	assert.WithinDuration(t, time.Now().Add(time.Hour), refreshed.ExpiresAt, time.Minute)
}

func TestSessionCache_GetByKey_Expired(t *testing.T) {
	cache := newTestCache()
	defer cache.Stop()

	session := createTestSession(t, cache)

	session.ExpiresAt = time.Now().Add(-30 * time.Minute)

	expired, err := cache.GetByKey(context.Background(), session.SessionKey)
	require.NoError(t, err)
	assert.Nil(t, expired)
}

func TestSessionCache_GetByKey_NotFound(t *testing.T) {
	cache := newTestCache()
	defer cache.Stop()

	got, err := cache.GetByKey(context.Background(), "non-existent-key")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestSessionCache_Delete(t *testing.T) {
	cache := newTestCache()
	defer cache.Stop()

	session := createTestSession(t, cache)

	require.NoError(t, cache.Delete(context.Background(), session.SessionKey))

	got, err := cache.GetByKey(context.Background(), session.SessionKey)
	require.NoError(t, err)
	assert.Nil(t, got)
}
