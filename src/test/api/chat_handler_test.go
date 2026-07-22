//go:build integration

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/src/api"
	"backend/src/database"
	"backend/src/domain"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupChatTest() (*database.ChatRoomRepositoryPGX, *database.MemberRepositoryPGX, *database.SessionCache, *api.ChatRoomHandler) {
	repo := database.NewChatRoomRepositoryPGX(testPool)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	sessionCache := database.NewSessionCache(24 * time.Hour)
	handler := api.NewChatRoomHandler(repo)
	return repo, memberRepo, sessionCache, handler
}

func createChatMember(t *testing.T, memberRepo *database.MemberRepositoryPGX, sessionCache *database.SessionCache) *domain.Member {
	t.Helper()

	member := domain.Member{
		Email:    "chat-" + t.Name() + "-" + uuid.New().String()[:8] + "@example.com",
		Password: "password",
		Name:     "Chat User",
	}
	err := memberRepo.Create(context.Background(), &member)
	require.NoError(t, err)

	session := domain.Session{MemberID: member.ID}
	err = sessionCache.Create(context.Background(), &session)
	require.NoError(t, err)

	return &member
}

func cleanupChat(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE read_receipts, chat_messages, chat_room_members, chat_rooms CASCADE")
	require.NoError(t, err)
}

func TestChatHandler_CreateRoom(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	body, _ := json.Marshal(domain.CreateRoomRequest{
		Name: "Group Chat",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.CreateRoom(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp domain.RoomResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Room.ID)
	assert.Equal(t, "Group Chat", resp.Room.Name)
	assert.True(t, resp.Room.IsMember)

	_ = chatRepo
}



func TestChatHandler_ListRooms(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	for _, name := range []string{"Room A", "Room B"} {
		err := chatRepo.CreateRoom(context.Background(), &domain.ChatRoom{
			Name: name, CreatedBy: member.ID,
		})
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/chat/rooms", nil)
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.ListRooms(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.RoomListResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Rooms, 2)
}

func TestChatHandler_GetRoom(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	room := domain.ChatRoom{
		Name: "Test Room", CreatedBy: member.ID,
	}
	err := chatRepo.CreateRoom(context.Background(), &room)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/chat/rooms/"+room.ID, nil)
	req = mux.SetURLVars(req, map[string]string{"roomId": room.ID})
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.GetRoom(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.RoomResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, room.ID, resp.Room.ID)
	assert.Equal(t, "Test Room", resp.Room.Name)
}

func TestChatHandler_GetRoom_NotFound(t *testing.T) {
	defer cleanupChat(t)
	_, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	req := httptest.NewRequest(http.MethodGet, "/api/chat/rooms/non-existent", nil)
	req = mux.SetURLVars(req, map[string]string{"roomId": "00000000-0000-0000-0000-000000000000"})
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.GetRoom(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestChatHandler_SendMessage(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	room := domain.ChatRoom{
		Name: "Test Room", CreatedBy: member.ID,
	}
	err := chatRepo.CreateRoom(context.Background(), &room)
	require.NoError(t, err)

	var b bytes.Buffer
	mw := multipart.NewWriter(&b)
	mw.WriteField("content", "Hello, World!")
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/chat/rooms/"+room.ID+"/messages", &b)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req = mux.SetURLVars(req, map[string]string{"roomId": room.ID})
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.SendMessage(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp domain.MessageResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message.ID)
	assert.Equal(t, "Hello, World!", resp.Message.Content)
	assert.Equal(t, member.ID, resp.Message.SenderID)
}

func TestChatHandler_ListMessages(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	room := domain.ChatRoom{
		Name: "Test Room", CreatedBy: member.ID,
	}
	err := chatRepo.CreateRoom(context.Background(), &room)
	require.NoError(t, err)

	chatRepo.SendMessage(context.Background(), &domain.ChatMessage{
		RoomID: room.ID, SenderID: member.ID, Content: "Message 1",
	})
	chatRepo.SendMessage(context.Background(), &domain.ChatMessage{
		RoomID: room.ID, SenderID: member.ID, Content: "Message 2",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/chat/rooms/"+room.ID+"/messages", nil)
	req = mux.SetURLVars(req, map[string]string{"roomId": room.ID})
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.ListMessages(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp domain.MessageListResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Messages, 2)
}

func TestChatHandler_DeleteMessage(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	room := domain.ChatRoom{
		Name: "Test Room", CreatedBy: member.ID,
	}
	err := chatRepo.CreateRoom(context.Background(), &room)
	require.NoError(t, err)

	msg := domain.ChatMessage{
		RoomID: room.ID, SenderID: member.ID, Content: "To be deleted",
	}
	err = chatRepo.SendMessage(context.Background(), &msg)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/chat/messages/"+msg.ID+"/delete", nil)
	req = mux.SetURLVars(req, map[string]string{"messageId": msg.ID})
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.DeleteMessage(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChatHandler_MarkAsRead(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()
	member := createChatMember(t, memberRepo, sessionCache)

	other := domain.Member{
		Email:    "markread-other-" + t.Name() + "-" + uuid.New().String()[:8] + "@example.com",
		Password: "password",
		Name:     "Other Read",
	}
	err := memberRepo.Create(context.Background(), &other)
	require.NoError(t, err)

	room := domain.ChatRoom{
		Name: "Test Room", CreatedBy: other.ID,
	}
	err = chatRepo.CreateRoom(context.Background(), &room)
	require.NoError(t, err)
	err = chatRepo.AddMembers(context.Background(), room.ID, []string{member.ID})
	require.NoError(t, err)

	msg := domain.ChatMessage{
		RoomID: room.ID, SenderID: other.ID, Content: "Read this",
	}
	err = chatRepo.SendMessage(context.Background(), &msg)
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{"message_id": msg.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/chat/messages/read", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(api.ContextWithMember(req.Context(), member))
	w := httptest.NewRecorder()
	handler.MarkAsRead(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func cleanupMembersFull(t *testing.T) {
	t.Helper()
	_, _ = testPool.Exec(context.Background(), "TRUNCATE TABLE read_receipts, chat_messages, chat_room_members, chat_rooms CASCADE")
	_, err := testPool.Exec(context.Background(), "DELETE FROM members WHERE id != '00000000-0000-0000-0000-000000000000'")
	require.NoError(t, err)
}

func TestChatHandler_ListAvailableMembers(t *testing.T) {
	defer cleanupChat(t)
	chatRepo, memberRepo, sessionCache, handler := setupChatTest()

	t.Run("list available members", func(t *testing.T) {
		cleanupMembersFull(t)

		member := createChatMember(t, memberRepo, sessionCache)
		memberRepo.Create(context.Background(), &domain.Member{
			Email:    "other1@" + uuid.New().String()[:8] + ".com",
			Password: "pw",
			Name:     "Other One",
		})
		memberRepo.Create(context.Background(), &domain.Member{
			Email:    "other2@" + uuid.New().String()[:8] + ".com",
			Password: "pw",
			Name:     "Other Two",
		})

		room := domain.ChatRoom{Name: "Test", CreatedBy: member.ID}
		require.NoError(t, chatRepo.CreateRoom(context.Background(), &room))

		body, _ := json.Marshal(domain.RoomMembersRequest{})
		req := httptest.NewRequest(http.MethodPost, "/api/chat/rooms/"+room.ID+"/available-members", bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"roomId": room.ID})
		req = req.WithContext(api.ContextWithMember(req.Context(), member))
		w := httptest.NewRecorder()
		handler.ListAvailableMembers(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp domain.MembersListResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		// System + Other1 + Other2 = 3 (member is in the room, excluded)
		assert.Equal(t, 3, resp.Total)
	})

	t.Run("unauthorized", func(t *testing.T) {
		cleanupChat(t)
		body, _ := json.Marshal(domain.RoomMembersRequest{})
		req := httptest.NewRequest(http.MethodPost, "/api/chat/rooms/some-room/available-members", bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"roomId": "some-room"})
		w := httptest.NewRecorder()
		handler.ListAvailableMembers(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
