package domain

import "errors"

var ErrProductNotFound = errors.New("product not found")
var ErrMemberNotFound = errors.New("member not found")
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrSessionNotFound = errors.New("session not found")
var ErrSessionExpired = errors.New("session expired")
var ErrDetailNotFound = errors.New("detail not found")
var ErrPriceNotFound = errors.New("price not found")
var ErrDeviceMismatch = errors.New("device fingerprint mismatch")
var ErrInventoryNotFound = errors.New("inventory not found")
var ErrInventoryItemNotFound = errors.New("inventory item not found")
