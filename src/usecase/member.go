package usecase

import (
	"backend/src/domain/model"
	"backend/src/domain/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type MemberUseCase struct {
	members  repository.Member
	sessions repository.Session
}

func NewMemberUseCase(members repository.Member, sessions repository.Session) *MemberUseCase {
	return &MemberUseCase{members: members, sessions: sessions}
}
func (u *MemberUseCase) Login(ctx context.Context, email, password string) (*model.Member, *model.Session, error) {
	m, err := u.members.GetByEmail(ctx, email)
	if err != nil || m == nil {
		return nil, nil, model.ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password)) != nil {
		return nil, nil, model.ErrInvalidCredentials
	}
	s := &model.Session{MemberID: m.ID}
	if err := u.sessions.Create(ctx, s); err != nil {
		return nil, nil, err
	}
	return m, s, nil
}
func (u *MemberUseCase) Register(ctx context.Context, m *model.Member, password string, register func(context.Context, *model.Member) error) error {
	if m.Email == "" || password == "" {
		return errors.New("email and password are required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.Password = string(hash)
	return register(ctx, m)
}
