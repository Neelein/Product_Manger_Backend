package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"backend/src/domain"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AnnouncementHandler struct {
	repo domain.AnnouncementRepository
}

func NewAnnouncementHandler(repo domain.AnnouncementRepository) *AnnouncementHandler {
	return &AnnouncementHandler{repo: repo}
}

func saveUploadedFile(r *http.Request, field string, uploadDir string) (string, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		return "", nil
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("creating upload directory: %w", err)
	}

	dst, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		return "", fmt.Errorf("creating file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	return filename, nil
}

func (h *AnnouncementHandler) CreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		writeError(w, http.StatusBadRequest, "title and content are required")
		return
	}

	imagePath := ""
	filename, err := saveUploadedFile(r, "image", filepath.Join(os.Getenv("MEDIA_ROOT"), "images/announcements"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filename != "" {
		imagePath = "/media/images/announcements/" + filename
	}

	announcement := domain.Announcement{
		Title:       title,
		Content:     content,
		ImagePath:   imagePath,
		PublisherID: member.ID,
	}

	if err := h.repo.Create(context.Background(), &announcement); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.AnnouncementResponse{Announcement: announcement})
}

func (h *AnnouncementHandler) ListAnnouncements(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	announcements, total, err := h.repo.List(context.Background(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.AnnouncementListResponse{
		Announcements: announcements,
		Total:         total,
		Page:          page,
		Limit:         limit,
	})
}

func (h *AnnouncementHandler) GetAnnouncement(w http.ResponseWriter, r *http.Request) {
	announcementID := mux.Vars(r)["announcementId"]

	announcement, err := h.repo.GetByID(context.Background(), announcementID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.AnnouncementResponse{Announcement: *announcement})
}

func (h *AnnouncementHandler) UpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	announcementID := mux.Vars(r)["announcementId"]

	announcement, err := h.repo.GetByID(context.Background(), announcementID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	imagePath := r.FormValue("image_path")

	if title != "" {
		announcement.Title = title
	}
	if content != "" {
		announcement.Content = content
	}

	filename, err := saveUploadedFile(r, "image", filepath.Join(os.Getenv("MEDIA_ROOT"), "images/announcements"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filename != "" {
		announcement.ImagePath = "/media/images/announcements/" + filename
	} else if imagePath != "" {
		announcement.ImagePath = imagePath
	}

	if err := h.repo.Update(context.Background(), announcement); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.AnnouncementResponse{Announcement: *announcement})
}

func (h *AnnouncementHandler) DeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	announcementID := mux.Vars(r)["announcementId"]

	if err := h.repo.Delete(context.Background(), announcementID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "announcement deleted"})
}
