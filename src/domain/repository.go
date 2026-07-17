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

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByKey(ctx context.Context, sessionKey string) (*Session, error)
	Rotate(ctx context.Context, oldSessionKey string, fingerprint string) (*Session, error)
	Delete(ctx context.Context, sessionKey string) error
	DeleteByMemberID(ctx context.Context, memberID string) error
}
