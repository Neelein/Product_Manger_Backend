package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend/src/domain/model"
	"backend/src/usecase"
	"github.com/gorilla/mux"
)

type DepartmentHandler struct {
	service               usecase.DepartmentService
	defaultDepartmentCode int
}

func NewDepartmentHandler(service usecase.DepartmentService, defaultDepartmentCode ...int) *DepartmentHandler {
	code := 1
	if len(defaultDepartmentCode) > 0 && defaultDepartmentCode[0] > 0 {
		code = defaultDepartmentCode[0]
	}
	return &DepartmentHandler{service: service, defaultDepartmentCode: code}
}

func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	actor := MemberFromContext(r.Context())
	departments, err := h.service.List(r.Context(), actor.ID)
	if err != nil {
		writeDepartmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, DepartmentListResponse{Departments: departments})
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req DepartmentRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actor := MemberFromContext(r.Context())
	department, err := h.service.Create(r.Context(), actor.ID, req.Name)
	if err != nil {
		writeDepartmentError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, department)
}

func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	code, err := strconv.Atoi(mux.Vars(r)["code"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid department code")
		return
	}
	var req DepartmentRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actor := MemberFromContext(r.Context())
	department, err := h.service.Update(r.Context(), actor.ID, code, req.Name)
	if err != nil {
		writeDepartmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, department)
}

func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	code, err := strconv.Atoi(mux.Vars(r)["code"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid department code")
		return
	}
	actor := MemberFromContext(r.Context())
	if err := h.service.Delete(r.Context(), actor.ID, code); err != nil {
		writeDepartmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "department deleted"})
}

func (h *DepartmentHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	actor := MemberFromContext(r.Context())
	page, limit, err := pageAndLimit(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	employees, total, err := h.service.ListEmployees(r.Context(), actor.ID, r.URL.Query().Get("query"), page, limit)
	if err != nil {
		writeDepartmentError(w, err)
		return
	}
	members := make([]MemberResponse, 0, len(employees))
	for _, employee := range employees {
		members = append(members, MemberResponse{ID: employee.ID, Email: employee.Email, Name: employee.Name, MemberType: employee.MemberType, Permission: employee.Permission, DepartmentCode: employee.DepartmentCode})
	}
	writeJSON(w, http.StatusOK, MembersListResponse{Members: members, Total: total})
}

func (h *DepartmentHandler) ListDepartmentEmployees(w http.ResponseWriter, r *http.Request) {
	actor := MemberFromContext(r.Context())
	departmentCode, err := strconv.Atoi(mux.Vars(r)["departmentCode"])
	if mux.Vars(r)["departmentCode"] == "" {
		departmentCode = h.defaultDepartmentCode
		err = nil
	}
	if err != nil || departmentCode <= 0 {
		writeError(w, http.StatusBadRequest, "invalid department code")
		return
	}
	page, limit, err := pageAndLimit(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	search := r.URL.Query().Get("search")
	if search == "" {
		search = r.URL.Query().Get("query")
	}
	employees, total, err := h.service.ListDepartmentEmployees(r.Context(), actor.ID, departmentCode, search, page, limit, r.URL.Query().Get("sort_by"), r.URL.Query().Get("sort_order"))
	if err != nil {
		writeDepartmentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, EmployeeDirectoryResponse{Employees: employees, Total: total, Page: page, Limit: limit})
}

func (h *DepartmentHandler) UpdateMemberDepartment(w http.ResponseWriter, r *http.Request) {
	var req DepartmentAssignmentRequest
	if json.NewDecoder(r.Body).Decode(&req) != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actor := MemberFromContext(r.Context())
	if err := h.service.UpdateMemberDepartment(r.Context(), actor.ID, mux.Vars(r)["memberId"], req.DepartmentCode); err != nil {
		writeDepartmentError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func pageAndLimit(r *http.Request) (int, int, error) {
	page, limit := model.DefaultPage, model.DefaultPageSize
	var err error
	if value := r.URL.Query().Get("page"); value != "" {
		page, err = strconv.Atoi(value)
		if err != nil || page < 1 {
			return 0, 0, errors.New("invalid page parameter")
		}
	}
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.Atoi(value)
		if err != nil || limit < 1 || limit > model.MaxPageSize {
			return 0, 0, errors.New("invalid limit parameter")
		}
	}
	return page, limit, nil
}

func writeDepartmentError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, model.ErrDepartmentNotFound), errors.Is(err, model.ErrMemberNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrDepartmentNameExists):
		status = http.StatusConflict
	case errors.Is(err, model.ErrInvalidDepartment):
		status = http.StatusBadRequest
	}
	writeError(w, status, err.Error())
}
