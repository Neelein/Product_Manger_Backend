package http

import (
	"encoding/json"
	"net/http"

	"backend/src/usecase"

	"github.com/gorilla/mux"
)

type RegistrationCodeHandler struct {
	service usecase.RegistrationCodeService
}

func NewRegistrationCodeHandler(service usecase.RegistrationCodeService) *RegistrationCodeHandler {
	return &RegistrationCodeHandler{service: service}
}

func (h *RegistrationCodeHandler) CreateCode(w http.ResponseWriter, r *http.Request) {
	member := memberFromRequest(r)
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRegistrationCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	code, err := h.service.CreateApplication(r.Context(), member.ID, req.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, RegistrationCodeResponse{Code: *code})
}

func (h *RegistrationCodeHandler) ListCodes(w http.ResponseWriter, r *http.Request) {
	codes, err := h.service.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, RegistrationCodeListResponse{Codes: codes})
}

func (h *RegistrationCodeHandler) DeleteCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	deleted, err := h.service.Delete(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "registration code not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
