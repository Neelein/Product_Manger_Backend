//go:build integration

package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"backend/src/database"
	"backend/src/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cleanupCategories(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE categories CASCADE")
	require.NoError(t, err)
}

func TestCategoryRepositoryPGX_Create(t *testing.T) {
	defer cleanupCategories(t)
	repo := database.NewCategoryRepositoryPGX(testPool)

	t.Run("create category", func(t *testing.T) {
		c, err := repo.Create(context.Background(), "electronics")
		assert.NoError(t, err)
		assert.NotEmpty(t, c.ID)
		assert.Equal(t, "electronics", c.Name)
		assert.False(t, c.CreatedAt.IsZero())
		assert.False(t, c.UpdatedAt.IsZero())
	})

	t.Run("duplicate name", func(t *testing.T) {
		_, err := repo.Create(context.Background(), "electronics")
		assert.ErrorIs(t, err, domain.ErrCategoryNameExists)
	})
}

func TestCategoryRepositoryPGX_List(t *testing.T) {
	defer cleanupCategories(t)
	repo := database.NewCategoryRepositoryPGX(testPool)

	t.Run("empty repository", func(t *testing.T) {
		categories, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, categories)
	})

	t.Run("sorted by created_at desc", func(t *testing.T) {
		first, err := repo.Create(context.Background(), "first")
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond)
		second, err := repo.Create(context.Background(), "second")
		require.NoError(t, err)

		categories, err := repo.List(context.Background())
		assert.NoError(t, err)
		require.Len(t, categories, 2)
		assert.Equal(t, second.ID, categories[0].ID)
		assert.Equal(t, first.ID, categories[1].ID)
	})
}

func TestCategoryRepositoryPGX_Update(t *testing.T) {
	defer cleanupCategories(t)
	repo := database.NewCategoryRepositoryPGX(testPool)

	t.Run("update existing category", func(t *testing.T) {
		c, err := repo.Create(context.Background(), "old-name")
		require.NoError(t, err)

		updated, err := repo.Update(context.Background(), c.ID, "new-name")
		assert.NoError(t, err)
		assert.True(t, updated)

		categories, err := repo.List(context.Background())
		require.NoError(t, err)
		require.Len(t, categories, 1)
		assert.Equal(t, "new-name", categories[0].Name)
	})

	t.Run("update non-existent category", func(t *testing.T) {
		updated, err := repo.Update(context.Background(), "00000000-0000-0000-0000-000000000000", "ghost")
		assert.NoError(t, err)
		assert.False(t, updated)
	})

	t.Run("update to duplicate name", func(t *testing.T) {
		_, err := repo.Create(context.Background(), "dup-a")
		require.NoError(t, err)
		second, err := repo.Create(context.Background(), "dup-b")
		require.NoError(t, err)

		updated, err := repo.Update(context.Background(), second.ID, "dup-a")
		assert.ErrorIs(t, err, domain.ErrCategoryNameExists)
		assert.False(t, updated)
	})
}

func TestCategoryRepositoryPGX_Delete(t *testing.T) {
	defer cleanupCategories(t)
	repo := database.NewCategoryRepositoryPGX(testPool)

	t.Run("delete non-existent category", func(t *testing.T) {
		deleted, err := repo.Delete(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.NoError(t, err)
		assert.False(t, deleted)
	})

	t.Run("delete existing category", func(t *testing.T) {
		c, err := repo.Create(context.Background(), fmt.Sprintf("to-delete-%d", categorySeq.Add(1)))
		require.NoError(t, err)

		deleted, err := repo.Delete(context.Background(), c.ID)
		assert.NoError(t, err)
		assert.True(t, deleted)

		categories, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Empty(t, categories)
	})

	t.Run("delete category in use", func(t *testing.T) {
		c, err := repo.Create(context.Background(), fmt.Sprintf("in-use-%d", categorySeq.Add(1)))
		require.NoError(t, err)

		productRepo := database.NewProductRepositoryPGX(testPool)
		p := domain.Product{
			Name:       "Referencing Product",
			Price:      10.0,
			CategoryID: c.ID,
		}
		require.NoError(t, productRepo.Create(context.Background(), &p))

		deleted, err := repo.Delete(context.Background(), c.ID)
		assert.ErrorIs(t, err, domain.ErrCategoryInUse)
		assert.False(t, deleted)

		categories, err := repo.List(context.Background())
		require.NoError(t, err)
		require.Len(t, categories, 1)
		assert.Equal(t, c.ID, categories[0].ID)
	})
}
