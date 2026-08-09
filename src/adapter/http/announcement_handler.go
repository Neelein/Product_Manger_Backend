package http

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"backend/src/usecase"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AnnouncementHandler struct {
	repo usecase.AnnouncementService
}

func NewAnnouncementHandler(service usecase.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{repo: service}
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
		imagePath = os.Getenv("API_DOMAIN") + "/media/images/announcements/" + filename
	}

	announcement := Announcement{
		Title:       title,
		Content:     content,
		ImagePath:   imagePath,
		PublisherID: member.ID,
	}

	if err := h.repo.Create(r.Context(), &announcement); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, AnnouncementResponse{Announcement: announcement})
}

func (h *AnnouncementHandler) ListAnnouncements(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

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

	// If year and month are provided, filter by month
	if yearStr != "" && monthStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}

		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}

		announcements, total, err := h.repo.ListByMonth(r.Context(), year, month, limit, offset)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, AnnouncementListResponse{
			Announcements: announcements,
			Total:         total,
			Page:          page,
			Limit:         limit,
		})
		return
	}

	announcements, total, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AnnouncementListResponse{
		Announcements: announcements,
		Total:         total,
		Page:          page,
		Limit:         limit,
	})
}

func (h *AnnouncementHandler) GetAnnouncement(w http.ResponseWriter, r *http.Request) {
	announcementID := mux.Vars(r)["announcementId"]

	announcement, err := h.repo.GetByID(r.Context(), announcementID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AnnouncementResponse{Announcement: *announcement})
}

func (h *AnnouncementHandler) UpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	announcementID := mux.Vars(r)["announcementId"]

	announcement, err := h.repo.GetByID(r.Context(), announcementID)
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
		announcement.ImagePath = os.Getenv("API_DOMAIN") + "/media/images/announcements/" + filename
	} else if imagePath != "" {
		announcement.ImagePath = imagePath
	}

	if err := h.repo.Update(r.Context(), announcement); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AnnouncementResponse{Announcement: *announcement})
}

func (h *AnnouncementHandler) DeleteAnnouncement(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	announcementID := mux.Vars(r)["announcementId"]

	if err := h.repo.Delete(r.Context(), announcementID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "announcement deleted"})
}
