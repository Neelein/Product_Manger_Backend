package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDepartmentsMigrationContract(t *testing.T) {
	path := filepath.Join("..", "..", "..", "db", "migrations", "025_create_departments.up.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ToLower(string(content))
	for _, required := range []string{
		"create table departments",
		"generated always as identity",
		"check (code > 0)",
		"btrim(name) <> ''",
		"unique index departments_name_trimmed_unique",
		"add column department_code integer not null default 0",
		"department_code_non_negative",
	} {
		if !strings.Contains(sql, required) {
			t.Errorf("migration missing %q", required)
		}
	}
	phonePath := filepath.Join("..", "..", "..", "db", "migrations", "026_add_member_phone.up.sql")
	phoneMigration, err := os.ReadFile(phonePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(phoneMigration)), "add column phone varchar(50) not null default ''") {
		t.Error("phone migration missing phone column")
	}
}
