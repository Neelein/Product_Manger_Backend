//go:build integration

package api_test

import (
	apphttp "backend/src/adapter/http"
	"backend/src/adapter/postgres"
	"backend/src/adapter/session"
	"backend/src/usecase"
)

func composeProductHandler(repo *postgres.ProductRepositoryPGX) *apphttp.ProductHandler {
	return apphttp.NewProductHandler(usecase.NewProductService(repo))
}

func composeInventoryHandler(repo *postgres.InventoryRepositoryPGX) *apphttp.InventoryHandler {
	return apphttp.NewInventoryHandler(usecase.NewInventoryService(repo))
}

func composeMemberHandler(member *postgres.MemberRepositoryPGX, sessions *session.SessionCache, codes *postgres.RegistrationCodeRepositoryPGX) *apphttp.MemberHandler {
	return apphttp.NewMemberHandler(
		usecase.NewMemberService(member, sessions, codes),
		usecase.NewSessionService(sessions),
		usecase.NewRegistrationCodeService(codes),
	)
}

func composeRegistrationCodeHandler(repo *postgres.RegistrationCodeRepositoryPGX) *apphttp.RegistrationCodeHandler {
	return apphttp.NewRegistrationCodeHandler(usecase.NewRegistrationCodeService(repo))
}

func composeCategoryHandler(repo *postgres.CategoryRepositoryPGX) *apphttp.CategoryHandler {
	return apphttp.NewCategoryHandler(usecase.NewCategoryService(repo))
}

func composeChatHandler(repo *postgres.ChatRoomRepositoryPGX) *apphttp.ChatRoomHandler {
	return apphttp.NewChatRoomHandler(usecase.NewChatService(repo))
}
