package schema_test

import (
	"os"
	"strings"
	"testing"
)

func TestVariantInventoryMigrationSafelyRemovesLegacyPriceColumn(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/021_link_price_inventory_by_variant.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(schema)
	for _, required := range []string{
		"WHERE i.product_variant_id IS NULL",
		"i.product_price_id IS DISTINCT FROM v.product_price_id",
		"RAISE EXCEPTION",
		"DROP CONSTRAINT IF EXISTS inventories_product_price_id_fkey",
		"DROP CONSTRAINT IF EXISTS inventories_product_price_id_key",
		"DROP COLUMN IF EXISTS product_price_id",
		"INSERT INTO inventories(product_variant_id, status)",
		"JOIN inventories i ON i.product_variant_id = v.id",
	} {
		if !strings.Contains(s, required) {
			t.Errorf("migration missing %q", required)
		}
	}
}

func TestVariantInventoryMigrationDoesNotUseLegacyColumnInFinalQueries(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/021_link_price_inventory_by_variant.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	final := string(schema)[strings.Index(string(schema), "CREATE OR REPLACE FUNCTION create_inventory"):]
	if strings.Contains(final, "i.product_price_id") || strings.Contains(final, "inventories.product_price_id") {
		t.Fatal("final migration function definitions must derive prices through product_variants")
	}
}

func TestVariantInventoryMigrationDoesNotExposeInventoryIDOnPrices(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/021_link_price_inventory_by_variant.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(schema)
	priceFunctions := s[strings.Index(s, "CREATE FUNCTION get_product_price_by_id"):]
	if strings.Contains(priceFunctions, "inventory_id") {
		t.Fatal("price function contracts must not expose inventory_id")
	}
}

func TestInventoryContractRepairMigrationDefinesTwelveColumns(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/022_repair_inventory_function_contract.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(schema)
	for _, function := range []string{"get_inventory_by_id", "get_inventory_by_price_id", "list_inventories"} {
		if !strings.Contains(s, "CREATE FUNCTION "+function) {
			t.Errorf("repair migration missing %s", function)
		}
	}
	if strings.Count(s, "variant_name VARCHAR") != 3 {
		t.Fatalf("repair migration must expose variant_name in all three inventory functions")
	}
	if strings.Count(s, "updated_at TIMESTAMPTZ") != 3 {
		t.Fatalf("repair migration must preserve the complete twelve-column inventory contract")
	}
}
