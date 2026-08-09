package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/src/usecase"

	"github.com/gorilla/mux"
)

type EventHandler struct {
	repo usecase.EventService
}

func NewEventHandler(service usecase.EventService) *EventHandler {
	return &EventHandler{repo: service}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	req.StartTime = req.StartTime.UTC()
	req.EndTime = req.EndTime.UTC()

	event := Event{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      req.Status,
		CreatedBy:   member.ID,
	}

	if err := h.repo.Create(r.Context(), &event); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, EventResponse{Event: event})
}

func (h *EventHandler) ListEventsByMonth(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	if yearStr == "" || monthStr == "" {
		writeError(w, http.StatusBadRequest, "year and month query parameters are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid year")
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		writeError(w, http.StatusBadRequest, "invalid month")
		return
	}

	events, err := h.repo.ListByMonth(r.Context(), year, month, member.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EventListResponse{Events: events})
}

func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	eventID := mux.Vars(r)["eventId"]

	event, err := h.repo.GetByID(r.Context(), eventID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EventResponse{Event: *event})
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	eventID := mux.Vars(r)["eventId"]

	event, err := h.repo.GetByID(r.Context(), eventID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if event.CreatedBy != member.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if !req.StartTime.IsZero() {
		event.StartTime = req.StartTime.UTC()
	}
	if !req.EndTime.IsZero() {
		event.EndTime = req.EndTime.UTC()
	}
	if req.Status != "" {
		event.Status = req.Status
	}

	if err := h.repo.Update(r.Context(), event); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, err := h.repo.GetByID(r.Context(), eventID, member.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EventResponse{Event: *updated})
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	eventID := mux.Vars(r)["eventId"]

	event, err := h.repo.GetByID(r.Context(), eventID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if event.CreatedBy != member.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.repo.Delete(r.Context(), eventID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "event deleted"})
}

func (h *EventHandler) AddViewer(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	eventID := mux.Vars(r)["eventId"]

	event, err := h.repo.GetByID(r.Context(), eventID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if event.CreatedBy != member.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req AddEventViewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MemberID == "" {
		writeError(w, http.StatusBadRequest, "member_id is required")
		return
	}

	if err := h.repo.AddViewer(r.Context(), eventID, req.MemberID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "viewer added"})
}

func (h *EventHandler) RemoveViewer(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	eventID := vars["eventId"]
	memberID := vars["memberId"]

	event, err := h.repo.GetByID(r.Context(), eventID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if event.CreatedBy != member.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.repo.RemoveViewer(r.Context(), eventID, memberID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "viewer removed"})
}

func (h *EventHandler) ListViewers(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	eventID := mux.Vars(r)["eventId"]

	event, err := h.repo.GetByID(r.Context(), eventID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if event.CreatedBy != member.ID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	viewers, err := h.repo.ListViewers(r.Context(), eventID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, EventViewerListResponse{Viewers: viewers})
}
