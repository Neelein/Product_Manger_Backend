package usecase

import (
	"backend/src/domain/model"
	"backend/src/domain/repository"
	"context"
)

type EventUseCase struct{ events repository.Event }

func NewEventUseCase(events repository.Event) *EventUseCase { return &EventUseCase{events: events} }
func (u *EventUseCase) Update(ctx context.Context, event *model.Event, memberID string, admin bool) error {
	if !admin && event.CreatedBy != memberID {
		return model.ErrNotEventOwner
	}
	return u.events.Update(ctx, event)
}
