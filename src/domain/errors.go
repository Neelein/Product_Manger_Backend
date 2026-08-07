package domain

import "errors"

var ErrProductNotFound = errors.New("product not found")
var ErrMemberNotFound = errors.New("member not found")
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrDetailNotFound = errors.New("detail not found")
var ErrPriceNotFound = errors.New("price not found")
var ErrInventoryNotFound = errors.New("inventory not found")
var ErrInventoryItemNotFound = errors.New("inventory item not found")
var ErrAnnouncementNotFound = errors.New("announcement not found")
var ErrChatRoomNotFound = errors.New("chat room not found")
var ErrChatMessageNotFound = errors.New("chat message not found")
var ErrNotRoomMember = errors.New("user is not a member of this room")
var ErrEventNotFound = errors.New("event not found")
var ErrNotEventOwner = errors.New("not the event owner")
var ErrEventViewerNotFound = errors.New("event viewer not found")
