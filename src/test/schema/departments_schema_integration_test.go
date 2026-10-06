//go:build integration

package schema

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDepartmentsSchemaIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must be set for integration schema tests")
	}
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var departmentTable, departmentColumn, memberColumn, phoneColumn string
	err = pool.QueryRow(context.Background(), `SELECT to_regclass('public.departments'), (SELECT column_name FROM information_schema.columns WHERE table_name = 'departments' AND column_name = 'code'), (SELECT column_name FROM information_schema.columns WHERE table_name = 'members' AND column_name = 'department_code'), (SELECT column_name FROM information_schema.columns WHERE table_name = 'members' AND column_name = 'phone')`).Scan(&departmentTable, &departmentColumn, &memberColumn, &phoneColumn)
	if err != nil {
		t.Fatal(err)
	}
	if departmentTable != "departments" || departmentColumn != "code" || memberColumn != "department_code" || phoneColumn != "phone" {
		t.Fatalf("unexpected department schema: table=%q department=%q member=%q phone=%q", departmentTable, departmentColumn, memberColumn, phoneColumn)
	}
}
