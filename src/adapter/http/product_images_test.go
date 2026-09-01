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
func (imageRepoFake) DeleteImage(context.Context, string, string) (*model.ProductImage, error) {
	return &model.ProductImage{ProductID: "product-id", Filename: "upload.jpg"}, nil
}

type imageStorageFake struct{}

func (imageStorageFake) Save(_ string, src io.Reader) error { _, _ = io.ReadAll(src); return nil }
func (imageStorageFake) Delete(string) error                { return nil }

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

func TestDeleteImageReturnsSuccessMessage(t *testing.T) {
	h := NewProductHandler(usecase.NewProductService(imageRepoFake{}, imageStorageFake{}))
	r := httptest.NewRequest(http.MethodPost, "/api/products/product-id/images/image-id/delete", nil)
	r = mux.SetURLVars(r, map[string]string{"productId": "product-id", "imageId": "image-id"})
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee-id", MemberType: "employee"}))
	recorder := httptest.NewRecorder()
	h.DeleteImage(recorder, r)
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"message":"product image deleted"}`, recorder.Body.String())
}

func TestDeleteImageRequiresAuthenticatedMember(t *testing.T) {
	h := NewProductHandler(usecase.NewProductService(imageRepoFake{}, imageStorageFake{}))
	r := httptest.NewRequest(http.MethodPost, "/api/products/product-id/images/image-id/delete", nil)
	r = mux.SetURLVars(r, map[string]string{"productId": "product-id", "imageId": "image-id"})
	recorder := httptest.NewRecorder()
	h.DeleteImage(recorder, r)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestDeleteImageMapsMissingImageToNotFound(t *testing.T) {
	service := usecase.NewProductService(&missingImageRepoFake{}, imageStorageFake{})
	h := NewProductHandler(service)
	r := httptest.NewRequest(http.MethodPost, "/api/products/product-id/images/missing/delete", nil)
	r = mux.SetURLVars(r, map[string]string{"productId": "product-id", "imageId": "missing"})
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee-id", MemberType: "employee"}))
	recorder := httptest.NewRecorder()
	h.DeleteImage(recorder, r)
	require.Equal(t, http.StatusNotFound, recorder.Code)
}

type missingImageRepoFake struct{ repository.Product }

func (missingImageRepoFake) DeleteImage(context.Context, string, string) (*model.ProductImage, error) {
	return nil, model.ErrProductImageNotFound
}
