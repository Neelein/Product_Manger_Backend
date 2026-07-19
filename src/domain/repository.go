package domain

import "context"

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	List(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id string) (*Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id string) error
	CreateDetail(ctx context.Context, detail *ProductDetail) error
	GetDetailByProductID(ctx context.Context, productID string) (*ProductDetail, error)
	UpdateDetail(ctx context.Context, detail *ProductDetail) error
	CreatePrice(ctx context.Context, price *ProductPrice) error
	GetPriceByID(ctx context.Context, id string) (*ProductPrice, error)
	GetPricesByDetailID(ctx context.Context, detailID string) ([]ProductPrice, error)
	UpdatePrice(ctx context.Context, price *ProductPrice) error
}

type MemberRepository interface {
	Create(ctx context.Context, member *Member) error
	GetByEmail(ctx context.Context, email string) (*Member, error)
	GetByID(ctx context.Context, id string) (*Member, error)
	Update(ctx context.Context, member *Member) error
}

type InventoryRepository interface {
	CreateInventory(ctx context.Context, inventory *Inventory) error
	GetInventoryByID(ctx context.Context, id string) (*Inventory, error)
	GetInventoryByPriceID(ctx context.Context, priceID string) (*Inventory, error)
	ListInventories(ctx context.Context) ([]Inventory, error)
	UpdateInventory(ctx context.Context, inventory *Inventory) error
	DeleteInventory(ctx context.Context, id string) error
	CreateItem(ctx context.Context, item *InventoryItem) error
	GetItemByID(ctx context.Context, id string) (*InventoryItem, error)
	ListItemsByInventoryID(ctx context.Context, inventoryID string) ([]InventoryItem, error)
	UpdateItem(ctx context.Context, item *InventoryItem) error
	DeleteItem(ctx context.Context, id string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByKey(ctx context.Context, sessionKey string) (*Session, error)
	Rotate(ctx context.Context, oldSessionKey string, fingerprint string) (*Session, error)
	Delete(ctx context.Context, sessionKey string) error
	DeleteByMemberID(ctx context.Context, memberID string) error
}

type AnnouncementRepository interface {
	Create(ctx context.Context, announcement *Announcement) error
	GetByID(ctx context.Context, id string) (*Announcement, error)
	List(ctx context.Context, limit, offset int) ([]Announcement, int, error)
	Update(ctx context.Context, announcement *Announcement) error
	Delete(ctx context.Context, id string) error
}
