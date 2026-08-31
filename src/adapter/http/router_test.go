package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestAnnouncementMediaServesJPEGWithImageContentType(t *testing.T) {
	mediaRoot := t.TempDir()
	announcementDir := filepath.Join(mediaRoot, "images", "announcements")
	require.NoError(t, os.MkdirAll(announcementDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(announcementDir, "image.jpg"), []byte("jpeg bytes"), 0644))
	t.Setenv("MEDIA_ROOT", mediaRoot)

	router := mux.NewRouter()
	RegisterAnnouncementRoutes(router, nil, nil, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/media/images/announcements/image.jpg", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "image/jpeg", recorder.Header().Get("Content-Type"))
}
