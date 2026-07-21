package api

import (
	"net/http"
	"os"
	"path/filepath"

	"backend/src/domain"

	"github.com/gorilla/mux"
)

func RegisterProductRoutes(r *mux.Router, repo domain.ProductRepository, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewProductHandler(repo)
	auth := AuthMiddleware(sessionRepo, memberRepo)

	r.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	r.Handle("/api/products", auth(http.HandlerFunc(h.CreateProduct))).Methods("POST")
	r.HandleFunc("/api/products/{productId}", h.GetProduct).Methods("GET")
	r.HandleFunc("/api/products/{productId}/update", h.UpdateProduct).Methods("POST")
	r.HandleFunc("/api/products/{productId}/delete", h.DeleteProduct).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail", h.GetDetail).Methods("GET")
	r.Handle("/api/products/{productId}/detail/update", auth(http.HandlerFunc(h.UpdateDetail))).Methods("POST")
	r.HandleFunc("/api/products/{productId}/detail/prices", h.ListPrices).Methods("GET")
	r.HandleFunc("/api/products/{productId}/detail/prices/{priceId}", h.GetPrice).Methods("GET")
	r.Handle("/api/products/{productId}/detail/prices/{priceId}/update", auth(http.HandlerFunc(h.UpdatePrice))).Methods("POST")
	r.Handle("/api/products/{productId}/details", auth(http.HandlerFunc(h.CreateDetail))).Methods("POST")
	r.Handle("/api/products/{productId}/details/{detailId}/prices", auth(http.HandlerFunc(h.CreatePrice))).Methods("POST")
}

func RegisterInventoryRoutes(r *mux.Router, repo domain.InventoryRepository, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewInventoryHandler(repo)
	auth := AuthMiddleware(sessionRepo, memberRepo)

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

func RegisterMemberRoutes(r *mux.Router, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewMemberHandler(memberRepo, sessionRepo)
	auth := AuthMiddleware(sessionRepo, memberRepo)

	r.HandleFunc("/api/members/register", h.RegisterMember).Methods("POST")
	r.HandleFunc("/api/members/login", h.LoginMember).Methods("POST")
	r.HandleFunc("/api/members/logout", h.LogoutMember).Methods("POST")
	r.Handle("/api/members/me", auth(http.HandlerFunc(h.GetCurrentMember))).Methods("GET")
	r.Handle("/api/members/update", auth(http.HandlerFunc(h.UpdateMember))).Methods("POST")
}

func RegisterAnnouncementRoutes(r *mux.Router, repo domain.AnnouncementRepository, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewAnnouncementHandler(repo)
	auth := AuthMiddleware(sessionRepo, memberRepo)

	r.HandleFunc("/api/announcements", h.ListAnnouncements).Methods("GET")
	r.Handle("/api/announcements", auth(http.HandlerFunc(h.CreateAnnouncement))).Methods("POST")
	r.HandleFunc("/api/announcements/{announcementId}", h.GetAnnouncement).Methods("GET")
	r.Handle("/api/announcements/{announcementId}/update", auth(http.HandlerFunc(h.UpdateAnnouncement))).Methods("POST")
	r.Handle("/api/announcements/{announcementId}/delete", auth(http.HandlerFunc(h.DeleteAnnouncement))).Methods("POST")

	r.PathPrefix("/media/images/announcements/").Handler(http.StripPrefix("/media/images/announcements/", http.FileServer(http.Dir(filepath.Join(os.Getenv("MEDIA_ROOT"), "images/announcements")))))
}

func RegisterChatRoutes(r *mux.Router, repo domain.ChatRoomRepository, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewChatRoomHandler(repo)
	auth := AuthMiddleware(sessionRepo, memberRepo)

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
}
