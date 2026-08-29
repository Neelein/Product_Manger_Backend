package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/src/domain/model"
	"backend/src/usecase"
	"github.com/gorilla/mux"
)

func memberFromRequest(r *http.Request) *Member {
	m := MemberFromContext(r.Context())
	if m == nil {
		return nil
	}
	return m
}

type MemberHandler struct {
	memberService  usecase.MemberService
	sessionService usecase.SessionService
	codeService    usecase.RegistrationCodeService
}

func NewMemberHandler(
	memberService usecase.MemberService,
	sessionService usecase.SessionService,
	codeService usecase.RegistrationCodeService,
) *MemberHandler {
	return &MemberHandler{
		memberService:  memberService,
		sessionService: sessionService,
		codeService:    codeService,
	}
}

func (h *MemberHandler) RegisterMember(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	member := Member{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := h.memberService.RegisterApplication(r.Context(), &member, req.Password, req.Code); err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			writeError(w, http.StatusConflict, "email already exists")
		case errors.Is(err, ErrInvalidRegistrationCode):
			writeError(w, http.StatusBadRequest, "invalid registration code")
		case errors.Is(err, ErrRegistrationCodeUsed):
			writeError(w, http.StatusBadRequest, "registration code already used")
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, MemberResponse{
		ID: member.ID, Email: member.Email, Name: member.Name,
		MemberType: member.MemberType, Permission: member.Permission,
	})
}

func (h *MemberHandler) UpdateMemberPermission(w http.ResponseWriter, r *http.Request) {
	actor := memberFromRequest(r)
	if actor == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req UpdateMemberPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.memberService.UpdatePermission(r.Context(), actor.ID, mux.Vars(r)["memberId"], req.Permission); err != nil {
		if errors.Is(err, model.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		if errors.Is(err, ErrMemberNotFound) {
			writeError(w, http.StatusNotFound, "member not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"permission": req.Permission})
}

func (h *MemberHandler) LoginMember(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	member, session, err := h.memberService.Authenticate(r.Context(), req.Email, req.Password)
	if err == model.ErrInvalidCredentials {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, sessionCookie(r, session))

	writeJSON(w, http.StatusOK, LoginResponse{
		Member: MemberResponse{
			ID:         member.ID,
			Email:      member.Email,
			Name:       member.Name,
			MemberType: member.MemberType, Permission: member.Permission,
		},
	})
}

func (h *MemberHandler) LogoutMember(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_key")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
		return
	}

	if err := h.sessionService.Delete(r.Context(), cookie.Value); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, clearSessionCookie(r))

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	member := memberFromRequest(r)
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.memberService.UpdateApplication(r.Context(), member, req.Email, req.Name); err != nil {
		if err.Error() == "email and name are required" {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		if errors.Is(err, ErrMemberNotFound) {
			writeError(w, http.StatusNotFound, "member not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MemberResponse{
		ID:         member.ID,
		Email:      member.Email,
		Name:       member.Name,
		MemberType: member.MemberType, Permission: member.Permission,
	})
}

func (h *MemberHandler) GetCurrentMember(w http.ResponseWriter, r *http.Request) {
	member := memberFromRequest(r)
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, MemberResponse{
		ID:         member.ID,
		Email:      member.Email,
		Name:       member.Name,
		MemberType: member.MemberType, Permission: member.Permission,
	})
}
