package usecase

import (
	"backend/src/domain/model"
	"context"
	"sort"
	"strings"
	"testing"
)

type departmentMemberRepo struct{ members map[string]*model.Member }

func (r *departmentMemberRepo) Create(context.Context, *model.Member) error { return nil }
func (r *departmentMemberRepo) GetByEmail(context.Context, string) (*model.Member, error) {
	return nil, nil
}
func (r *departmentMemberRepo) GetByID(_ context.Context, id string) (*model.Member, error) {
	return r.members[id], nil
}
func (r *departmentMemberRepo) Update(context.Context, *model.Member) error          { return nil }
func (r *departmentMemberRepo) UpdatePassword(context.Context, string, string) error { return nil }

type departmentDirectory struct {
	updatedCode                       int
	entries                           []model.EmployeeDirectoryEntry
	gotCode, gotPage, gotLimit        int
	gotQuery, gotSortBy, gotSortOrder string
}

func (d *departmentDirectory) ListEmployees(context.Context, string, int, int) ([]model.Member, int, error) {
	return []model.Member{}, 0, nil
}
func (d *departmentDirectory) ListEmployeesByDepartment(_ context.Context, code int, query string, page, limit int, sortBy, sortOrder string) ([]model.EmployeeDirectoryEntry, int, error) {
	d.gotCode, d.gotQuery, d.gotPage, d.gotLimit = code, query, page, limit
	d.gotSortBy, d.gotSortOrder = sortBy, sortOrder
	query = strings.ToLower(query)
	entries := make([]model.EmployeeDirectoryEntry, 0, len(d.entries))
	for _, entry := range d.entries {
		if entry.Department.Code == 0 || entry.Department.Code != code {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(entry.ID+entry.Name+entry.Email+entry.Phone), query) {
			continue
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		left, right := entries[i].Name, entries[j].Name
		if sortBy == "email" {
			left, right = entries[i].Email, entries[j].Email
		}
		if sortBy == "phone" {
			left, right = entries[i].Phone, entries[j].Phone
		}
		if strings.EqualFold(sortOrder, "desc") {
			return left > right
		}
		return left < right
	})
	total := len(entries)
	start := (page - 1) * limit
	if start >= total {
		return []model.EmployeeDirectoryEntry{}, total, nil
	}
	end := start + limit
	if end > len(entries) {
		end = len(entries)
	}
	entries = entries[start:end]
	return entries, total, nil
}
func (d *departmentDirectory) UpdateDepartment(_ context.Context, _ string, code int) error {
	d.updatedCode = code
	return nil
}

type departmentRepo struct{ departments []model.Department }

func (r *departmentRepo) List(context.Context) ([]model.Department, error) { return r.departments, nil }
func (r *departmentRepo) Create(_ context.Context, name string) (*model.Department, error) {
	return &model.Department{Code: 1, Name: name}, nil
}
func (r *departmentRepo) Update(_ context.Context, code int, name string) (*model.Department, error) {
	return &model.Department{Code: code, Name: name}, nil
}
func (r *departmentRepo) Delete(context.Context, int) error { return nil }

func TestDepartmentService_AuthorizationAndAssignment(t *testing.T) {
	repo := &departmentMemberRepo{members: map[string]*model.Member{
		"admin":    {ID: "admin", MemberType: "employee", Permission: "admin"},
		"staff":    {ID: "staff", MemberType: "employee", Permission: "staff"},
		"employee": {ID: "employee", MemberType: "employee"},
	}}
	directory := &departmentDirectory{}
	service := NewDepartmentService(repo, directory, &departmentRepo{departments: []model.Department{{Code: 2, Name: "Sales"}}})

	tests := []struct {
		name     string
		actor    string
		member   string
		code     int
		wantErr  error
		wantCode int
	}{
		{name: "admin assigns", actor: "admin", member: "employee", code: 2, wantCode: 2},
		{name: "admin removes with zero", actor: "admin", member: "employee", code: 0, wantCode: 0},
		{name: "non admin forbidden", actor: "staff", member: "employee", code: 2, wantErr: model.ErrForbidden},
		{name: "unknown department", actor: "admin", member: "employee", code: 3, wantErr: model.ErrDepartmentNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateMemberDepartment(context.Background(), tt.actor, tt.member, tt.code)
			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && directory.updatedCode != tt.wantCode {
				t.Fatalf("code = %d, want %d", directory.updatedCode, tt.wantCode)
			}
		})
	}
}

func TestDepartmentService_TrimsAndValidatesName(t *testing.T) {
	service := NewDepartmentService(
		&departmentMemberRepo{members: map[string]*model.Member{"admin": {MemberType: "employee", Permission: "admin"}}},
		&departmentDirectory{},
		&departmentRepo{},
	)
	department, err := service.Create(context.Background(), "admin", "  Sales  ")
	if err != nil || department.Name != "Sales" {
		t.Fatalf("department = %#v, error = %v", department, err)
	}
	_, err = service.Create(context.Background(), "admin", "   ")
	if err != model.ErrInvalidDepartment {
		t.Fatalf("error = %v, want invalid department", err)
	}
}

func TestDepartmentService_ListDepartmentEmployees_AllowsAnyEmployeeAndScopesDepartment(t *testing.T) {
	directory := &departmentDirectory{entries: []model.EmployeeDirectoryEntry{
		{ID: "sales-1", Name: "Alice", Email: "alice@example.com", Department: model.Department{Code: 4, Name: "Sales"}},
		{ID: "sales-2", Name: "Bob", Department: model.Department{Code: 4, Name: "Sales"}},
		{ID: "other-1", Name: "Alice", Department: model.Department{Code: 9, Name: "Other"}},
		{ID: "ungrouped", Name: "Alice", Department: model.Department{}},
	}}
	service := NewDepartmentService(
		&departmentMemberRepo{members: map[string]*model.Member{"staff": {MemberType: "employee", Permission: "staff"}}},
		directory,
		&departmentRepo{departments: []model.Department{{Code: 4, Name: "Sales"}}},
	)
	allEntries, allTotal, err := service.ListDepartmentEmployees(context.Background(), "staff", 4, "", 1, 20, "name", "asc")
	if err != nil || allTotal != 2 || len(allEntries) != 2 || allEntries[0].ID != "sales-1" || allEntries[1].ID != "sales-2" {
		t.Fatalf("scoped entries = %#v, total = %d, error = %v", allEntries, allTotal, err)
	}
	entries, total, err := service.ListDepartmentEmployees(context.Background(), "staff", 4, "  alice ", 1, 5, "email", "desc")
	if err != nil {
		t.Fatalf("list employees error = %v", err)
	}
	if len(entries) != 1 || total != 1 || entries[0].ID != "sales-1" {
		t.Fatalf("entries = %#v, total = %d", entries, total)
	}
	if directory.gotCode != 4 || directory.gotQuery != "alice" || directory.gotPage != 1 || directory.gotLimit != 5 || directory.gotSortBy != "email" || directory.gotSortOrder != "desc" {
		t.Fatalf("repository arguments = code %d, query %q, page %d, limit %d, sort %q/%q", directory.gotCode, directory.gotQuery, directory.gotPage, directory.gotLimit, directory.gotSortBy, directory.gotSortOrder)
	}
	pageTwo, total, err := service.ListDepartmentEmployees(context.Background(), "staff", 4, "", 2, 1, "name", "asc")
	if err != nil || total != 2 || len(pageTwo) != 1 || pageTwo[0].ID != "sales-2" {
		t.Fatalf("page two = %#v, total = %d, error = %v", pageTwo, total, err)
	}
	_, _, err = service.ListDepartmentEmployees(context.Background(), "staff", 9, "", 1, 20, "name", "asc")
	if err != model.ErrDepartmentNotFound {
		t.Fatalf("missing department error = %v, want %v", err, model.ErrDepartmentNotFound)
	}
}

func TestDepartmentService_ListDepartmentEmployees_RejectsUnauthenticatedAndNonEmployees(t *testing.T) {
	service := NewDepartmentService(
		&departmentMemberRepo{members: map[string]*model.Member{
			"customer": {MemberType: "customer"}, "employee": {MemberType: "employee"},
		}}, &departmentDirectory{}, &departmentRepo{departments: []model.Department{{Code: 4}}},
	)
	for _, actor := range []string{"", "customer"} {
		_, _, err := service.ListDepartmentEmployees(context.Background(), actor, 4, "", 1, 20, "name", "asc")
		if err != model.ErrForbidden {
			t.Errorf("actor %q error = %v, want %v", actor, err, model.ErrForbidden)
		}
	}
}

func TestDepartmentService_ListDepartmentEmployees_RejectsInvalidPagination(t *testing.T) {
	service := NewDepartmentService(
		&departmentMemberRepo{members: map[string]*model.Member{"employee": {MemberType: "employee"}}},
		&departmentDirectory{}, &departmentRepo{departments: []model.Department{{Code: 4}}},
	)
	for _, values := range [][2]int{{0, 20}, {1, 0}, {1, 101}} {
		_, _, err := service.ListDepartmentEmployees(context.Background(), "employee", 4, "", values[0], values[1], "name", "asc")
		if err != model.ErrInvalidDepartment {
			t.Errorf("page/limit %v error = %v, want %v", values, err, model.ErrInvalidDepartment)
		}
	}
}
