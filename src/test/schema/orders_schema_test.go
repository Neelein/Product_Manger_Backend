package schema_test

import (
	"os"
	"strings"
	"testing"
)

func TestOrdersMigrationContainsIntegrityConstraints(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/018_members_orders.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(schema)
	for _, required := range []string{
		"CREATE TABLE orders", "order_no VARCHAR(64)", "UPDATE orders SET order_no = 'ORD-' || id::text",
		"orders_order_no_key UNIQUE (order_no)", "customer_snapshot JSONB NOT NULL", "shipping_address_snapshot JSONB NOT NULL",
		"CREATE TABLE order_items", "quantity INTEGER NOT NULL CHECK (quantity > 0)",
		"CREATE TABLE order_status_history", "CREATE TABLE inventory_reservations",
		"inventory_reservations_order_item_order_fkey", "FOREIGN KEY (order_item_id, order_id)",
		"CREATE UNIQUE INDEX inventory_reservations_one_active_item", "WHERE status = 'active'",
	} {
		if !strings.Contains(s, required) {
			t.Errorf("migration missing %q", required)
		}
	}
}

func TestOrdersMigrationAllowsCustomerAndEmployeeOrderOwners(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/018_members_orders.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(schema)
	if strings.Contains(s, "orders_customer_type_trigger") || strings.Contains(s, "ensure_order_customer_is_customer") {
		t.Fatal("orders must allow both customer and employee members as customer_id")
	}
}

func TestMemberMigrationUsesSafeLegacyRoleMapping(t *testing.T) {
	schema, err := os.ReadFile("../../../db/migrations/018_members_orders.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	s := string(schema)
	if strings.Contains(s, "SET member_type = 'employee', permission = COALESCE(permission, role)") {
		t.Fatal("legacy member rows must not all become employees")
	}
	for _, required := range []string{"lower(role) IN ('admin', 'employee')", "lower(role) = 'admin'", "ELSE 'customer'"} {
		if !strings.Contains(s, required) {
			t.Errorf("migration missing safe mapping %q", required)
		}
	}
}

func TestRegistrationCodeConsumptionIsNotClearedByMemberDelete(t *testing.T) {
	for _, path := range []string{"../../../db/migrations/015_create_registration_codes.up.sql", "../../../db/migrations/018_members_orders.up.sql"} {
		schema, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(schema), "used_by") || !strings.Contains(string(schema), "ON DELETE RESTRICT") {
			t.Errorf("%s does not preserve consumed registration codes", path)
		}
	}
	registration, err := os.ReadFile("../../../db/migrations/015_create_registration_codes.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(registration), "WHERE id = p_id AND used_at IS NULL") {
		t.Fatal("used registration codes must not be deletable and reusable")
	}
}
