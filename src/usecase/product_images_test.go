package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"backend/src/domain/model"
	"backend/src/domain/repository"
	"github.com/stretchr/testify/require"
)

type productImageRepoFake struct {
	repository.Product
	images    []model.ProductImage
	deleted   []string
	deleteErr error
}

func (f *productImageRepoFake) ListImages(context.Context, string) ([]model.ProductImage, error) {
	return f.images, nil
}
func (f *productImageRepoFake) CreateImage(_ context.Context, image *model.ProductImage) error {
	image.ID = "image-id"
	f.images = append(f.images, *image)
	return nil
}
func (f *productImageRepoFake) DeleteImage(_ context.Context, productID, imageID string) (*model.ProductImage, error) {
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	for i, image := range f.images {
		if image.ProductID == productID && image.ID == imageID {
			f.images = append(f.images[:i], f.images[i+1:]...)
			f.deleted = append(f.deleted, imageID)
			return &image, nil
		}
	}
	return nil, model.ErrProductImageNotFound
}

func TestProductServiceUploadImagesPersistsUploads(t *testing.T) {
	repo := &productImageRepoFake{}
	store := &readerStorage{saved: new(int)}
	service := NewProductService(repo, store)

	images, err := service.UploadImages(context.Background(), "product-id", []UploadInput{
		{Directory: "/tmp/product", Filename: "one.jpg", Content: bytes.NewReader([]byte("one"))},
		{Directory: "/tmp/product", Filename: "two.png", Content: bytes.NewReader([]byte("two"))},
	})
	require.NoError(t, err)
	require.Len(t, images, 2)
	require.Len(t, repo.images, 2)
	require.Equal(t, "one.jpg", repo.images[0].Filename)
}

func TestProductServiceUploadImagesRejectsMoreThanThreeTotal(t *testing.T) {
	repo := &productImageRepoFake{images: []model.ProductImage{{}, {}, {}}}
	service := NewProductService(repo, &readerStorage{saved: new(int)})

	_, err := service.UploadImages(context.Background(), "product-id", []UploadInput{{Filename: "four.jpg", Content: bytes.NewReader(nil)}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "limit exceeded")
}

type readerStorage struct{ saved *int }

func (s *readerStorage) Save(_ string, src io.Reader) error {
	_, _ = src.Read(make([]byte, 32))
	*s.saved++
	return nil
}

type deleteTrackingStorage struct {
	deleted []string
	err     error
}

func (s *deleteTrackingStorage) Save(_ string, src io.Reader) error {
	_, _ = io.ReadAll(src)
	return nil
}
func (s *deleteTrackingStorage) Delete(path string) error {
	s.deleted = append(s.deleted, path)
	return s.err
}

func TestProductServiceDeleteImageDeletesDatabaseRowBeforeStoredFile(t *testing.T) {
	repo := &productImageRepoFake{images: []model.ProductImage{{ID: "image-id", ProductID: "product-id", Filename: "one.jpg"}}}
	store := &deleteTrackingStorage{}
	t.Setenv("MEDIA_ROOT", "/media")
	service := NewProductService(repo, store)

	require.NoError(t, service.DeleteImage(context.Background(), "product-id", "image-id"))
	require.Equal(t, []string{"image-id"}, repo.deleted)
	require.Equal(t, []string{"/media/images/products/product-id/one.jpg"}, store.deleted)
}

func TestProductServiceDeleteImageDoesNotDeleteFileWhenDatabaseDeleteFails(t *testing.T) {
	repo := &productImageRepoFake{deleteErr: errors.New("database unavailable")}
	store := &deleteTrackingStorage{}
	service := NewProductService(repo, store)

	require.Error(t, service.DeleteImage(context.Background(), "product-id", "image-id"))
	require.Empty(t, store.deleted)
}

func TestProductServiceDeleteImageReturnsStorageFailureAfterDatabaseDelete(t *testing.T) {
	repo := &productImageRepoFake{images: []model.ProductImage{{ID: "image-id", ProductID: "product-id", Filename: "one.jpg"}}}
	store := &deleteTrackingStorage{err: errors.New("storage unavailable")}
	service := NewProductService(repo, store)

	require.Error(t, service.DeleteImage(context.Background(), "product-id", "image-id"))
	require.Equal(t, []string{"image-id"}, repo.deleted)
	require.Len(t, store.deleted, 1)
}
