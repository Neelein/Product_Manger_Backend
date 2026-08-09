package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/src/usecase"

	"golang.org/x/crypto/bcrypt"
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

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "registration code is required")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	member := Member{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	if err := h.codeService.RegisterMemberWithCode(r.Context(), &member, req.Code); err != nil {
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
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
		Role:  member.Role,
	})
}

func (h *MemberHandler) LoginMember(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	member, err := h.memberService.GetByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if member == nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	session := Session{
		MemberID: member.ID,
	}
	if err := h.sessionService.Create(r.Context(), &session); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, sessionCookie(r, &session))

	writeJSON(w, http.StatusOK, LoginResponse{
		Member: MemberResponse{
			ID:    member.ID,
			Email: member.Email,
			Name:  member.Name,
			Role:  member.Role,
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

	if req.Email == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email and name are required")
		return
	}

	member.Email = req.Email
	member.Name = req.Name

	if err := h.memberService.Update(r.Context(), member); err != nil {
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
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
		Role:  member.Role,
	})
}

func (h *MemberHandler) GetCurrentMember(w http.ResponseWriter, r *http.Request) {
	member := memberFromRequest(r)
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, MemberResponse{
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
		Role:  member.Role,
	})
}
