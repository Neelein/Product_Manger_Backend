package model

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
