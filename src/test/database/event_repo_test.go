//go:build integration

package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	database "backend/src/adapter/postgres"
	domain "backend/src/domain/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var eventMemberCounter int

func createEventMember(t *testing.T, memberRepo *database.MemberRepositoryPGX) domain.Member {
	t.Helper()
	eventMemberCounter++
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	m := domain.Member{
		Email:    fmt.Sprintf("event_test_%d@example.com", eventMemberCounter),
		Password: string(hash),
		Name:     "Event User",
	}
	err = memberRepo.Create(context.Background(), &m)
	require.NoError(t, err)
	return m
}

func cleanupEvents(t *testing.T) {
	t.Helper()
	require.NoError(t, testHarness.Reset(context.Background()))
}

func TestEventRepositoryPGX_Create(t *testing.T) {
	defer cleanupEvents(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	repo := database.NewEventRepositoryPGX(testPool)

	creator := createEventMember(t, memberRepo)

	event := domain.Event{
		Title:       "Team Meeting",
		Description: "Weekly sync",
		StartTime:   time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2026, 7, 25, 11, 0, 0, 0, time.UTC),
		Status:      "active",
		CreatedBy:   creator.ID,
	}

	err := repo.Create(context.Background(), &event)
	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID)
	assert.False(t, event.CreatedAt.IsZero())
	assert.False(t, event.UpdatedAt.IsZero())
	assert.Equal(t, "Team Meeting", event.Title)
}

func TestEventRepositoryPGX_GetByID(t *testing.T) {
	defer cleanupEvents(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	repo := database.NewEventRepositoryPGX(testPool)

	creator := createEventMember(t, memberRepo)

	created := domain.Event{
		Title:       "Sprint Review",
		Description: "Demo new features",
		StartTime:   time.Date(2026, 8, 1, 14, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2026, 8, 1, 15, 0, 0, 0, time.UTC),
		Status:      "active",
		CreatedBy:   creator.ID,
	}
	err := repo.Create(context.Background(), &created)
	require.NoError(t, err)

	t.Run("existing event by creator", func(t *testing.T) {
		got, err := repo.GetByID(context.Background(), created.ID, creator.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "Sprint Review", got.Title)
		assert.Equal(t, "Demo new features", got.Description)
		assert.Equal(t, creator.ID, got.CreatedBy)
		assert.Equal(t, creator.Name, got.CreatorName)
		assert.Equal(t, "active", got.Status)
	})

	t.Run("non-existent event", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), "00000000-0000-0000-0000-000000000000", creator.ID)
		assert.ErrorIs(t, err, domain.ErrEventNotFound)
	})
}

func TestEventRepositoryPGX_ListByMonth(t *testing.T) {
	defer cleanupEvents(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	repo := database.NewEventRepositoryPGX(testPool)

	creator := createEventMember(t, memberRepo)
	viewer := createEventMember(t, memberRepo)

	t.Run("list events in month", func(t *testing.T) {
		e1 := domain.Event{
			Title: "July Event", StartTime: time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC),
			EndTime: time.Date(2026, 7, 15, 11, 0, 0, 0, time.UTC), Status: "active", CreatedBy: creator.ID,
		}
		err := repo.Create(context.Background(), &e1)
		require.NoError(t, err)

		e2 := domain.Event{
			Title: "August Event", StartTime: time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC),
			EndTime: time.Date(2026, 8, 1, 11, 0, 0, 0, time.UTC), Status: "active", CreatedBy: creator.ID,
		}
		err = repo.Create(context.Background(), &e2)
		require.NoError(t, err)

		events, err := repo.ListByMonth(context.Background(), 2026, 7, creator.ID)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, "July Event", events[0].Title)
	})

	t.Run("empty month returns empty list", func(t *testing.T) {
		events, err := repo.ListByMonth(context.Background(), 2026, 12, creator.ID)
		assert.NoError(t, err)
		assert.Empty(t, events)
	})

	t.Run("viewer sees event with viewer access", func(t *testing.T) {
		e := domain.Event{
			Title: "Viewer Event", StartTime: time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC),
			EndTime: time.Date(2026, 7, 20, 11, 0, 0, 0, time.UTC), Status: "active", CreatedBy: creator.ID,
		}
		err := repo.Create(context.Background(), &e)
		require.NoError(t, err)

		err = repo.AddViewer(context.Background(), e.ID, viewer.ID)
		require.NoError(t, err)

		events, err := repo.ListByMonth(context.Background(), 2026, 7, viewer.ID)
		assert.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, "Viewer Event", events[0].Title)
	})
}

func TestEventRepositoryPGX_Update(t *testing.T) {
	defer cleanupEvents(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	repo := database.NewEventRepositoryPGX(testPool)

	creator := createEventMember(t, memberRepo)

	created := domain.Event{
		Title: "Original Title", Description: "Original Desc",
		StartTime: time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 7, 25, 11, 0, 0, 0, time.UTC),
		Status:    "active", CreatedBy: creator.ID,
	}
	err := repo.Create(context.Background(), &created)
	require.NoError(t, err)
	originalUpdatedAt := created.UpdatedAt

	t.Run("update existing event", func(t *testing.T) {
		created.Title = "Updated Title"
		created.Description = "Updated Desc"
		created.Status = "completed"

		err := repo.Update(context.Background(), &created)
		assert.NoError(t, err)
		assert.True(t, created.UpdatedAt.After(originalUpdatedAt))

		got, err := repo.GetByID(context.Background(), created.ID, creator.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", got.Title)
		assert.Equal(t, "Updated Desc", got.Description)
		assert.Equal(t, "completed", got.Status)
	})

	t.Run("update non-existent event", func(t *testing.T) {
		e := domain.Event{
			ID: "00000000-0000-0000-0000-000000000000",
		}
		err := repo.Update(context.Background(), &e)
		assert.ErrorIs(t, err, domain.ErrEventNotFound)
	})
}

func TestEventRepositoryPGX_Delete(t *testing.T) {
	defer cleanupEvents(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	repo := database.NewEventRepositoryPGX(testPool)

	creator := createEventMember(t, memberRepo)

	created := domain.Event{
		Title: "To Delete", StartTime: time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC),
		EndTime: time.Date(2026, 7, 25, 11, 0, 0, 0, time.UTC), Status: "active", CreatedBy: creator.ID,
	}
	err := repo.Create(context.Background(), &created)
	require.NoError(t, err)

	t.Run("delete existing event", func(t *testing.T) {
		err := repo.Delete(context.Background(), created.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(context.Background(), created.ID, creator.ID)
		assert.ErrorIs(t, err, domain.ErrEventNotFound)
	})

	t.Run("delete non-existent event", func(t *testing.T) {
		err := repo.Delete(context.Background(), "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrEventNotFound)
	})
}

func TestEventRepositoryPGX_Viewers(t *testing.T) {
	defer cleanupEvents(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	repo := database.NewEventRepositoryPGX(testPool)

	creator := createEventMember(t, memberRepo)
	viewer1 := createEventMember(t, memberRepo)
	viewer2 := createEventMember(t, memberRepo)

	created := domain.Event{
		Title: "Viewer Test", StartTime: time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC),
		EndTime: time.Date(2026, 7, 25, 11, 0, 0, 0, time.UTC), Status: "active", CreatedBy: creator.ID,
	}
	err := repo.Create(context.Background(), &created)
	require.NoError(t, err)

	t.Run("add viewer", func(t *testing.T) {
		err := repo.AddViewer(context.Background(), created.ID, viewer1.ID)
		assert.NoError(t, err)

		err = repo.AddViewer(context.Background(), created.ID, viewer2.ID)
		assert.NoError(t, err)
	})

	t.Run("list viewers", func(t *testing.T) {
		viewers, err := repo.ListViewers(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Len(t, viewers, 2)

		ids := make(map[string]string)
		for _, v := range viewers {
			ids[v.MemberID] = v.MemberName
		}
		assert.Equal(t, viewer1.Name, ids[viewer1.ID])
		assert.Equal(t, viewer2.Name, ids[viewer2.ID])
	})

	t.Run("viewer can see event", func(t *testing.T) {
		got, err := repo.GetByID(context.Background(), created.ID, viewer1.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, creator.Name, got.CreatorName)
	})

	t.Run("non-viewer cannot see event", func(t *testing.T) {
		nonViewer := createEventMember(t, memberRepo)
		_, err := repo.GetByID(context.Background(), created.ID, nonViewer.ID)
		assert.ErrorIs(t, err, domain.ErrEventNotFound)
	})

	t.Run("remove viewer", func(t *testing.T) {
		err := repo.RemoveViewer(context.Background(), created.ID, viewer1.ID)
		assert.NoError(t, err)

		viewers, err := repo.ListViewers(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Len(t, viewers, 1)
		assert.Equal(t, viewer2.ID, viewers[0].MemberID)
	})

	t.Run("remove non-existent viewer", func(t *testing.T) {
		err := repo.RemoveViewer(context.Background(), created.ID, "00000000-0000-0000-0000-000000000000")
		assert.ErrorIs(t, err, domain.ErrEventViewerNotFound)
	})
}
