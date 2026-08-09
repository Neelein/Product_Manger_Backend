package usecase

import (
	"bytes"
	"context"
	"io"
	"testing"

	"backend/src/domain/model"
	"backend/src/domain/repository"
)

type serviceProductRepo struct {
	repository.Product
	detail        *model.ProductDetail
	option        *model.ProductOption
	created       bool
	updatedOption bool
}

func (r *serviceProductRepo) GetDetailByProductID(context.Context, string) (*model.ProductDetail, error) {
	return r.detail, nil
}
func (r *serviceProductRepo) GetOptionByID(context.Context, string) (*model.ProductOption, error) {
	return r.option, nil
}
func (r *serviceProductRepo) CreateVariant(context.Context, *model.ProductVariant) error {
	r.created = true
	return nil
}
func (r *serviceProductRepo) UpdateOption(context.Context, *model.ProductOption) error {
	r.updatedOption = true
	return nil
}

func TestProductServiceOwnsVariantWorkflow(t *testing.T) {
	repo := &serviceProductRepo{
		option: &model.ProductOption{ID: "00000000-0000-0000-0000-000000000002", ProductDetailID: "detail"},
	}
	service := NewProductService(repo)
	variant := &model.ProductVariant{
		ProductDetailID: "detail",
		ProductPriceID:  "00000000-0000-0000-0000-000000000001",
		OptionIDs:       []string{"00000000-0000-0000-0000-000000000002"},
	}
	if err := service.CreateVariant(context.Background(), variant); err != nil {
		t.Fatal(err)
	}
	if !repo.created {
		t.Fatal("repository was not called after application validation")
	}
}

func TestProductServiceOwnsOptionMutation(t *testing.T) {
	repo := &serviceProductRepo{option: &model.ProductOption{ID: "option", ProductDetailID: "detail"}}
	if _, err := NewProductService(repo).UpdateOptionApplication(context.Background(), "detail", "option", ProductOptionUpdateInput{Name: "size", Value: "M"}); err != nil {
		t.Fatal(err)
	}
	if !repo.updatedOption {
		t.Fatal("option update did not reach repository")
	}
}

type serviceInventoryRepo struct {
	repository.Inventory
	created   bool
	updated   bool
	inventory *model.Inventory
}

func (r *serviceInventoryRepo) CreateInventory(context.Context, *model.Inventory) error {
	r.created = true
	return nil
}
func (r *serviceInventoryRepo) GetInventoryByID(context.Context, string) (*model.Inventory, error) {
	return r.inventory, nil
}
func (r *serviceInventoryRepo) UpdateInventory(context.Context, *model.Inventory) error {
	r.updated = true
	return nil
}

func TestInventoryServiceValidatesVariantBeforePersistence(t *testing.T) {
	repo := &serviceInventoryRepo{}
	service := NewInventoryService(repo)
	if err := service.CreateInventory(context.Background(), &model.Inventory{ProductVariantID: "bad"}); err == nil {
		t.Fatal("expected invalid variant error")
	}
	if repo.created {
		t.Fatal("invalid inventory reached repository")
	}
}

func TestInventoryServiceOwnsUpdateMutation(t *testing.T) {
	repo := &serviceInventoryRepo{inventory: &model.Inventory{Status: "old"}}
	result, err := NewInventoryService(repo).UpdateInventoryApplication(context.Background(), "inventory", InventoryUpdateInput{Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.updated || result.Status != "active" {
		t.Fatalf("updated=%v status=%q", repo.updated, result.Status)
	}
}

type serviceCategoryRepo struct {
	repository.Category
	name string
}

func (r *serviceCategoryRepo) Create(_ context.Context, name string) (*model.Category, error) {
	r.name = name
	return &model.Category{Name: name}, nil
}

func TestCategoryServiceNormalizesName(t *testing.T) {
	repo := &serviceCategoryRepo{}
	if _, err := NewCategoryService(repo).Create(context.Background(), "  office  "); err != nil {
		t.Fatal(err)
	}
	if repo.name != "office" {
		t.Fatalf("name = %q", repo.name)
	}
}

type memoryStorage struct{ data []byte }

func (s *memoryStorage) Save(_ string, content io.Reader) error {
	buffer := bytes.Buffer{}
	_, err := buffer.ReadFrom(content)
	s.data = buffer.Bytes()
	return err
}

func TestAnnouncementAndChatServicesOwnUploadPersistence(t *testing.T) {
	for _, store := range []struct {
		name   string
		upload func(UploadInput) error
	}{
		{"announcement", NewAnnouncementService(nil, &memoryStorage{}).StoreUpload},
		{"chat", NewChatService(nil, &memoryStorage{}).StoreUpload},
	} {
		t.Run(store.name, func(t *testing.T) {
			if err := store.upload(UploadInput{Directory: "uploads", Filename: "file", Content: bytes.NewBufferString("content")}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type serviceAnnouncementRepo struct {
	repository.Announcement
	announcement *model.Announcement
	updated      bool
}

func (r *serviceAnnouncementRepo) GetByID(context.Context, string) (*model.Announcement, error) {
	return r.announcement, nil
}
func (r *serviceAnnouncementRepo) Update(context.Context, *model.Announcement) error {
	r.updated = true
	return nil
}

func TestAnnouncementServiceOwnsUpdateMutation(t *testing.T) {
	repo := &serviceAnnouncementRepo{announcement: &model.Announcement{Title: "old", Content: "content"}}
	result, err := NewAnnouncementService(repo).UpdateAnnouncementApplication(context.Background(), "announcement", AnnouncementUpdateInput{Title: "new"})
	if err != nil {
		t.Fatal(err)
	}
	if !repo.updated || result.Title != "new" {
		t.Fatalf("updated=%v title=%q", repo.updated, result.Title)
	}
}

type serviceEventRepo struct {
	repository.EventOperations
	event *model.Event
}

func (r *serviceEventRepo) GetByID(context.Context, string, string) (*model.Event, error) {
	return r.event, nil
}
func (r *serviceEventRepo) Update(context.Context, *model.Event) error { return nil }

func TestEventServiceOwnsAuthorization(t *testing.T) {
	service := NewEventService(&serviceEventRepo{event: &model.Event{ID: "event", CreatedBy: "owner"}})
	_, err := service.UpdateApplication(context.Background(), "event", "other", false, EventUpdateInput{Title: "changed"})
	if err != model.ErrNotEventOwner {
		t.Fatalf("err = %v", err)
	}
}
