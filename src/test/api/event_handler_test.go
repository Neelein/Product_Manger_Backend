package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "backend/src/adapter/http"
	domain "backend/src/adapter/http"
	"backend/src/domain/model"
	"backend/src/usecase"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockEventRepo struct {
	createFunc       func(ctx context.Context, event *domain.Event) error
	getByIDFunc      func(ctx context.Context, id string, memberID string) (*domain.Event, error)
	listByMonthFunc  func(ctx context.Context, year, month int, memberID string) ([]domain.Event, error)
	updateFunc       func(ctx context.Context, event *domain.Event) error
	deleteFunc       func(ctx context.Context, id string) error
	addViewerFunc    func(ctx context.Context, eventID string, memberID string) error
	removeViewerFunc func(ctx context.Context, eventID string, memberID string) error
	listViewersFunc  func(ctx context.Context, eventID string) ([]domain.EventViewer, error)
}

func (m *mockEventRepo) Create(ctx context.Context, event *domain.Event) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, event)
	}
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()
	return nil
}

func (m *mockEventRepo) GetByID(ctx context.Context, id string, memberID string) (*domain.Event, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id, memberID)
	}
	return &domain.Event{ID: id, Title: "Test Event", CreatedBy: "creator-id"}, nil
}

func (m *mockEventRepo) ListByMonth(ctx context.Context, year, month int, memberID string) ([]domain.Event, error) {
	if m.listByMonthFunc != nil {
		return m.listByMonthFunc(ctx, year, month, memberID)
	}
	return []domain.Event{}, nil
}
func (m *mockEventRepo) ListByMonthApplication(ctx context.Context, memberID string, year, month int) ([]domain.Event, error) {
	if year == 0 || month == 0 {
		return nil, usecase.ErrEventMonthRequired
	}
	if month < 1 || month > 12 {
		return nil, usecase.ErrEventMonthInvalid
	}
	return m.ListByMonth(ctx, year, month, memberID)
}

func (m *mockEventRepo) Update(ctx context.Context, event *domain.Event) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, event)
	}
	event.UpdatedAt = time.Now()
	return nil
}

func (m *mockEventRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockEventRepo) AddViewer(ctx context.Context, eventID string, memberID string) error {
	if m.addViewerFunc != nil {
		return m.addViewerFunc(ctx, eventID, memberID)
	}
	return nil
}

func (m *mockEventRepo) RemoveViewer(ctx context.Context, eventID string, memberID string) error {
	if m.removeViewerFunc != nil {
		return m.removeViewerFunc(ctx, eventID, memberID)
	}
	return nil
}

func (m *mockEventRepo) ListViewers(ctx context.Context, eventID string) ([]domain.EventViewer, error) {
	if m.listViewersFunc != nil {
		return m.listViewersFunc(ctx, eventID)
	}
	return []domain.EventViewer{}, nil
}

func (m *mockEventRepo) CreateApplication(ctx context.Context, memberID string, input usecase.EventCreateInput) (*domain.Event, error) {
	event := &domain.Event{Title: input.Title, Description: input.Description, Status: input.Status, CreatedBy: memberID, StartTime: input.StartTime, EndTime: input.EndTime}
	if err := m.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}
func (m *mockEventRepo) UpdateApplication(ctx context.Context, id, memberID string, admin bool, input usecase.EventUpdateInput) (*domain.Event, error) {
	event, err := m.GetByID(ctx, id, memberID)
	if err != nil {
		return nil, err
	}
	if !admin && event.CreatedBy != memberID {
		return nil, model.ErrNotEventOwner
	}
	if input.Title != "" {
		event.Title = input.Title
	}
	if input.Description != "" {
		event.Description = input.Description
	}
	if err := m.Update(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}
func (m *mockEventRepo) DeleteApplication(ctx context.Context, id, memberID string, admin bool) error {
	event, err := m.GetByID(ctx, id, memberID)
	if err != nil {
		return err
	}
	if !admin && event.CreatedBy != memberID {
		return model.ErrNotEventOwner
	}
	return m.Delete(ctx, id)
}
func (m *mockEventRepo) AddViewerApplication(ctx context.Context, id, memberID, viewerID string, admin bool) error {
	event, err := m.GetByID(ctx, id, memberID)
	if err != nil {
		return err
	}
	if !admin && event.CreatedBy != memberID {
		return model.ErrNotEventOwner
	}
	return m.AddViewer(ctx, id, viewerID)
}
func (m *mockEventRepo) RemoveViewerApplication(ctx context.Context, id, memberID, viewerID string, admin bool) error {
	return m.RemoveViewer(ctx, id, viewerID)
}
func (m *mockEventRepo) ListViewersApplication(ctx context.Context, id, memberID string, admin bool) ([]domain.EventViewer, error) {
	return m.ListViewers(ctx, id)
}

func TestEvent_CreateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := &mockEventRepo{}
		handler := api.NewEventHandler(mock)

		body, _ := json.Marshal(domain.CreateEventRequest{
			Title:       "Test Event",
			Description: "A description",
			StartTime:   time.Date(2026, 7, 25, 18, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			EndTime:     time.Date(2026, 7, 25, 20, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			Status:      "active",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id", Name: "Test User"}))
		w := httptest.NewRecorder()
		handler.CreateEvent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp domain.EventResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.Event.ID)
		assert.Equal(t, "Test Event", resp.Event.Title)
		assert.Equal(t, "A description", resp.Event.Description)
		assert.Equal(t, "active", resp.Event.Status)
		assert.True(t, resp.Event.StartTime.Equal(time.Date(2026, 7, 25, 10, 0, 0, 0, time.UTC)), "StartTime should be UTC")
		assert.True(t, resp.Event.EndTime.Equal(time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)), "EndTime should be UTC")
	})

	t.Run("unauthorized", func(t *testing.T) {
		mock := &mockEventRepo{}
		handler := api.NewEventHandler(mock)

		body, _ := json.Marshal(domain.CreateEventRequest{
			Title: "Test Event",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.CreateEvent(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		mock := &mockEventRepo{}
		handler := api.NewEventHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewReader([]byte("{invalid}")))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.CreateEvent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEvent_ListEventsByMonth(t *testing.T) {
	t.Run("success with events", func(t *testing.T) {
		expected := []domain.Event{
			{ID: uuid.New().String(), Title: "Event 1", StartTime: time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)},
			{ID: uuid.New().String(), Title: "Event 2", StartTime: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)},
		}
		mock := &mockEventRepo{
			listByMonthFunc: func(ctx context.Context, year, month int, memberID string) ([]domain.Event, error) {
				assert.Equal(t, 2026, year)
				assert.Equal(t, 7, month)
				return expected, nil
			},
		}
		handler := api.NewEventHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/api/events?year=2026&month=7", nil)
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.ListEventsByMonth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.EventListResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Len(t, resp.Events, 2)
		assert.Equal(t, "Event 1", resp.Events[0].Title)
		assert.Equal(t, "Event 2", resp.Events[1].Title)
	})

	t.Run("missing year/month params", func(t *testing.T) {
		mock := &mockEventRepo{}
		handler := api.NewEventHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.ListEventsByMonth(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestEvent_GetEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		eventID := uuid.New().String()
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				assert.Equal(t, eventID, id)
				return &domain.Event{
					ID: id, Title: "Test Event", Description: "Desc",
					Status: "active", CreatedBy: "creator-id", CreatorName: "Creator",
				}, nil
			},
		}
		handler := api.NewEventHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID, nil)
		req = mux.SetURLVars(req, map[string]string{"eventId": eventID})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.GetEvent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.EventResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, eventID, resp.Event.ID)
		assert.Equal(t, "Test Event", resp.Event.Title)
		assert.Equal(t, "Creator", resp.Event.CreatorName)
	})

	t.Run("not found", func(t *testing.T) {
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				return nil, domain.ErrEventNotFound
			},
		}
		handler := api.NewEventHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/api/events/non-existent", nil)
		req = mux.SetURLVars(req, map[string]string{"eventId": "00000000-0000-0000-0000-000000000000"})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.GetEvent(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestEvent_UpdateEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		eventID := uuid.New().String()
		stored := &domain.Event{ID: eventID, Title: "Original", CreatedBy: "member-id"}
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				return stored, nil
			},
			updateFunc: func(ctx context.Context, event *domain.Event) error {
				stored = event
				return nil
			},
		}
		handler := api.NewEventHandler(mock)

		body, _ := json.Marshal(domain.UpdateEventRequest{Title: "Updated"})
		req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"eventId": eventID})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.UpdateEvent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.EventResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, "Updated", resp.Event.Title)
	})

	t.Run("not owner", func(t *testing.T) {
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				return &domain.Event{ID: id, Title: "Original", CreatedBy: "other-user"}, nil
			},
		}
		handler := api.NewEventHandler(mock)

		body, _ := json.Marshal(domain.UpdateEventRequest{Title: "Updated"})
		req := httptest.NewRequest(http.MethodPost, "/api/events/some-id/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"eventId": "some-id"})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.UpdateEvent(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestEvent_DeleteEvent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		eventID := uuid.New().String()
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				return &domain.Event{ID: id, Title: "To Delete", CreatedBy: "member-id"}, nil
			},
			deleteFunc: func(ctx context.Context, id string) error {
				return nil
			},
		}
		handler := api.NewEventHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/delete", nil)
		req = mux.SetURLVars(req, map[string]string{"eventId": eventID})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.DeleteEvent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not owner", func(t *testing.T) {
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				return &domain.Event{ID: id, Title: "To Delete", CreatedBy: "other-user"}, nil
			},
		}
		handler := api.NewEventHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/api/events/some-id/delete", nil)
		req = mux.SetURLVars(req, map[string]string{"eventId": "some-id"})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.DeleteEvent(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestEvent_AddViewer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		eventID := uuid.New().String()
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				return &domain.Event{ID: id, Title: "Test", CreatedBy: "member-id"}, nil
			},
			addViewerFunc: func(ctx context.Context, eventID string, memberID string) error {
				return nil
			},
		}
		handler := api.NewEventHandler(mock)

		body, _ := json.Marshal(domain.AddEventViewerRequest{MemberID: "viewer-id"})
		req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/viewers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"eventId": eventID})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.AddViewer(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("not owner", func(t *testing.T) {
		eventID := uuid.New().String()
		mock := &mockEventRepo{
			getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
				return &domain.Event{ID: id, Title: "Test", CreatedBy: "other-user"}, nil
			},
		}
		handler := api.NewEventHandler(mock)

		body, _ := json.Marshal(domain.AddEventViewerRequest{MemberID: "viewer-id"})
		req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/viewers", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{"eventId": eventID})
		req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
		w := httptest.NewRecorder()
		handler.AddViewer(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestEvent_RemoveViewer(t *testing.T) {
	eventID := uuid.New().String()
	mock := &mockEventRepo{
		getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
			return &domain.Event{ID: id, Title: "Test", CreatedBy: "member-id"}, nil
		},
		removeViewerFunc: func(ctx context.Context, eventID string, memberID string) error {
			return nil
		},
	}
	handler := api.NewEventHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/api/events/"+eventID+"/viewers/viewer-id/remove", nil)
	req = mux.SetURLVars(req, map[string]string{"eventId": eventID, "memberId": "viewer-id"})
	req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
	w := httptest.NewRecorder()
	handler.RemoveViewer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEvent_ListViewers(t *testing.T) {
	eventID := uuid.New().String()
	expected := []domain.EventViewer{
		{MemberID: "viewer-1", MemberName: "Viewer One"},
		{MemberID: "viewer-2", MemberName: "Viewer Two"},
	}
	mock := &mockEventRepo{
		getByIDFunc: func(ctx context.Context, id string, memberID string) (*domain.Event, error) {
			return &domain.Event{ID: id, Title: "Test", CreatedBy: "member-id"}, nil
		},
		listViewersFunc: func(ctx context.Context, id string) ([]domain.EventViewer, error) {
			assert.Equal(t, eventID, id)
			return expected, nil
		},
	}
	handler := api.NewEventHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/api/events/"+eventID+"/viewers", nil)
	req = mux.SetURLVars(req, map[string]string{"eventId": eventID})
	req = req.WithContext(api.ContextWithMember(req.Context(), &domain.Member{ID: "member-id"}))
	w := httptest.NewRecorder()
	handler.ListViewers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.EventViewerListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Viewers, 2)
	assert.Equal(t, "viewer-1", resp.Viewers[0].MemberID)
	assert.Equal(t, "Viewer One", resp.Viewers[0].MemberName)
}
