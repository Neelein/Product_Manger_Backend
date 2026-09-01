package usecase

import (
	"backend/src/domain/model"
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type fakeMembers struct {
	member          *model.Member
	updatedPassword string
}

func (f *fakeMembers) Create(context.Context, *model.Member) error { return nil }
func (f *fakeMembers) GetByEmail(context.Context, string) (*model.Member, error) {
	return f.member, nil
}
func (f *fakeMembers) GetByID(context.Context, string) (*model.Member, error) { return f.member, nil }
func (f *fakeMembers) Update(context.Context, *model.Member) error            { return nil }
func (f *fakeMembers) UpdatePassword(_ context.Context, _ string, hash string) error {
	f.updatedPassword = hash
	f.member.Password = hash
	return nil
}
func (f *fakeMembers) UpdatePermission(_ context.Context, _ string, permission string) error {
	f.member.Permission = permission
	return nil
}

type fakeSessions struct {
	session       *model.Session
	deletedMember string
}

func (f *fakeSessions) Create(_ context.Context, s *model.Session) error { f.session = s; return nil }
func (f *fakeSessions) GetByKey(context.Context, string) (*model.Session, error) {
	return f.session, nil
}
func (f *fakeSessions) Delete(context.Context, string) error { return nil }
func (f *fakeSessions) DeleteByMemberID(_ context.Context, memberID string) error {
	f.deletedMember = memberID
	return nil
}

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

func TestMemberServiceChangePassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("currentpw"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	members := &fakeMembers{member: &model.Member{ID: "member-1", Password: string(hash)}}
	sessions := &fakeSessions{}
	service := NewMemberService(members, sessions)

	if err := service.ChangePassword(context.Background(), "member-1", "currentpw", "newpass1"); err != nil {
		t.Fatal(err)
	}
	if members.updatedPassword == "newpass1" || bcrypt.CompareHashAndPassword([]byte(members.updatedPassword), []byte("newpass1")) != nil {
		t.Fatal("password was not persisted as a bcrypt hash")
	}
	if sessions.deletedMember != "member-1" {
		t.Fatalf("deleted member = %q, want member-1", sessions.deletedMember)
	}
}

func TestMemberServiceChangePasswordValidation(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("currentpw"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		newPassword string
		wantErr     error
	}{
		{name: "seven runes", newPassword: "1234567", wantErr: ErrInvalidPasswordChange},
		{name: "eight runes", newPassword: "12345678"},
		{name: "sixteen runes", newPassword: "1234567890123456"},
		{name: "seventeen runes", newPassword: "12345678901234567", wantErr: ErrInvalidPasswordChange},
		{name: "unicode eight runes", newPassword: "密碼密碼密碼密碼"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			members := &fakeMembers{member: &model.Member{ID: "member-1", Password: string(hash)}}
			sessions := &fakeSessions{}
			service := NewMemberService(members, sessions)
			err := service.ChangePassword(context.Background(), "member-1", "currentpw", tt.newPassword)
			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && members.updatedPassword != "" {
				t.Fatal("invalid password was persisted")
			}
		})
	}
}

func TestMemberServiceChangePasswordWrongCurrentPasswordDoesNotUpdate(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("currentpw"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	members := &fakeMembers{member: &model.Member{ID: "member-1", Password: string(hash)}}
	service := NewMemberService(members, &fakeSessions{})

	err = service.ChangePassword(context.Background(), "member-1", "wrongpw", "newpass1")
	if err != model.ErrInvalidCredentials {
		t.Fatalf("error = %v, want invalid credentials", err)
	}
	if members.updatedPassword != "" {
		t.Fatal("wrong current password updated the member")
	}
}
