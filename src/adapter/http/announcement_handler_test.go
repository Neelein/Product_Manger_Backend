package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"strings"
	"testing"

	"backend/src/domain/model"
	"backend/src/domain/repository"
	"backend/src/usecase"
	"github.com/stretchr/testify/require"
)

type announcementRepoFake struct {
	repository.Announcement
	created *model.Announcement
}

func (r *announcementRepoFake) Create(_ context.Context, announcement *model.Announcement) error {
	announcement.ID = "announcement-id"
	r.created = announcement
	return nil
}

type announcementStorageFake struct {
	filename string
}

func (s *announcementStorageFake) Save(path string, src io.Reader) error {
	s.filename = path
	_, err := io.ReadAll(src)
	return err
}

func TestCreateAnnouncementAcceptsJPEGAndUsesDetectedExtension(t *testing.T) {
	repo := &announcementRepoFake{}
	storage := &announcementStorageFake{}
	handler := NewAnnouncementHandler(usecase.NewAnnouncementService(repo, storage))
	content, err := base64.StdEncoding.DecodeString("/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////2wBDAf//////////////////////////////////////////////////////////////////////////////////////wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAX/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIQAxAAAAH/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/9oACAEBAAEFAqf/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oACAEDAQE/AX//xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oACAECAQE/AX//xAAUEAEAAAAAAAAAAAAAAAAAAAAA/9oACAEBAAY/Aqf/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/9oACAEBAAE/Iqf/2gAMAwEAAgADAAAAEP/EABQRAQAAAAAAAAAAAAAAAAAAABD/2gAIAQMBAT8QH//EABQRAQAAAAAAAAAAAAAAAAAAABD/2gAIAQIBAT8QH//EABQQAQAAAAAAAAAAAAAAAAAAABD/2gAIAQEAAT8QH//Z")
	require.NoError(t, err)
	request := announcementMultipartRequest(t, "photo.heic", content, "Title", "Body", "")
	recorder := httptest.NewRecorder()
	handler.CreateAnnouncement(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.True(t, strings.HasSuffix(storage.filename, ".jpg"))
	require.Equal(t, "announcement-id", repo.created.ID)
	require.Contains(t, repo.created.ImagePath, "/media/images/announcements/")
}

func TestCreateAnnouncementRejectsUnsupportedContentAndHEIC(t *testing.T) {
	for _, test := range []struct {
		name     string
		filename string
		content  []byte
	}{
		{name: "plain content with forged jpg MIME", filename: "photo.jpg", content: []byte("not an image")},
		{name: "heic", filename: "photo.heic", content: []byte("....ftypheic....")},
	} {
		t.Run(test.name, func(t *testing.T) {
			handler := NewAnnouncementHandler(usecase.NewAnnouncementService(&announcementRepoFake{}, &announcementStorageFake{}))
			recorder := httptest.NewRecorder()
			handler.CreateAnnouncement(recorder, announcementMultipartRequest(t, test.filename, test.content, "Title", "Body", "image/jpeg"))
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			require.Contains(t, recorder.Body.String(), "only JPEG, PNG, and WebP")
		})
	}
}

func TestCreateAnnouncementAcceptsPNGAndWebP(t *testing.T) {
	images := []struct {
		name     string
		filename string
		ext      string
		encoded  string
	}{
		{name: "png", filename: "photo.any", ext: ".png", encoded: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="},
		{name: "webp", filename: "photo.jpg", ext: ".webp", encoded: "UklGRiIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEAAUAmJaQAA3AA/v89WAAAAA=="},
	}
	for _, image := range images {
		t.Run(image.name, func(t *testing.T) {
			content, err := base64.StdEncoding.DecodeString(image.encoded)
			require.NoError(t, err)
			repo := &announcementRepoFake{}
			storage := &announcementStorageFake{}
			handler := NewAnnouncementHandler(usecase.NewAnnouncementService(repo, storage))
			recorder := httptest.NewRecorder()
			handler.CreateAnnouncement(recorder, announcementMultipartRequest(t, image.filename, content, "Title", "Body", ""))
			require.Equal(t, http.StatusCreated, recorder.Code)
			require.Equal(t, image.ext, filepath.Ext(storage.filename))
		})
	}
}

func announcementMultipartRequest(t *testing.T, filename string, content []byte, title string, bodyText string, contentType string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("title", title))
	require.NoError(t, writer.WriteField("content", bodyText))
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="image"; filename="`+filename+`"`)
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	part, err := writer.CreatePart(header)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	request := httptest.NewRequest(http.MethodPost, "/api/announcements", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request.WithContext(ContextWithMember(request.Context(), &Member{ID: "employee-id", MemberType: "employee"}))
}
