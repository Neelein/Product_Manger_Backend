package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileStorage interface {
	Save(path string, src io.Reader) error
}

type LocalFileStorage struct{}

func (LocalFileStorage) Save(path string, src io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating upload directory: %w", err)
	}
	return Save(path, src)
}

func (LocalFileStorage) Delete(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func Save(path string, src io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, src)
	return err
}
