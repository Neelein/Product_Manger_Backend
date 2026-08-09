package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalFileStorageSavesNestedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "file.txt")
	if err := (LocalFileStorage{}).Save(path, strings.NewReader("content")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Fatalf("content = %q", data)
	}
}
