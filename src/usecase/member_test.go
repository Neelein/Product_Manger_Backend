package usecase

import (
	"backend/src/domain/model"
	"context"
	"testing"
)

type fakeMembers struct{ member *model.Member }

func (f *fakeMembers) Create(context.Context, *model.Member) error { return nil }
func (f *fakeMembers) GetByEmail(context.Context, string) (*model.Member, error) {
	return f.member, nil
}
func (f *fakeMembers) GetByID(context.Context, string) (*model.Member, error) { return f.member, nil }
func (f *fakeMembers) Update(context.Context, *model.Member) error            { return nil }
func (f *fakeMembers) UpdatePermission(_ context.Context, _ string, permission string) error {
	f.member.Permission = permission
	return nil
}

type fakeSessions struct{ session *model.Session }

func (f *fakeSessions) Create(_ context.Context, s *model.Session) error { f.session = s; return nil }
func (f *fakeSessions) GetByKey(context.Context, string) (*model.Session, error) {
	return f.session, nil
}
func (f *fakeSessions) Delete(context.Context, string) error { return nil }

func TestMemberUseCaseLoginAndRegister(t *testing.T) {
	f := &fakeMembers{}
	u := NewMemberUseCase(f, &fakeSessions{})
	member := &model.Member{Email: "a@example.com", Name: "A"}
	if err := u.Register(context.Background(), member, "secret", func(_ context.Context, m *model.Member) error { f.member = m; return nil }); err != nil {
		t.Fatal(err)
	}
	if _, _, err := u.Login(context.Background(), member.Email, "secret"); err != nil {
		t.Fatal(err)
	}
}

func TestMemberServiceUpdatePermissionRequiresAdminEmployee(t *testing.T) {
	repo := &fakeMembers{member: &model.Member{ID: "admin", MemberType: "employee", Permission: "admin"}}
	service := NewMemberService(repo)
	if err := service.UpdatePermission(context.Background(), "admin", "employee", "catalog_manager"); err != nil {
		t.Fatal(err)
	}
	if repo.member.Permission != "catalog_manager" {
		t.Fatalf("permission was not persisted")
	}
	if err := service.UpdatePermission(context.Background(), "admin", "admin", "employee"); err != model.ErrForbidden {
		t.Fatalf("expected self-permission change to be forbidden, got %v", err)
	}
	repo.member.MemberType = "customer"
	if err := service.UpdatePermission(context.Background(), "admin", "employee", "admin"); err != model.ErrForbidden {
		t.Fatalf("expected customer actor to be forbidden, got %v", err)
	}
}
