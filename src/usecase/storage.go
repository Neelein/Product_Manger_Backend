package usecase

import "io"

// FileStorage is the application port for persisted uploads.
type FileStorage interface {
	Save(path string, src io.Reader) error
}

type FileDeleter interface {
	Delete(path string) error
}

type UploadInput struct {
	Directory string
	Filename  string
	Content   io.Reader
}
