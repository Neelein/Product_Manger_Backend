package http

import (
	"bytes"
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
	service usecase.AnnouncementService
}

func NewAnnouncementHandler(service usecase.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{service: service}
}

func prepareUploadedFile(r *http.Request, field string, uploadDir string) (string, *usecase.UploadInput, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		return "", nil, nil
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext

	content, err := io.ReadAll(file)
	if err != nil {
		return "", nil, fmt.Errorf("reading file: %w", err)
	}
	return filename, &usecase.UploadInput{Directory: uploadDir, Filename: filename, Content: bytes.NewReader(content)}, nil
}

func prepareAnnouncementImage(r *http.Request, uploadDir string) (string, *usecase.UploadInput, error) {
	file, _, err := r.FormFile("image")
	if err != nil {
		return "", nil, nil
	}
	defer file.Close()

	content, err := io.ReadAll(io.LimitReader(file, 10<<20+1))
	if err != nil {
		return "", nil, fmt.Errorf("reading image: %w", err)
	}
	if len(content) > 10<<20 {
		return "", nil, fmt.Errorf("announcement image must be 10 MB or smaller")
	}

	ext := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}[http.DetectContentType(content)]
	if ext == "" {
		return "", nil, fmt.Errorf("only JPEG, PNG, and WebP images are supported")
	}

	filename := uuid.NewString() + ext
	return filename, &usecase.UploadInput{
		Directory: uploadDir,
		Filename:  filename,
		Content:   bytes.NewReader(content),
	}, nil
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

	imagePath := ""
	filename, upload, err := prepareAnnouncementImage(r, filepath.Join(os.Getenv("MEDIA_ROOT"), "images/announcements"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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

	if upload != nil {
		if err := h.service.CreateApplication(r.Context(), &announcement, *upload); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else if err := h.service.CreateApplication(r.Context(), &announcement); err != nil {
		if err == usecase.ErrInvalidAnnouncement {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
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

	page := 0
	limit := 0
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	var year, month *int
	if yearStr != "" && monthStr != "" {
		y, err := strconv.Atoi(yearStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}

		m, err := strconv.Atoi(monthStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}
		year, month = &y, &m
	}
	result, err := h.service.ListPage(r.Context(), page, limit, year, month)
	if err != nil {
		if err == usecase.ErrEventMonthInvalid {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AnnouncementListResponse{
		Announcements: result.Announcements,
		Total:         result.Total,
		Page:          result.Page,
		Limit:         result.Limit,
	})
}

func (h *AnnouncementHandler) GetAnnouncement(w http.ResponseWriter, r *http.Request) {
	announcementID := mux.Vars(r)["announcementId"]

	announcement, err := h.service.GetByID(r.Context(), announcementID)
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

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	title := r.FormValue("title")
	content := r.FormValue("content")
	imagePath := r.FormValue("image_path")

	filename, upload, err := prepareAnnouncementImage(r, filepath.Join(os.Getenv("MEDIA_ROOT"), "images/announcements"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if filename != "" {
		imagePath = os.Getenv("API_DOMAIN") + "/media/images/announcements/" + filename
	}
	input := usecase.AnnouncementUpdateInput{Title: title, Content: content, ImagePath: imagePath}
	var announcement *Announcement
	if upload != nil {
		announcement, err = h.service.UpdateAnnouncementApplication(r.Context(), announcementID, input, *upload)
	} else {
		announcement, err = h.service.UpdateAnnouncementApplication(r.Context(), announcementID, input)
	}
	if err != nil {
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

	if err := h.service.DeleteAnnouncementApplication(r.Context(), announcementID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "announcement deleted"})
}
