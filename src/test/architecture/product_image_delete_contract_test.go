package architecture_test

import (
	"os"
	"strings"
	"testing"
)

func TestProductImageDeleteMigrationReturnsScopedImageMetadata(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/023_delete_product_image.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(schema)
	for _, required := range []string{
		"CREATE FUNCTION delete_product_image",
		"product_images.product_id = p_product_id",
		"product_images.id = p_image_id",
		"RETURNING product_images.id, product_images.product_id, product_images.filename, product_images.created_at",
	} {
		if !strings.Contains(s, required) {
			t.Errorf("migration missing %q", required)
		}
	}
}
