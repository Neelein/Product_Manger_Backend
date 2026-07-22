package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRoomRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewChatRoomRepositoryPGX(pool *pgxpool.Pool) *ChatRoomRepositoryPGX {
	return &ChatRoomRepositoryPGX{pool: pool}
}

func (r *ChatRoomRepositoryPGX) CreateRoom(ctx context.Context, room *domain.ChatRoom) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_chat_room($1, $2)",
		room.Name, room.CreatedBy,
	).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating chat room: %w", err)
	}

	_, err = r.pool.Exec(ctx, "SELECT * FROM add_chat_room_members($1, $2::uuid[])", room.ID, []string{room.CreatedBy})
	if err != nil {
		return fmt.Errorf("adding creator to chat room: %w", err)
	}

	return nil
}

func (r *ChatRoomRepositoryPGX) GetRoomByID(ctx context.Context, roomID string, memberID string) (*domain.ChatRoomWithMeta, error) {
	var cr domain.ChatRoomWithMeta
	err := r.pool.QueryRow(ctx, "SELECT * FROM get_chat_room_by_id($1, $2)", roomID, memberID).Scan(&cr.ID, &cr.Name, &cr.CreatedBy, &cr.CreatedAt, &cr.UpdatedAt, &cr.IsMember)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrChatRoomNotFound
		}
		return nil, fmt.Errorf("getting chat room by ID: %w", err)
	}
	return &cr, nil
}

func (r *ChatRoomRepositoryPGX) ListRoomsByMember(ctx context.Context, memberID string) ([]domain.ChatRoomWithMeta, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_chat_rooms_by_member($1)", memberID)
	if err != nil {
		return nil, fmt.Errorf("listing chat rooms by member: %w", err)
	}

	rooms, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ChatRoomWithMeta, error) {
		var cr domain.ChatRoomWithMeta
		var lastMessageContent *string
		var lastMessageSenderID *string
		var lastMessageCreatedAt *time.Time

		err := row.Scan(
			&cr.ID, &cr.Name, &cr.CreatedBy,
			&cr.CreatedAt, &cr.UpdatedAt,
			&lastMessageContent, &lastMessageSenderID, &lastMessageCreatedAt,
			&cr.UnreadCount,
		)
		if err != nil {
			return cr, err
		}

		if lastMessageContent != nil {
			cr.LastMessage = *lastMessageContent
		}
		if lastMessageSenderID != nil {
			cr.LastMessageSender = *lastMessageSenderID
		}
		cr.LastMessageAt = lastMessageCreatedAt

		return cr, nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing chat rooms by member: %w", err)
	}

	if rooms == nil {
		rooms = []domain.ChatRoomWithMeta{}
	}

	return rooms, nil
}

func (r *ChatRoomRepositoryPGX) UpdateRoom(ctx context.Context, roomID string, name string) error {
	var updatedAt time.Time
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_chat_room($1, $2)", roomID, name).Scan(&updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrChatRoomNotFound
		}
		return fmt.Errorf("updating chat room: %w", err)
	}
	return nil
}

func (r *ChatRoomRepositoryPGX) DeleteRoom(ctx context.Context, roomID string) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_chat_room($1)", roomID).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting chat room: %w", err)
	}
	if !deleted {
		return domain.ErrChatRoomNotFound
	}
	return nil
}

func (r *ChatRoomRepositoryPGX) AddMembers(ctx context.Context, roomID string, memberIDs []string) error {
	_, err := r.pool.Exec(ctx, "SELECT * FROM add_chat_room_members($1, $2::uuid[])", roomID, memberIDs)
	if err != nil {
		return fmt.Errorf("adding chat room members: %w", err)
	}
	return nil
}

func (r *ChatRoomRepositoryPGX) RemoveMember(ctx context.Context, roomID string, memberID string) error {
	var removed bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM remove_chat_room_member($1, $2)", roomID, memberID).Scan(&removed)
	if err != nil {
		return fmt.Errorf("removing chat room member: %w", err)
	}
	if !removed {
		return domain.ErrChatRoomNotFound
	}
	return nil
}

func (r *ChatRoomRepositoryPGX) SendMessage(ctx context.Context, msg *domain.ChatMessage) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM send_message($1, $2, $3, $4, $5)",
		msg.RoomID, msg.SenderID, msg.Content, msg.ImagePath, msg.FilePath,
	).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}
	return nil
}

func (r *ChatRoomRepositoryPGX) ListMessages(ctx context.Context, roomID string, beforeID string, limit int) ([]domain.ChatMessage, error) {
	var beforeIDParam interface{}
	if beforeID == "" {
		beforeIDParam = nil
	} else {
		beforeIDParam = beforeID
	}

	rows, err := r.pool.Query(ctx, "SELECT * FROM list_messages($1, $2, $3)", roomID, beforeIDParam, limit)
	if err != nil {
		return nil, fmt.Errorf("listing messages: %w", err)
	}

	messages, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ChatMessage, error) {
		var msg domain.ChatMessage
		err := row.Scan(
			&msg.ID, &msg.RoomID, &msg.SenderID,
			&msg.Content, &msg.ImagePath, &msg.FilePath,
			&msg.CreatedAt, &msg.SenderName,
		)
		return msg, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing messages: %w", err)
	}

	if messages == nil {
		messages = []domain.ChatMessage{}
	}

	return messages, nil
}

func (r *ChatRoomRepositoryPGX) DeleteMessage(ctx context.Context, messageID string) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_message($1)", messageID).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting message: %w", err)
	}
	if !deleted {
		return domain.ErrChatMessageNotFound
	}
	return nil
}

func (r *ChatRoomRepositoryPGX) MarkAsRead(ctx context.Context, messageID string, memberID string) error {
	var readAt time.Time
	err := r.pool.QueryRow(ctx, "SELECT * FROM mark_message_read($1, $2)", messageID, memberID).Scan(&readAt)
	if err != nil {
		return fmt.Errorf("marking message as read: %w", err)
	}
	return nil
}

func (r *ChatRoomRepositoryPGX) GetReadBy(ctx context.Context, messageID string) ([]domain.ReadReceipt, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM get_message_read_by($1)", messageID)
	if err != nil {
		return nil, fmt.Errorf("getting read by: %w", err)
	}

	receipts, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ReadReceipt, error) {
		var rr domain.ReadReceipt
		err := row.Scan(&rr.MemberID, &rr.MemberName, &rr.ReadAt)
		rr.MessageID = messageID
		return rr, err
	})
	if err != nil {
		return nil, fmt.Errorf("getting read by: %w", err)
	}

	if receipts == nil {
		receipts = []domain.ReadReceipt{}
	}

	return receipts, nil
}

func (r *ChatRoomRepositoryPGX) CountUnread(ctx context.Context, roomID string, memberID string) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT * FROM count_room_unread($1, $2)", roomID, memberID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting unread messages: %w", err)
	}
	return count, nil
}
