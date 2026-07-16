//go:build integration

package database_test

import (
	"context"
	"testing"

	"backend/src/database"
	"backend/src/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func cleanupMembers(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE members CASCADE")
	require.NoError(t, err)
}

func createTestMember(t *testing.T, repo *database.MemberRepositoryPGX) domain.Member {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	m := domain.Member{
		Email:    "test@example.com",
		Password: string(hash),
		Name:     "Test User",
	}
	err = repo.Create(context.Background(), &m)
	require.NoError(t, err)
	return m
}

func TestMemberRepositoryPGX_Create(t *testing.T) {
	defer cleanupMembers(t)
	repo := database.NewMemberRepositoryPGX(testPool)

	member := domain.Member{
		Email:    "new@example.com",
		Password: "hashedpassword",
		Name:     "New User",
	}

	err := repo.Create(context.Background(), &member)
	assert.NoError(t, err)
	assert.NotEmpty(t, member.ID)
	assert.Equal(t, "new@example.com", member.Email)
	assert.False(t, member.CreatedAt.IsZero())
	assert.False(t, member.UpdatedAt.IsZero())
}

func TestMemberRepositoryPGX_Create_DuplicateEmail(t *testing.T) {
	defer cleanupMembers(t)
	repo := database.NewMemberRepositoryPGX(testPool)

	member := domain.Member{
		Email:    "duplicate@example.com",
		Password: "hash1",
		Name:     "User 1",
	}
	err := repo.Create(context.Background(), &member)
	require.NoError(t, err)

	duplicate := domain.Member{
		Email:    "duplicate@example.com",
		Password: "hash2",
		Name:     "User 2",
	}
	err = repo.Create(context.Background(), &duplicate)
	assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
}

func TestMemberRepositoryPGX_GetByEmail(t *testing.T) {
	defer cleanupMembers(t)
	repo := database.NewMemberRepositoryPGX(testPool)

	t.Run("existing email", func(t *testing.T) {
		created := createTestMember(t, repo)
		got, err := repo.GetByEmail(context.Background(), created.Email)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "Test User", got.Name)
	})

	t.Run("non-existent email", func(t *testing.T) {
		got, err := repo.GetByEmail(context.Background(), "nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestMemberRepositoryPGX_Update(t *testing.T) {
	defer cleanupMembers(t)
	repo := database.NewMemberRepositoryPGX(testPool)

	created := createTestMember(t, repo)

	t.Run("update existing member", func(t *testing.T) {
		created.Email = "updated@example.com"
		created.Name = "Updated Name"

		err := repo.Update(context.Background(), &created)
		assert.NoError(t, err)

		got, err := repo.GetByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, "updated@example.com", got.Email)
		assert.Equal(t, "Updated Name", got.Name)
		assert.False(t, got.UpdatedAt.Equal(created.CreatedAt))
	})

	t.Run("update non-existent member", func(t *testing.T) {
		m := domain.Member{
			ID:    "00000000-0000-0000-0000-000000000000",
			Email: "nobody@example.com",
			Name:  "Nobody",
		}
		err := repo.Update(context.Background(), &m)
		assert.ErrorIs(t, err, domain.ErrMemberNotFound)
	})

	t.Run("update to duplicate email", func(t *testing.T) {
		other := domain.Member{
			Email:    "other@example.com",
			Password: "hash",
			Name:     "Other",
		}
		err := repo.Create(context.Background(), &other)
		require.NoError(t, err)

		created.Email = other.Email
		created.Name = "Original Name"
		err = repo.Update(context.Background(), &created)
		assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
	})
}

func TestMemberRepositoryPGX_GetByID(t *testing.T) {
	defer cleanupMembers(t)
	repo := database.NewMemberRepositoryPGX(testPool)

	t.Run("existing ID", func(t *testing.T) {
		created := createTestMember(t, repo)
		got, err := repo.GetByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, created.Email, got.Email)
	})

	t.Run("non-existent ID", func(t *testing.T) {
		got, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}