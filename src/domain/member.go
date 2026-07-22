package domain

import "time"

type Member struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID                string    `json:"id"`
	MemberID          string    `json:"member_id"`
	SessionKey        string    `json:"session_key"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	CreatedAt         time.Time `json:"created_at"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Member MemberResponse `json:"member"`
}

type UpdateMemberRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type MemberResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type MembersListResponse struct {
	Members []MemberResponse `json:"members"`
	Total   int              `json:"total"`
}

