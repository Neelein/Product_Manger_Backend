package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"backend/src/domain"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ChatRoomHandler struct {
	repo domain.ChatRoomRepository
}

func NewChatRoomHandler(repo domain.ChatRoomRepository) *ChatRoomHandler {
	return &ChatRoomHandler{repo: repo}
}

func (h *ChatRoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req domain.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	room := domain.ChatRoom{
		ID:        uuid.New().String(),
		Name:      req.Name,
		CreatedBy: member.ID,
	}

	if err := h.repo.CreateRoom(context.Background(), &room); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	roomWithMeta, err := h.repo.GetRoomByID(context.Background(), room.ID, member.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.RoomResponse{Room: *roomWithMeta})
}

func (h *ChatRoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	rooms, err := h.repo.ListRoomsByMember(context.Background(), member.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.RoomListResponse{Rooms: rooms})
}

func (h *ChatRoomHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	room, err := h.repo.GetRoomByID(context.Background(), roomID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.RoomResponse{Room: *room})
}

func (h *ChatRoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	var req domain.UpdateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := h.repo.UpdateRoom(context.Background(), roomID, req.Name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	room, err := h.repo.GetRoomByID(context.Background(), roomID, member.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.RoomResponse{Room: *room})
}

func (h *ChatRoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	if err := h.repo.DeleteRoom(context.Background(), roomID); err != nil {
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

	if len(req.MemberIDs) == 0 {
		writeError(w, http.StatusBadRequest, "member_ids is required")
		return
	}

	if err := h.repo.AddMembers(context.Background(), roomID, req.MemberIDs); err != nil {
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

	if err := h.repo.RemoveMember(context.Background(), roomID, memberID); err != nil {
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

	messages, err := h.repo.ListMessages(context.Background(), roomID, beforeID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.MessageListResponse{Messages: messages})
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
	filename, err := saveUploadedFile(r, "image", filepath.Join(os.Getenv("MEDIA_ROOT"), "images/chat"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filename != "" {
		imagePath = os.Getenv("API_DOMAIN") + "/media/images/chat/" + filename
	}

	filePath := ""
	filename, err = saveUploadedFile(r, "file", filepath.Join(os.Getenv("MEDIA_ROOT"), "files/chat"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filename != "" {
		filePath = os.Getenv("API_DOMAIN") + "/media/files/chat/" + filename
	}

	msg := domain.ChatMessage{
		ID:         uuid.New().String(),
		RoomID:     roomID,
		SenderID:   member.ID,
		SenderName: member.Name,
		Content:    content,
		ImagePath:  imagePath,
		FilePath:   filePath,
	}

	if err := h.repo.SendMessage(context.Background(), &msg); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, domain.MessageResponse{Message: msg})
}

func (h *ChatRoomHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	messageID := mux.Vars(r)["messageId"]

	if err := h.repo.DeleteMessage(context.Background(), messageID); err != nil {
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

	if req.MessageID == "" {
		writeError(w, http.StatusBadRequest, "message_id is required")
		return
	}

	if err := h.repo.MarkAsRead(context.Background(), req.MessageID, member.ID); err != nil {
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

	readBy, err := h.repo.GetReadBy(context.Background(), messageID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.ReadByResponse{ReadBy: readBy})
}

func (h *ChatRoomHandler) CountUnread(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	count, err := h.repo.CountUnread(context.Background(), roomID, member.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, domain.UnreadCountResponse{UnreadCount: count})
}

func (h *ChatRoomHandler) ListAvailableMembers(w http.ResponseWriter, r *http.Request) {
	member := MemberFromContext(r.Context())
	if member == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roomID := mux.Vars(r)["roomId"]

	var req domain.RoomMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	if req.Page < 1 {
		writeError(w, http.StatusBadRequest, "invalid page parameter")
		return
	}
	if req.Limit < 1 || req.Limit > 100 {
		writeError(w, http.StatusBadRequest, "invalid limit parameter")
		return
	}

	offset := (req.Page - 1) * req.Limit

	members, err := h.repo.ListMembersNotInRoom(r.Context(), roomID, req.Limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	total, err := h.repo.CountMembersNotInRoom(r.Context(), roomID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	memberResponses := make([]domain.MemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = domain.MemberResponse{
			ID:    m.ID,
			Email: m.Email,
			Name:  m.Name,
		}
	}

	writeJSON(w, http.StatusOK, domain.MembersListResponse{
		Members: memberResponses,
		Total:   total,
	})
}
