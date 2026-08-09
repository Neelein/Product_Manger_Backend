package usecase

import "backend/src/domain/repository"

type ProductService interface{ repository.Product }
type InventoryService interface{ repository.Inventory }
type MemberService interface{ repository.Member }
type SessionService interface{ repository.Session }
type RegistrationCodeService interface{ repository.RegistrationCode }
type CategoryService interface{ repository.Category }
type AnnouncementService interface{ repository.Announcement }
type ChatService interface{ repository.ChatRoom }
type EventService interface{ repository.EventOperations }

type productService struct{ repository.Product }

func NewProductService(repo repository.Product) ProductService { return &productService{repo} }

type inventoryService struct{ repository.Inventory }

func NewInventoryService(repo repository.Inventory) InventoryService { return &inventoryService{repo} }

type memberService struct{ repository.Member }

func NewMemberService(repo repository.Member) MemberService { return &memberService{repo} }

type sessionService struct{ repository.Session }

func NewSessionService(repo repository.Session) SessionService {
	return &sessionService{repo}
}

type registrationCodeService struct{ repository.RegistrationCode }

func NewRegistrationCodeService(repo repository.RegistrationCode) RegistrationCodeService {
	return &registrationCodeService{repo}
}

type categoryService struct{ repository.Category }

func NewCategoryService(repo repository.Category) CategoryService { return &categoryService{repo} }

type announcementService struct{ repository.Announcement }

func NewAnnouncementService(repo repository.Announcement) AnnouncementService {
	return &announcementService{repo}
}

type chatService struct{ repository.ChatRoom }

func NewChatService(repo repository.ChatRoom) ChatService { return &chatService{repo} }

type eventService struct{ repository.EventOperations }

func NewEventService(repo repository.EventOperations) EventService { return &eventService{repo} }
