package http

import (
	"backend/src/domain/model"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

type fakeDepartmentService struct {
	employees                          []model.Member
	directory                          []model.EmployeeDirectoryEntry
	directoryErr                       error
	gotCode, gotPage, gotLimit         int
	gotSearch, gotSortBy, gotSortOrder string
}

func (f *fakeDepartmentService) List(context.Context, string) ([]model.Department, error) {
	return []model.Department{}, nil
}
func (f *fakeDepartmentService) Create(context.Context, string, string) (*model.Department, error) {
	return &model.Department{Code: 1, Name: "Sales"}, nil
}
func (f *fakeDepartmentService) Update(context.Context, string, int, string) (*model.Department, error) {
	return &model.Department{}, nil
}
func (f *fakeDepartmentService) Delete(context.Context, string, int) error { return nil }
func (f *fakeDepartmentService) ListEmployees(context.Context, string, string, int, int) ([]model.Member, int, error) {
	return f.employees, len(f.employees), nil
}
func (f *fakeDepartmentService) ListDepartmentEmployees(_ context.Context, _ string, code int, search string, page, limit int, sortBy, sortOrder string) ([]model.EmployeeDirectoryEntry, int, error) {
	f.gotCode, f.gotSearch, f.gotPage, f.gotLimit = code, search, page, limit
	f.gotSortBy, f.gotSortOrder = sortBy, sortOrder
	if f.directoryErr != nil {
		return nil, 0, f.directoryErr
	}
	return f.directory, len(f.directory), nil
}

func TestDepartmentHandler_ListDepartmentEmployees(t *testing.T) {
	handler := NewDepartmentHandler(&fakeDepartmentService{directory: []model.EmployeeDirectoryEntry{{ID: "1", Name: "A", Email: "a@example.com", Phone: "0912", Department: model.Department{Code: 4, Name: "Sales"}}}})
	request := httptest.NewRequest(http.MethodGet, "/api/departments/4/employees?search=a&page=2&limit=5&sort_by=email&sort_order=desc", nil)
	request = mux.SetURLVars(request, map[string]string{"departmentCode": "4"})
	request = request.WithContext(ContextWithMember(request.Context(), &Member{ID: "staff", MemberType: "employee", Permission: "staff"}))
	recorder := httptest.NewRecorder()
	handler.ListDepartmentEmployees(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	for _, expected := range []string{`"id":"1"`, `"phone":"0912"`, `"code":4`, `"name":"Sales"`, `"page":2`, `"limit":5`} {
		if !strings.Contains(recorder.Body.String(), expected) {
			t.Fatalf("response missing %s: %s", expected, recorder.Body.String())
		}
	}
}

func TestDepartmentHandlerUsesInjectedDefaultDepartmentCode(t *testing.T) {
	service := &fakeDepartmentService{}
	handler := NewDepartmentHandler(service, 7)
	req := httptest.NewRequest(http.MethodGet, "/api/departments/employees", nil)
	req = req.WithContext(ContextWithMember(req.Context(), &Member{ID: "employee", MemberType: "employee"}))
	w := httptest.NewRecorder()
	handler.ListDepartmentEmployees(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if service.gotCode != 7 {
		t.Fatalf("department code = %d, want 7", service.gotCode)
	}
}

func TestDepartmentHandler_ListDepartmentEmployees_PassesQueryAndRedactsSensitiveFields(t *testing.T) {
	service := &fakeDepartmentService{directory: []model.EmployeeDirectoryEntry{{ID: "1", Name: "A", Email: "a@example.com", Phone: "0912", Department: model.Department{Code: 4, Name: "Sales"}}}}
	handler := NewDepartmentHandler(service)
	r := httptest.NewRequest(http.MethodGet, "/api/departments/4/employees?search=alice&page=2&limit=5&sort_by=phone&sort_order=desc", nil)
	r = mux.SetURLVars(r, map[string]string{"departmentCode": "4"})
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "staff", MemberType: "employee"}))
	recorder := httptest.NewRecorder()
	handler.ListDepartmentEmployees(recorder, r)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if service.gotCode != 4 || service.gotSearch != "alice" || service.gotPage != 2 || service.gotLimit != 5 || service.gotSortBy != "phone" || service.gotSortOrder != "desc" {
		t.Fatalf("service arguments = code %d, search %q, page %d, limit %d, sort %q/%q", service.gotCode, service.gotSearch, service.gotPage, service.gotLimit, service.gotSortBy, service.gotSortOrder)
	}
	body := recorder.Body.String()
	for _, field := range []string{`"phone":"0912"`, `"id":"1"`, `"department"`} {
		if !strings.Contains(body, field) {
			t.Fatalf("response missing %s: %s", field, body)
		}
	}
	for _, field := range []string{"password", "permission", "created_at", "updated_at", "member_type"} {
		if strings.Contains(body, field) {
			t.Fatalf("response leaked %s: %s", field, body)
		}
	}
}

func TestDepartmentHandler_ListDepartmentEmployees_AllowsRepositorySortFallbackForIllegalValues(t *testing.T) {
	service := &fakeDepartmentService{}
	handler := NewDepartmentHandler(service)
	r := httptest.NewRequest(http.MethodGet, "/api/departments/4/employees?sort_by=password&sort_order=sideways", nil)
	r = mux.SetURLVars(r, map[string]string{"departmentCode": "4"})
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee", MemberType: "employee"}))
	recorder := httptest.NewRecorder()
	handler.ListDepartmentEmployees(recorder, r)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if service.gotSortBy != "password" || service.gotSortOrder != "sideways" {
		t.Fatalf("sort arguments = %q/%q", service.gotSortBy, service.gotSortOrder)
	}
}

func TestDepartmentHandler_ListDepartmentEmployees_RejectsInvalidParametersAndMissingDepartment(t *testing.T) {
	for _, query := range []string{"?page=0", "?limit=101"} {
		handler := NewDepartmentHandler(&fakeDepartmentService{})
		r := httptest.NewRequest(http.MethodGet, "/api/departments/4/employees"+query, nil)
		r = mux.SetURLVars(r, map[string]string{"departmentCode": "4"})
		r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee", MemberType: "employee"}))
		recorder := httptest.NewRecorder()
		handler.ListDepartmentEmployees(recorder, r)
		if recorder.Code != http.StatusBadRequest {
			t.Errorf("query %s status = %d, want %d", query, recorder.Code, http.StatusBadRequest)
		}
	}
	handler := NewDepartmentHandler(&fakeDepartmentService{directoryErr: model.ErrDepartmentNotFound})
	r := httptest.NewRequest(http.MethodGet, "/api/departments/999/employees", nil)
	r = mux.SetURLVars(r, map[string]string{"departmentCode": "999"})
	r = r.WithContext(ContextWithMember(r.Context(), &Member{ID: "employee", MemberType: "employee"}))
	recorder := httptest.NewRecorder()
	handler.ListDepartmentEmployees(recorder, r)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("missing department status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDepartmentHandler_ListDepartmentEmployees_RequireEmployeeRejectsUnauthenticatedAndCustomer(t *testing.T) {
	for _, member := range []*Member{nil, {ID: "customer", MemberType: "customer"}} {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		handler := RequireEmployee(func(h http.Handler) http.Handler { return h })(next)
		r := httptest.NewRequest(http.MethodGet, "/api/departments/4/employees", nil)
		if member != nil {
			r = r.WithContext(ContextWithMember(r.Context(), member))
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, r)
		want := http.StatusUnauthorized
		if member != nil {
			want = http.StatusForbidden
		}
		if recorder.Code != want {
			t.Errorf("member %#v status = %d, want %d", member, recorder.Code, want)
		}
	}
}
func (f *fakeDepartmentService) UpdateMemberDepartment(context.Context, string, string, int) error {
	return nil
}

func TestDepartmentHandler_ListEmployees(t *testing.T) {
	handler := NewDepartmentHandler(&fakeDepartmentService{employees: []model.Member{{ID: "1", Email: "a@example.com", Name: "A", MemberType: "employee", DepartmentCode: 4}}})
	request := httptest.NewRequest(http.MethodGet, "/api/members?query=a&page=2&limit=5", nil)
	request = request.WithContext(ContextWithMember(request.Context(), &Member{ID: "admin", MemberType: "employee", Permission: "admin"}))
	recorder := httptest.NewRecorder()
	handler.ListEmployees(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"department_code":4`) {
		t.Fatalf("response missing department code: %s", recorder.Body.String())
	}
}

func TestDepartmentHandler_Delete(t *testing.T) {
	handler := NewDepartmentHandler(&fakeDepartmentService{})
	request := httptest.NewRequest(http.MethodDelete, "/api/departments/4", nil)
	request = mux.SetURLVars(request, map[string]string{"code": "4"})
	request = request.WithContext(ContextWithMember(request.Context(), &Member{ID: "admin", MemberType: "employee", Permission: "admin"}))
	recorder := httptest.NewRecorder()

	handler.Delete(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var response map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["message"] != "department deleted" {
		t.Fatalf("message = %q, want %q", response["message"], "department deleted")
	}
}
