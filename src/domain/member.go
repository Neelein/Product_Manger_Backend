package domain

import "time"

type Member struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID         string    `json:"id"`
	MemberID   string    `json:"member_id"`
	SessionKey string    `json:"session_key"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Code     string `json:"code"`
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
	Role  string `json:"role"`
}

type RegistrationCode struct {
	ID              string     `json:"id"`
	Code            string     `json:"code"`
	CreatedBy       string     `json:"created_by"`
	CreatedByEmail  string     `json:"created_by_email"`
	UsedBy          string     `json:"used_by"`
	UsedByEmail     string     `json:"used_by_email"`
	UsedAt          *time.Time `json:"used_at"`
	CreatedAt       time.Time  `json:"created_at"`
	Status          string     `json:"status"`
}

type CreateRegistrationCodeRequest struct {
	Code string `json:"code"`
}

type RegistrationCodeResponse struct {
	Code RegistrationCode `json:"code"`
}

type RegistrationCodeListResponse struct {
	Codes []RegistrationCode `json:"codes"`
}

type MembersListResponse struct {
	Members []MemberResponse `json:"members"`
	Total   int              `json:"total"`
}
