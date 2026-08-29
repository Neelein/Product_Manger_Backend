package usecase

import (
	"bytes"
	"context"
	"io"
	"testing"

	"backend/src/domain/model"
	"backend/src/domain/repository"
	"github.com/stretchr/testify/require"
)

type productImageRepoFake struct {
	repository.Product
	images []model.ProductImage
}

func (f *productImageRepoFake) ListImages(context.Context, string) ([]model.ProductImage, error) {
	return f.images, nil
}
func (f *productImageRepoFake) CreateImage(_ context.Context, image *model.ProductImage) error {
	image.ID = "image-id"
	f.images = append(f.images, *image)
	return nil
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
