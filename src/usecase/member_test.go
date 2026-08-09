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
