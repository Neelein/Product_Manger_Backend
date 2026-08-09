package usecase

import (
	"backend/src/domain/model"
	"context"
)

type ProductOptionUpdateInput struct{ Name, Value string }
type ProductVariantUpdateInput struct {
	ProductPriceID string
	SKU            *string
	Status         string
	OptionIDs      []string
}
type ProductDetailUpdateInput struct{ Introduction, UsageInstructions, ReturnPolicy string }
type ProductPriceUpdateInput struct {
	Label, Currency string
	Amount          float64
	SortOrder       int
}

func (s *productService) UpdateOptionApplication(ctx context.Context, detailID, optionID string, input ProductOptionUpdateInput) (*model.ProductOption, error) {
	o, err := s.GetOptionForDetail(ctx, detailID, optionID)
	if err != nil {
		return nil, err
	}
	o.Name, o.Value = input.Name, input.Value
	if err := s.Product.UpdateOption(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}
func (s *productService) DeleteOptionApplication(ctx context.Context, detailID, optionID string) error {
	o, err := s.GetOptionForDetail(ctx, detailID, optionID)
	if err != nil {
		return err
	}
	return s.Product.DeleteOption(ctx, o.ID)
}
func (s *productService) UpdateVariantApplication(ctx context.Context, detailID, variantID string, input ProductVariantUpdateInput) (*model.ProductVariant, error) {
	v, err := s.GetVariantForDetail(ctx, detailID, variantID)
	if err != nil {
		return nil, err
	}
	v.ProductPriceID, v.SKU, v.Status, v.OptionIDs = input.ProductPriceID, input.SKU, input.Status, input.OptionIDs
	if err := ValidateProductVariantIDs(v.ProductPriceID, v.OptionIDs); err != nil {
		return nil, err
	}
	if err := s.Product.UpdateVariant(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}
func (s *productService) DeleteVariantApplication(ctx context.Context, detailID, variantID string) error {
	v, err := s.GetVariantForDetail(ctx, detailID, variantID)
	if err != nil {
		return err
	}
	return s.Product.DeleteVariant(ctx, v.ID)
}
func (s *productService) UpdateDetailApplication(ctx context.Context, productID string, input ProductDetailUpdateInput) (*model.ProductDetail, error) {
	d, err := s.Product.GetDetailByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}
	d.Introduction, d.UsageInstructions, d.ReturnPolicy = input.Introduction, input.UsageInstructions, input.ReturnPolicy
	if err := s.Product.UpdateDetail(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}
func (s *productService) UpdatePriceApplication(ctx context.Context, priceID string, input ProductPriceUpdateInput) (*model.ProductPrice, error) {
	p, err := s.Product.GetPriceByID(ctx, priceID)
	if err != nil {
		return nil, err
	}
	p.Label, p.Amount, p.Currency, p.SortOrder = input.Label, input.Amount, input.Currency, input.SortOrder
	if err := s.Product.UpdatePrice(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

type InventoryUpdateInput struct{ Status string }
type InventoryItemUpdateInput struct {
	ItemCode, Status, DateAdded string
	Cost                        float64
}

func (s *inventoryService) UpdateInventoryApplication(ctx context.Context, id string, input InventoryUpdateInput) (*model.Inventory, error) {
	i, err := s.Inventory.GetInventoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	i.Status = input.Status
	if err := s.Inventory.UpdateInventory(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}
func (s *inventoryService) DeleteInventoryApplication(ctx context.Context, id string) error {
	return s.Inventory.DeleteInventory(ctx, id)
}
func (s *inventoryService) UpdateItemApplication(ctx context.Context, id string, input InventoryItemUpdateInput) (*model.InventoryItem, error) {
	i, err := s.Inventory.GetItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	i.ItemCode, i.Status, i.Cost, i.DateAdded = input.ItemCode, input.Status, input.Cost, input.DateAdded
	if err := s.Inventory.UpdateItem(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}
func (s *inventoryService) DeleteItemApplication(ctx context.Context, id string) error {
	return s.Inventory.DeleteItem(ctx, id)
}

type AnnouncementUpdateInput struct{ Title, Content, ImagePath string }

func (s *announcementService) UpdateAnnouncementApplication(ctx context.Context, id string, input AnnouncementUpdateInput, uploads ...UploadInput) (*model.Announcement, error) {
	a, err := s.Announcement.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Title != "" {
		a.Title = input.Title
	}
	if input.Content != "" {
		a.Content = input.Content
	}
	if input.ImagePath != "" {
		a.ImagePath = input.ImagePath
	}
	if err := s.persistUploads(uploads); err != nil {
		return nil, err
	}
	if err := s.Announcement.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}
func (s *announcementService) DeleteAnnouncementApplication(ctx context.Context, id string) error {
	return s.Announcement.Delete(ctx, id)
}
