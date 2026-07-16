//go:build integration

package database_test

import (
	"context"
	"testing"
	"time"

	"backend/src/database"
	"backend/src/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func cleanupSessions(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE sessions CASCADE")
	require.NoError(t, err)
	_, err = testPool.Exec(context.Background(), "TRUNCATE TABLE members CASCADE")
	require.NoError(t, err)
}

func createTestMemberForSession(t *testing.T) domain.Member {
	t.Helper()
	repo := database.NewMemberRepositoryPGX(testPool)
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	m := domain.Member{
		Email:    "session-test@example.com",
		Password: string(hash),
		Name:     "Session Test",
	}
	err = repo.Create(context.Background(), &m)
	require.NoError(t, err)
	return m
}

func TestSessionRepositoryPGX_Create(t *testing.T) {
	defer cleanupSessions(t)
	repo := database.NewSessionRepositoryPGX(testPool)
	member := createTestMemberForSession(t)

	session := domain.Session{MemberID: member.ID}
	err := repo.Create(context.Background(), &session)
	assert.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.NotEmpty(t, session.SessionKey)
	assert.False(t, session.CreatedAt.IsZero())
	assert.False(t, session.ExpiresAt.IsZero())
	assert.True(t, session.ExpiresAt.After(time.Now()))
}

func TestSessionRepositoryPGX_GetByKey(t *testing.T) {
	defer cleanupSessions(t)
	repo := database.NewSessionRepositoryPGX(testPool)
	member := createTestMemberForSession(t)

	created := domain.Session{MemberID: member.ID}
	err := repo.Create(context.Background(), &created)
	require.NoError(t, err)

	t.Run("existing key", func(t *testing.T) {
		got, err := repo.GetByKey(context.Background(), created.SessionKey)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, member.ID, got.MemberID)
	})

	t.Run("non-existent key", func(t *testing.T) {
		got, err := repo.GetByKey(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestSessionRepositoryPGX_Delete(t *testing.T) {
	defer cleanupSessions(t)
	repo := database.NewSessionRepositoryPGX(testPool)
	member := createTestMemberForSession(t)

	created := domain.Session{MemberID: member.ID}
	err := repo.Create(context.Background(), &created)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), created.SessionKey)
	assert.NoError(t, err)

	got, err := repo.GetByKey(context.Background(), created.SessionKey)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestSessionRepositoryPGX_DeleteByMemberID(t *testing.T) {
	defer cleanupSessions(t)
	repo := database.NewSessionRepositoryPGX(testPool)
	member := createTestMemberForSession(t)

	for i := 0; i < 3; i++ {
		session := domain.Session{MemberID: member.ID}
		err := repo.Create(context.Background(), &session)
		require.NoError(t, err)
	}

	err := repo.DeleteByMemberID(context.Background(), member.ID)
	assert.NoError(t, err)
}

func TestSessionRepositoryPGX_GetByKey_Expired(t *testing.T) {
	defer cleanupSessions(t)
	repo := database.NewSessionRepositoryPGX(testPool)
	member := createTestMemberForSession(t)

	created := domain.Session{MemberID: member.ID}
	err := repo.Create(context.Background(), &created)
	require.NoError(t, err)

	_, err = testPool.Exec(context.Background(),
		`UPDATE sessions SET expires_at = $1 WHERE id = $2`,
		time.Now().Add(-1*time.Hour), created.ID)
	require.NoError(t, err)

	got, err := repo.GetByKey(context.Background(), created.SessionKey)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.True(t, got.ExpiresAt.Before(time.Now()))
}
