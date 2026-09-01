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
	employeeAuth := RequireEmployee(auth)

	r.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	r.Handle("/api/products", employeeAuth(http.HandlerFunc(h.CreateProduct))).Methods("POST")
	r.HandleFunc("/api/products/{productId}", h.GetProduct).Methods("GET")
	r.HandleFunc("/api/products/{productId}/images", h.ListImages).Methods("GET")
	r.Handle("/api/products/{productId}/images", employeeAuth(http.HandlerFunc(h.UploadImages))).Methods("POST")
	r.Handle("/api/products/{productId}/images/{imageId}/delete", employeeAuth(http.HandlerFunc(h.DeleteImage))).Methods("POST")
	r.Handle("/api/products/{productId}/update", employeeAuth(http.HandlerFunc(h.UpdateProduct))).Methods("POST")
	r.Handle("/api/products/{productId}/delete", employeeAuth(http.HandlerFunc(h.DeleteProduct))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail", h.GetDetail).Methods("GET")
	r.Handle("/api/products/{productId}/detail/update", employeeAuth(http.HandlerFunc(h.UpdateDetail))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/prices", h.ListPrices).Methods("GET")
	r.HandleFunc("/api/products/{productId}/detail/prices/{priceId}", h.GetPrice).Methods("GET")
	r.Handle("/api/products/{productId}/detail/prices/{priceId}/update", employeeAuth(http.HandlerFunc(h.UpdatePrice))).Methods("POST")
	r.Handle("/api/products/{productId}/details", employeeAuth(http.HandlerFunc(h.CreateDetail))).Methods("POST")
	r.Handle("/api/products/{productId}/details/{detailId}/prices", employeeAuth(http.HandlerFunc(h.CreatePrice))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/options", h.ListOptions).Methods("GET")
	r.Handle("/api/products/{productId}/detail/options", employeeAuth(http.HandlerFunc(h.CreateOption))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/options/{optionId}", h.GetOption).Methods("GET")
	r.Handle("/api/products/{productId}/detail/options/{optionId}/update", employeeAuth(http.HandlerFunc(h.UpdateOption))).Methods("POST")
	r.Handle("/api/products/{productId}/detail/options/{optionId}/delete", employeeAuth(http.HandlerFunc(h.DeleteOption))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/variants", h.ListVariants).Methods("GET")
	r.Handle("/api/products/{productId}/detail/variants", employeeAuth(http.HandlerFunc(h.CreateVariant))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/variants/{variantId}", h.GetVariant).Methods("GET")
	r.PathPrefix("/media/images/products/").Handler(http.StripPrefix("/media/images/products/", http.FileServer(http.Dir(filepath.Join(os.Getenv("MEDIA_ROOT"), "images/products")))))
	r.Handle("/api/products/{productId}/detail/variants/{variantId}/update", employeeAuth(http.HandlerFunc(h.UpdateVariant))).Methods("POST")
	r.Handle("/api/products/{productId}/detail/variants/{variantId}/delete", employeeAuth(http.HandlerFunc(h.DeleteVariant))).Methods("POST")
}

func RegisterInventoryRoutes(r *mux.Router, service usecase.InventoryService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewInventoryHandler(service)
	auth := AuthMiddleware(sessionService, memberService)
	employeeAuth := RequireEmployee(auth)

	r.HandleFunc("/api/inventories", h.ListInventories).Methods("GET")
	r.Handle("/api/inventories", employeeAuth(http.HandlerFunc(h.CreateInventory))).Methods("POST")
	r.HandleFunc("/api/inventories/{inventoryId}", h.GetInventory).Methods("GET")
	r.Handle("/api/inventories/{inventoryId}/update", employeeAuth(http.HandlerFunc(h.UpdateInventory))).Methods("POST")
	r.Handle("/api/inventories/{inventoryId}/delete", employeeAuth(http.HandlerFunc(h.DeleteInventory))).Methods("POST")
	r.HandleFunc("/api/inventories/{inventoryId}/items", h.ListItems).Methods("GET")
	r.Handle("/api/inventories/{inventoryId}/items", employeeAuth(http.HandlerFunc(h.CreateItem))).Methods("POST")
	r.HandleFunc("/api/inventories/{inventoryId}/items/{itemId}", h.GetItem).Methods("GET")
	r.Handle("/api/inventories/{inventoryId}/items/{itemId}/update", employeeAuth(http.HandlerFunc(h.UpdateItem))).Methods("POST")
	r.Handle("/api/inventories/{inventoryId}/items/{itemId}/delete", employeeAuth(http.HandlerFunc(h.DeleteItem))).Methods("POST")
}

func RegisterMemberRoutes(r *mux.Router, memberService usecase.MemberService, sessionService usecase.SessionService, codeService usecase.RegistrationCodeService) {
	h := NewMemberHandler(memberService, sessionService, codeService)
	auth := AuthMiddleware(sessionService, memberService)

	r.HandleFunc("/api/members/register", h.RegisterMember).Methods("POST")
	r.HandleFunc("/api/members/login", h.LoginMember).Methods("POST")
	r.HandleFunc("/api/members/logout", h.LogoutMember).Methods("POST")
	r.Handle("/api/members/me", auth(http.HandlerFunc(h.GetCurrentMember))).Methods("GET")
	r.Handle("/api/members/update", auth(http.HandlerFunc(h.UpdateMember))).Methods("POST")
	r.Handle("/api/members/{memberId}/permission", auth(http.HandlerFunc(h.UpdateMemberPermission))).Methods("POST")
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
	employeeAuth := RequireEmployee(auth)

	r.HandleFunc("/api/categories", h.ListCategories).Methods("GET")
	r.Handle("/api/categories", employeeAuth(http.HandlerFunc(h.CreateCategory))).Methods("POST")
	r.Handle("/api/categories/{id}/update", employeeAuth(http.HandlerFunc(h.UpdateCategory))).Methods("POST")
	r.Handle("/api/categories/{id}/delete", employeeAuth(http.HandlerFunc(h.DeleteCategory))).Methods("POST")
}

func RegisterAnnouncementRoutes(r *mux.Router, service usecase.AnnouncementService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewAnnouncementHandler(service)
	auth := AuthMiddleware(sessionService, memberService)
	employeeAuth := RequireEmployee(auth)

	r.HandleFunc("/api/announcements", h.ListAnnouncements).Methods("GET")
	r.Handle("/api/announcements", employeeAuth(http.HandlerFunc(h.CreateAnnouncement))).Methods("POST")
	r.HandleFunc("/api/announcements/{announcementId}", h.GetAnnouncement).Methods("GET")
	r.Handle("/api/announcements/{announcementId}/update", employeeAuth(http.HandlerFunc(h.UpdateAnnouncement))).Methods("POST")
	r.Handle("/api/announcements/{announcementId}/delete", employeeAuth(http.HandlerFunc(h.DeleteAnnouncement))).Methods("POST")

	r.PathPrefix("/media/images/announcements/").Handler(http.StripPrefix("/media/images/announcements/", http.FileServer(http.Dir(filepath.Join(os.Getenv("MEDIA_ROOT"), "images/announcements")))))
}

func RegisterChatRoutes(r *mux.Router, service usecase.ChatService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewChatRoomHandler(service)
	auth := AuthMiddleware(sessionService, memberService)
	employeeAuth := RequireEmployee(auth)

	r.Handle("/api/chat/rooms", employeeAuth(http.HandlerFunc(h.CreateRoom))).Methods("POST")
	r.Handle("/api/chat/rooms", employeeAuth(http.HandlerFunc(h.ListRooms))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}", employeeAuth(http.HandlerFunc(h.GetRoom))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/update", employeeAuth(http.HandlerFunc(h.UpdateRoom))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/delete", employeeAuth(http.HandlerFunc(h.DeleteRoom))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/members", employeeAuth(http.HandlerFunc(h.AddMembers))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/members/{memberId}/remove", employeeAuth(http.HandlerFunc(h.RemoveMember))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/messages", employeeAuth(http.HandlerFunc(h.ListMessages))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/messages", employeeAuth(http.HandlerFunc(h.SendMessage))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/messages/{messageId}/delete", employeeAuth(http.HandlerFunc(h.DeleteMessage))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/read", employeeAuth(http.HandlerFunc(h.MarkAsRead))).Methods("POST")
	r.Handle("/api/chat/rooms/{roomId}/messages/{messageId}/read-by", employeeAuth(http.HandlerFunc(h.GetReadBy))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/unread", employeeAuth(http.HandlerFunc(h.CountUnread))).Methods("GET")
	r.Handle("/api/chat/rooms/{roomId}/available-members", employeeAuth(http.HandlerFunc(h.ListAvailableMembers))).Methods("POST")
}

func RegisterEventRoutes(r *mux.Router, service usecase.EventService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewEventHandler(service)
	auth := AuthMiddleware(sessionService, memberService)
	employeeAuth := RequireEmployee(auth)

	r.Handle("/api/events", auth(http.HandlerFunc(h.ListEventsByMonth))).Methods("GET")
	r.Handle("/api/events", employeeAuth(http.HandlerFunc(h.CreateEvent))).Methods("POST")
	r.Handle("/api/events/{eventId}", auth(http.HandlerFunc(h.GetEvent))).Methods("GET")
	r.Handle("/api/events/{eventId}/update", employeeAuth(http.HandlerFunc(h.UpdateEvent))).Methods("POST")
	r.Handle("/api/events/{eventId}/delete", employeeAuth(http.HandlerFunc(h.DeleteEvent))).Methods("POST")
	r.Handle("/api/events/{eventId}/viewers", employeeAuth(http.HandlerFunc(h.AddViewer))).Methods("POST")
	r.Handle("/api/events/{eventId}/viewers/{memberId}/remove", employeeAuth(http.HandlerFunc(h.RemoveViewer))).Methods("POST")
	r.Handle("/api/events/{eventId}/viewers", auth(http.HandlerFunc(h.ListViewers))).Methods("GET")
}

func RegisterOrderRoutes(r *mux.Router, service usecase.OrderService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewOrderHandler(service)
	auth := AuthMiddleware(sessionService, memberService)
	employee := RequireEmployee(auth)
	r.Handle("/api/orders", auth(http.HandlerFunc(h.Create))).Methods("POST")
	r.Handle("/api/orders", auth(http.HandlerFunc(h.List))).Methods("GET")
	r.Handle("/api/orders/{orderId}", auth(http.HandlerFunc(h.Get))).Methods("GET")
	r.Handle("/api/orders/{orderId}/cancel", auth(http.HandlerFunc(h.Cancel))).Methods("POST")
	r.Handle("/api/orders/{orderId}/status", employee(http.HandlerFunc(h.Status))).Methods("POST")
	r.Handle("/api/orders/{orderId}/history", auth(http.HandlerFunc(h.History))).Methods("GET")
}

func RegisterPaymentRoutes(r *mux.Router, service usecase.PaymentService, memberService usecase.MemberService, sessionService usecase.SessionService) {
	h := NewPaymentHandler(service)
	auth := AuthMiddleware(sessionService, memberService)
	r.Handle("/api/orders/{orderId}/payments", auth(http.HandlerFunc(h.Pay))).Methods("POST")
}
