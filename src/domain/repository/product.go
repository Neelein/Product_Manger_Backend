package repository

import (
	"backend/src/domain/model"
	"context"
)

type Product interface {
	Create(context.Context, *model.Product) error
	List(context.Context) ([]model.Product, error)
	Search(context.Context, string, string) ([]model.Product, error)
	GetByID(context.Context, string) (*model.Product, error)
	Update(context.Context, *model.Product) error
	Delete(context.Context, string) error
	CreateDetail(context.Context, *model.ProductDetail) error
	GetDetailByProductID(context.Context, string) (*model.ProductDetail, error)
	UpdateDetail(context.Context, *model.ProductDetail) error
	CreatePrice(context.Context, *model.ProductPrice) error
	GetPriceByID(context.Context, string) (*model.ProductPrice, error)
	GetPricesByDetailID(context.Context, string) ([]model.ProductPrice, error)
	UpdatePrice(context.Context, *model.ProductPrice) error
	CreateOption(context.Context, *model.ProductOption) error
	GetOptionByID(context.Context, string) (*model.ProductOption, error)
	ListOptionsByDetailID(context.Context, string) ([]model.ProductOption, error)
	UpdateOption(context.Context, *model.ProductOption) error
	DeleteOption(context.Context, string) error
	CreateVariant(context.Context, *model.ProductVariant) error
	GetVariantByID(context.Context, string) (*model.ProductVariant, error)
	ListVariantsByDetailID(context.Context, string) ([]model.ProductVariant, error)
	UpdateVariant(context.Context, *model.ProductVariant) error
	DeleteVariant(context.Context, string) error
	CreateImage(context.Context, *model.ProductImage) error
	ListImages(context.Context, string) ([]model.ProductImage, error)
}

type ProductVariants interface {
	GetDetailByProductID(context.Context, string) (*model.ProductDetail, error)
	GetOptionByID(context.Context, string) (*model.ProductOption, error)
	CreateVariant(context.Context, *model.ProductVariant) error
}
