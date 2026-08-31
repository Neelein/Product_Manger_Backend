package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductPriceJSONDoesNotExposeInventoryID(t *testing.T) {
	price := ProductPrice{ID: "price-1", ProductVariantID: stringPtr("variant-1")}

	payload, err := json.Marshal(price)
	require.NoError(t, err)
	require.NotContains(t, string(payload), "inventory_id")
	require.Contains(t, string(payload), "product_variant_id")
}

func stringPtr(value string) *string { return &value }
