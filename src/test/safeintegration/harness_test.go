package safeintegration

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestSchemaIdentifierIsQuoted(t *testing.T) {
	got := pgx.Identifier{"integration_example"}.Sanitize()
	if got != `"integration_example"` {
		t.Fatalf("schema identifier = %q", got)
	}
}
