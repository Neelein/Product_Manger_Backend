package http

import (
	"backend/src/domain/model"
	"time"
)

type Member = model.Member
type Session = model.Session
type RegistrationCode = model.RegistrationCode
type Product = model.Product
type ProductDetail = model.ProductDetail
type ProductPrice = model.ProductPrice
type ProductOption = model.ProductOption
type ProductVariant = model.ProductVariant
type Inventory = model.Inventory
type InventoryItem = model.InventoryItem
type Category = model.Category
type Announcement = model.Announcement
type Event = model.Event
type EventViewer = model.EventViewer
type ChatRoom = model.ChatRoom
type ChatRoomWithMeta = model.ChatRoomWithMeta
type ChatMessage = model.ChatMessage
type ReadReceipt = model.ReadReceipt

type ErrorResponse struct {
	Error string `json:"error"`
}
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Code     string `json:"code"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Member MemberResponse `json:"member"`
}
type UpdateMemberRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
type MemberResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}
type CreateRegistrationCodeRequest struct {
	Code string `json:"code"`
}
type RegistrationCodeResponse struct {
	Code RegistrationCode `json:"code"`
}
type RegistrationCodeListResponse struct {
	Codes []RegistrationCode `json:"codes"`
}
type MembersListResponse struct {
	Members []MemberResponse `json:"members"`
	Total   int              `json:"total"`
}
type CreateProductRequest struct {
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Price      float64 `json:"price"`
	CategoryID string  `json:"category_id"`
}
type UpdateProductRequest = CreateProductRequest
type ProductResponse struct {
	Product Product `json:"product"`
}
type ProductListResponse struct {
	Products []Product `json:"products"`
}
type CreateDetailRequest struct {
	Introduction      string `json:"introduction"`
	UsageInstructions string `json:"usage_instructions"`
	ReturnPolicy      string `json:"return_policy"`
}
type UpdateDetailRequest = CreateDetailRequest
type DetailResponse struct {
	Detail ProductDetail `json:"detail"`
}
type CreatePriceRequest struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	SortOrder int     `json:"sort_order"`
}
type UpdatePriceRequest = CreatePriceRequest
type PriceResponse struct {
	Price ProductPrice `json:"price"`
}
type PriceListResponse struct {
	Prices []ProductPrice `json:"prices"`
}
type CreateProductOptionRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type UpdateProductOptionRequest = CreateProductOptionRequest
type ProductOptionResponse struct {
	Option ProductOption `json:"option"`
}
type ProductOptionListResponse struct {
	Options []ProductOption `json:"options"`
}
type CreateProductVariantRequest struct {
	ProductPriceID string   `json:"product_price_id"`
	SKU            *string  `json:"sku"`
	Status         string   `json:"status"`
	OptionIDs      []string `json:"option_ids"`
}
type UpdateProductVariantRequest = CreateProductVariantRequest
type ProductVariantResponse struct {
	Variant ProductVariant `json:"variant"`
}
type ProductVariantListResponse struct {
	Variants []ProductVariant `json:"variants"`
}
type CreateInventoryRequest struct {
	ProductPriceID   string `json:"product_price_id"`
	ProductVariantID string `json:"product_variant_id"`
	Status           string `json:"status"`
}
type UpdateInventoryRequest struct {
	Status string `json:"status"`
}
type CreateInventoryItemRequest struct {
	ItemCode  string  `json:"item_code"`
	Status    string  `json:"status"`
	Cost      float64 `json:"cost"`
	DateAdded string  `json:"date_added"`
}
type UpdateInventoryItemRequest = CreateInventoryItemRequest
type InventoryResponse struct {
	Inventory Inventory `json:"inventory"`
}
type InventoryListResponse struct {
	Inventories []Inventory `json:"inventories"`
}
type InventoryItemResponse struct {
	Item InventoryItem `json:"item"`
}
type InventoryItemListResponse struct {
	Items []InventoryItem `json:"items"`
}
type CreateCategoryRequest struct {
	Name string `json:"name"`
}
type UpdateCategoryRequest = CreateCategoryRequest
type CategoryResponse struct {
	Category Category `json:"category"`
}
type CategoryListResponse struct {
	Categories []Category `json:"categories"`
}
type CreateAnnouncementRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	ImagePath string `json:"image_path"`
}
type UpdateAnnouncementRequest = CreateAnnouncementRequest
type AnnouncementResponse struct {
	Announcement Announcement `json:"announcement"`
}
type AnnouncementListResponse struct {
	Announcements      []Announcement `json:"announcements"`
	Total, Page, Limit int
}
type CreateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
}
type UpdateEventRequest = CreateEventRequest
type EventResponse struct {
	Event Event `json:"event"`
}
type EventListResponse struct {
	Events []Event `json:"events"`
}
type AddEventViewerRequest struct {
	MemberID string `json:"member_id"`
}
type EventViewerListResponse struct {
	Viewers []EventViewer `json:"viewers"`
}
type CreateRoomRequest struct {
	Name string `json:"name"`
}
type SendMessageRequest struct {
	Content   string `json:"content"`
	ImagePath string `json:"image_path"`
	FilePath  string `json:"file_path"`
}
type UpdateRoomRequest = CreateRoomRequest
type RoomResponse struct {
	Room ChatRoomWithMeta `json:"room"`
}
type RoomListResponse struct {
	Rooms []ChatRoomWithMeta `json:"rooms"`
}
type MessageResponse struct {
	Message ChatMessage `json:"message"`
}
type MessageListResponse struct {
	Messages []ChatMessage `json:"messages"`
}
type ReadByResponse struct {
	ReadBy []ReadReceipt `json:"read_by"`
}
type UnreadCountResponse struct {
	UnreadCount int64 `json:"unread_count"`
}
type RoomMembersRequest struct {
	RoomID string `json:"room_id"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

var (
	ErrProductNotFound         = model.ErrProductNotFound
	ErrMemberNotFound          = model.ErrMemberNotFound
	ErrEmailAlreadyExists      = model.ErrEmailAlreadyExists
	ErrInvalidCredentials      = model.ErrInvalidCredentials
	ErrDetailNotFound          = model.ErrDetailNotFound
	ErrPriceNotFound           = model.ErrPriceNotFound
	ErrInventoryNotFound       = model.ErrInventoryNotFound
	ErrInventoryItemNotFound   = model.ErrInventoryItemNotFound
	ErrAnnouncementNotFound    = model.ErrAnnouncementNotFound
	ErrChatRoomNotFound        = model.ErrChatRoomNotFound
	ErrChatMessageNotFound     = model.ErrChatMessageNotFound
	ErrNotRoomMember           = model.ErrNotRoomMember
	ErrEventNotFound           = model.ErrEventNotFound
	ErrNotEventOwner           = model.ErrNotEventOwner
	ErrEventViewerNotFound     = model.ErrEventViewerNotFound
	ErrInvalidRegistrationCode = model.ErrInvalidRegistrationCode
	ErrRegistrationCodeUsed    = model.ErrRegistrationCodeUsed
	ErrCategoryNotFound        = model.ErrCategoryNotFound
	ErrCategoryInUse           = model.ErrCategoryInUse
	ErrCategoryNameExists      = model.ErrCategoryNameExists
	ErrForbidden               = model.ErrForbidden
	ErrProductOptionNotFound   = model.ErrProductOptionNotFound
	ErrProductVariantNotFound  = model.ErrProductVariantNotFound
	ErrDuplicateProductVariant = model.ErrDuplicateProductVariant
	ErrInvalidProductVariant   = model.ErrInvalidProductVariant
)
