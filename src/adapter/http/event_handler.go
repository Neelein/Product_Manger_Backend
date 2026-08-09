package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend/src/domain/model"
	"backend/src/usecase"

	"github.com/gorilla/mux"
)

type EventHandler struct {
	service usecase.EventService
}

func eventError(w http.ResponseWriter, err error, fallback int) bool {
	if errors.Is(err, model.ErrNotEventOwner) {
		writeError(w, http.StatusForbidden, "forbidden")
		return true
	}
	if errors.Is(err, model.ErrEventNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return true
	}
	writeError(w, fallback, err.Error())
	return true
}

func NewEventHandler(service usecase.EventService) *EventHandler {
	return &EventHandler{service: service}
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

	event, err := h.service.CreateApplication(r.Context(), member.ID, usecase.EventCreateInput{
		Title: req.Title, Description: req.Description, StartTime: req.StartTime, EndTime: req.EndTime, Status: req.Status,
	})
	if err != nil {
		if err == usecase.ErrEventTitleRequired {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, EventResponse{Event: *event})
}

func (h *EventHandler) ListEventsByMonth(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		year = 0
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		month = 0
	}
	if yearStr != "" && year == 0 {
		writeError(w, http.StatusBadRequest, "invalid year")
		return
	}
	if monthStr != "" && month == 0 {
		writeError(w, http.StatusBadRequest, "invalid month")
		return
	}
	events, err := h.service.ListByMonthApplication(r.Context(), member.ID, year, month)
	if err != nil {
		if err == usecase.ErrEventMonthRequired {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if year == 0 {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}
		if err == usecase.ErrEventMonthInvalid {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}
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

	event, err := h.service.GetByID(r.Context(), eventID, member.ID)
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

	var req UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updated, err := h.service.UpdateApplication(r.Context(), eventID, member.ID, member.Role == "admin", usecase.EventUpdateInput{
		Title: req.Title, Description: req.Description, StartTime: req.StartTime, EndTime: req.EndTime, Status: req.Status,
	})
	if err != nil {
		eventError(w, err, http.StatusInternalServerError)
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

	if err := h.service.DeleteApplication(r.Context(), eventID, member.ID, member.Role == "admin"); err != nil {
		eventError(w, err, http.StatusInternalServerError)
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

	var req AddEventViewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.AddViewerApplication(r.Context(), eventID, member.ID, req.MemberID, member.Role == "admin"); err != nil {
		if err == model.ErrInvalidProductVariant {
			writeError(w, http.StatusBadRequest, "member_id is required")
			return
		}
		eventError(w, err, http.StatusInternalServerError)
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

	if err := h.service.RemoveViewerApplication(r.Context(), eventID, member.ID, memberID, member.Role == "admin"); err != nil {
		eventError(w, err, http.StatusInternalServerError)
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

	viewers, err := h.service.ListViewersApplication(r.Context(), eventID, member.ID, member.Role == "admin")
	if err != nil {
		eventError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, EventViewerListResponse{Viewers: viewers})
}
