package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository = ProductRepositoryPGX
type InventoryRepository = InventoryRepositoryPGX
type MemberRepository = MemberRepositoryPGX
type RegistrationCodeRepository = RegistrationCodeRepositoryPGX
type CategoryRepository = CategoryRepositoryPGX
type AnnouncementRepository = AnnouncementRepositoryPGX
type ChatRoomRepository = ChatRoomRepositoryPGX
type EventRepository = EventRepositoryPGX
type OrderRepository = OrderRepositoryPGX

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return NewProductRepositoryPGX(pool)
}
func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return NewInventoryRepositoryPGX(pool)
}
func NewMemberRepository(pool *pgxpool.Pool) *MemberRepository {
	return NewMemberRepositoryPGX(pool)
}
func NewRegistrationCodeRepository(pool *pgxpool.Pool) *RegistrationCodeRepository {
	return NewRegistrationCodeRepositoryPGX(pool)
}
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return NewCategoryRepositoryPGX(pool)
}
func NewAnnouncementRepository(pool *pgxpool.Pool) *AnnouncementRepository {
	return NewAnnouncementRepositoryPGX(pool)
}
func NewChatRoomRepository(pool *pgxpool.Pool) *ChatRoomRepository {
	return NewChatRoomRepositoryPGX(pool)
}
func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return NewEventRepositoryPGX(pool)
}
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository { return NewOrderRepositoryPGX(pool) }
