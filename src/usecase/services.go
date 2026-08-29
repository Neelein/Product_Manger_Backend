package usecase

import (
	"backend/src/domain/model"
	"backend/src/domain/repository"
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Application services expose operation contracts to transport adapters. Repository
// ports are intentionally kept behind the concrete implementations below.
type ProductService interface {
	UpdateOptionApplication(context.Context, string, string, ProductOptionUpdateInput) (*model.ProductOption, error)
	DeleteOptionApplication(context.Context, string, string) error
	UpdateVariantApplication(context.Context, string, string, ProductVariantUpdateInput) (*model.ProductVariant, error)
	DeleteVariantApplication(context.Context, string, string) error
	UpdateDetailApplication(context.Context, string, ProductDetailUpdateInput) (*model.ProductDetail, error)
	UpdatePriceApplication(context.Context, string, ProductPriceUpdateInput) (*model.ProductPrice, error)
	GetDetailForRoute(context.Context, string, string) (*model.ProductDetail, error)
	GetOptionForDetail(context.Context, string, string) (*model.ProductOption, error)
	GetVariantForDetail(context.Context, string, string) (*model.ProductVariant, error)
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
	UploadImages(context.Context, string, []UploadInput) ([]model.ProductImage, error)
	ListImages(context.Context, string) ([]model.ProductImage, error)
}

type InventoryService interface {
	UpdateInventoryApplication(context.Context, string, InventoryUpdateInput) (*model.Inventory, error)
	DeleteInventoryApplication(context.Context, string) error
	UpdateItemApplication(context.Context, string, InventoryItemUpdateInput) (*model.InventoryItem, error)
	DeleteItemApplication(context.Context, string) error
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

type MemberService interface {
	RegisterApplication(context.Context, *model.Member, string, string) error
	Authenticate(context.Context, string, string) (*model.Member, *model.Session, error)
	UpdateApplication(context.Context, *model.Member, string, string) error
	Create(context.Context, *model.Member) error
	GetByEmail(context.Context, string) (*model.Member, error)
	GetByID(context.Context, string) (*model.Member, error)
	Update(context.Context, *model.Member) error
	UpdatePermission(context.Context, string, string, string) error
}

type SessionService interface {
	Create(context.Context, *model.Session) error
	GetByKey(context.Context, string) (*model.Session, error)
	Delete(context.Context, string) error
}

type RegistrationCodeService interface {
	CreateApplication(context.Context, string, string) (*model.RegistrationCode, error)
	RegisterMemberWithCode(context.Context, *model.Member, string) error
	Create(context.Context, string, string) (*model.RegistrationCode, error)
	List(context.Context) ([]model.RegistrationCode, error)
	Delete(context.Context, string) (bool, error)
}

type CategoryService interface {
	List(context.Context) ([]model.Category, error)
	Create(context.Context, string) (*model.Category, error)
	Update(context.Context, string, string) (bool, error)
	Delete(context.Context, string) (bool, error)
}

type AnnouncementService interface {
	UpdateAnnouncementApplication(context.Context, string, AnnouncementUpdateInput, ...UploadInput) (*model.Announcement, error)
	DeleteAnnouncementApplication(context.Context, string) error
	CreateApplication(context.Context, *model.Announcement, ...UploadInput) error
	UpdateApplication(context.Context, *model.Announcement, ...UploadInput) error
	ListPage(context.Context, int, int, *int, *int) (AnnouncementPage, error)
	Create(context.Context, *model.Announcement) error
	GetByID(context.Context, string) (*model.Announcement, error)
	List(context.Context, int, int) ([]model.Announcement, int, error)
	ListByMonth(context.Context, int, int, int, int) ([]model.Announcement, int, error)
	Update(context.Context, *model.Announcement) error
	Delete(context.Context, string) error
	StoreUpload(UploadInput) error
}

type ChatService interface {
	CreateRoomApplication(context.Context, *model.ChatRoom, string) (*model.ChatRoomWithMeta, error)
	UpdateRoomApplication(context.Context, string, string, string) (*model.ChatRoomWithMeta, error)
	ListMessagesApplication(context.Context, string, string, int) ([]model.ChatMessage, error)
	MarkAsReadApplication(context.Context, string, string) error
	ListAvailableMembers(context.Context, string, int, int) ([]model.Member, int, error)
	CreateRoom(context.Context, *model.ChatRoom) error
	GetRoomByID(context.Context, string, string) (*model.ChatRoomWithMeta, error)
	ListRoomsByMember(context.Context, string) ([]model.ChatRoomWithMeta, error)
	ListRoomsByMemberByMonth(context.Context, string, int, int) ([]model.ChatRoomWithMeta, error)
	UpdateRoom(context.Context, string, string) error
	DeleteRoom(context.Context, string) error
	AddMembers(context.Context, string, []string) error
	RemoveMember(context.Context, string, string) error
	SendMessage(context.Context, *model.ChatMessage) error
	SendMessageApplication(context.Context, *model.ChatMessage, ...UploadInput) error
	ListMessages(context.Context, string, string, int) ([]model.ChatMessage, error)
	DeleteMessage(context.Context, string) error
	MarkAsRead(context.Context, string, string) error
	GetReadBy(context.Context, string) ([]model.ReadReceipt, error)
	CountUnread(context.Context, string, string) (int64, error)
	ListMembersNotInRoom(context.Context, string, int, int) ([]model.Member, error)
	CountMembersNotInRoom(context.Context, string) (int, error)
	StoreUpload(UploadInput) error
}

type EventService interface {
	ListByMonthApplication(context.Context, string, int, int) ([]model.Event, error)
	CreateApplication(context.Context, string, EventCreateInput) (*model.Event, error)
	UpdateApplication(context.Context, string, string, bool, EventUpdateInput) (*model.Event, error)
	DeleteApplication(context.Context, string, string, bool) error
	AddViewerApplication(context.Context, string, string, string, bool) error
	RemoveViewerApplication(context.Context, string, string, string, bool) error
	ListViewersApplication(context.Context, string, string, bool) ([]model.EventViewer, error)
	Create(context.Context, *model.Event) error
	GetByID(context.Context, string, string) (*model.Event, error)
	ListByMonth(context.Context, int, int, string) ([]model.Event, error)
	Update(context.Context, *model.Event) error
	Delete(context.Context, string) error
	AddViewer(context.Context, string, string) error
	RemoveViewer(context.Context, string, string) error
	ListViewers(context.Context, string) ([]model.EventViewer, error)
}

type OrderCreateItem struct {
	ProductPriceID string
	Quantity       int
}

type OrderCustomerInput struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type OrderCreateInput struct {
	Items           []OrderCreateItem
	Customer        OrderCustomerInput
	DeliveryMethod  string
	ShippingAddress string
}

type OrderService interface {
	Create(context.Context, *model.Member, OrderCreateInput) (*model.Order, error)
	GetByID(context.Context, *model.Member, string) (*model.Order, error)
	List(context.Context, *model.Member, string, int, int) ([]model.Order, int, error)
	Cancel(context.Context, *model.Member, string) error
	UpdateStatus(context.Context, *model.Member, string, string) (*model.Order, error)
	History(context.Context, *model.Member, string) ([]model.OrderStatusHistory, error)
}

type productService struct {
	repository.Product
	storage FileStorage
}

func NewProductService(repo repository.Product, stores ...FileStorage) ProductService {
	var store FileStorage
	if len(stores) > 0 {
		store = stores[0]
	}
	return &productService{Product: repo, storage: store}
}
func (s *productService) UploadImages(ctx context.Context, productID string, uploads []UploadInput) ([]model.ProductImage, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("file storage is not configured")
	}
	existing, err := s.Product.ListImages(ctx, productID)
	if err != nil {
		return nil, err
	}
	if len(existing)+len(uploads) > 3 {
		return nil, fmt.Errorf("product image limit exceeded")
	}
	images := make([]model.ProductImage, 0, len(uploads))
	for _, upload := range uploads {
		if err := s.storage.Save(filepath.Join(upload.Directory, upload.Filename), upload.Content); err != nil {
			return nil, err
		}
		image := model.ProductImage{ProductID: productID, Filename: upload.Filename}
		if err := s.Product.CreateImage(ctx, &image); err != nil {
			return nil, err
		}
		images = append(images, image)
	}
	return images, nil
}
func (s *productService) GetDetailForRoute(ctx context.Context, productID, detailID string) (*model.ProductDetail, error) {
	detail, err := s.Product.GetDetailByProductID(ctx, productID)
	if err != nil || (detailID != "" && detail.ID != detailID) {
		return nil, model.ErrDetailNotFound
	}
	return detail, nil
}
func (s *productService) GetOptionForDetail(ctx context.Context, detailID, optionID string) (*model.ProductOption, error) {
	option, err := s.Product.GetOptionByID(ctx, optionID)
	if err != nil || option.ProductDetailID != detailID {
		return nil, model.ErrProductOptionNotFound
	}
	return option, nil
}
func (s *productService) GetVariantForDetail(ctx context.Context, detailID, variantID string) (*model.ProductVariant, error) {
	variant, err := s.Product.GetVariantByID(ctx, variantID)
	if err != nil || variant.ProductDetailID != detailID {
		return nil, model.ErrProductVariantNotFound
	}
	return variant, nil
}

func (s *productService) CreateVariant(ctx context.Context, variant *model.ProductVariant) error {
	if err := ValidateProductVariantIDs(variant.ProductPriceID, variant.OptionIDs); err != nil {
		return err
	}
	for _, optionID := range variant.OptionIDs {
		option, err := s.GetOptionByID(ctx, optionID)
		if err != nil {
			return err
		}
		if option.ProductDetailID != variant.ProductDetailID {
			return model.ErrInvalidProductVariant
		}
	}
	return s.Product.CreateVariant(ctx, variant)
}

type inventoryService struct{ repository.Inventory }

func NewInventoryService(repo repository.Inventory) InventoryService { return &inventoryService{repo} }
func (s *inventoryService) CreateInventory(ctx context.Context, inventory *model.Inventory) error {
	if err := ValidateInventoryVariantID(inventory.ProductVariantID); err != nil {
		return ErrInvalidInventoryVariant
	}
	return s.Inventory.CreateInventory(ctx, inventory)
}

type memberService struct {
	repository.Member
	sessions repository.Session
	codes    repository.RegistrationCode
}

func NewMemberService(repo repository.Member, dependencies ...interface{}) MemberService {
	service := &memberService{Member: repo}
	for _, dependency := range dependencies {
		switch value := dependency.(type) {
		case repository.Session:
			service.sessions = value
		case repository.RegistrationCode:
			service.codes = value
		}
	}
	return service
}
func (s *memberService) RegisterApplication(ctx context.Context, member *model.Member, password, code string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	member.Password = string(hash)
	member.Permission = ""
	if code == "" {
		member.MemberType = "customer"
		return s.Member.Create(ctx, member)
	}
	if s.codes == nil {
		return fmt.Errorf("registration service is not configured")
	}
	return s.codes.RegisterMemberWithCode(ctx, member, code)
}

func (s *memberService) UpdatePermission(ctx context.Context, actorID, targetID, permission string) error {
	actor, err := s.Member.GetByID(ctx, actorID)
	if err != nil || actor == nil || actor.MemberType != "employee" || actor.Permission != "admin" || actorID == targetID {
		return model.ErrForbidden
	}
	target, err := s.Member.GetByID(ctx, targetID)
	if err != nil || target == nil || target.MemberType != "employee" {
		return model.ErrMemberNotFound
	}
	repo, ok := s.Member.(repository.MemberPermission)
	if !ok {
		return fmt.Errorf("member permission repository is not configured")
	}
	return repo.UpdatePermission(ctx, targetID, permission)
}
func (s *memberService) Authenticate(ctx context.Context, email, password string) (*model.Member, *model.Session, error) {
	member, err := s.Member.GetByEmail(ctx, email)
	if err != nil || member == nil || bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(password)) != nil {
		return nil, nil, model.ErrInvalidCredentials
	}
	if s.sessions == nil {
		return nil, nil, fmt.Errorf("session service is not configured")
	}
	session := &model.Session{MemberID: member.ID}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, nil, err
	}
	return member, session, nil
}
func (s *memberService) UpdateApplication(ctx context.Context, member *model.Member, email, name string) error {
	if strings.TrimSpace(email) == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("email and name are required")
	}
	member.Email, member.Name = email, name
	return s.Member.Update(ctx, member)
}

type sessionService struct{ repository.Session }

func NewSessionService(repo repository.Session) SessionService { return &sessionService{repo} }

type registrationCodeService struct{ repository.RegistrationCode }

func NewRegistrationCodeService(repo repository.RegistrationCode) RegistrationCodeService {
	return &registrationCodeService{repo}
}
func (s *registrationCodeService) CreateApplication(ctx context.Context, memberID, code string) (*model.RegistrationCode, error) {
	return s.RegistrationCode.Create(ctx, memberID, code)
}

type categoryService struct{ repository.Category }

func NewCategoryService(repo repository.Category) CategoryService { return &categoryService{repo} }
func (s *categoryService) Create(ctx context.Context, name string) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if err := ValidateCategoryName(name); err != nil {
		return nil, ErrInvalidCategoryName
	}
	return s.Category.Create(ctx, name)
}
func (s *categoryService) Update(ctx context.Context, id, name string) (bool, error) {
	name = strings.TrimSpace(name)
	if err := ValidateCategoryName(name); err != nil {
		return false, ErrInvalidCategoryName
	}
	return s.Category.Update(ctx, id, name)
}

type announcementService struct {
	repository.Announcement
	storage FileStorage
}

func (s *announcementService) persistUploads(uploads []UploadInput) error {
	for _, upload := range uploads {
		if err := s.StoreUpload(upload); err != nil {
			return err
		}
	}
	return nil
}
func (s *announcementService) CreateApplication(ctx context.Context, announcement *model.Announcement, uploads ...UploadInput) error {
	if strings.TrimSpace(announcement.Title) == "" || strings.TrimSpace(announcement.Content) == "" {
		return ErrInvalidAnnouncement
	}
	if err := s.persistUploads(uploads); err != nil {
		return err
	}
	return s.Announcement.Create(ctx, announcement)
}
func (s *announcementService) UpdateApplication(ctx context.Context, announcement *model.Announcement, uploads ...UploadInput) error {
	if err := s.persistUploads(uploads); err != nil {
		return err
	}
	return s.Announcement.Update(ctx, announcement)
}

type AnnouncementPage struct {
	Announcements []model.Announcement
	Total         int
	Page, Limit   int
}

func (s *announcementService) ListPage(ctx context.Context, page, limit int, year, month *int) (AnnouncementPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit
	var announcements []model.Announcement
	var total int
	var err error
	if year != nil && month != nil {
		if *month < 1 || *month > 12 {
			return AnnouncementPage{}, ErrEventMonthInvalid
		}
		announcements, total, err = s.Announcement.ListByMonth(ctx, *year, *month, limit, offset)
	} else {
		announcements, total, err = s.Announcement.List(ctx, limit, offset)
	}
	return AnnouncementPage{Announcements: announcements, Total: total, Page: page, Limit: limit}, err
}

func NewAnnouncementService(repo repository.Announcement, stores ...FileStorage) AnnouncementService {
	var store FileStorage
	if len(stores) > 0 {
		store = stores[0]
	}
	return &announcementService{Announcement: repo, storage: store}
}
func (s *announcementService) StoreUpload(input UploadInput) error {
	if s.storage == nil {
		return fmt.Errorf("file storage is not configured")
	}
	return s.storage.Save(filepath.Join(input.Directory, input.Filename), input.Content)
}

type chatService struct {
	repository.ChatRoom
	storage FileStorage
}

func NewChatService(repo repository.ChatRoom, stores ...FileStorage) ChatService {
	var store FileStorage
	if len(stores) > 0 {
		store = stores[0]
	}
	return &chatService{ChatRoom: repo, storage: store}
}
func (s *chatService) CreateRoom(ctx context.Context, room *model.ChatRoom) error {
	if strings.TrimSpace(room.Name) == "" {
		return ErrInvalidChatName
	}
	return s.ChatRoom.CreateRoom(ctx, room)
}
func (s *chatService) CreateRoomApplication(ctx context.Context, room *model.ChatRoom, memberID string) (*model.ChatRoomWithMeta, error) {
	if err := s.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return s.ChatRoom.GetRoomByID(ctx, room.ID, memberID)
}
func (s *chatService) UpdateRoom(ctx context.Context, id, name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidChatName
	}
	return s.ChatRoom.UpdateRoom(ctx, id, name)
}
func (s *chatService) UpdateRoomApplication(ctx context.Context, id, name, memberID string) (*model.ChatRoomWithMeta, error) {
	if err := s.UpdateRoom(ctx, id, name); err != nil {
		return nil, err
	}
	return s.ChatRoom.GetRoomByID(ctx, id, memberID)
}
func (s *chatService) AddMembers(ctx context.Context, roomID string, memberIDs []string) error {
	if len(memberIDs) == 0 {
		return ErrInvalidChatMembers
	}
	return s.ChatRoom.AddMembers(ctx, roomID, memberIDs)
}
func (s *chatService) SendMessageApplication(ctx context.Context, message *model.ChatMessage, uploads ...UploadInput) error {
	for _, upload := range uploads {
		if err := s.StoreUpload(upload); err != nil {
			return err
		}
	}
	return s.ChatRoom.SendMessage(ctx, message)
}
func (s *chatService) ListMessagesApplication(ctx context.Context, roomID, beforeID string, limit int) ([]model.ChatMessage, error) {
	if limit < 1 {
		limit = 20
	}
	return s.ChatRoom.ListMessages(ctx, roomID, beforeID, limit)
}
func (s *chatService) MarkAsReadApplication(ctx context.Context, messageID, memberID string) error {
	if strings.TrimSpace(messageID) == "" {
		return ErrInvalidMessageID
	}
	return s.ChatRoom.MarkAsRead(ctx, messageID, memberID)
}
func (s *chatService) ListRoomsByMemberByMonth(ctx context.Context, memberID string, year, month int) ([]model.ChatRoomWithMeta, error) {
	if month < 1 || month > 12 {
		return nil, ErrEventMonthInvalid
	}
	return s.ChatRoom.ListRoomsByMemberByMonth(ctx, memberID, year, month)
}
func (s *chatService) ListAvailableMembers(ctx context.Context, roomID string, page, limit int) ([]model.Member, int, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 20
	}
	if page < 1 {
		return nil, 0, ErrInvalidChatPage
	}
	if limit < 1 || limit > 100 {
		return nil, 0, ErrInvalidChatLimit
	}
	offset := (page - 1) * limit
	members, err := s.ChatRoom.ListMembersNotInRoom(ctx, roomID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.ChatRoom.CountMembersNotInRoom(ctx, roomID)
	return members, total, err
}
func (s *chatService) StoreUpload(input UploadInput) error {
	if s.storage == nil {
		return fmt.Errorf("file storage is not configured")
	}
	return s.storage.Save(filepath.Join(input.Directory, input.Filename), input.Content)
}

type eventService struct{ repository.EventOperations }

type orderService struct{ repository.Order }

func NewOrderService(repo repository.Order) OrderService { return &orderService{Order: repo} }

func (s *orderService) Create(ctx context.Context, member *model.Member, input OrderCreateInput) (*model.Order, error) {
	if member == nil || (member.MemberType != "customer" && member.MemberType != "employee") || len(input.Items) == 0 {
		return nil, model.ErrInvalidOrder
	}
	customer := OrderCustomerInput{Name: strings.TrimSpace(input.Customer.Name), Phone: strings.TrimSpace(input.Customer.Phone), Email: strings.TrimSpace(input.Customer.Email)}
	if customer.Name == "" || customer.Phone == "" || customer.Email == "" {
		return nil, model.ErrInvalidOrder
	}
	parsedEmail, err := mail.ParseAddress(customer.Email)
	if err != nil || parsedEmail.Address != customer.Email {
		return nil, model.ErrInvalidOrder
	}
	if input.DeliveryMethod != "email" && input.DeliveryMethod != "home_address" {
		return nil, model.ErrInvalidOrder
	}
	address := strings.TrimSpace(input.ShippingAddress)
	if input.DeliveryMethod == "home_address" && address == "" {
		return nil, model.ErrInvalidOrder
	}
	items := make([]model.OrderItem, len(input.Items))
	for i, item := range input.Items {
		if item.ProductPriceID == "" || item.Quantity <= 0 {
			return nil, model.ErrInvalidOrder
		}
		items[i].ProductPriceID, items[i].Quantity = item.ProductPriceID, item.Quantity
	}
	customerSnapshot, _ := json.Marshal(customer)
	shippingSnapshot := map[string]string{"delivery_method": input.DeliveryMethod}
	if address != "" {
		shippingSnapshot["address"] = address
	}
	shipping, _ := json.Marshal(shippingSnapshot)
	order := &model.Order{CustomerID: member.ID, CustomerSnapshot: customerSnapshot, ShippingAddressSnapshot: shipping}
	if err := s.Order.Create(ctx, order, items); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetByID(ctx context.Context, member *model.Member, id string) (*model.Order, error) {
	if member == nil {
		return nil, model.ErrForbidden
	}
	return s.Order.GetByID(ctx, id, member.ID, member.MemberType == "employee")
}
func (s *orderService) List(ctx context.Context, member *model.Member, status string, page, size int) ([]model.Order, int, error) {
	if member == nil {
		return nil, 0, model.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return s.Order.List(ctx, member.ID, status, page, size, member.MemberType == "employee")
}
func (s *orderService) Cancel(ctx context.Context, member *model.Member, id string) error {
	if member == nil {
		return model.ErrForbidden
	}
	return s.Order.Cancel(ctx, id, member.ID, member.MemberType == "employee")
}
func (s *orderService) UpdateStatus(ctx context.Context, member *model.Member, id, status string) (*model.Order, error) {
	if member == nil || member.MemberType != "employee" {
		return nil, model.ErrForbidden
	}
	return s.Order.UpdateStatus(ctx, id, member.ID, status)
}
func (s *orderService) History(ctx context.Context, member *model.Member, id string) ([]model.OrderStatusHistory, error) {
	if member == nil {
		return nil, model.ErrForbidden
	}
	return s.Order.History(ctx, id, member.ID, member.MemberType == "employee")
}

func NewEventService(repo repository.EventOperations) EventService { return &eventService{repo} }

func (s *eventService) ListByMonthApplication(ctx context.Context, memberID string, year, month int) ([]model.Event, error) {
	if year == 0 || month == 0 {
		return nil, ErrEventMonthRequired
	}
	if month < 1 || month > 12 {
		return nil, ErrEventMonthInvalid
	}
	return s.EventOperations.ListByMonth(ctx, year, month, memberID)
}

type EventCreateInput struct {
	Title, Description, Status string
	StartTime, EndTime         time.Time
}

type EventUpdateInput struct {
	Title, Description, Status string
	StartTime, EndTime         time.Time
}

func (s *eventService) CreateApplication(ctx context.Context, memberID string, input EventCreateInput) (*model.Event, error) {
	if strings.TrimSpace(input.Title) == "" {
		return nil, ErrEventTitleRequired
	}
	if input.Status == "" {
		input.Status = "active"
	}
	event := &model.Event{Title: input.Title, Description: input.Description, Status: input.Status, CreatedBy: memberID, StartTime: input.StartTime.UTC(), EndTime: input.EndTime.UTC()}
	if err := s.EventOperations.Create(ctx, event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *eventService) owned(ctx context.Context, eventID, memberID string, admin bool) (*model.Event, error) {
	event, err := s.EventOperations.GetByID(ctx, eventID, memberID)
	if err != nil {
		return nil, err
	}
	if !admin && event.CreatedBy != memberID {
		return nil, model.ErrNotEventOwner
	}
	return event, nil
}
func (s *eventService) UpdateApplication(ctx context.Context, eventID, memberID string, admin bool, input EventUpdateInput) (*model.Event, error) {
	event, err := s.owned(ctx, eventID, memberID, admin)
	if err != nil {
		return nil, err
	}
	if input.Title != "" {
		event.Title = input.Title
	}
	if input.Description != "" {
		event.Description = input.Description
	}
	if !input.StartTime.IsZero() {
		event.StartTime = input.StartTime.UTC()
	}
	if !input.EndTime.IsZero() {
		event.EndTime = input.EndTime.UTC()
	}
	if input.Status != "" {
		event.Status = input.Status
	}
	if err := s.EventOperations.Update(ctx, event); err != nil {
		return nil, err
	}
	return s.EventOperations.GetByID(ctx, eventID, memberID)
}
func (s *eventService) DeleteApplication(ctx context.Context, eventID, memberID string, admin bool) error {
	if _, err := s.owned(ctx, eventID, memberID, admin); err != nil {
		return err
	}
	return s.EventOperations.Delete(ctx, eventID)
}
func (s *eventService) AddViewerApplication(ctx context.Context, eventID, memberID, viewerID string, admin bool) error {
	if _, err := s.owned(ctx, eventID, memberID, admin); err != nil {
		return err
	}
	if viewerID == "" {
		return model.ErrInvalidProductVariant
	}
	return s.EventOperations.AddViewer(ctx, eventID, viewerID)
}
func (s *eventService) RemoveViewerApplication(ctx context.Context, eventID, memberID, viewerID string, admin bool) error {
	if _, err := s.owned(ctx, eventID, memberID, admin); err != nil {
		return err
	}
	return s.EventOperations.RemoveViewer(ctx, eventID, viewerID)
}
func (s *eventService) ListViewersApplication(ctx context.Context, eventID, memberID string, admin bool) ([]model.EventViewer, error) {
	if _, err := s.owned(ctx, eventID, memberID, admin); err != nil {
		return nil, err
	}
	return s.EventOperations.ListViewers(ctx, eventID)
}
