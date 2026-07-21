//go:build integration

package database_test

import (
	"context"
	"fmt"
	"testing"

	"backend/src/database"
	"backend/src/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func cleanupChat(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE read_receipts, chat_messages, chat_room_members, chat_rooms CASCADE")
	require.NoError(t, err)
	_, err = testPool.Exec(context.Background(), "DELETE FROM members WHERE id != '00000000-0000-0000-0000-000000000000'")
	require.NoError(t, err)
}

var chatMemberCounter int

func createTestMemberForChat(t *testing.T, repo *database.MemberRepositoryPGX) domain.Member {
	t.Helper()
	chatMemberCounter++
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	m := domain.Member{
		Email:    fmt.Sprintf("chat_test_%d@example.com", chatMemberCounter),
		Password: string(hash),
		Name:     "Test User",
	}
	err = repo.Create(context.Background(), &m)
	require.NoError(t, err)
	return m
}

func createTestRoom(t *testing.T, chatRepo *database.ChatRoomRepositoryPGX, _ *database.MemberRepositoryPGX, creator *domain.Member, memberIDs []string) domain.ChatRoom {
	t.Helper()
	room := domain.ChatRoom{
		Name:      "Test Room",
		Type:      "group",
		CreatedBy: creator.ID,
	}
	err := chatRepo.CreateRoom(context.Background(), &room, memberIDs)
	require.NoError(t, err)
	return room
}

func TestChatRoomRepositoryPGX_CreateRoom(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	member2 := createTestMemberForChat(t, memberRepo)

	t.Run("create group room with members", func(t *testing.T) {
		room := domain.ChatRoom{
			Name:      "Test Group",
			Type:      "group",
			CreatedBy: creator.ID,
		}
		err := chatRepo.CreateRoom(context.Background(), &room, []string{member2.ID})
		assert.NoError(t, err)
		assert.NotEmpty(t, room.ID)
		assert.False(t, room.CreatedAt.IsZero())
		assert.False(t, room.UpdatedAt.IsZero())
	})
}

func TestChatRoomRepositoryPGX_GetRoomByID(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, nil)

	t.Run("existing room", func(t *testing.T) {
		got, err := chatRepo.GetRoomByID(context.Background(), room.ID, creator.ID)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, room.ID, got.ID)
		assert.Equal(t, room.Name, got.Name)
		assert.Equal(t, room.Type, got.Type)
		assert.Equal(t, room.CreatedBy, got.CreatedBy)
		assert.True(t, got.IsMember)
	})
}

func TestChatRoomRepositoryPGX_GetRoomByID_NotFound(t *testing.T) {
	defer cleanupChat(t)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	t.Run("non-existent room", func(t *testing.T) {
		got, err := chatRepo.GetRoomByID(context.Background(), "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000")
		assert.Nil(t, got)
		assert.ErrorIs(t, err, domain.ErrChatRoomNotFound)
	})
}

func TestChatRoomRepositoryPGX_ListRoomsByMember(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)

	t.Run("list rooms for member with 2 rooms", func(t *testing.T) {
		createTestRoom(t, chatRepo, memberRepo, &creator, nil)
		createTestRoom(t, chatRepo, memberRepo, &creator, nil)

		rooms, err := chatRepo.ListRoomsByMember(context.Background(), creator.ID)
		assert.NoError(t, err)
		assert.Len(t, rooms, 2)
	})
}

func TestChatRoomRepositoryPGX_SendMessage(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, nil)

	t.Run("send message and verify fields", func(t *testing.T) {
		msg := domain.ChatMessage{
			RoomID:   room.ID,
			SenderID: creator.ID,
			Content:  "Hello, world!",
		}
		err := chatRepo.SendMessage(context.Background(), &msg)
		assert.NoError(t, err)
		assert.NotEmpty(t, msg.ID)
		assert.False(t, msg.CreatedAt.IsZero())
	})
}

func TestChatRoomRepositoryPGX_ListMessages(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, nil)

	t.Run("list messages returns messages in DESC order", func(t *testing.T) {
		msg1 := domain.ChatMessage{
			RoomID:   room.ID,
			SenderID: creator.ID,
			Content:  "First message",
		}
		err := chatRepo.SendMessage(context.Background(), &msg1)
		require.NoError(t, err)

		msg2 := domain.ChatMessage{
			RoomID:   room.ID,
			SenderID: creator.ID,
			Content:  "Second message",
		}
		err = chatRepo.SendMessage(context.Background(), &msg2)
		require.NoError(t, err)

		messages, err := chatRepo.ListMessages(context.Background(), room.ID, "", 10)
		assert.NoError(t, err)
		assert.Len(t, messages, 2)
		assert.Equal(t, "Second message", messages[0].Content)
		assert.Equal(t, "First message", messages[1].Content)
	})
}

func TestChatRoomRepositoryPGX_DeleteMessage(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, nil)

	t.Run("delete existing message", func(t *testing.T) {
		msg := domain.ChatMessage{
			RoomID:   room.ID,
			SenderID: creator.ID,
			Content:  "To be deleted",
		}
		err := chatRepo.SendMessage(context.Background(), &msg)
		require.NoError(t, err)

		err = chatRepo.DeleteMessage(context.Background(), msg.ID)
		assert.NoError(t, err)
	})
}

func TestChatRoomRepositoryPGX_MarkAsRead(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	member2 := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, []string{member2.ID})

	t.Run("mark message as read", func(t *testing.T) {
		msg := domain.ChatMessage{
			RoomID:   room.ID,
			SenderID: creator.ID,
			Content:  "Read me",
		}
		err := chatRepo.SendMessage(context.Background(), &msg)
		require.NoError(t, err)

		err = chatRepo.MarkAsRead(context.Background(), msg.ID, member2.ID)
		assert.NoError(t, err)

		readBy, err := chatRepo.GetReadBy(context.Background(), msg.ID)
		assert.NoError(t, err)
		assert.Len(t, readBy, 1)
		assert.Equal(t, member2.ID, readBy[0].MemberID)
		assert.False(t, readBy[0].ReadAt.IsZero())
	})
}

func TestChatRoomRepositoryPGX_CountUnread(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	member2 := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, []string{member2.ID})

	t.Run("unread count for other member", func(t *testing.T) {
		msg := domain.ChatMessage{
			RoomID:   room.ID,
			SenderID: creator.ID,
			Content:  "Unread message",
		}
		err := chatRepo.SendMessage(context.Background(), &msg)
		require.NoError(t, err)

		count, err := chatRepo.CountUnread(context.Background(), room.ID, member2.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		countCreator, err := chatRepo.CountUnread(context.Background(), room.ID, creator.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), countCreator)
	})
}

func TestChatRoomRepositoryPGX_UpdateRoom(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, nil)

	t.Run("update room name", func(t *testing.T) {
		err := chatRepo.UpdateRoom(context.Background(), room.ID, "Updated Room Name")
		assert.NoError(t, err)
	})
}

func TestChatRoomRepositoryPGX_DeleteRoom(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, nil)

	t.Run("delete existing room", func(t *testing.T) {
		err := chatRepo.DeleteRoom(context.Background(), room.ID)
		assert.NoError(t, err)
	})
}

func TestChatRoomRepositoryPGX_AddMembers(t *testing.T) {
	defer cleanupChat(t)
	memberRepo := database.NewMemberRepositoryPGX(testPool)
	chatRepo := database.NewChatRoomRepositoryPGX(testPool)

	creator := createTestMemberForChat(t, memberRepo)
	member2 := createTestMemberForChat(t, memberRepo)
	member3 := createTestMemberForChat(t, memberRepo)
	room := createTestRoom(t, chatRepo, memberRepo, &creator, []string{member2.ID})

	t.Run("add member then list their rooms", func(t *testing.T) {
		err := chatRepo.AddMembers(context.Background(), room.ID, []string{member3.ID})
		assert.NoError(t, err)

		rooms, err := chatRepo.ListRoomsByMember(context.Background(), member3.ID)
		assert.NoError(t, err)
		assert.Len(t, rooms, 1)
		assert.Equal(t, room.ID, rooms[0].ID)
	})
}
