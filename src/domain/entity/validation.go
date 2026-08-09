package entity

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrRequiredName = errors.New("name is required")
	ErrInvalidUUID  = errors.New("invalid UUID")
	ErrInvalidPage  = errors.New("invalid page parameter")
	ErrInvalidLimit = errors.New("invalid limit parameter")
)

func RequiredName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrRequiredName
	}
	return nil
}

func UUID(value string) error {
	if uuid.Validate(value) != nil {
		return ErrInvalidUUID
	}
	return nil
}

func PageLimit(page, limit int) error {
	if page < 1 {
		return ErrInvalidPage
	}
	if limit < 1 || limit > 100 {
		return ErrInvalidLimit
	}
	return nil
}
