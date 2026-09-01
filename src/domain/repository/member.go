package repository

import (
	"backend/src/domain/model"
	"context"
	"time"
)

type Member interface {
	Create(context.Context, *model.Member) error
	GetByEmail(context.Context, string) (*model.Member, error)
	GetByID(context.Context, string) (*model.Member, error)
	Update(context.Context, *model.Member) error
	UpdatePassword(context.Context, string, string) error
}

type MemberPermission interface {
	UpdatePermission(context.Context, string, string) error
}
type Session interface {
	Create(context.Context, *model.Session) error
	GetByKey(context.Context, string) (*model.Session, error)
	Delete(context.Context, string) error
}
type SessionOperations interface {
	Session
	DeleteByMemberID(context.Context, string) error
}

type RegistrationCode interface {
	RegisterMemberWithCode(context.Context, *model.Member, string) error
	Create(context.Context, string, string) (*model.RegistrationCode, error)
	List(context.Context) ([]model.RegistrationCode, error)
	Delete(context.Context, string) (bool, error)
}

type Category interface {
	List(context.Context) ([]model.Category, error)
	Create(context.Context, string) (*model.Category, error)
	Update(context.Context, string, string) (bool, error)
	Delete(context.Context, string) (bool, error)
}

type Inventory interface {
	CreateInventory(context.Context, *model.Inventory) error
	GetInventoryByID(context.Context, string) (*model.Inventory, error)
	GetInventoryByPriceID(context.Context, string) (*model.Inventory, error)
	ListInventories(context.Context) ([]model.Inventory, error)
	UpdateInventory(context.Context, *model.Inventory) error
	DeleteInventory(context.Context, string) error
	CreateItem(context.Context, *model.InventoryItem) error
	GetItemByID(context.Context, string) (*model.InventoryItem, error)
	ListItemsByInventoryID(context.Context, string) ([]model.InventoryItem, error)
	UpdateItem(context.Context, *model.InventoryItem) error
	DeleteItem(context.Context, string) error
}

type Order interface {
	Create(context.Context, *model.Order, []model.OrderItem) error
	GetByID(context.Context, string, string, bool) (*model.Order, error)
	List(context.Context, string, string, int, int, bool) ([]model.Order, int, error)
	Cancel(context.Context, string, string, bool) error
	UpdateStatus(context.Context, string, string, string) (*model.Order, error)
	History(context.Context, string, string, bool) ([]model.OrderStatusHistory, error)
}

type Payment interface {
	Pay(context.Context, string, string, string, string, string, time.Time) (*model.Payment, error)
	ExpirePending(context.Context, time.Time, time.Time) (int, error)
}

type Announcement interface {
	Create(context.Context, *model.Announcement) error
	GetByID(context.Context, string) (*model.Announcement, error)
	List(context.Context, int, int) ([]model.Announcement, int, error)
	ListByMonth(context.Context, int, int, int, int) ([]model.Announcement, int, error)
	Update(context.Context, *model.Announcement) error
	Delete(context.Context, string) error
}

type Event interface {
	GetByID(context.Context, string, string) (*model.Event, error)
	Update(context.Context, *model.Event) error
	Delete(context.Context, string) error
}
type EventOperations interface {
	Event
	Create(context.Context, *model.Event) error
	GetByID(context.Context, string, string) (*model.Event, error)
	ListByMonth(context.Context, int, int, string) ([]model.Event, error)
	Update(context.Context, *model.Event) error
	Delete(context.Context, string) error
	AddViewer(context.Context, string, string) error
	RemoveViewer(context.Context, string, string) error
	ListViewers(context.Context, string) ([]model.EventViewer, error)
}

type ChatRoom interface {
	CreateRoom(context.Context, *model.ChatRoom) error
	GetRoomByID(context.Context, string, string) (*model.ChatRoomWithMeta, error)
	ListRoomsByMember(context.Context, string) ([]model.ChatRoomWithMeta, error)
	ListRoomsByMemberByMonth(context.Context, string, int, int) ([]model.ChatRoomWithMeta, error)
	UpdateRoom(context.Context, string, string) error
	DeleteRoom(context.Context, string) error
	AddMembers(context.Context, string, []string) error
	RemoveMember(context.Context, string, string) error
	SendMessage(context.Context, *model.ChatMessage) error
	ListMessages(context.Context, string, string, int) ([]model.ChatMessage, error)
	DeleteMessage(context.Context, string) error
	MarkAsRead(context.Context, string, string) error
	GetReadBy(context.Context, string) ([]model.ReadReceipt, error)
	CountUnread(context.Context, string, string) (int64, error)
	ListMembersNotInRoom(context.Context, string, int, int) ([]model.Member, error)
	CountMembersNotInRoom(context.Context, string) (int, error)
}
