package api

import (
	"encoding/json"
	"net/http"

	"backend/src/domain"

	"github.com/gorilla/mux"
)

type RegistrationCodeHandler struct {
	codeRepo domain.RegistrationCodeRepository
}

func NewRegistrationCodeHandler(codeRepo domain.RegistrationCodeRepository) *RegistrationCodeHandler {
	return &RegistrationCodeHandler{codeRepo: codeRepo}
}

func (h *RegistrationCodeHandler) CreateCode(w http.ResponseWriter, r *http.Request) {
	member := memberFromRequest(r)
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.CreateRegistrationCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	code, err := h.codeRepo.Create(r.Context(), member.ID, req.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.RegistrationCodeResponse{Code: *code})
}

func (h *RegistrationCodeHandler) ListCodes(w http.ResponseWriter, r *http.Request) {
	codes, err := h.codeRepo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, domain.RegistrationCodeListResponse{Codes: codes})
}

func (h *RegistrationCodeHandler) DeleteCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	deleted, err := h.codeRepo.Delete(r.Context(), id)
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