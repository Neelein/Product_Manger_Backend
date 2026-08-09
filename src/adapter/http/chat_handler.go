package http

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"backend/src/usecase"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ChatRoomHandler struct {
	service usecase.ChatService
}

func NewChatRoomHandler(service usecase.ChatService) *ChatRoomHandler {
	return &ChatRoomHandler{service: service}
}

func (h *ChatRoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room := ChatRoom{
		ID:        uuid.New().String(),
		Name:      req.Name,
		CreatedBy: member.ID,
	}

	roomWithMeta, err := h.service.CreateRoomApplication(r.Context(), &room, member.ID)
	if err != nil {
		if err == usecase.ErrInvalidChatName {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, RoomResponse{Room: *roomWithMeta})
}

func (h *ChatRoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	// If year and month are provided, filter by month
	if yearStr != "" && monthStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid year")
			return
		}

		month, err := strconv.Atoi(monthStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid month")
			return
		}

		rooms, err := h.service.ListRoomsByMemberByMonth(r.Context(), member.ID, year, month)
		if err != nil {
			if err == usecase.ErrEventMonthInvalid {
				writeError(w, http.StatusBadRequest, "invalid month")
				return
			}
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, RoomListResponse{Rooms: rooms})
		return
	}

	rooms, err := h.service.ListRoomsByMember(r.Context(), member.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, RoomListResponse{Rooms: rooms})
}

func (h *ChatRoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	room, err := h.service.GetRoomByID(r.Context(), roomID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, RoomResponse{Room: *room})
}

func (h *ChatRoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	var req UpdateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, err := h.service.UpdateRoomApplication(r.Context(), roomID, req.Name, member.ID)
	if err != nil {
		if err == usecase.ErrInvalidChatName {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, RoomResponse{Room: *room})
}

func (h *ChatRoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	if err := h.service.DeleteRoom(r.Context(), roomID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "room deleted"})
}

func (h *ChatRoomHandler) AddMembers(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	var req struct {
		MemberIDs []string `json:"member_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.AddMembers(r.Context(), roomID, req.MemberIDs); err != nil {
		if err == usecase.ErrInvalidChatMembers {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "members added"})
}

func (h *ChatRoomHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	vars := mux.Vars(r)
	roomID := vars["roomId"]
	memberID := vars["memberId"]

	if err := h.service.RemoveMember(r.Context(), roomID, memberID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

func (h *ChatRoomHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]
	beforeID := r.URL.Query().Get("before_id")
	limitStr := r.URL.Query().Get("limit")

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	messages, err := h.service.ListMessagesApplication(r.Context(), roomID, beforeID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, MessageListResponse{Messages: messages})
}

func (h *ChatRoomHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}

	content := r.FormValue("content")

	imagePath := ""
	filename, imageUpload, err := prepareUploadedFile(r, "image", filepath.Join(os.Getenv("MEDIA_ROOT"), "images/chat"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filename != "" {
		imagePath = os.Getenv("API_DOMAIN") + "/media/images/chat/" + filename
	}

	filePath := ""
	filename, fileUpload, err := prepareUploadedFile(r, "file", filepath.Join(os.Getenv("MEDIA_ROOT"), "files/chat"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filename != "" {
		filePath = os.Getenv("API_DOMAIN") + "/media/files/chat/" + filename
	}

	msg := ChatMessage{
		ID:         uuid.New().String(),
		RoomID:     roomID,
		SenderID:   member.ID,
		SenderName: member.Name,
		Content:    content,
		ImagePath:  imagePath,
		FilePath:   filePath,
	}

	uploads := make([]usecase.UploadInput, 0, 2)
	if imageUpload != nil {
		uploads = append(uploads, *imageUpload)
	}
	if fileUpload != nil {
		uploads = append(uploads, *fileUpload)
	}
	if err := h.service.SendMessageApplication(r.Context(), &msg, uploads...); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, MessageResponse{Message: msg})
}

func (h *ChatRoomHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	messageID := mux.Vars(r)["messageId"]

	if err := h.service.DeleteMessage(r.Context(), messageID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "message deleted"})
}

func (h *ChatRoomHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		MessageID string `json:"message_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.MarkAsReadApplication(r.Context(), req.MessageID, member.ID); err != nil {
		if err == usecase.ErrInvalidMessageID {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "marked as read"})
}

func (h *ChatRoomHandler) GetReadBy(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	messageID := mux.Vars(r)["messageId"]

	readBy, err := h.service.GetReadBy(r.Context(), messageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ReadByResponse{ReadBy: readBy})
}

func (h *ChatRoomHandler) CountUnread(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	count, err := h.service.CountUnread(r.Context(), roomID, member.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, UnreadCountResponse{UnreadCount: count})
}

func (h *ChatRoomHandler) ListAvailableMembers(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	var req RoomMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	members, total, err := h.service.ListAvailableMembers(r.Context(), roomID, req.Page, req.Limit)
	if err != nil {
		switch err {
		case usecase.ErrInvalidChatPage:
			writeError(w, http.StatusBadRequest, "invalid page parameter")
			return
		case usecase.ErrInvalidChatLimit:
			writeError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	memberResponses := make([]MemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = MemberResponse{
			ID:    m.ID,
			Email: m.Email,
			Name:  m.Name,
		}
	}

	writeJSON(w, http.StatusOK, MembersListResponse{
		Members: memberResponses,
		Total:   total,
	})
}
