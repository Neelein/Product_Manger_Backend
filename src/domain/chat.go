package domain

import "time"

type ChatRoom struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChatMessage struct {
	ID         string    `json:"id"`
	RoomID     string    `json:"room_id"`
	SenderID   string    `json:"sender_id"`
	SenderName string    `json:"sender_name"`
	Content    string    `json:"content"`
	ImagePath  string    `json:"image_path"`
	FilePath   string    `json:"file_path"`
	CreatedAt  time.Time `json:"created_at"`
}

type ChatRoomMember struct {
	RoomID   string    `json:"room_id"`
	MemberID string    `json:"member_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type ReadReceipt struct {
	MessageID  string    `json:"message_id"`
	MemberID   string    `json:"member_id"`
	MemberName string    `json:"member_name,omitempty"`
	ReadAt     time.Time `json:"read_at"`
}

type ChatRoomWithMeta struct {
	ChatRoom
	IsMember          bool       `json:"is_member"`
	LastMessage       string     `json:"last_message_content,omitempty"`
	LastMessageAt     *time.Time `json:"last_message_at,omitempty"`
	LastMessageSender string     `json:"last_message_sender_id,omitempty"`
	UnreadCount       int64      `json:"unread_count"`
}

type CreateRoomRequest struct {
	Name string `json:"name"`
}

type SendMessageRequest struct {
	Content   string `json:"content"`
	ImagePath string `json:"image_path"`
	FilePath  string `json:"file_path"`
}

type UpdateRoomRequest struct {
	Name string `json:"name"`
}

type RoomResponse struct {
	Room ChatRoomWithMeta `json:"room"`
}

type RoomListResponse struct {
	Rooms []ChatRoomWithMeta `json:"rooms"`
}

type MessageResponse struct {
	Message ChatMessage `json:"message"`
}

type MessageListResponse struct {
	Messages []ChatMessage `json:"messages"`
}

type ReadByResponse struct {
	ReadBy []ReadReceipt `json:"read_by"`
}

type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
