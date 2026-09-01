package usecase

import "errors"

var (
	ErrInvalidCategoryName      = errors.New("category name is required")
	ErrInvalidInventoryVariant  = errors.New("product_variant_id is required")
	ErrEventTitleRequired       = errors.New("title is required")
	ErrEventMonthRequired       = errors.New("year and month query parameters are required")
	ErrEventMonthInvalid        = errors.New("invalid month")
	ErrEventYearInvalid         = errors.New("invalid year")
	ErrInvalidChatName          = errors.New("name is required")
	ErrInvalidChatMembers       = errors.New("member_ids is required")
	ErrInvalidChatPage          = errors.New("invalid page parameter")
	ErrInvalidChatLimit         = errors.New("invalid limit parameter")
	ErrRegistrationCodeRequired = errors.New("registration code is required")
	ErrInvalidAnnouncement      = errors.New("title and content are required")
	ErrInvalidMessageID         = errors.New("message_id is required")
	ErrInvalidPasswordChange    = errors.New("new password must be 8-16 characters")
	ErrPasswordConfirmation     = errors.New("new passwords do not match")
)
