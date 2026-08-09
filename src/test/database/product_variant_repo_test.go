//go:build integration

package database_test

import (
	"context"
	"testing"

	database "backend/src/adapter/postgres"
	domain "backend/src/domain/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductVariantRepositoryPGX_OptionsAndSharedPrice(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product := createTestProduct(t, repo)
	detail := domain.ProductDetail{ProductID: product.ID}
	require.NoError(t, repo.CreateDetail(context.Background(), &detail))
	price := domain.ProductPrice{ProductDetailID: detail.ID, Label: "same price", Amount: 100}
	require.NoError(t, repo.CreatePrice(context.Background(), &price))

	small := domain.ProductOption{ProductDetailID: detail.ID, Name: "size", Value: "S"}
	large := domain.ProductOption{ProductDetailID: detail.ID, Name: "size", Value: "L"}
	color := domain.ProductOption{ProductDetailID: detail.ID, Name: "color", Value: "red"}
	for _, option := range []*domain.ProductOption{&small, &large, &color} {
		require.NoError(t, repo.CreateOption(context.Background(), option))
	}
	options, err := repo.ListOptionsByDetailID(context.Background(), detail.ID)
	require.NoError(t, err)
	assert.Len(t, options, 3)

	v1 := domain.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: price.ID, OptionIDs: []string{small.ID, color.ID}}
	v2 := domain.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: price.ID, OptionIDs: []string{large.ID, color.ID}}
	require.NoError(t, repo.CreateVariant(context.Background(), &v1))
	require.NoError(t, repo.CreateVariant(context.Background(), &v2))
	assert.Equal(t, price.ID, v1.ProductPriceID)
	assert.Equal(t, price.ID, v2.ProductPriceID)

	err = repo.CreateVariant(context.Background(), &domain.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: price.ID, OptionIDs: []string{small.ID, color.ID}})
	assert.Error(t, err)
	v2.OptionIDs = []string{small.ID, color.ID}
	assert.Error(t, repo.UpdateVariant(context.Background(), &v2))
	unchanged, err := repo.GetVariantByID(context.Background(), v2.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{large.ID, color.ID}, unchanged.OptionIDs)

	got, err := repo.GetVariantByID(context.Background(), v1.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{small.ID, color.ID}, got.OptionIDs)

	small.Value = "small"
	require.NoError(t, repo.UpdateOption(context.Background(), &small))
	v1.OptionIDs = []string{small.ID}
	require.NoError(t, repo.UpdateVariant(context.Background(), &v1))
	require.NoError(t, repo.DeleteVariant(context.Background(), v2.ID))
	require.NoError(t, repo.DeleteOption(context.Background(), large.ID))
}

func TestProductVariantRepositoryPGX_InventoryByVariant(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product := createTestProduct(t, repo)
	detail := domain.ProductDetail{ProductID: product.ID}
	require.NoError(t, repo.CreateDetail(context.Background(), &detail))
	price := domain.ProductPrice{ProductDetailID: detail.ID, Label: "shared", Amount: 50}
	require.NoError(t, repo.CreatePrice(context.Background(), &price))
	option := domain.ProductOption{ProductDetailID: detail.ID, Name: "edition", Value: "standard"}
	require.NoError(t, repo.CreateOption(context.Background(), &option))
	variant := domain.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: price.ID, OptionIDs: []string{option.ID}}
	require.NoError(t, repo.CreateVariant(context.Background(), &variant))

	invRepo := database.NewInventoryRepositoryPGX(testPool)
	inv := domain.Inventory{ProductVariantID: variant.ID}
	require.NoError(t, invRepo.CreateInventory(context.Background(), &inv))
	got, err := invRepo.GetInventoryByID(context.Background(), inv.ID)
	require.NoError(t, err)
	assert.Equal(t, variant.ID, got.ProductVariantID)
	assert.Equal(t, price.ID, got.ProductPriceID)
}

func TestProductVariantRepositoryPGX_MapsDuplicateSKU(t *testing.T) {
	defer cleanupProducts(t)
	repo := database.NewProductRepositoryPGX(testPool)
	product := createTestProduct(t, repo)
	detail := domain.ProductDetail{ProductID: product.ID}
	require.NoError(t, repo.CreateDetail(context.Background(), &detail))
	price := domain.ProductPrice{ProductDetailID: detail.ID, Label: "standard", Amount: 10}
	require.NoError(t, repo.CreatePrice(context.Background(), &price))
	firstOption := domain.ProductOption{ProductDetailID: detail.ID, Name: "size", Value: "S"}
	require.NoError(t, repo.CreateOption(context.Background(), &firstOption))
	sku := "SKU-DUPLICATE"
	first := domain.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: price.ID, SKU: &sku, OptionIDs: []string{firstOption.ID}}
	require.NoError(t, repo.CreateVariant(context.Background(), &first))
	option := domain.ProductOption{ProductDetailID: detail.ID, Name: "size", Value: "M"}
	require.NoError(t, repo.CreateOption(context.Background(), &option))

	second := domain.ProductVariant{ProductDetailID: detail.ID, ProductPriceID: price.ID, SKU: first.SKU, OptionIDs: []string{option.ID}}
	err := repo.CreateVariant(context.Background(), &second)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrDuplicateSKU)
	assert.NotContains(t, err.Error(), "R0023")
	assert.NotContains(t, err.Error(), "SQLSTATE")
}
