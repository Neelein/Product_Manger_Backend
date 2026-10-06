package usecase

import (
	"backend/src/domain/model"
	"backend/src/domain/repository"
	"context"
	"strings"
)

type DepartmentService interface {
	List(context.Context, string) ([]model.Department, error)
	Create(context.Context, string, string) (*model.Department, error)
	Update(context.Context, string, int, string) (*model.Department, error)
	Delete(context.Context, string, int) error
	ListEmployees(context.Context, string, string, int, int) ([]model.Member, int, error)
	ListDepartmentEmployees(context.Context, string, int, string, int, int, string, string) ([]model.EmployeeDirectoryEntry, int, error)
	UpdateMemberDepartment(context.Context, string, string, int) error
}

type departmentService struct {
	members     repository.Member
	directory   repository.MemberDirectory
	departments repository.Department
}

func NewDepartmentService(members repository.Member, directory repository.MemberDirectory, departments repository.Department) DepartmentService {
	return &departmentService{members: members, directory: directory, departments: departments}
}

func (s *departmentService) authorize(ctx context.Context, actorID string) error {
	actor, err := s.members.GetByID(ctx, actorID)
	if err != nil || actor == nil || actor.MemberType != string(model.MemberTypeEmployee) || actor.Permission != string(model.PermissionAdmin) {
		return model.ErrForbidden
	}
	return nil
}

func (s *departmentService) authorizeEmployee(ctx context.Context, actorID string) error {
	actor, err := s.members.GetByID(ctx, actorID)
	if err != nil || actor == nil || actor.MemberType != string(model.MemberTypeEmployee) {
		return model.ErrForbidden
	}
	return nil
}

func (s *departmentService) List(ctx context.Context, actorID string) ([]model.Department, error) {
	if err := s.authorize(ctx, actorID); err != nil {
		return nil, err
	}
	return s.departments.List(ctx)
}

func (s *departmentService) Create(ctx context.Context, actorID, name string) (*model.Department, error) {
	if err := s.authorize(ctx, actorID); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, model.ErrInvalidDepartment
	}
	return s.departments.Create(ctx, name)
}

func (s *departmentService) Update(ctx context.Context, actorID string, code int, name string) (*model.Department, error) {
	if err := s.authorize(ctx, actorID); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if code <= 0 || name == "" {
		return nil, model.ErrInvalidDepartment
	}
	return s.departments.Update(ctx, code, name)
}

func (s *departmentService) Delete(ctx context.Context, actorID string, code int) error {
	if err := s.authorize(ctx, actorID); err != nil {
		return err
	}
	if code <= 0 {
		return model.ErrInvalidDepartment
	}
	return s.departments.Delete(ctx, code)
}

func (s *departmentService) ListEmployees(ctx context.Context, actorID, query string, page, limit int) ([]model.Member, int, error) {
	if err := s.authorize(ctx, actorID); err != nil {
		return nil, 0, err
	}
	if page < model.DefaultPage || limit < 1 || limit > model.MaxPageSize {
		return nil, 0, model.ErrInvalidDepartment
	}
	return s.directory.ListEmployees(ctx, strings.TrimSpace(query), page, limit)
}

func (s *departmentService) ListDepartmentEmployees(ctx context.Context, actorID string, departmentCode int, query string, page, limit int, sortBy, sortOrder string) ([]model.EmployeeDirectoryEntry, int, error) {
	if err := s.authorizeEmployee(ctx, actorID); err != nil {
		return nil, 0, err
	}
	if departmentCode <= 0 || page < model.DefaultPage || limit < 1 || limit > model.MaxPageSize {
		return nil, 0, model.ErrInvalidDepartment
	}
	departments, err := s.departments.List(ctx)
	if err != nil {
		return nil, 0, err
	}
	found := false
	for _, department := range departments {
		if department.Code == departmentCode {
			found = true
			break
		}
	}
	if !found {
		return nil, 0, model.ErrDepartmentNotFound
	}
	return s.directory.ListEmployeesByDepartment(ctx, departmentCode, strings.TrimSpace(query), page, limit, sortBy, sortOrder)
}

func (s *departmentService) UpdateMemberDepartment(ctx context.Context, actorID, memberID string, code int) error {
	if err := s.authorize(ctx, actorID); err != nil {
		return err
	}
	if code < 0 {
		return model.ErrInvalidDepartment
	}
	member, err := s.members.GetByID(ctx, memberID)
	if err != nil || member == nil || member.MemberType != string(model.MemberTypeEmployee) {
		return model.ErrMemberNotFound
	}
	if code > 0 {
		departments, listErr := s.departments.List(ctx)
		if listErr != nil {
			return listErr
		}
		found := false
		for _, department := range departments {
			if department.Code == code {
				found = true
				break
			}
		}
		if !found {
			return model.ErrDepartmentNotFound
		}
	}
	return s.directory.UpdateDepartment(ctx, memberID, code)
}
