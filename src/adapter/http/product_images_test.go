package http

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/src/domain/model"
	"backend/src/domain/repository"
	"backend/src/usecase"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type imageRepoFake struct{ repository.Product }

func (imageRepoFake) ListImages(context.Context, string) ([]model.ProductImage, error) {
	return nil, nil
}
func (imageRepoFake) CreateImage(_ context.Context, image *model.ProductImage) error {
	image.ID = "image-id"
	return nil
}

type imageStorageFake struct{}

func (imageStorageFake) Save(_ string, src io.Reader) error { _, _ = io.ReadAll(src); return nil }

func imageRequest(t *testing.T, contentType string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("images", "upload")
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	r := httptest.NewRequest(http.MethodPost, "/api/products/product-id/images", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	if contentType != "" {
		r.Header.Set("X-Test-Content-Type", contentType)
	}
	return r
}

func TestUploadImagesRequiresAuthenticatedMember(t *testing.T) {
	h := NewProductHandler(usecase.NewProductService(imageRepoFake{}, imageStorageFake{}))
	r := imageRequest(t, "", []byte("not an image"))
	r = mux.SetURLVars(r, map[string]string{"productId": "product-id"})
	recorder := httptest.NewRecorder()
	h.UploadImages(recorder, r)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUploadImagesRejectsUnsupportedContent(t *testing.T) {
	h := NewProductHandler(usecase.NewProductService(imageRepoFake{}, imageStorageFake{}))
	r := imageRequest(t, "", []byte("plain text"))
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee-id", MemberType: "employee"}))
	r = mux.SetURLVars(r, map[string]string{"productId": "product-id"})
	recorder := httptest.NewRecorder()
	h.UploadImages(recorder, r)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestUploadImagesRejectsEmptyRequest(t *testing.T) {
	h := NewProductHandler(usecase.NewProductService(imageRepoFake{}, imageStorageFake{}))
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.Close())
	r := httptest.NewRequest(http.MethodPost, "/api/products/product-id/images", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee-id", MemberType: "employee"}))
	r = mux.SetURLVars(r, map[string]string{"productId": "product-id"})
	recorder := httptest.NewRecorder()
	h.UploadImages(recorder, r)
	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
