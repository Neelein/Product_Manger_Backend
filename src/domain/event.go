package domain

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

type CreateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
}

type UpdateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
}

type EventResponse struct {
	Event Event `json:"event"`
}

type EventListResponse struct {
	Events []Event `json:"events"`
}

type AddEventViewerRequest struct {
	MemberID string `json:"member_id"`
}

type EventViewerListResponse struct {
	Viewers []EventViewer `json:"viewers"`
}
