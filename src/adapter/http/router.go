package http

import (
	"net/http"
	"os"
	"path/filepath"

	"backend/src/usecase"

	"github.com/gorilla/mux"
)

func RegisterProductRoutes(r *mux.Router, service usecase.ProductService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewProductHandler(service)
	auth := AuthMiddleware(sessionService, memberService)

	r.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	r.Handle("/api/products", auth(http.HandlerFunc(h.CreateProduct))).Methods("POST")
	r.HandleFunc("/api/products/{productId}", h.GetProduct).Methods("GET")
	r.Handle("/api/products/{productId}/update", auth(http.HandlerFunc(h.UpdateProduct))).Methods("POST")
	r.Handle("/api/products/{productId}/delete", auth(http.HandlerFunc(h.DeleteProduct))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail", h.GetDetail).Methods("GET")
	r.Handle("/api/products/{productId}/detail/update", auth(http.HandlerFunc(h.UpdateDetail))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/prices", h.ListPrices).Methods("GET")
	r.HandleFunc("/api/products/{productId}/detail/prices/{priceId}", h.GetPrice).Methods("GET")
	r.Handle("/api/products/{productId}/detail/prices/{priceId}/update", auth(http.HandlerFunc(h.UpdatePrice))).Methods("POST")
	r.Handle("/api/products/{productId}/details", auth(http.HandlerFunc(h.CreateDetail))).Methods("POST")
	r.Handle("/api/products/{productId}/details/{detailId}/prices", auth(http.HandlerFunc(h.CreatePrice))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/options", h.ListOptions).Methods("GET")
	r.Handle("/api/products/{productId}/detail/options", auth(http.HandlerFunc(h.CreateOption))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/options/{optionId}", h.GetOption).Methods("GET")
	r.Handle("/api/products/{productId}/detail/options/{optionId}/update", auth(http.HandlerFunc(h.UpdateOption))).Methods("POST")
	r.Handle("/api/products/{productId}/detail/options/{optionId}/delete", auth(http.HandlerFunc(h.DeleteOption))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/variants", h.ListVariants).Methods("GET")
	r.Handle("/api/products/{productId}/detail/variants", auth(http.HandlerFunc(h.CreateVariant))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/variants/{variantId}", h.GetVariant).Methods("GET")
	r.Handle("/api/products/{productId}/detail/variants/{variantId}/update", auth(http.HandlerFunc(h.UpdateVariant))).Methods("POST")
	r.Handle("/api/products/{productId}/detail/variants/{variantId}/delete", auth(http.HandlerFunc(h.DeleteVariant))).Methods("POST")
}

func RegisterInventoryRoutes(r *mux.Router, service usecase.InventoryService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewInventoryHandler(service)
	auth := AuthMiddleware(sessionService, memberService)

	r.HandleFunc("/api/inventories", h.ListInventories).Methods("GET")
	r.Handle("/api/inventories", auth(http.HandlerFunc(h.CreateInventory))).Methods("POST")
	r.HandleFunc("/api/inventories/{inventoryId}", h.GetInventory).Methods("GET")
	r.Handle("/api/inventories/{inventoryId}/update", auth(http.HandlerFunc(h.UpdateInventory))).Methods("POST")
	r.Handle("/api/inventories/{inventoryId}/delete", auth(http.HandlerFunc(h.DeleteInventory))).Methods("POST")
	r.HandleFunc("/api/inventories/{inventoryId}/items", h.ListItems).Methods("GET")
	r.Handle("/api/inventories/{inventoryId}/items", auth(http.HandlerFunc(h.CreateItem))).Methods("POST")
	r.HandleFunc("/api/inventories/{inventoryId}/items/{itemId}", h.GetItem).Methods("GET")
	r.Handle("/api/inventories/{inventoryId}/items/{itemId}/update", auth(http.HandlerFunc(h.UpdateItem))).Methods("POST")
	r.Handle("/api/inventories/{inventoryId}/items/{itemId}/delete", auth(http.HandlerFunc(h.DeleteItem))).Methods("POST")
}

func RegisterMemberRoutes(r *mux.Router, memberService usecase.MemberService, sessionService usecase.SessionService, codeService usecase.RegistrationCodeService) {
	h := NewMemberHandler(memberService, sessionService, codeService)
	auth := AuthMiddleware(sessionService, memberService)

	r.HandleFunc("/api/members/register", h.RegisterMember).Methods("POST")
	r.HandleFunc("/api/members/login", h.LoginMember).Methods("POST")
	r.HandleFunc("/api/members/logout", h.LogoutMember).Methods("POST")
	r.Handle("/api/members/me", auth(http.HandlerFunc(h.GetCurrentMember))).Methods("GET")
	r.Handle("/api/members/update", auth(http.HandlerFunc(h.UpdateMember))).Methods("POST")
}

func RegisterRegistrationCodeRoutes(r *mux.Router, codeService usecase.RegistrationCodeService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewRegistrationCodeHandler(codeService)
	auth := AuthMiddleware(sessionService, memberService)
	adminAuth := RequireRole("admin", auth)

	r.Handle("/api/registration-codes", adminAuth(http.HandlerFunc(h.CreateCode))).Methods("POST")
	r.Handle("/api/registration-codes", adminAuth(http.HandlerFunc(h.ListCodes))).Methods("GET")
	r.Handle("/api/registration-codes/{id}", adminAuth(http.HandlerFunc(h.DeleteCode))).Methods("DELETE")
}

func RegisterCategoryRoutes(r *mux.Router, categoryService usecase.CategoryService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewCategoryHandler(categoryService)
	auth := AuthMiddleware(sessionService, memberService)

	r.HandleFunc("/api/categories", h.ListCategories).Methods("GET")
	r.Handle("/api/categories", auth(http.HandlerFunc(h.CreateCategory))).Methods("POST")
	r.Handle("/api/categories/{id}/update", auth(http.HandlerFunc(h.UpdateCategory))).Methods("POST")
	r.Handle("/api/categories/{id}/delete", auth(http.HandlerFunc(h.DeleteCategory))).Methods("POST")
}

func RegisterAnnouncementRoutes(r *mux.Router, service usecase.AnnouncementService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewAnnouncementHandler(service)
	auth := AuthMiddleware(sessionService, memberService)

	r.HandleFunc("/api/announcements", h.ListAnnouncements).Methods("GET")
	r.Handle("/api/announcements", auth(http.HandlerFunc(h.CreateAnnouncement))).Methods("POST")
	r.HandleFunc("/api/announcements/{announcementId}", h.GetAnnouncement).Methods("GET")
	r.Handle("/api/announcements/{announcementId}/update", auth(http.HandlerFunc(h.UpdateAnnouncement))).Methods("POST")
	r.Handle("/api/announcements/{announcementId}/delete", auth(http.HandlerFunc(h.DeleteAnnouncement))).Methods("POST")

	r.PathPrefix("/media/images/announcements/").Handler(http.StripPrefix("/media/images/announcements/", http.FileServer(http.Dir(filepath.Join(os.Getenv("MEDIA_ROOT"), "images/announcements")))))
}

func RegisterChatRoutes(r *mux.Router, service usecase.ChatService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewChatRoomHandler(service)
	auth := AuthMiddleware(sessionService, memberService)

	r.Handle("/api/chat/rooms", auth(http.HandlerFunc(h.CreateRoom))).Methods("POST")
	r.Handle("/api/chat/rooms", auth(http.HandlerFunc(h.ListRooms))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}", auth(http.HandlerFunc(h.GetRoom))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/update", auth(http.HandlerFunc(h.UpdateRoom))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/delete", auth(http.HandlerFunc(h.DeleteRoom))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/members", auth(http.HandlerFunc(h.AddMembers))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/members/{memberId}/remove", auth(http.HandlerFunc(h.RemoveMember))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/messages", auth(http.HandlerFunc(h.ListMessages))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/messages", auth(http.HandlerFunc(h.SendMessage))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/messages/{messageId}/delete", auth(http.HandlerFunc(h.DeleteMessage))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/read", auth(http.HandlerFunc(h.MarkAsRead))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/messages/{messageId}/read-by", auth(http.HandlerFunc(h.GetReadBy))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/unread", auth(http.HandlerFunc(h.CountUnread))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/available-members", auth(http.HandlerFunc(h.ListAvailableMembers))).Methods("POST")
}

func RegisterEventRoutes(r *mux.Router, service usecase.EventService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewEventHandler(service)
	auth := AuthMiddleware(sessionService, memberService)

	r.Handle("/api/events", auth(http.HandlerFunc(h.ListEventsByMonth))).Methods("GET")
	r.Handle("/api/events", auth(http.HandlerFunc(h.CreateEvent))).Methods("POST")
	r.Handle("/api/events/{eventId}", auth(http.HandlerFunc(h.GetEvent))).Methods("GET")
	r.Handle("/api/events/{eventId}/update", auth(http.HandlerFunc(h.UpdateEvent))).Methods("POST")
	r.Handle("/api/events/{eventId}/delete", auth(http.HandlerFunc(h.DeleteEvent))).Methods("POST")
	r.Handle("/api/events/{eventId}/viewers", auth(http.HandlerFunc(h.AddViewer))).Methods("POST")
	r.Handle("/api/events/{eventId}/viewers/{memberId}/remove", auth(http.HandlerFunc(h.RemoveViewer))).Methods("POST")
	r.Handle("/api/events/{eventId}/viewers", auth(http.HandlerFunc(h.ListViewers))).Methods("GET")
}
