package domain

import "time"

type Announcement struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	ImagePath     string    `json:"image_path"`
	PublisherID   string    `json:"publisher_id"`
	PublisherName string    `json:"publisher_name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateAnnouncementRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	ImagePath string `json:"image_path"`
}

type UpdateAnnouncementRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	ImagePath string `json:"image_path"`
}

type AnnouncementResponse struct {
	Announcement Announcement `json:"announcement"`
}

type AnnouncementListResponse struct {
	Announcements []Announcement `json:"announcements"`
	Total         int            `json:"total"`
	Page          int            `json:"page"`
	Limit         int            `json:"limit"`
}
