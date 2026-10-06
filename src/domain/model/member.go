package model

import "time"

type Member struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Password       string    `json:"-"`
	Name           string    `json:"name"`
	Phone          string    `json:"phone"`
	MemberType     string    `json:"member_type"`
	Permission     string    `json:"permission"`
	DepartmentCode int       `json:"department_code"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// EmployeeDirectoryEntry is the read-only, department-scoped employee contract.
type EmployeeDirectoryEntry struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Department Department `json:"department"`
	Phone      string     `json:"phone"`
}
type Session struct {
	ID         string    `json:"id"`
	MemberID   string    `json:"member_id"`
	SessionKey string    `json:"session_key"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}
type RegistrationCode struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	CreatedBy      string     `json:"created_by"`
	CreatedByEmail string     `json:"created_by_email"`
	UsedBy         string     `json:"used_by"`
	UsedByEmail    string     `json:"used_by_email"`
	UsedAt         *time.Time `json:"used_at"`
	CreatedAt      time.Time  `json:"created_at"`
	Status         string     `json:"status"`
}
