package usecase

import (
	"backend/src/domain/model"
	"context"
	"testing"
)

type fakeEvents struct{ updated bool }

func (f *fakeEvents) GetByID(context.Context, string, string) (*model.Event, error) { return nil, nil }
func (f *fakeEvents) Update(context.Context, *model.Event) error                    { f.updated = true; return nil }
func (f *fakeEvents) Delete(context.Context, string) error                          { return nil }

func TestEventUseCaseRequiresOwnerUnlessAdmin(t *testing.T) {
	f := &fakeEvents{}
	u := NewEventUseCase(f)
	e := &model.Event{CreatedBy: "owner"}
	if err := u.Update(context.Background(), e, "other", false); err != model.ErrNotEventOwner {
		t.Fatalf("err = %v", err)
	}
	if err := u.Update(context.Background(), e, "admin", true); err != nil {
		t.Fatal(err)
	}
}
