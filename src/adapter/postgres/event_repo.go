package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "backend/src/domain/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewEventRepositoryPGX(pool *pgxpool.Pool) *EventRepositoryPGX {
	return &EventRepositoryPGX{pool: pool}
}

func (r *EventRepositoryPGX) Create(ctx context.Context, event *domain.Event) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_event($1, $2, $3, $4, $5, $6)",
		event.Title, event.Description, event.StartTime, event.EndTime, event.Status, event.CreatedBy,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating event: %w", err)
	}
	return nil
}

func (r *EventRepositoryPGX) GetByID(ctx context.Context, id string, memberID string) (*domain.Event, error) {
	var e domain.Event
	err := r.pool.QueryRow(ctx, "SELECT * FROM get_event_by_id($1, $2)", id, memberID).Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.Status, &e.CreatedBy, &e.CreatorName, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEventNotFound
		}
		return nil, fmt.Errorf("getting event by ID: %w", err)
	}
	return &e, nil
}

func (r *EventRepositoryPGX) ListByMonth(ctx context.Context, year, month int, memberID string) ([]domain.Event, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_events_by_month($1, $2, $3)", year, month, memberID)
	if err != nil {
		return nil, fmt.Errorf("listing events by month: %w", err)
	}

	events, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Event, error) {
		var e domain.Event
		err := row.Scan(&e.ID, &e.Title, &e.Description, &e.StartTime, &e.EndTime, &e.Status, &e.CreatedBy, &e.CreatorName, &e.CreatedAt, &e.UpdatedAt)
		return e, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing events by month: %w", err)
	}

	if events == nil {
		events = []domain.Event{}
	}

	return events, nil
}

func (r *EventRepositoryPGX) Update(ctx context.Context, event *domain.Event) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_event($1, $2, $3, $4, $5, $6)",
		event.ID, event.Title, event.Description, event.StartTime, event.EndTime, event.Status,
	).Scan(&event.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrEventNotFound
		}
		return fmt.Errorf("updating event: %w", err)
	}
	return nil
}

func (r *EventRepositoryPGX) Delete(ctx context.Context, id string) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_event($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting event: %w", err)
	}
	if !deleted {
		return domain.ErrEventNotFound
	}
	return nil
}

func (r *EventRepositoryPGX) AddViewer(ctx context.Context, eventID string, memberID string) error {
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, "SELECT * FROM add_event_viewer($1, $2)", eventID, memberID).Scan(&createdAt)
	if err != nil {
		return fmt.Errorf("adding event viewer: %w", err)
	}
	return nil
}

func (r *EventRepositoryPGX) RemoveViewer(ctx context.Context, eventID string, memberID string) error {
	var removed bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM remove_event_viewer($1, $2)", eventID, memberID).Scan(&removed)
	if err != nil {
		return fmt.Errorf("removing event viewer: %w", err)
	}
	if !removed {
		return domain.ErrEventViewerNotFound
	}
	return nil
}

func (r *EventRepositoryPGX) ListViewers(ctx context.Context, eventID string) ([]domain.EventViewer, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_event_viewers($1)", eventID)
	if err != nil {
		return nil, fmt.Errorf("listing event viewers: %w", err)
	}

	viewers, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.EventViewer, error) {
		var v domain.EventViewer
		err := row.Scan(&v.MemberID, &v.MemberName)
		return v, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing event viewers: %w", err)
	}

	if viewers == nil {
		viewers = []domain.EventViewer{}
	}

	return viewers, nil
}
