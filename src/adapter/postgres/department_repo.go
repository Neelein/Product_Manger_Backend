package postgres

import (
	"backend/src/domain/model"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DepartmentRepositoryPGX struct{ pool *pgxpool.Pool }

func NewDepartmentRepositoryPGX(pool *pgxpool.Pool) *DepartmentRepositoryPGX {
	return &DepartmentRepositoryPGX{pool: pool}
}

func (r *DepartmentRepositoryPGX) List(ctx context.Context) ([]model.Department, error) {
	rows, err := r.pool.Query(ctx, `SELECT code, name FROM departments ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("listing departments: %w", err)
	}
	departments, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Department])
	if err != nil {
		return nil, fmt.Errorf("collecting departments: %w", err)
	}
	if departments == nil {
		departments = []model.Department{}
	}
	return departments, nil
}

func (r *DepartmentRepositoryPGX) Create(ctx context.Context, name string) (*model.Department, error) {
	var department model.Department
	err := r.pool.QueryRow(ctx, `INSERT INTO departments(name) VALUES ($1) RETURNING code, name`, name).Scan(&department.Code, &department.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, model.ErrDepartmentNameExists
		}
		return nil, fmt.Errorf("creating department: %w", err)
	}
	return &department, nil
}

func (r *DepartmentRepositoryPGX) Update(ctx context.Context, code int, name string) (*model.Department, error) {
	var department model.Department
	err := r.pool.QueryRow(ctx, `UPDATE departments SET name = $2 WHERE code = $1 RETURNING code, name`, code, name).Scan(&department.Code, &department.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrDepartmentNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, model.ErrDepartmentNameExists
		}
		return nil, fmt.Errorf("updating department: %w", err)
	}
	return &department, nil
}

func (r *DepartmentRepositoryPGX) Delete(ctx context.Context, code int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning department deletion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, `UPDATE members SET department_code = 0, updated_at = now() WHERE department_code = $1 AND member_type = 'employee'`, code); err != nil {
		return fmt.Errorf("updating employee departments: %w", err)
	}
	result, err := tx.Exec(ctx, `DELETE FROM departments WHERE code = $1`, code)
	if err != nil {
		return fmt.Errorf("deleting department: %w", err)
	}
	if result.RowsAffected() == 0 {
		return model.ErrDepartmentNotFound
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing department deletion: %w", err)
	}
	return nil
}

func (r *MemberRepositoryPGX) ListEmployees(ctx context.Context, query string, page, limit int) ([]model.Member, int, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM members WHERE member_type = 'employee' AND ($1 = '%' OR lower(email) LIKE $1 OR lower(name) LIKE $1 OR id::text ILIKE $1)`, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting employees: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT id, email, name, member_type, COALESCE(permission, ''), department_code, created_at, updated_at FROM members WHERE member_type = 'employee' AND ($1 = '%' OR lower(email) LIKE $1 OR lower(name) LIKE $1 OR id::text ILIKE $1) ORDER BY lower(name), id LIMIT $2 OFFSET $3`, pattern, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, fmt.Errorf("listing employees: %w", err)
	}
	employees, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Member, error) {
		var member model.Member
		err := row.Scan(&member.ID, &member.Email, &member.Name, &member.MemberType, &member.Permission, &member.DepartmentCode, &member.CreatedAt, &member.UpdatedAt)
		return member, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collecting employees: %w", err)
	}
	if employees == nil {
		employees = []model.Member{}
	}
	return employees, total, nil
}

func (r *MemberRepositoryPGX) ListEmployeesByDepartment(ctx context.Context, departmentCode int, query string, page, limit int, sortBy, sortOrder string) ([]model.EmployeeDirectoryEntry, int, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	orderColumn := map[string]string{"id": "m.id", "name": "lower(m.name)", "email": "lower(m.email)", "phone": "lower(m.phone)"}[sortBy]
	if orderColumn == "" {
		orderColumn = "lower(m.name)"
	}
	direction := "ASC"
	if strings.EqualFold(sortOrder, "desc") {
		direction = "DESC"
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM members m WHERE m.member_type = 'employee' AND m.department_code = $1 AND ($2 = '%' OR lower(m.email) LIKE $2 OR lower(m.name) LIKE $2 OR lower(m.phone) LIKE $2 OR m.id::text ILIKE $2)`, departmentCode, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting department employees: %w", err)
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`SELECT m.id, m.name, m.email, d.code, d.name, m.phone FROM members m JOIN departments d ON d.code = m.department_code WHERE m.member_type = 'employee' AND m.department_code = $1 AND ($2 = '%%' OR lower(m.email) LIKE $2 OR lower(m.name) LIKE $2 OR lower(m.phone) LIKE $2 OR m.id::text ILIKE $2) ORDER BY %s %s, m.id ASC LIMIT $3 OFFSET $4`, orderColumn, direction), departmentCode, pattern, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, fmt.Errorf("listing department employees: %w", err)
	}
	entries, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.EmployeeDirectoryEntry, error) {
		var entry model.EmployeeDirectoryEntry
		err := row.Scan(&entry.ID, &entry.Name, &entry.Email, &entry.Department.Code, &entry.Department.Name, &entry.Phone)
		return entry, err
	})
	if err != nil {
		return nil, 0, fmt.Errorf("collecting department employees: %w", err)
	}
	if entries == nil {
		entries = []model.EmployeeDirectoryEntry{}
	}
	return entries, total, nil
}

func (r *MemberRepositoryPGX) UpdateDepartment(ctx context.Context, memberID string, code int) error {
	result, err := r.pool.Exec(ctx, `UPDATE members SET department_code = $2, updated_at = now() WHERE id = $1 AND member_type = 'employee'`, memberID, code)
	if err != nil {
		return fmt.Errorf("updating member department: %w", err)
	}
	if result.RowsAffected() == 0 {
		return model.ErrMemberNotFound
	}
	return nil
}
