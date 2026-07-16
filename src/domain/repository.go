package domain

import "context"

type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	List(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id string) (*Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id string) error
}
