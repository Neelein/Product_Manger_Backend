package architecture

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionImportsUseBoundedDomainPackages(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	production := []string{"main.go", "src/adapter", "src/infrastructure", "src/usecase"}
	for _, path := range production {
		abs := filepath.Join(root, path)
		files := []string{abs}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			files, err = filepath.Glob(filepath.Join(abs, "*.go"))
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, file := range files {
			f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, imp := range f.Imports {
				path := strings.Trim(imp.Path.Value, `"`)
				if path == "backend/src/domain" || path == "backend/src/api" || path == "backend/src/database" {
					t.Fatalf("production file %s imports forbidden legacy package %s", file, path)
				}
			}
		}
	}
}

func TestLegacyPackagesAreRemoved(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"src/api", "src/database", "src/domain"} {
		if _, err := os.Stat(filepath.Join(root, path)); err == nil {
			if path == "src/domain" {
				entries, readErr := os.ReadDir(filepath.Join(root, path))
				if readErr != nil {
					t.Fatal(readErr)
				}
				for _, entry := range entries {
					if entry.IsDir() && (entry.Name() == "model" || entry.Name() == "repository" || entry.Name() == "entity") {
						continue
					}
					t.Fatalf("legacy domain entry remains: %s", filepath.Join(path, entry.Name()))
				}
				continue
			}
			t.Fatalf("legacy package directory remains: %s", path)
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
}

func TestAllGoImportsAvoidLegacyPackages(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			pkg := strings.Trim(imp.Path.Value, `"`)
			if pkg == "backend/src/api" || pkg == "backend/src/database" || pkg == "backend/src/domain" {
				return fmt.Errorf("%s imports removed legacy package %s", path, pkg)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func source(t *testing.T, path ...string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(append([]string{root}, path...)...))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRuntimeAdaptersDoNotImportLegacyPackages(t *testing.T) {
	for _, file := range []string{
		"main.go",
		"src/adapter/http/router.go",
		"src/adapter/postgres/repositories.go",
		"src/infrastructure/database.go",
		"src/adapter/session/cache.go",
	} {
		if strings.Contains(source(t, file), `"backend/src/api"`) || strings.Contains(source(t, file), `"backend/src/database"`) {
			t.Fatalf("runtime file %s imports a legacy package", file)
		}
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "src", "adapter", "postgres", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), `"backend/src/database"`) {
			t.Fatalf("postgres adapter %s imports the legacy database package", file)
		}
	}
}

func TestHTTPHandlersDoNotDependOnPostgresConcreteTypes(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "src", "adapter", "http", "*_handler.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "pgxpool") || strings.Contains(string(data), "src/adapter/postgres") {
			t.Fatalf("HTTP handler %s depends on a PostgreSQL concrete type", file)
		}
	}
}

func TestHTTPHandlersUseApplicationServicesOnly(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "src", "adapter", "http", "*_handler.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, forbidden := range []string{"h.repo", "repo usecase.", "codeRepo", "backend/src/domain/repository"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("HTTP handler %s exposes repository-shaped dependency %s", file, forbidden)
			}
		}
		if strings.Contains(text, "usecase.Validate") {
			t.Fatalf("HTTP handler %s calls an exported validation function directly", file)
		}
	}
	checks := map[string][]string{
		"event_handler.go":   {"CreatedBy !="},
		"product_handler.go": {"detail.ID !=", "routeDetail"},
		"chat_handler.go":    {"req.Page <", "req.Limit <", "len(req.MemberIDs) == 0"},
	}
	mutationCalls := map[string][]string{
		"announcement_handler.go": {"h.service.Update(r.Context(), announcement)", "h.service.Delete(r.Context(), announcementID)"},
		"inventory_handler.go":    {"h.service.UpdateInventory(r.Context(), inventory)", "h.service.UpdateItem(r.Context(), item)", "h.service.DeleteInventory(r.Context(), inventoryID)", "h.service.DeleteItem(r.Context(), itemID)"},
		"product_handler.go":      {"h.service.UpdateOption(r.Context(), o)", "h.service.UpdateVariant(r.Context(), v)", "h.service.UpdateDetail(r.Context(), detail)", "h.service.UpdatePrice(r.Context(), price)", "h.service.DeleteOption(r.Context(), o.ID)", "h.service.DeleteVariant(r.Context(), v.ID)"},
	}
	for name, forbidden := range mutationCalls {
		text := source(t, "src/adapter/http", name)
		for _, pattern := range forbidden {
			if strings.Contains(text, pattern) {
				t.Fatalf("handler %s still performs legacy fetch/mutate workflow: %s", name, pattern)
			}
		}
	}
	for name, forbidden := range checks {
		text := source(t, "src/adapter/http", name)
		for _, pattern := range forbidden {
			if strings.Contains(text, pattern) {
				t.Fatalf("HTTP handler %s contains business-rule pattern %q", name, pattern)
			}
		}
	}
	services := source(t, "src/usecase/services.go")
	if strings.Contains(services, "interface{ repository.") {
		t.Fatal("application service interface embeds a repository port")
	}
}

func TestProductionLayerBoundaries(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	router := source(t, "src/adapter/http/router.go")
	if strings.Contains(router, "backend/src/domain") || strings.Contains(router, "backend/src/api") {
		t.Fatal("HTTP router imports a legacy or root domain package")
	}

	files, err := filepath.Glob(filepath.Join(root, "src", "adapter", "http", "*_handler.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if strings.Contains(text, "domain.*Repository") || strings.Contains(text, "backend/src/api") || strings.Contains(text, "backend/src/database") {
			t.Fatalf("HTTP handler %s depends on a forbidden adapter", file)
		}
	}

	files, err = filepath.Glob(filepath.Join(root, "src", "usecase", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		forbidden := []string{"net/http", "github.com/jackc/pgx", "github.com/gorilla/mux"}
		for _, importPath := range forbidden {
			if strings.Contains(text, importPath) {
				t.Fatalf("usecase %s imports %s", file, importPath)
			}
		}
	}
}

func TestDomainEntitiesHaveNoAdapterDependencies(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "src", "domain", "entity", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		for _, forbidden := range []string{"backend/src/adapter", "backend/src/usecase", "net/http", "pgx"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("domain entity %s depends on %s", file, forbidden)
			}
		}
	}
}
