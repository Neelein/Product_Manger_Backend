package usecase

import (
	"backend/src/domain/model"
	"context"
	"testing"
)

type fakeVariants struct {
	detail  *model.ProductDetail
	options map[string]*model.ProductOption
	created *model.ProductVariant
}

func (f *fakeVariants) GetDetailByProductID(context.Context, string) (*model.ProductDetail, error) {
	return f.detail, nil
}
func (f *fakeVariants) GetOptionByID(_ context.Context, id string) (*model.ProductOption, error) {
	return f.options[id], nil
}
func (f *fakeVariants) CreateVariant(_ context.Context, v *model.ProductVariant) error {
	f.created = v
	return nil
}
func TestProductUseCaseRejectsDuplicateOptions(t *testing.T) {
	f := &fakeVariants{detail: &model.ProductDetail{ID: "d"}, options: map[string]*model.ProductOption{"o": {ID: "o", ProductDetailID: "d"}}}
	if err := NewProductUseCase(nil, f).CreateVariant(context.Background(), "p", "price", []string{"o", "o"}); err != model.ErrDuplicateProductVariant {
		t.Fatalf("err = %v", err)
	}
}
func TestProductUseCaseRejectsForeignOption(t *testing.T) {
	f := &fakeVariants{detail: &model.ProductDetail{ID: "d"}, options: map[string]*model.ProductOption{"o": {ID: "o", ProductDetailID: "other"}}}
	if err := NewProductUseCase(nil, f).CreateVariant(context.Background(), "p", "price", []string{"o"}); err == nil {
		t.Fatal("expected ownership error")
	}
}
