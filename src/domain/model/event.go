package model

import "time"

type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatorName string    `json:"creator_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type EventViewer struct {
	MemberID   string    `json:"member_id"`
	MemberName string    `json:"member_name"`
	CreatedAt  time.Time `json:"created_at"`
}
