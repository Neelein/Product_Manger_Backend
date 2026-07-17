package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"backend/src/domain"

	"golang.org/x/crypto/bcrypt"
)

func memberFromRequest(r *http.Request) *domain.Member {
	m := MemberFromContext(r.Context())
	if m == nil {
		return nil
	}
	return m
}

type MemberHandler struct {
	memberRepo  domain.MemberRepository
	sessionRepo domain.SessionRepository
}

func NewMemberHandler(
	memberRepo domain.MemberRepository,
	sessionRepo domain.SessionRepository,
) *MemberHandler {
	return &MemberHandler{
		memberRepo:  memberRepo,
		sessionRepo: sessionRepo,
	}
}

func (h *MemberHandler) RegisterMember(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	member := domain.Member{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	if err := h.memberRepo.Create(context.Background(), &member); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.MemberResponse{
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
	})
}

func (h *MemberHandler) LoginMember(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	member, err := h.memberRepo.GetByEmail(context.Background(), req.Email)
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

	session := domain.Session{MemberID: member.ID}
	if err := h.sessionRepo.Create(context.Background(), &session); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    session.SessionKey,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})

	writeJSON(w, http.StatusOK, domain.LoginResponse{
		Member: domain.MemberResponse{
			ID:    member.ID,
			Email: member.Email,
			Name:  member.Name,
		},
	})
}

func (h *MemberHandler) LogoutMember(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_key")
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
		return
	}

	if err := h.sessionRepo.Delete(context.Background(), cookie.Value); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_key",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	member := memberFromRequest(r)
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.UpdateMemberRequest
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

	if err := h.memberRepo.Update(context.Background(), member); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		if errors.Is(err, domain.ErrMemberNotFound) {
			writeError(w, http.StatusNotFound, "member not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.MemberResponse{
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
	})
}

func (h *MemberHandler) GetCurrentMember(w http.ResponseWriter, r *http.Request) {
	member := memberFromRequest(r)
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeJSON(w, http.StatusOK, domain.MemberResponse{
		ID:    member.ID,
		Email: member.Email,
		Name:  member.Name,
	})
}


